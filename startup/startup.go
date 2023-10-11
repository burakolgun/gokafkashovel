package startup

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/burakolgun/gogenviper"
	"github.com/burakolgun/gokafkashovel/producer"
	"github.com/burakolgun/gokafkashovel/shovel"
	"github.com/burakolgun/gokafkashovel/startup/container"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/diode"
)

func Start() {
	initLogger()
	initCfg()
	initCommonProducer()
	initRedisClient()
	go shovelManager()

	wg := sync.WaitGroup{}
	wg.Add(1)

	container.GetCommonProducer().StartProducer(context.Background())
	wg.Wait()
}

func initCfg() {
	watcher, err := gogenviper.Init("./resources/config", "config", "json", container.GetCfg())
	if err != nil {
		panic(err)
	}

	watcher.Watch()
}

func initCommonProducer() {
	p, err := producer.New(container.GetLogger(), producer.Config{
		ProducerName: "shovel common producer",
		ProducerConfig: producer.KafkaProducerConfig{
			Broker: "localhost:9092",
			Acks:   "1",
		},
	})

	if err != nil {
		return
	}

	container.SetCommonProducer(p)
}

func initLogger() {
	wr := diode.NewWriter(os.Stdout, 1000, 10*time.Millisecond, func(missed int) {
		fmt.Printf("Logger Dropped %d messages", missed)
	})
	l := zerolog.New(wr).With().Timestamp().Logger()

	container.SetLogger(l)
}

func initRedisClient() {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	ctx := context.Background()

	err := rdb.Process(ctx, rdb.Ping(ctx))

	if err != nil {
		panic(err)
	}

	container.SetRedisClient(rdb)
}

func shovelManager() {
	fmt.Println("before shovelManager cycle")
	ctx, cancelFn := context.WithTimeout(context.Background(), time.Second*time.Duration(container.GetCfg().ShovelIntervalInSec))
	runShovel(ctx, cancelFn)
	fmt.Println("waiting runShovel()")
	fmt.Println("after shovelManager cycle")
	shovelManager()
}
func runShovel(ctx context.Context, cancelFn context.CancelFunc) {
	brokerList := container.GetCfg().Kafka.BrokerList

	shovelWg := sync.WaitGroup{}
	for _, app := range container.GetCfg().ShovelList {
		c, err := kafka.NewConsumer(&kafka.ConfigMap{
			"bootstrap.servers": brokerList,
			"group.id":          app.GroupId,
			"auto.offset.reset": app.AutoOffsetReset,
		})

		if err != nil {
			panic(err)
		}

		q := shovel.New(shovel.Config{
			SourceTopicConsumer: c,
			Logger:              container.GetLogger(),
			Rdb:                 container.GetRedisClient(),
			SourceTopicName:     app.SourceTopicName,
			TargetTopicName:     app.TargetTopicName,
			PoisonedTopicName:   app.PoisonedTopicName,
			Producer:            container.GetCommonProducer(),
			Name:                app.Name,
			MaxErrorCount:       app.MaxErrorCount,
		})

		shovelWg.Add(1)
		go q.ConsumeWithContext(ctx, &shovelWg)
		fmt.Printf("%s consumer up\n", app.Name)
	}

	fmt.Println("before complete the cycle")
	shovelWg.Wait()
	fmt.Println("after complete the cycle")
	fmt.Println("waiting interval starting...")
	<-time.After(time.Second * time.Duration(container.GetCfg().ShovelWaitingIntervalInSec))
	fmt.Println("waiting interval completed.")
}
