package main

import (
	"flag"
	"fmt"
	"go-challenge2"
	"log"
	"net"
	"os"
)

func main() {
	port := flag.Int("l", 0, "Listen mode. Specify port")
	flag.Parse()

	// Server mode
	if *port != 0 {
		l, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
		if err != nil {
			log.Fatal(err)
		}
		defer l.Close()
		log.Fatal(gc2.Serve(l))
	}

	// Client mode
	if len(os.Args) != 2 {
		log.Fatalf("Usage: %s <port>", os.Args[0])
	}
	conn, err := gc2.Dial("localhost:" + os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	dsp := newDisplay(os.Stdout, "> ")
	kb := newKeyboard(os.Stdin)
	machine := conn
	term := NewTerminal(kb, dsp, machine)

	log.Fatal(term.Run())
}
