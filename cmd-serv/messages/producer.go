package messages

import (
	"flag"
	"fmt"
	"github.com/shopify/sarama"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"
)

var (
	key         = flag.String("key", "", "The key of the message to produce. Can be empty")
	partitioner = flag.String("partitioner", "random", "The partitioning schema to use. Can be 'hash', 'manual' or 'random'")
	partition   = flag.Int("partition", -1, "The partition to produce to")
)

type Producer struct {
	syncProducer sarama.SyncProducer
	config       *sarama.Config
}

func newProducer() *Producer {
	p := Producer{}

	p.config.Producer.RequiredAcks = sarama.WaitForLocal // only wait for the leader to ack
	p.config.Producer.Return.Successes = true
	p.config.Producer.Compression = sarama.CompressionSnappy  // Compress messages
	p.config.Producer.FlushFrequency = 500 * time.Millisecond // Flush batches every 500ms

	switch *partitioner {
	case "":
		if *partition >= 0 {
			p.config.Producer.Partitioner = sarama.NewManualPartitioner
		} else {
			p.config.Producer.Partitioner = sarama.NewHashPartitioner
		}
	case "hash":
		p.config.Producer.Partitioner = sarama.NewHashPartitioner
	case "random":
		p.config.Producer.Partitioner = sarama.NewRandomPartitioner
	case "manual":
		p.config.Producer.Partitioner = sarama.NewManualPartitioner
		if *partition == -1 {
			log.Fatal("-partition is required when partitioning manually")
		}
	default:
		log.Fatal(fmt.Sprintf("Partitioner %s not supported.", *partitioner))
	}

	producer, err := sarama.NewSyncProducer(strings.Split(*brokerList, ","), p.config)
	if err != nil {
		log.Fatalf("Failed to open Kafka producer: %s", err)
	}

	p.syncProducer = producer

	return &p
}

func (p *Producer) NewMessage(msg string) *sarama.ProducerMessage {
	message := &sarama.ProducerMessage{Topic: *topic, Partition: int32(*partition)}

	if *key != "" {
		message.Key = sarama.StringEncoder(*key)
	}

	if msg != "" {
		message.Value = sarama.StringEncoder(msg)
	} else if stdinAvailable() {
		bytes, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			log.Fatalf("Failed to read data from the standard input: %s", err)
		}
		message.Value = sarama.ByteEncoder(bytes)
	} else {
		log.Fatal("-value is required, or you have to provide the value on stdin")
	}
	return message
}

func (p *Producer) Sent(message *sarama.ProducerMessage) {

	partition, offset, err := p.syncProducer.SendMessage(message)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Message is stored in topic(%s)/partition(%d)/offset(%d)\n", *topic, partition, offset)
}

func stdinAvailable() bool {
	stat, _ := os.Stdin.Stat()
	return (stat.Mode() & os.ModeCharDevice) == 0
}

func (p *Producer) Close() error {
	log.Println("Producer.Close()")
	if err := p.syncProducer.Close(); err != nil {
		return err
	}
	return nil
}
