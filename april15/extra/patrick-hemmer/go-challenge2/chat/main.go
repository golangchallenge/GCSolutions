// This is a simple multi-user chat server & client.
//
// In server mode, it listens for incoming connections, and then listens for
// messages on those connections. When a message is received, it is broadcasted
// out to all other connected clients.
//
// In client mode, it reads complete lines from STDIN, and sends them to the
// server. Messages from the server are dumped to STDOUT.
// Note that there is no fancy prompt here (rather complex to implement), and
// if a message comes in while you are typing, it'll garble things up.
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

func mainChatServer(port int) {
	l, err := Listen(fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()

	console := struct {
		io.Reader
		io.Writer
	}{os.Stdin, os.Stdout}

	hub := NewChatHub()
	go hub.AddClient("console", console)

	log.Fatal(hub.Listen(l))
}

func mainChatClient(addr string, message ...string) {
	conn, err := Dial(addr)
	if err != nil {
		log.Fatal(err)
	}

	if len(message) > 0 {
		// have a message. send to hub and exit
		_, err := fmt.Fprintf(conn, "%s\n", strings.Join(message, " "))
		if err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}

	errChan := make(chan error)
	go func() {
		_, err := io.Copy(conn, os.Stdin)
		errChan <- err
	}()
	go func() {
		_, err := io.Copy(os.Stdout, conn)
		errChan <- err
	}()
	if err := <-errChan; err != nil {
		log.Fatal(err)
	}

	os.Exit(0)
}

func main() {
	port := flag.Int("l", 0, "Listen mode. Specify port")
	flag.Parse()

	// Server mode
	if *port != 0 {
		mainChatServer(*port)
	}

	// Client mode
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <port> [message]", os.Args[0])
	}

	mainChatClient("localhost:"+os.Args[1], os.Args[2:]...)
}
