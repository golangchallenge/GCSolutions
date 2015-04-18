package main

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"golang.org/x/crypto/nacl/box"
	"io"
	"log"
	"net"
	"os"
)

const (
	nonceSize      = 24 // matches size used by NaCl box
	maxMessageSize = 32 * 1024
	keySize        = 32 // matches size used by NaCl box
)

/*
secureReader reads sealed messages from its underlying reader and opens
them using a precomputed shared key (see box.Precompute). It returns bytes
from the opened messages to its caller.

The secureReader must deal with its caller not reading the whole opened
message.  This means it requires a buffer and a pointer to track what has
been returned to its caller.
*/
type secureReader struct {
	r             io.Reader
	sharedKey     [keySize]byte
	openedBuffer  []byte
	openedPointer uint16
}

func (sr *secureReader) Read(p []byte) (n int, err error) {
	if !sr.hasOpenedMessage() {
		err = sr.readAndOpenMessage()
		if err != nil {
			return 0, err
		}
	}
	// return minimum of (len(p), remaining opened bytes)
	var bytesToReturn int
	openedBytesRemaining := len(sr.openedBuffer) - int(sr.openedPointer)
	if len(p) < openedBytesRemaining {
		bytesToReturn = len(p)
	} else {
		bytesToReturn = openedBytesRemaining
	}

	// copy bytes to supplied buffer and advance pointer
	sliceFrom := int(sr.openedPointer)
	sliceTo := sliceFrom + bytesToReturn
	copy(p[:bytesToReturn], sr.openedBuffer[sliceFrom:sliceTo])
	sr.openedPointer += uint16(bytesToReturn)

	return bytesToReturn, nil
}

func (sr *secureReader) hasOpenedMessage() bool {
	return sr.openedBuffer != nil &&
		int(sr.openedPointer) < len(sr.openedBuffer)
}

func (sr *secureReader) readAndOpenMessage() error {
	var totalLength uint16

	// read total length
	err := binary.Read(sr.r, binary.BigEndian, &totalLength)
	if err != nil {
		return err
	}

	// read nonce
	var nonce [nonceSize]byte
	_, err = io.ReadFull(sr.r, nonce[:])
	if err != nil {
		return err
	}

	// read sealed message (length = total length - nonce)
	sealedMessage := make([]byte, totalLength-nonceSize)
	_, err = io.ReadFull(sr.r, sealedMessage)
	if err != nil {
		return err
	}

	// open sealed message
	sr.openedBuffer = sr.openedBuffer[0:0]
	var ok bool
	sr.openedBuffer, ok = box.OpenAfterPrecomputation(sr.openedBuffer, sealedMessage, &nonce, &sr.sharedKey)
	if !ok {
		return errors.New("failed to open NaCl box")
	}
	sr.openedPointer = 0
	return nil
}

// NewSecureReader instantiates a new SecureReader
func NewSecureReader(r io.Reader, priv, pub *[keySize]byte) io.Reader {
	var sr secureReader

	sr.r = r
	// pre-size the buffer to the max message size
	sr.openedBuffer = make([]byte, 0, maxMessageSize)
	sr.openedPointer = 0
	box.Precompute(&sr.sharedKey, pub, priv)
	return &sr
}

/*
secureWriter receives bytes from its caller and writes sealed messages to its
underlying writer using a precomputed shared key (see box.Precompute).

Since the NaCl box operates on whole messages instead of a byte stream, the
writer must translate the input byte stream to a series of sealed messages.

We take the easy way out by translating a single call to Write to a single
sealed message. A single sealed message looks like this on the wire:

2 bytes - unsigned network byte order: total length (nonce + sealed message)
24 bytes: nonce (obtained from crypto.rand.Reader)
(total length - 24) bytes: sealed message

2 bytes are enough for total length as the problem states that max message
size is 32KB. Adding the nonce and overhead does not take it over 64KB.

We can obtain the length of the sealed message from the plaintext message
by adding box.Overhead.
*/
type secureWriter struct {
	w         io.Writer
	sharedKey [keySize]byte
	buffer    []byte
}

