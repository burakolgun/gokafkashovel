package producer

import (
	"context"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/rs/zerolog"
	"time"
)

type CommonProducer struct {
	producer *kafka.Producer
	logger   *zerolog.Logger
	config   Config
	mChannel chan Message
}

type Producer interface {
	Produce(key []byte, message []byte, headers []kafka.Header, topic string)
}

type KafkaProducerConfig struct {
	Broker string
	Acks   string
}

type Message struct {
	key     []byte
	message []byte
	headers []kafka.Header
	topic   string
}

type Config struct {
	ProducerName   string
	ProducerConfig KafkaProducerConfig
}

func New(logger *zerolog.Logger, config Config) (*CommonProducer, error) {
	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": config.ProducerConfig.Broker,
		"acks":              config.ProducerConfig.Acks,
	})

	if err != nil {
		return nil, err
	}

	return &CommonProducer{
		producer: producer,
		logger:   logger,
		config:   config,
	}, err
}

func (p *CommonProducer) Produce(key []byte, msg []byte, headers []kafka.Header, topic string) {

	p.mChannel <- Message{
		key:     key,
		message: msg,
		headers: headers,
		topic:   topic,
	}
}

func (p *CommonProducer) StartProducer(ctx context.Context) {
	go p.start(ctx)
}

func (p *CommonProducer) start(ctx context.Context) {
	p.mChannel = make(chan Message, 5000)

	go func() {
		for e := range p.producer.Events() {

			switch ev := e.(type) {
			case *kafka.Message:

				if ev.TopicPartition.Error != nil {
					p.logger.Error().Msg(fmt.Sprintf("Delivery failed: %v \n %s", ev.TopicPartition.Error.Error(), *ev.TopicPartition.Topic))
				} else {
					p.logger.Info().Msg(fmt.Sprintf("Delivered message to topic %s [%d] at offset %d\n",
						*ev.TopicPartition.Topic, ev.TopicPartition.Partition, ev.TopicPartition.Offset))
				}

			}
		}
	}()

	p.logger.Info().Msg(p.config.ProducerName + " Producer UP && RUN")

	for {
		select {
		case m := <-p.mChannel:
			err := p.producer.Produce(&kafka.Message{
				TopicPartition: kafka.TopicPartition{Topic: &m.topic, Partition: kafka.PartitionAny},
				Value:          m.message,
				Headers:        m.headers,
				Key:            m.key,
			}, nil)

			if err != nil {
				p.logger.Error().Msg(fmt.Sprintf("an error occurred while enqueue. dc: %s, topic: %s, err: %v", p.config.ProducerName, m.topic, err))
			}

		case <-ctx.Done():
			t := time.Tick(time.Second * 10)
			<-t
			uFMC := p.producer.Flush(15 * 1000)
			p.logger.Warn().Msg(fmt.Sprintf("flush comleted, and unflushed msg count: %d", uFMC))
			return
		}
	}
}
