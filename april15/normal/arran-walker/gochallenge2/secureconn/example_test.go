package secureconn_test

import (
	"bufio"
	"fmt"
	"net"
	"os"

	"gochallenge2/secureconn"
)

// Secure chat server
func ExampleSecureConn_server() error {
	// create listener
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		return err
	}

	// connected clients
	clients := make([]*secureconn.SecureConn, 0, 20)

	type message struct {
		text   string
		client *secureconn.SecureConn
	}

	// channels used for client joins/leaves and messages
	messaging := make(chan message)
	joining := make(chan *secureconn.SecureConn)
	leaving := make(chan *secureconn.SecureConn)

	// control goroutine for channels
	go func() {
		for {
			select {
			case msg := <-messaging:
				// send all messages to all clients connected
				for _, c := range clients {
					if c == msg.client {
						// don't send clients their own messages
						continue
					}
					c.Write([]byte(msg.text + "\n"))
				}

			case join := <-joining:
				// add client
				clients = append(clients, join)

			case leave := <-leaving:
				for idx, c := range clients {
					if c == leave {
						// remove client
						clients = append(clients[:idx], clients[idx+1:]...)
						break
					}
				}
			}
		}
	}()

	// accept incoming connections
	for {
		conn, err := ln.Accept()
		if err != nil {
			return err
		}
		// handle accepted connection
		go func(c net.Conn) {
			defer c.Close()

			// secure the connection
			sc, err := secureconn.New(c)
			if err != nil {
				return
			}
			defer sc.Close()

			// handle joining and leaving chat
			joining <- sc
			defer func() {
				leaving <- sc
			}()

			// start reading incoming messages
			scanner := bufio.NewScanner(sc)
			for scanner.Scan() {
				// send message to messaging channel
				messaging <- message{scanner.Text(), sc}
			}
		}(conn)
	}
}

// Secure chat client
func ExampleSecureConn_client() error {
	// connect to remote server
	c, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		return err
	}
	defer c.Close()

	// wrap insecure client, perform key exchange and retrieve
	// a secured connection
	sc, err := secureconn.New(c)
	if err != nil {
		return err
	}
	defer sc.Close()

	// read messages from the server line by line
	go func() {
		scanner := bufio.NewScanner(sc)
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
	}()

	// read messages from command line to send to server
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		_, err := sc.Write([]byte(scanner.Text() + "\n"))
		if err != nil {
			return err
		}
	}

	return nil
}
