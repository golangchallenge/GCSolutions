package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"sync"
)

type client struct {
	name      string
	rw        io.ReadWriter
	msgChan   chan *message
	closeChan chan bool
	hub       *ChatHub
}

type message struct {
	sender  *client
	content string
}

// ChatHub is a chat server which receives messages, and broadcasts them out
// to all connected clients.
type ChatHub struct {
	msgChan chan *message

	clients      []*client
	clientsMutex sync.Mutex
}

// NewChatHub creates and starts a ChatHub server.
func NewChatHub() *ChatHub {
	ch := &ChatHub{
		msgChan: make(chan *message, 10),
	}

	go ch.broadcaster()

	return ch
}

// broadcaster listens for messages and broadcasts the message to all connected
// clients.
//
// If a client's message buffer fills up (client not reading fast enough),
// the client will be kicked off.
func (ch *ChatHub) broadcaster() {
	for {
		msg, ok := <-ch.msgChan
		if !ok {
			// channel closed
			return
		}

		ch.clientsMutex.Lock()
		for _, cl := range ch.clients {
			if cl == msg.sender {
				continue
			}

			select {
			case cl.msgChan <- msg:
			default:
				// client's msgChan buffer is full. Kick them off
				go ch.removeClient(cl)
			}
		}
		ch.clientsMutex.Unlock()
	}
}

// Listen listens for incoming connections and adds them to the hub.
func (ch *ChatHub) Listen(l net.Listener) error {
	for {
		c, err := l.Accept()
		if err != nil {
			return err
		}

		clientName := c.RemoteAddr().String()
		go ch.AddClient(clientName, c)
	}
}

// AddClient builds a new client object, adds it to the hub so that it can
// receive messages, and starts reading from the client sending messages to the
// hub.
// AddClient returns when the client disconnects, or is kicked off the hub.
func (ch *ChatHub) AddClient(name string, rw io.ReadWriter) {
	cl := &client{
		name:      name,
		rw:        rw,
		msgChan:   make(chan *message, 10),
		closeChan: make(chan bool),
	}

	ch.clientsMutex.Lock()
	ch.clients = append(ch.clients, cl)
	ch.clientsMutex.Unlock()

	errChan := make(chan error)
	go func() {
		errChan <- cl.sendToHub(ch.msgChan)
	}()
	go func() {
		errChan <- cl.sendToClient()
	}()

	<-errChan
	// ignore error
	ch.removeClient(cl)
}

// removeClient removes the client from the hub and then closes it.
func (ch *ChatHub) removeClient(cl *client) {
	ch.clientsMutex.Lock()
	for i, hcl := range ch.clients {
		if hcl != cl {
			continue
		}

		copy(ch.clients[i:], ch.clients[i+1:])
		ch.clients[len(ch.clients)-1] = nil
		ch.clients = ch.clients[:len(ch.clients)-1]
		break
	}
	ch.clientsMutex.Unlock()

	cl.Close()
}

// sendToHub reads from the client and sends messages to the hub.
func (cl *client) sendToHub(msgChan chan *message) error {
	scanner := bufio.NewScanner(cl.rw)

	for scanner.Scan() {
		msg := &message{
			sender:  cl,
			content: scanner.Text(),
		}

		select {
		case msgChan <- msg:
		case <-cl.closeChan:
			return nil
		}
	}

	return scanner.Err()
}

// sendToClient listes for messages from the hub and writes them to the client.
func (cl *client) sendToClient() error {
	for {
		select {
		case msg := <-cl.msgChan:
			if msg == nil {
				return nil
			}
			_, err := fmt.Fprintf(cl.rw, "%s: %s\n", msg.sender.name, msg.content)
			if err != nil {
				return err
			}
		case <-cl.closeChan:
			return nil
		}
	}
}

// Close disconnects the client (if the client satisfies io.Closer), and closes
// all its channels.
func (cl *client) Close() {
	if closer, ok := cl.rw.(io.Closer); ok {
		closer.Close()
	}
	close(cl.closeChan)
	close(cl.msgChan)
}
