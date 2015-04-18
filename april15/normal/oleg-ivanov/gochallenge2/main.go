// Secure chat based on public key cryptography from crypto/nacl.
//
// The chat client and server follow a simple communication protocol, which
// begins with a handshake to exchange public keys. Handshake is followed by
// exchange of message packets, which include metadata and encrypted
// message text itself.
//
// Handshake phase
//
// Upon establishing the connection, the first step for both the client and the
// server is the handshake, with the following two steps (in exactly this order):
// 1. Server sends its public key to the client.
// 2. Client sends its public key to the server.
//
// Keys are 32 bytes long, and are sent as binary data.
//
// Message protocol
//
// Once the client and server have exchanged public keys, they can begin
// sending messages to each other. Each message packet has the following
// structure:
//
//   type                  | description
//   -----------------------------------------------
//   [24]byte              | nonce for this message
//   uint16, little-endian | length of the ciphertext
//   []byte                | message ciphertext
//
package main

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"golang.org/x/crypto/nacl/box"
)

// message length, capped at 32KB by the challenge setup
const msglen = 32767

// nonce length, as set by crypto/nacl
const noncelen = 24

// key length, as set by crypto/nacl
const keylen = 32

// endianess used for binary operations
var endianess = binary.LittleEndian

// metadata packet structure. Length field size can be limited to uint16,
// as challenge setup capped message length at 32KB - so it fits even with
// encryption overhead
type packet struct {
	Nonce [noncelen]byte
	Len   uint16
}

// NewSecureReader instantiates a new SecureReader
func NewSecureReader(r io.Reader, priv, pub *[keylen]byte) io.Reader {
	br := readbox{in: r}
	// precompute and store shared key in the readbox
	box.Precompute(&br.key, pub, priv)
	return br
}

// NewSecureWriter instantiates a new SecureWriter
func NewSecureWriter(w io.Writer, priv, pub *[keylen]byte) io.Writer {
	bw := writebox{out: w}
	// precompute and store shared key in the writebox
	box.Precompute(&bw.key, pub, priv)
	return bw
}

// Dial generates a private/public key pair,
// connects to the server, perform the handshake
// and return a reader/writer.
func Dial(addr string) (io.ReadWriteCloser, error) {
	con, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	return clientHandshake(con)
}

// Serve starts a secure echo server on the given listener.
func Serve(l net.Listener) error {
	for {
		c, err := l.Accept()
		if err != nil {
			return err
		}
		go func(c net.Conn) {
			defer c.Close()
			// perform handshake and receive secure ReadWriter
			rw, err := serverHandshake(c)
			if err != nil {
				log.Print(err)
				return
			}
			// launch echo server on prepared ReadWriter
			if err = echo(rw); err != nil && err != io.EOF {
				log.Print(err)
			}
		}(c)
	}
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
	if err != nil && err != io.EOF {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", buf[:n])
}

// clientHandshake implements a client side of handshake protocol.
// It generates client's keypair, reads server's public key, and writes
// its own public key to the server. ReadWriteCloser that it returns is
// composed of initialised secure reader and writer.
func clientHandshake(c net.Conn) (io.ReadWriteCloser, error) {
	mypub, mypriv, err := box.GenerateKey(rand.Reader)

	// actual handshake - reading server key, writing client key
	theirpub, err := readKey(err, c)
	err = writeKey(err, c, mypub)
	if err != nil {
		return nil, err
	}

	return struct {
		io.Reader
		io.Writer
		io.Closer
	}{
		NewSecureReader(c, mypriv, theirpub),
		NewSecureWriter(c, mypriv, theirpub),
		c,
	}, nil
}

// serverHandshake implements server side of handshake protocol.
// It generates server's keypair, and then writes server's public key into
// the connection, and reads client's key. Returned ReadWriter is composed
// of a secure reader and writer, initialised with the proper keys.
func serverHandshake(c net.Conn) (io.ReadWriter, error) {
	mypub, mypriv, err := box.GenerateKey(rand.Reader)

	// actual handshake - writing server key, reading client key
	err = writeKey(err, c, mypub)
	theirpub, err := readKey(err, c)
	if err != nil {
		return nil, err
	}

	return struct {
		io.Reader
		io.Writer
	}{
		NewSecureReader(c, mypriv, theirpub),
		NewSecureWriter(c, mypriv, theirpub),
	}, nil
}

// echo server on top of a generic io.ReadWriter. Reads from the reader,
// and writes received data back to the writer. To implement secure echo,
// rw must be composed of a secure reader and writer, as produced by
// serverHandshake.
func echo(rw io.ReadWriter) error {
	var (
		n   int
		err error
	)
	buf := make([]byte, msglen)

	for {
		if n, err = rw.Read(buf); err != nil {
			return err
		}
		if _, err := rw.Write(buf[:n]); err != nil {
			return err
		}
	}
}

// readbox implements io.Reader interface, allowing its clients
// to read from a secure communication channel
type readbox struct {
	key [keylen]byte
	in  io.Reader
}

// Read reads and decrypts next encrypted message
func (br readbox) Read(b []byte) (int, error) {
	var pkt packet

	// reading metadata packet first
	if err := binary.Read(br.in, endianess, &pkt); err != nil {
		return 0, err
	}

	// as metadata specifies message len, create its buffer and
	// read the actual message ciphertext
	msg := make([]byte, pkt.Len)
	if _, err := io.ReadFull(br.in, msg); err != nil {
		return 0, err
	}

	// with message nonce (received with metadata packet), and
	// the ciphertext we can perform decryption
	txt, ok := box.OpenAfterPrecomputation([]byte{}, msg, &pkt.Nonce,
		&br.key)
	if !ok {
		return 0, errors.New("failed to decrypt the message")
	}
	return copy(b, txt), nil
}

// writebox implements io.Writer interface, allowing its clients to write
// data to a secure communication channel
type writebox struct {
	key [keylen]byte
	out io.Writer
}

// Write encrypts given data and writes it to the output writer
func (bw writebox) Write(b []byte) (int, error) {
	// a unique nonce must be generated for every message
	nn, err := nonce()
	if err != nil {
		return 0, err
	}
	// with nonce generated (and shared key precomputed),
	// we can encrypt the message
	msg := box.SealAfterPrecomputation([]byte{}, b, nn, &bw.key)

	// producing and writing message metadata packet first
	pkt := packet{
		Len:   uint16(len(msg)),
		Nonce: *nn,
	}
	if err := binary.Write(bw.out, endianess, pkt); err != nil {
		return 0, err
	}
	// and writing the message ciphertext itself
	return bw.out.Write(msg)
}

// nonce generates and returns a unique cryptographic nonce
func nonce() (*[noncelen]byte, error) {
	buf := make([]byte, noncelen)
	if _, err := rand.Read(buf); err != nil {
		return nil, err
	}

	// convert slice to an array, as nacl/box's crypto primitives
	// operate on arrays
	var n [noncelen]byte
	copy(n[:], buf[0:noncelen])

	return &n, nil
}

// writeKey writes a key into the given io.Writer. No-op if error
// argument is not nil, returning the same error.
func writeKey(err error, c io.Writer, key *[keylen]byte) error {
	if err != nil {
		return err
	}

	_, err = c.Write(key[:])
	return err
}

// readKey reads a key from io.Reader. No-op if error argument is not nil,
// returning the same error.
func readKey(err error, c io.Reader) (*[keylen]byte, error) {
	if err != nil {
		return nil, err
	}

	key := &[keylen]byte{}
	err = binary.Read(c, endianess, key)
	return key, err
}
