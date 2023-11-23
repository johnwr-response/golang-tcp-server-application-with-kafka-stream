# How to Build Golang TCP Server Application with Kafka Stream# How to Build Golang TCP Server Application with Kafka Stream

## Lesson 1: Getting Started How to create Go TCP server and Kafka in GO

### Introduction

I need to create a simple framework to provide my endpoint devices (doesn't matter which platform they run on) the option to send and receive message from my backend.  
I require those messages to be managed by a message broker so that they can be processed in an asynchronous way.  
The system contains of 4 main layers, this article section is mainly about the first one:

  1. TCP servers - Needs to maintain as many TCP sockets in sync with the endpoints as possible. All the endpoints messages will be processed in a different layer by the message broker. This keeps the TCP servers layer very thin and effective. I also want to keep as many concurrent connections as possible, and Go is a good choice for this.
  2. Message broker - Responsible for delivering the messages between the TCP servers layer and the workers layer. I chose Apache Kafka for that purpose.
  3. Workers layers - will process the messages through services exposed in the backend layer.
  4. Backend services layer - An encapsulation of services required by your application such as DB, Authentication, Logging, external APIs and more.

So, this Go server:
  1. Communicated with its endpoint clients by TCP sockets.
  2. Queues the messages in the message broker
  3. Receives back messages from the broker after they were processed and sends response acknowledgement and/or errors to the TCP clients.

I have also included a Dockerfile and a build script to push  the image to your Docker repository.  
The article is divided to sections representing the components of the system. Each component should be decoupled from the others in a way that allows you to read about a single component in a straight forward manner.

#### TCP Client

Its role is to represent a TCP client communicating with our TCP server

````
type Client struct {
    ID                string
    Conn              net.Conn
    onMessageReceived func(clientID string, msg []byte)
    close Connection  func(clientID string)
}
````

#### Kafka Consumer

Its role is to consume messages from our Kafka broker, and to broadcast them back to relevant clients by their UIDs.
In this example we are consuming from multiple topics using the cluster implementation of sarama.
Let's define our Consumer struct:

````
type Consumer struct {
    consumer  sarama.Consumer
	wg        sync.WaitGroup
	messages  chan *sarama.ConsumerMessage
	closing   chan struct{}
	callbacks Callbacks
}
````

The constructor receives the callbacks and relevant details to connect to the topic:

````
func NewConsumer(callbacks Callbacks) *Consumer {
    consumer := Consumer{
	    callbacks: callbacks,
	}
    consumer.messages = make(chan *sarama.ConsumerMessage, *bufferSize)
    consumer.closing = make(chan struct{})
	saramaConsumer, err := sarama.NewConsumer(strings.Split(*brokerList, ","),nil)
	if (err != nil) {
	    log.Fatalf("Failed to start consumer %s", err)
	}
	consumer.consumer = saramaConsumer
	return &consumer
	
}
````

It will consume permanently on a new goroutine inside the Consume() function.
It reads from the Messages() channel for new messages and the Notifications() channel for events.

#### Kafka Producer

Its role is to produce messages to our Kafka broker.
In this example we are producing to a single topic.
Let's define our Producer struct:


````
type Producer struct {
    syncProducer sarama.syncProducer
	config       *sarama.Config
}
````

Producer is constructed with the callbacks for Error,and the details to connect to the Kafka broker.

````
func NewProducer() *Producer {
    p := Producer()
	p.config = sarama.NewConfig()
	p.config.Producer.RequiredAcks = sarama.waitForLocal // Only wait for the leader to ack
	p.config.Producer.Return.Successes := true
	p.config.Producer.Compression = sarama.CompressionSnappy // Compress messages
	p.config.Producer.Flush.Frequency = 500 * time.Millisecond // Flush batches every 500ms
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
		        log.fatal("-partitionis required when partitioning manually")
			}
	    default:
		    log.fatal(fmt.Sprintf("Partitioner %s not supported"), 'partitioner)
	}
	producer, err := sarama.NewSyncProducer(strings.Split(*brokerList,","), p.config)
	if (err != nil) {
	    log.Fatal("Failed to open Kafka producer: %s", err)
	}
	p.syncProducer = producer
	return %p
}
````

