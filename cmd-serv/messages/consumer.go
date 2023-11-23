package messages

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
)

type Consumer struct {
	consumer sarama.Consumer
	wg       sync.WaitGroup

	messages chan *sarama.ConsumerMessage
	closing  chan struct{}

	callbacks Callbacks
}

var (
	brokerList = flag.String("brokers", "127.0.0.1:9092", "The comma separated list of brokers in the Kafka cluster.")
	topic      = flag.String("topic", "Topic1", "REQUIRED: the topic to consume.")
	offset     = flag.String("offset", "newest", "The offset to start with. Can be 'oldest', 'newest'.")
	bufferSize = flag.Int("buffer-size", 256, "The buffer size of the message channel.")
)

func NewConsumer(callbacks Callbacks) *Consumer {
	consumer := Consumer{
		callbacks: callbacks,
	}
	consumer.messages = make(chan *sarama.ConsumerMessage, *bufferSize)
	consumer.closing = make(chan struct{})

	saramaConsumer, err := sarama.NewConsumer(strings.Split(*brokerList, ","), nil)
	if err != nil {
		log.Fatalf("Failed to start consumer: %s", err)
	}

	consumer.consumer = saramaConsumer

	return &consumer
}

func (c *Consumer) Consume() {
	consumers := make(chan *sarama.ConsumerMessage)
	errors := make(chan *sarama.ConsumerError)

	var initialOffset int64
	switch *offset {
	case "oldest":
		initialOffset = sarama.OffsetOldest
	case "newest":
		initialOffset = sarama.OffsetNewest
	}

	partitions, _ := c.consumer.Partitions(*topic)
	// this only consumes partition no 1,you would probably want to consume all partitions
	consumer, err := c.consumer.ConsumePartition(*topic, partitions[0], initialOffset)
	if nil != err {
		fmt.Printf("Topic: %v Partitions: %v", *topic, partitions)
		panic(err)
	}
	fmt.Println(" Start consuming topic ", *topic)

	go func(topic string, consumer sarama.PrtitionConsumer) {
		for {
			select {
			case consumerError := <-consumer.Errors():
				errors <- consumerError
				fmt.Println("consumerError: ", consumerError.Err)
			case msg, ok := <-consumer.Messages():
				if ok {
					consumers <- msg
					fmt.Println("Got message on topic ", topic, msg.Value)
				}
			}
		}
	}(*topic, consumer)

	go func() {
		for {
			select {
			case msg := <-consumers:
				if c.callbacks.OnMessageConsumed != nil {
					fmt.Println("Received messages", string(msg.Key), string(msg.Value))
					c.callbacks.OnMessageConsumed(msg.Value)
				}
			case consumerError := <-errrors:
				fmt.Println("Received consumerError ", string(consumerError.Topic), string(consumerError.Partition), string(consumerError.Something))
			}
		}
	}()
}
func (c *Consumer) Close() {
	go func() {
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, os.Kill, os.Interrupt)
		<-signals
		log.Println("Initiating sutdown of consumer...")
		close(c.closing)
	}()
}
