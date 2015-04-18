package main

import (
	"fmt"
	"net"
	"testing"
)

// The tests are modified slightly to listen on localhost instead of
// 0.0.0.0. This stops the OS X firewall from triggering a warning.

func TestSecureEchoServer(t *testing.T) {
	// Create a random listener.
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	// Start the server.
	go Serve(l)

	conn, err := Dial(l.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	expected := "hello world\n"
	if _, err := fmt.Fprintf(conn, expected); err != nil {
		t.Fatal(err)
	}

	buf := make([]byte, 2048)
	n, err := conn.Read(buf)
	if err != nil {
		t.Fatal(err)
	}

	if got := string(buf[:n]); got != expected {
		t.Fatalf("Unexpected result:\nGot:\t\t%s\nExpected:\t%s\n", got, expected)
	}
}

func TestSecureOrdering(t *testing.T) {
	// Create a random listener.
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	// Start the server.
	go Serve(l)

	conn, err := Dial(l.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	data := []string{"one", "two", "three"}
	for _, want := range data {
		if _, err := fmt.Fprint(conn, want); err != nil {
			t.Fatal(err)
		}
	}

	for _, want := range data {
		buf := make([]byte, 2048)
		n, err := conn.Read(buf)
		if err != nil {
			t.Fatal(err)
		}

		if got := string(buf[:n]); got != want {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}
