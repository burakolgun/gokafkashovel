package main

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

func main() {

	wg := sync.WaitGroup{}
	wg.Add(1)
	go RunConsumer()

	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": "localhost"})
	if err != nil {
		panic(err)
	}

	defer p.Close()

	// Delivery report handler for produced messages
	go func() {
		for e := range p.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					fmt.Printf("Delivery failed: %v\n", ev.TopicPartition)
				} else {
					fmt.Printf("Delivered message to %v\n", ev.TopicPartition)
				}
			}
		}
	}()

	topic := "example-topic-name.error"
	for i := 0; i < 5; i++ {

		p.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
			Key:            []byte(strconv.Itoa(i)),
			Value:          []byte(strconv.Itoa(i)),
		}, nil)
	}
	// Wait for message deliveries before shutting down
	p.Flush(15 * 1000)

	wg.Wait()
}

func RunConsumer() {

	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": "localhost",
		"group.id":          "example-group",
		"auto.offset.reset": "earliest",
	})

	if err != nil {
		panic(err)
	}

	c.SubscribeTopics([]string{"example-topic-name.retry"}, nil)

	run := true
	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": "localhost"})

	if err != nil {
		panic(err)
	}

	for run {
		msg, err := c.ReadMessage(time.Second)
		if err == nil {
			fmt.Printf("Message on %s: %s\n", msg.TopicPartition, string(msg.Value))
			time.Sleep(time.Millisecond * 1000)

			topic := "example-topic-name.error"
			p.Produce(&kafka.Message{
				TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
				Key:            msg.Key,
				Value:          msg.Value,
				Headers:        msg.Headers,
			}, nil)
		} else if !err.(kafka.Error).IsTimeout() {
			fmt.Printf("Consumer error: %v (%v)\n", err, msg)
		}
	}

	c.Close()
}