func (sw *secureWriter) Write(p []byte) (n int, err error) {
	// TODO is there anything we can do to have a sensible
	// bytes-written to return if we error out somewhere in the
	// middle?
	var totalLength uint16
	var nonce [nonceSize]byte

	// we need to ensure the total length still fits in uint16.
	// alternatively, it can be split into multiple messages but
	// that seems like too much work
	if len(p) > maxMessageSize {
		return 0, errors.New("maximum message size exceeded")
	}

	totalLength = uint16(nonceSize + box.Overhead + len(p))
	err = binary.Write(sw.w, binary.BigEndian, totalLength)
	if err != nil {
		return 0, err
	}

	_, err = rand.Read(nonce[:])
	if err != nil {
		// TODO error might not make sense in Write context
		return 0, err
	}
	_, err = sw.w.Write(nonce[:])
	if err != nil {
		return 0, err
	}

	// since SealAfterPrecomputation appends, we force it to
	// start at the beginning, reusing the buffer
	sw.buffer = sw.buffer[0:0]
	sw.buffer = box.SealAfterPrecomputation(
		sw.buffer, p, &nonce, &sw.sharedKey)
	_, err = sw.w.Write(sw.buffer)
	if err != nil {
		return 0, err
	}
	return len(p), err
}

// NewSecureWriter instantiates a new SecureWriter
func NewSecureWriter(w io.Writer, priv, pub *[keySize]byte) io.Writer {
	var sw secureWriter

	sw.w = w
	// pre-size the buffer to the max sealed message size
	sw.buffer = make([]byte, 0, box.Overhead+maxMessageSize)
	box.Precompute(&sw.sharedKey, pub, priv)
	return &sw
}

// secureConn wraps a net.Conn with secureReader and secureWriter. It
// implements io.ReadWriteCloser.
type secureConn struct {
	sr io.Reader
	sw io.Writer
	c  net.Conn
}

func (sc *secureConn) Read(p []byte) (n int, err error) {
	return sc.sr.Read(p)
}

func (sc *secureConn) Write(p []byte) (n int, err error) {
	return sc.sw.Write(p)
}

func (sc *secureConn) Close() error {
	return sc.c.Close()
}

// exchangeKeys performs a (unauthenticated) key exchange, writing our 32 byte
// public key and reading the peer's 32 byte public key.
func exchangeKeys(c net.Conn, pub *[keySize]byte) (peerPub *[keySize]byte, err error) {
	var ec = make(chan error)
	// do key exchange
	// TODO not sure writing our pubkey needs to be async
	// will writing 32 bytes ever block??
	go (func(ec chan error) {
		_, err := c.Write(pub[:])
		ec <- err
	})(ec)

	peerPub = new([keySize]byte)
	_, err = io.ReadFull(c, peerPub[:])
	if err != nil {
		_ = <-ec
		return nil, err
	}

	err = <-ec
	if err != nil {
		return nil, err
	}

	return peerPub, nil
}

// secure wraps the given net.Conn in a secureConn. Reads and writes will use
// the given private key and peer public key.
func secure(c net.Conn, priv, peerPub *[keySize]byte) io.ReadWriteCloser {
	return &secureConn{
		NewSecureReader(c, priv, peerPub),
		NewSecureWriter(c, priv, peerPub), c}
}

// Dial generates a private/public key pair,
// connects to the server, perform the handshake
// and return a reader/writer.
func Dial(addr string) (io.ReadWriteCloser, error) {
	pub, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	peerPubKey, err := exchangeKeys(conn, pub)
	if err != nil {
		conn.Close()
		return nil, err
	}

	sc := secure(conn, priv, peerPubKey)
	return sc, nil
}

// Serve starts a secure echo server on the given listener.
func Serve(l net.Listener) error {
	pub, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		// just echo
		go func(c net.Conn) {
			peerPubKey, err := exchangeKeys(c, pub)
			if err != nil {
				log.Fatal(err)
			}
			sc := secure(c, priv, peerPubKey)
			io.Copy(sc, sc)
			sc.Close()
		}(conn)
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
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", buf[:n])
}
