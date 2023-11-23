package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func main() {

	// connect to this socket
	conn, err := net.Dial("tcp", "127.0.0.1:8081")
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		defer func() {
			conn.Close()
			log.Fatal("Connection closes ", conn.LocalAddr().String())
		}()

		for {
			message := make([]byte, 256)
			n, err := conn.Read(message)

			if n > 0 {
				log.Println("RECEIVED: " + strings.TrimRight(string(message[:n]), "\n"))
				log.Print("Text to send: ")
			}
			if err == io.EOF {
				conn.Close()
				log.Print("Connection closed: ", conn.LocalAddr().String())
			}
			if err != nil {
				log.Println("Client: Read", err)
				conn.Close()
				return
			}

		}
	}()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGKILL)

	go func() {
		for {
			s := <-signalChan
			switch s {
			case syscall.SIGINT:
				fmt.Println("syscall.SIGINT")
				conn.Close()
				os.Exit(0)
			default:
				fmt.Println("Unknown signal")
			}
		}
	}()

	log.Print("Text to send:")
	for {
		// read in input fromstdin
		reader := bufio.NewReader(os.Stdin)
		text, _ := reader.ReadString('\n')
		// send to socket
		conn.Write([]byte(strings.TrimRight(text, "\n")))
	}

}
