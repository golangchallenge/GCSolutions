package main

import (

	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"

)

const MAX_BYTE_LENGTH = 32768 // The maximum lenght in bytes of the string to send

// ------------------- SERVER  -------------------------------------

// Serve starts a secure echo server on the given listener.
func Serve(l net.Listener) error {
	fmt.Println("Starting server: ", l.Addr())
	for {
		conn, err := l.Accept()
		defer l.Close()		
		fmt.Println("Incoming connection: ", l.Addr())
		if err != nil {
			return err
		}
		go handleServerConn(conn)
	}
}

// HandleServerConn reads the public key from a connecting client,
// Generates a public and a private key, and sends the public key to
// the client. Creates an instance of secureReader and secureWriter
// using the client's connection, decodes the data coming from the 
// client using SecureReader.Read, and encodes the result again using
// SecureWriter.Write to echo it back to the client.
func handleServerConn(c net.Conn) {	
	defer c.Close()
	key := make([]byte, 32)	
	n, err := c.Read(key)	

	if err != nil {
		log.Println("Error when reading Client's public key ", err)
		return
	}

	var cpub [32]byte
	copy(cpub[:], key)

	// Generate our server side keys
	pub, priv, err := GenerateKeys()
	// Send the public key to the client	
	n, err = c.Write(pub[:])	

	sr := NewSecureReader(c, priv, &cpub)
	buff := make([]byte, MAX_BYTE_LENGTH)
	n, err = sr.Read(buff)	
	if !(err == nil || err == io.EOF) {
		log.Println("Error when decoding data: ", err)
		return
	}
	buff = buff[:n]
	sw := NewSecureWriter(c, priv, &cpub)
	n, err = sw.Write(buff)	
	return
}


// Dial generates a private/public key pair,
// connects to the server, perform the handshake
// and return a reader/writer.
func Dial(addr string) (io.ReadWriteCloser, error) {
	pub, priv, err := GenerateKeys()
	if err != nil {
		return nil, err
	}

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	var spub *[32]byte
	spub, err = doHandShake(conn, pub)

	if err != nil {
		return nil, err
	}
	w := NewSecureWriter(conn, priv, spub)
	r := NewSecureReader(conn, priv, spub)
	rwc := secureRWC{Reader: r, Writer: w, Closer: conn}
	return rwc, nil
}

// Do handShake performs a handshake with a Sever.
// First it sends the client's public key to the server,
// then it receives the sever public key.
// Returns the server public key, and an error if any.
func doHandShake(conn net.Conn, pub *[32]byte) (sp *[32]byte, err error) {
	n, err := conn.Write(pub[:])
	if err != nil {
		return nil, err
	}

	buff := make([]byte, 32)
	n, err = conn.Read(buff)	
	if err != nil {
		return nil, err
	}

	if n != 32 {
		return nil, errors.New("The server's public key is not valid")
	}
	var spublic [32]byte
	copy(spublic[:], buff)
	return &spublic, nil
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
