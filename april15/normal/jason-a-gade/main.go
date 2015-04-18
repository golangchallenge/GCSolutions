package main

import (
	"flag"
	"fmt"
	"io"
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
	conn, err := Dial("localhost:" + os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	var n int
	if n, err = conn.Write([]byte(os.Args[2])); err != nil {
		log.Fatal(err)
	}
	buf := make([]byte, n)
	n, err = conn.Read(buf)
	if err != nil && err != io.EOF {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", buf[:n])
}
