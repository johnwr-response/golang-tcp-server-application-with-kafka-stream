package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

var consumer = *messages.Consumer
var producer = *messages.Producer
var tcpServer = *messages.TcpServer

type ConsumerMsg struct {
	ID  string
	msg string
}

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

func onMessageReceivedFromClient(clientID string, msg []byte) {
	cm := ConsumerMsg{
		ID:  clientID,
		Msg: strings.ToUpper(string(msg)),
	}
	b, _ := json.Marshal(cm)
	m := producer.NewMessage(string(b))

	producer.Sent(m)

	//tcpServer.Connections[cm.ID].Send(msg)

}

func onMessageSendToClient(msg []byte) {
	var cm ConsumerMsg
	json.Unmarshal(msg, &cm)
	fmt.Println("Writeto connection: ", cm.ID, string(msg))

	if tcpServer.Connections[cm.ID] != nil {
		err := tcpServer.Connections[cm.ID].Send(msg)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		log.Println(fmt.Sprint("No connection with clientID ", cm.ID))
	}
}

func cleanup() {
	tcpServer.Close()
	producer.Close()
	consumer.Close()
	os.Exit(0)
}
