package messages

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
)

type Client struct {
	ID                string
	Conn              net.Conn
	onMessageReceived func(clientID string, msg []byte)
	closeConnection   func(clientID string)
}

func NewClient(conn net.Conn, onMessageReceived func(clientID string, msg []byte), closeConnection func(clientID string)) *Client {
	return &Client{
		Conn:              conn,
		onMessageReceived: onMessageReceived,
		closeConnection:   closeConnection,
	}
}

func (c *Client) listen() {
	defer func() {
		fmt.Println("Closing connection from", c.Conn.RemoteAddr())
		c.Close()
		c.closeConnection(c.ID)
	}()

	for {
		reader := bufio.NewReader(c.Conn)
		buf := make([]byte, 1024)

		for {
			n, err := reader.Read(buf)
			switch err {
			case io.EOF:
				// connection terminated
				c.Close()
				c.closeConnection(c.ID)
				return
			case nil:
				// output message received
				log.Print("Message received:", string(buf[:n]))
				// send new string back to client
				c.onMessageReceived(c.ID, buf[:n])
			default:
				log.Fatalf("Receive data failed:%s", err)
				c.Close()
				c.closeConnection(c.ID)
				return
			}
		}
	}
}

func (c *Client) Send(b []byte) error {
	_, err := c.Conn.Write(b)
	return err
}

func (c *Client) Close() error {
	if c.Conn != nil {
		return c.Conn.Close()
	}
	return nil
}
