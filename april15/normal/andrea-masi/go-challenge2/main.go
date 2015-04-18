// Go Challenge 2, Andrea Masi's solution
//
//	http://golang-challenge.com/go-challenge2/
//
// Use "-tags debug" when building to enable debug mode es:
//
//	go build -tags debug -o challenge2
//	go test -tags debug -v
//
// Debug messeges are written to stderr.
package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"golang.org/x/crypto/nacl/box"
)

// A secureMessageR wraps an io.Reader and decrypt NaCl boxes
// read from underlying Read method.
type secureMessageR struct {
	reader    io.Reader
	priv, pub *[32]byte
}

// A secureMessageW wraps an io.Writer and encrypt NaCl boxes
// before calling underlying Write method.
type secureMessageW struct {
	writer    io.Writer
	priv, pub *[32]byte
}

// A secureConnection is a mockup struct that embeds
// required interfaces to implements io.ReadWriteCloser returned in Dial.
type secureConnection struct {
	io.Reader
	io.Writer
	io.Closer
}

// Read makes secureMessageR implent io.Reader.
// It decrypts NaCl boxes read from underlying Read method.
func (r *secureMessageR) Read(decryptedData []byte) (n int, err error) {
	if len(decryptedData) == 0 {
		return 0, nil
	}
	encData := make([]byte, 32+box.Overhead+24)
	var nonce [24]byte
	n, err = r.reader.Read(encData)
	if err != nil {
		return
	}
	debugPrintln("[DEBUG] encrypted data read:", encData[:n], "bytes read:", n)
	copy(nonce[:], encData[:24])
	localDecryptedData, ok := box.Open([]byte{}, encData[24:n], &nonce, r.pub, r.priv)
	if !ok {
		return n, fmt.Errorf("error decrypting NaCl box")
	}
	n = copy(decryptedData, localDecryptedData)
	debugPrintln("[DEBUG] decrypted message:", string(decryptedData[:n]))
	return
}

// Writer makes secureMessageW implement io.Writer.
// It encrypts NaCl boxes before calling underlying Write method.
func (w *secureMessageW) Write(dataToEnc []byte) (n int, err error) {
	l := len(dataToEnc)
	if l > 32 {
		fmt.Fprintln(os.Stderr, "[WARNING] Message is too long, it will be truncated at 32th byte.")
		l = 32
	}
	var nonce [24]byte
	encData := make([]byte, len(dataToEnc)+box.Overhead+24)
	_, err = rand.Read(nonce[:])
	if err != nil {
		return
	}
	copy(encData, box.Seal(nonce[:], dataToEnc[:l], &nonce, w.pub, w.priv))
	n, err = w.writer.Write(encData)
	if err != nil {
		return
	}
	debugPrintln("[DEBUG] encrypted data to write:", encData, "bytes written:", n)
	return
}

// NewSecureReader instantiates a new secureMessageW.
func NewSecureReader(r io.Reader, priv, pub *[32]byte) io.Reader {
	return &secureMessageR{r, priv, pub}
}

// NewSecureWriter instantiates a new secureMessageW.
func NewSecureWriter(w io.Writer, priv, pub *[32]byte) io.Writer {
	return &secureMessageW{w, priv, pub}
}

// Dial generates a private/public key pair,
// connects to the server, perform the handshake
// and return a reader/writer.
func Dial(addr string) (io.ReadWriteCloser, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	pub, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	// send our public key
	_, err = conn.Write(pub[:])
	if err != nil {
		return nil, err
	}
	debugPrintln("[DEBUG] sended pub key @Dial:", *pub)
	buff := make([]byte, 32)
	receivedPubKey := [32]byte{}
	// receive remote public key
	n, err := conn.Read(buff)
	if err != nil {
		return nil, err
	}
	debugPrintln("[DEBUG] received pub key @Dial:", buff[:n])
	copy(receivedPubKey[:], buff[:n])
	r := NewSecureReader(conn, priv, &receivedPubKey)
	w := NewSecureWriter(conn, priv, &receivedPubKey)
	return secureConnection{r, w, conn}, nil
}

// Serve starts a secure echo server on the given listener.
// During handshake Serve expects to receive public key
// from client before sending out its own.
func Serve(l net.Listener) error {
	conn, err := l.Accept()
	if err != nil {
		return err
	}
	defer conn.Close()
	pub, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}
	buff := make([]byte, 32)
	receivedPubKey := [32]byte{}
	// receive sender's public key
	n, err := conn.Read(buff)
	if err != nil {
		return err
	}
	copy(receivedPubKey[:], buff[:n])
	debugPrintln("[DEBUG] received pub key @Serve:", receivedPubKey)
	// send our public key
	_, err = conn.Write(pub[:])
	if err != nil {
		return err
	}
	debugPrintln("[DEBUG] sended pub key @Serve:", *pub)
	dec := NewSecureReader(conn, priv, &receivedPubKey)
	n, err = dec.Read(buff)
	if err != nil {
		debugPrintln("[DEBUG] error reading from client, Serve returns.", err)
		return err
	}
	debugPrintln("[DEBUG] decrypted message received:", string(buff[:n]))
	enc := NewSecureWriter(conn, priv, &receivedPubKey)
	_, err = enc.Write(buff[:n])
	if err != nil {
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
