package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"testing"
)

func TestSecureSession_interface(t *testing.T) {
	var _ io.ReadWriteCloser = &SecureSession{}
}

func TestSecureListener_interface(t *testing.T) {
	var _ net.Listener = &SecureListener{}
}

func TestSecureConn_interface(t *testing.T) {
	var _ net.Conn = &SecureConn{}
}

func TestSecureSession(t *testing.T) {
	l, err := Listen("127.0.0.1:0")
	if err != nil {
		t.Fatalf("Error listening: %s", err)
	}

	conn1, err := Dial(l.Addr().String())
	if err != nil {
		t.Fatalf("Error dialing: %s", err)
	}

	conn2, err := l.Accept()
	if err != nil {
		t.Fatalf("Error accepting: %s", err)
	}

	if err := testSecureSessionMessage(conn1, conn2); err != nil {
		t.Fatalf("Error writing through conn1: %s", err)
	}

	if err := testSecureSessionMessage(conn2, conn1); err != nil {
		t.Fatalf("Error writing through conn2: %s", err)
	}

	if err := conn1.Close(); err != nil {
		t.Fatalf("Error closing conn1: %s", err)
	}

	if err := testSecureSessionMessage(conn1, conn2); err == nil {
		t.Fatalf("Expected error writing to closed connection. None received.")
	}
	if err := testSecureSessionMessage(conn2, conn1); err == nil {
		t.Fatalf("Expected error writing to closed connection. None received.")
	}

	if err := conn2.Close(); err != nil {
		t.Fatalf("Error closing conn2: %s", err)
	}
}

func testSecureSessionMessage(conn1, conn2 io.ReadWriter) error {
	pt := []byte("hello peer!")
	nw, err := conn1.Write(pt)
	if err != nil {
		return fmt.Errorf("write error: %s", err)
	}

	ptr := make([]byte, len(pt))
	nr, err := conn2.Read(ptr)
	if err != nil {
		return fmt.Errorf("read error: %s", err)
	}

	if nr != nw {
		return fmt.Errorf("bytes read not equal to bytes written: %d != %d", nr, nw)
	}

	if !bytes.Equal(ptr, pt) {
		return fmt.Errorf("message read not equal to message written: %q != %q", ptr, pt)
	}

	return nil
}