And finally, we provide the functions to produce the message and close the producer:

````
func NewProducer(p *Producer) NewMessage(msg String) *sarama.ProducerMessage{
    message := &sarama.ProducerMessage{Topic: *topic, Partition: int32(*partition)}
	if *key != "" {
        message.Key = sarama.StringEncoder(*key)
	}
	if msg != "" {
	    message.Value = sarama.StringEncoder(msg)
	} else if stdinAvailable() {
	    bytes, err := ioutil.ReadAll(os.Stdin)
		if (err != nil) {
		    log.Fatalf("Failed to read data from the standard input: %s", err)
		}
		message.Value = sarama.ByteEncoder(bytes)
	} else {
        log.Fatalf("value is required, or you have to provide the value on stdin:")
	}
	return message
}
````

#### TCP Server

Its role is to obtain and manage a set of Client, and send and receive messages from them.

````
type TcpServer struct {
    listener    net.Listener
	Connections map[string]*Client
	connLock    sync.RWMutex
	callbacks   Callbacks
}
````

It is constructed simply with an address to bind to and the callbacks to send:

````
func NewTcpServer(callbacks Callbacks) *TcpServer {
    tcpServer := TcpServer{
	    callbacks := Callbacks
	}
	tcpServer.Connections = make(map[string]*Client}
	return *tcpServer
}
````

When a connection even occurs we process it and handle it, if it's a new event we attach a new UIDto the client.
If connection is terminated we delete this client.
In both case we send the callbacks to notify of those events.
TcpServer will listen permanently for new connections and new data with Listen(), and support a graceful shutdown with Close(). 
We provide one option to send data to our clients, by the client id which is generated in our system with Send.

#### Main function - putting at all together

Obtains and manages all the other components in this system. It will include the TCP server that holds an array of TCP clients, and a connection to the Kafka broker for consuming and sending messages to it. Here are the main parts of main.go file:

````
func main() {
    log.Println("Launching server...")
	flag.Parse()
	tcpServer = messages.NewTcpServer(messages.Callbacks{
	    OnMessageReceived: onMessageReceivedFromClient,
	})
	producer = messages.NewProducer()
	consumer = messages.NewConsumer(messages.Callbacks{
	    OnMessageConsumed: onMessageSendToClient,
	})
	consumer.Consume()
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan,
	  syscall.SIGINT,
	  syscall.SIGTERM,
	  syscall.SIGQUIT,
	  syscall.SIGKILL)
	go func() {
	    for {
		    s := <-signalChan
			switch s {
			case syscall.SIGINT:
			    fmt.Println("syscall.SIGINT")
				cleanup()
			case syscall.SIGTERM:
			    fmt.Println("syscall.SIGTERM")
				cleanup()
			case syscall.SIGQUIT:
			    fmt.Println("syscall.SIGQUIT")
				cleanup()
			case syscall.SIGKILL:
			    fmt.Println("syscall.SIGKILL")
				cleanup()
			default:
			    fmt.Println("Unknown signal.")
			}
		}
	}()
    tcpServer.listen()
}
````

### How build docker

- Run Kafka

````
docker volume create my-kafka-vol
cd docker
docker-compose up -d
````

- Create and check a topic

````
# Log in to you Kafka broker
docker exec -it docker-kafka-1 /bin/bash

# Check version of Kafka
kafka-topics.sh --version

# List all current topics
kafka-topics.sh --bootstrap-server localhost:9092 --list

# Create a test-topic
kafka-topics.sh --bootstrap-server localhost:9092 --create --replication-factor 1 --partitions 2 --topic test_topic

# Describe the test-topic
kafka-topics.sh --bootstrap-server localhost:9092 --describe --topic test_topic
````

- Same but deployed as Kubernetes StatefulSet (Incomplete)
````
kubectl create namespace infrastructure
kubectl apply -f k8s/kafka.yml
````

### How build run server

### How build run client

### How build working example

- Run server:  
```go run main.go```

- Run client:  
  ```go run client.go```

### How build tcp server

### How build kafka producer

### How build kafka consumer

### How build a client

