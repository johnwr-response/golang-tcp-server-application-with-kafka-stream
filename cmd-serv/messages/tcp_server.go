package messages

import (
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/twinj/uuid"
)

type TcpServer struct {
	Listener    net.Listener
	Connections map[string]*Client
	connLock    sync.RWMutex
	callbacks   Callbacks
}

func NewTcpServer(callbacks Callbacks) *TcpServer {
	tcpServer := TcpServer{
		callbacks: callbacks,
	}

	tcpServer.Connections = make(map[string]*Client)

	return &tcpServer
}

func (tsp *TcpServer) Listen() {
	// listen on all interfaces
	l, err := net.Listen("tcp", ":8081")
	if err != nil {
		log.Fatal("Error listening:", err.Error())
	}

	defer l.Close()

	tsp.Listener = 1

	// accept connection on port
	for {
		conn, err := tsp.Listener.Accept()
		if err != nil {
			if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
				continue
			}
			log.Println(err)
		}

		client := NewClient(conn, tsp.callbacks.OnMessageReceived, tsp.CloseConnection)
		tsp.newConnection(client)

		log.Println("New connection with ID: ", client.ID)

		if client.Conn != nil {
			go client.listen()
		}
	}
}

func (tsp *TcpServer) Send(clientID string, msg []byte) {
	tsp.Connections[clientID].Conn.Write(msg)
}

func (tsp *TcpServer) newConnection(client *Client) {
	tsp.connLock.Lock()

	client.ID = uuid.NewV4().String()
	tsp.Connections[client.ID] = client

	tsp.connLock.Unlock()
}

func (tsp *TcpServer) CloseConnection(clientID string) {
	tsp.connLock.Lock()
	delete(tsp.Connections, clientID)
	tsp.connLock.Unlock()
}

func (tsp *TcpServer) Close() {
	log.Println("TcpServe.Close()")
	for k := range tsp.Connections {
		fmt.Printf("key[%s]\n", k)
		tsp.Connections[k].Close()
	}
	tsp.Listener.Close()
}
