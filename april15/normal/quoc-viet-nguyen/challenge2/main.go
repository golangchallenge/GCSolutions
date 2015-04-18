/*
Package main includes a simple echo server and client that utilizes NaCl to
securely transmitting messages.

Examples of running server and client:

  $ ./challenge2 -l 8080 &
  $ ./challenge2 8080 "hello world"
  hello world
*/
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"
)

// readWriteCloser wraps io.Reader, io.Writer and io.Closer.
type readWriteCloser struct {
	io.Reader
	io.Writer
	io.Closer
}

// Dial generates a private/public key pair,
// connects to the server, perform the handshake
// and return a reader/writer.
func Dial(addr string) (io.ReadWriteCloser, error) {
	priv, pub, err := generateKeyPair()
	if err != nil {
		return nil, err
	}

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	if err = handShake(conn, priv, pub); err != nil {
		conn.Close()
		return nil, err
	}
	// We should have a NewSecureReadWriter to avoid recalculation of
	// shared key.
	rwc := &readWriteCloser{
		NewSecureReader(conn, priv, pub),
		NewSecureWriter(conn, priv, pub),
		conn,
	}
	return rwc, nil
}

// Serve starts a secure echo server on the given listener.
func Serve(l net.Listener) error {
	var tempDelay time.Duration // how long to sleep on accept failure
	for {
		conn, err := l.Accept()
		if err != nil {
			// This error handling is taken from http.Server.Serve.
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				log.Printf("accept error: %v; retrying in %v", err, tempDelay)
				time.Sleep(tempDelay)
				continue
			}
			return err
		}
		tempDelay = 0
		go handleClient(conn)
	}
}

// handleClient generates a key pair, performs the handshake and echoes client's
// messages.
func handleClient(conn net.Conn) {
	defer conn.Close()

	priv, pub, err := generateKeyPair()
	if err != nil {
		log.Printf("%v: could not generate keypair: %v", conn.RemoteAddr(), err)
		return
	}
	if err = handShake(conn, priv, pub); err != nil {
		log.Printf("%v: %v", conn.RemoteAddr(), err)
		return
	}
	reader := NewSecureReader(conn, priv, pub)
	writer := NewSecureWriter(conn, priv, pub)
	// Echo messages.
	// Since maxMessageSize equals to the size of buffer used in io.Copy,
	// we can utilize that function to implement a echo server.
	// If fault tolerance is required, we can also flush pending data
	// in the socket and continue listening if encounted ErrInvalidHeader
	// or ErrInvalidMessage. But for now, just close the connection.
	if _, err = io.Copy(writer, reader); err != nil {
		if err != io.EOF {
			log.Printf("%v: %v", conn.RemoteAddr(), err)
		}
	}
}

// handShake exchanges public key between client and server.
// As pub will no longer be used, it will be overridden with peer public key.
func handShake(rw io.ReadWriter, priv, pub *[keySize]byte) error {
	if _, err := rw.Write(pub[:]); err != nil {
		return err
	}
	if _, err := io.ReadFull(rw, pub[:]); err != nil {
		return err
	}
	return nil
}

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
	if len(os.Args) < 3 {
		log.Fatalf("Usage: %s <port> <message>...", os.Args[0])
	}
	conn, err := Dial("localhost:" + os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	// Modified to support multiple messages.
	for i := 2; i < len(os.Args); i++ {
		if _, err := conn.Write([]byte(os.Args[i])); err != nil {
			log.Fatal(err)
		}

		buf := make([]byte, len(os.Args[i]))
		n, err := conn.Read(buf)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s\n", buf[:n])
	}
}
