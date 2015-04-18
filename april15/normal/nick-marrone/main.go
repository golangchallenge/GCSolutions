package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

// Dial generates a private/public key pair,
// connects to the server, perform the handshake
// and return a reader/writer.
func Dial(addr string) (io.ReadWriteCloser, error) {
	pub, priv, err := GenerateKeyPair()
	if err != nil {
		return nil, err
	}

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	// Exchange public keys. First, send key
	// to server. Then receive key from server.
	n, err := conn.Write(pub[:])
	if err != nil {
		return nil, err
	}

	// Receive public key from the server
	buf := make([]byte, KeyLength)
	n, err = conn.Read(buf)
	if err != nil || n == 0 {
		return nil, err
	}
	serverPub := new([KeyLength]byte)
	for i := 0; i < n; i++ {
		serverPub[i] = buf[i]
	}

	return NewSecureReaderWriter(conn, priv, serverPub), nil
}

// Serve starts a secure echo server on the given listener.
func Serve(l net.Listener) error {
	// Server needs its own public/private keypair
	pub, priv, err := GenerateKeyPair()
	if err != nil {
		return err
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}

		go handleConnection(conn, priv, pub)
	}
}

func handleConnection(c net.Conn, priv, pub *[KeyLength]byte) {
	buf := make([]byte, maxMsgLen)

	// Exchange public keys. First, receive public key
	// from dialer. Then send our key.
	dialerPub := new([KeyLength]byte)
	n, err := c.Read(buf)
	if err != nil || n == 0 {
		c.Close()
		log.Printf("Did not receive public key from %v.", c.RemoteAddr())
		return
	}
	for i, x := range buf[:n] {
		dialerPub[i] = x
	}

	// Send our public key to dialer
	n, err = c.Write(pub[:])
	if err != nil {
		c.Close()
		log.Printf("Unable to send public key to %v.", c.RemoteAddr())
		return
	}

	secureReaderWriter := NewSecureReaderWriter(c, priv, dialerPub)

	for {
		n, err := secureReaderWriter.Read(buf)
		if err != nil || n == 0 {
			c.Close()
			break
		}

		n, err = secureReaderWriter.Write(buf[0:n])
		if err != nil {
			c.Close()
			break
		}
	}
	log.Printf("Connection from %v closed.", c.RemoteAddr())
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
	if len(os.Args) != 3 {
		log.Fatalf("Usage: %s <port> <message>", os.Args[0])
	}
	conn, err := Dial("localhost:" + os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	if _, err := conn.Write([]byte(os.Args[2])); err != nil {
		log.Fatal(err)
	}
	buf := make([]byte, len(os.Args[2]))
	n, err := conn.Read(buf)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", buf[:n])
}
