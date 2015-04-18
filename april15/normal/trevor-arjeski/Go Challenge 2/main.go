// Go Challenge #2
//
// Author: Trevor Arjeski - github.com/trevorarjeski
//
// I never wrote in Go before this. Pretty fun crash course
// into the language.

package main

import (
	"flag"
	"fmt"
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
		log.Fatal(Serve(l))
	}

	// Client mode
	if len(os.Args) != 3 {
		log.Fatalf("Usage: %s <port> <message>", os.Args[0])
	}
	// Dial in like the old AOL days
	conn, err := Dial("localhost:" + os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	// Write to the server
	if _, err := conn.Write([]byte(os.Args[2])); err != nil {
		log.Fatal(err)
	}
	buf := make([]byte, len(os.Args[2]))
	// Read from the server
	n, err := conn.Read(buf)
	if err != nil {
		log.Fatal(err)
	}
	// Print the message recieved...
	fmt.Printf("%s\n", buf[:n])
}
