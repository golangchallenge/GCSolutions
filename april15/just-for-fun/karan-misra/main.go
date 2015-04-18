package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
)

func main() {
	host := flag.String("h", "", "Specify host")
	port := flag.Int("l", 0, "Listen mode. Specify port")
	flag.Parse()

	// Server mode.
	if *port != 0 {
		l, err := net.Listen("tcp", fmt.Sprintf("%s:%d", *host, *port))
		if err != nil {
			log.Fatal(err)
		}
		defer l.Close()
		log.Fatal(Serve(l))
	}

	// Client mode.
	if flag.NArg() < 1 {
		log.Fatalf("Usage: %s <port> <message>. If <message> is not given, it will read from STDIN.", os.Args[0])
	}

	conn, err := Dial("localhost:" + flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	if flag.NArg() > 1 {
		if _, err := conn.Write([]byte(flag.Arg(1))); err != nil {
			log.Fatal(err)
		}

		buf := make([]byte, len(flag.Arg(1)))
		n, err := conn.Read(buf)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("%s\n", buf[:n])

		return
	}

	// Read file from STDIN.
	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}

	v(2).Printf("main: read %v bytes from stdin", len(data))

	if _, err := conn.Write(data); err != nil {
		log.Fatal(err)
	}

	buf := make([]byte, len(data))
	n, err := conn.Read(buf)
	if err != nil {
		log.Fatal(err)
	}

	v(2).Printf("main: read %v bytes back from the server", n)

	if _, err := os.Stdout.Write(buf[:n]); err != nil {
		log.Fatal(err)
	}
}
