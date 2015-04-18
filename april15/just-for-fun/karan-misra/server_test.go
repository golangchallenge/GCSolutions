package main

import (
	"fmt"
	"net"
	"testing"
)

// The tests are modified slightly to listen on localhost instead of
// 0.0.0.0. This stops the OS X firewall from triggering a warning.

func TestSecureServe(t *testing.T) {
	// Create a random listener.
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	// Start the server.
	go Serve(l)

	conn, err := net.Dial("tcp", l.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	unexpected := "hello world\n"
	if _, err := fmt.Fprintf(conn, unexpected); err != nil {
		t.Fatal(err)
	}
	buf := make([]byte, 2048)
	n, err := conn.Read(buf)
	if err != nil {
		t.Fatal(err)
	}
	if got := string(buf[:n]); got == unexpected {
		t.Fatalf("Unexpected result:\nGot raw data instead of serialized key")
	}
}
