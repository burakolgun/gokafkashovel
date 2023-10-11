package shovel

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/burakolgun/gokafkashovel/constants"
	"github.com/burakolgun/gokafkashovel/producer"
	"github.com/burakolgun/gokafkashovel/utils/kafka_utils"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/hashicorp/go-uuid"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

type Shovel struct {
	sourceTopicConsumer *kafka.Consumer
	logger              *zerolog.Logger
	rdb                 *redis.Client
	producer            *producer.CommonProducer

	sourceTopicName   string
	targetTopicName   string
	poisonedTopicName string
	name              string
	maxErrorCount     int
	isCompleted       bool
}

type Config struct {
	SourceTopicConsumer *kafka.Consumer
	Logger              *zerolog.Logger
	Rdb                 *redis.Client
	Producer            *producer.CommonProducer
	SourceTopicName     string
	TargetTopicName     string
	PoisonedTopicName   string
	Name                string
	MaxErrorCount       int
}

func New(cfg Config) *Shovel {
	return &Shovel{
		sourceTopicConsumer: cfg.SourceTopicConsumer,
		logger:              cfg.Logger,
		rdb:                 cfg.Rdb,
		sourceTopicName:     cfg.SourceTopicName,
		targetTopicName:     cfg.TargetTopicName,
		poisonedTopicName:   cfg.PoisonedTopicName,
		producer:            cfg.Producer,
		name:                cfg.Name,
		maxErrorCount:       cfg.MaxErrorCount,
	}
}

func (consumer *Shovel) ConsumeWithContext(ctx context.Context, wg *sync.WaitGroup) {
	err := consumer.sourceTopicConsumer.Subscribe(consumer.sourceTopicName, nil)

	if err != nil {
		consumer.logger.Error().Msg(fmt.Sprintf("%s: could not subscribed to topic: %v err:%v", consumer.name, consumer.sourceTopicName, err))
		panic(err)
	}

	go consumer.consumeSourceTopic()

	<-ctx.Done()

	consumer.close()
	wg.Done()
}

func (consumer *Shovel) close() {
	consumer.isCompleted = true
	consumer.logger.Warn().Msg(fmt.Sprintf("consumer.close: consumer closing... topic: %s", consumer.sourceTopicName))
	<-time.After(time.Second * 5)
	err := consumer.sourceTopicConsumer.Pause([]kafka.TopicPartition{{Topic: &consumer.sourceTopicName, Partition: kafka.PartitionAny}})
	consumer.logger.Warn().Msg(fmt.Sprintf("consumer.close: consumer paused... topic: %s", consumer.sourceTopicName))
	if err != nil {
		consumer.logger.Error().Msg(fmt.Sprintf("consumer.closeConsumer: consumer can't paused. err: %s", err.Error()))
	}

	consumer.logger.Warn().Msg(fmt.Sprintf("consumer.close: trying to close... topic: %s", consumer.sourceTopicName))
	err = consumer.sourceTopicConsumer.Close()
	if err != nil {
		if err.Error() == "Operation not allowed on closed client" {
			consumer.logger.Info().Msg(fmt.Sprintf("consumer.closeConsumer: consumer already closed err: %s", err.Error()))
			return
		}

		consumer.logger.Error().Msg(fmt.Sprintf("consumer.closeConsumer: consumer could not closed err: %s", err.Error()))
	}

	consumer.logger.Info().Msg(fmt.Sprintf("consumer.close: consumer closed... topic: %s", consumer.sourceTopicName))
}

func (consumer *Shovel) consumeSourceTopic() {
	for {
		msg, err := consumer.sourceTopicConsumer.ReadMessage(time.Millisecond * 100)

		if err != nil {
			if err.Error() == "Operation not allowed on closed client" {
				consumer.logger.Warn().Msg("consumer already closed")
				return
			} else if !err.(kafka.Error).IsTimeout() {
				consumer.logger.Error().Msg(fmt.Sprintf("Consumer returned error, err: %s", err.Error()))
				panic(err)
			}
		}

		if err == nil {
			err := consumer.Process(msg)

			if err != nil && err.Error() == constants.ConsumerClosed {
				consumer.logger.Info().Msg(fmt.Sprintf("%s consumer will close because cycle is completed \n", consumer.name))
				return
			}

			if consumer.isCompleted {
				consumer.logger.Info().Msg(fmt.Sprintf("%s consume wont continue because consumer is closed \n", consumer.name))
				return
			}
		} else if err.Error() == "Operation not allowed on closed client" {
			consumer.logger.Info().Msg("consumer closed")
			return
		} else if !err.(kafka.Error).IsTimeout() {
			consumer.logger.Error().Msg(fmt.Sprintf("Consumer returned error, err: %s", err.Error()))
			panic(err)
		}
	}
}

func (consumer *Shovel) Process(msg *kafka.Message) error {
	ctx := context.Background()
	headers := msg.Headers
	var err error
	requestId := kafka_utils.GetFieldByNameFromHeader(headers, constants.RequestIdKey)

	if requestId == "" {
		requestId, err = uuid.GenerateUUID()

		if err != nil {
			return err
		}

		consumer.logger.Warn().Msg(fmt.Sprintf("requestID not found, key: %s, topic: %s", msg.Key, *msg.TopicPartition.Topic))
		headers = kafka_utils.AddFieldToHeaderByFieldName(headers, constants.RequestIdKey, requestId)
	} else {
		m := consumer.rdb.Get(ctx, fmt.Sprintf("%s-%s", consumer.name, requestId))

		if m.Err() != redis.Nil {
			consumer.producer.Produce(msg.Key, msg.Value, headers, consumer.sourceTopicName)
			consumer.logger.Warn().Msg("shovel cycle is completed")
			return errors.New(constants.ConsumerClosed)
		}

	}

	headers, errorCount := increaseErrorCount(headers)

	if errorCount > consumer.maxErrorCount {
		consumer.producer.Produce(msg.Key, msg.Value, headers, consumer.poisonedTopicName)
		return err
	}

	consumer.rdb.Set(context.Background(), fmt.Sprintf("%s-%s", consumer.name, kafka_utils.GetFieldByNameFromHeader(headers, constants.RequestIdKey)), kafka_utils.GetFieldByNameFromHeader(headers, constants.RequestIdKey), time.Second*59).Err()
	consumer.producer.Produce(msg.Key, msg.Value, headers, consumer.targetTopicName)
	return err
}

func increaseErrorCount(headers []kafka.Header) ([]kafka.Header, int) {
	isFound := false
	count := 1
	var err error

	for i, header := range headers {
		if header.Key == constants.KafkaErrorCountKey {
			isFound = true
			count, err = strconv.Atoi(string(header.Value))

			if err != nil {
				panic(err)
			}

			headers[i] = kafka.Header{Key: constants.KafkaErrorCountKey, Value: []byte(strconv.Itoa(count + 1))}
		}
	}

	if !isFound {
		headers = kafka_utils.AddFieldToHeaderByFieldName(headers, constants.KafkaErrorCountKey, strconv.Itoa(count))
	}

	return headers, count
}
