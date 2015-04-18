package main

import (
	"bytes"
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

// MaxMessageSize is the maximum expected size of a message in bytes.
const MaxMessageSize = 32 * 1024

// messageHeader represents the 26 byte header prepended to all
// encrypted messages.
// It contains (in clear) the unique nonce and the length of the
// encrypted message.
type messageHeader struct {
	Nonce   [24]byte
	DataLen uint16
}

// A SecureReader box-decrypts all data read from the underlying io.Reader.
type SecureReader struct {
	r            io.Reader
	sharedKey    *[32]byte
	header       messageHeader // Reusable messageHeader
	box          []byte        // Scratch slice for reading encrypted messages
	openBox      []byte        // Scratch slice for decrypting messages
	decryptedBuf *bytes.Buffer // Buffer for unread decrypted bytes
}

// Read implements the io.Reader interface, allowing seamless
// decrypted reads.
func (sr *SecureReader) Read(p []byte) (int, error) {
	// check for any unread bytes from previous decryptions
	if sr.decryptedBuf.Len() > 0 {
		return sr.decryptedBuf.Read(p)
	}

	// otherwise, read next encrypted message from underlying io.Reader
	// decode message header first
	if err := binary.Read(sr.r, binary.LittleEndian, &sr.header); err != nil {
		return 0, errors.New("could not decode message header: " + err.Error())
	}

	// re-slice our scratch slice to expected encrypted message length
	sr.box = sr.box[:sr.header.DataLen]
	// read this next encrypted message entirely
	if _, err := io.ReadFull(sr.r, sr.box); err != nil {
		return 0, errors.New("could not read encrypted message: " + err.Error())
	}

	// if provided slice is big enough, use it directly
	// as decrypt destination, avoiding a copy
	decryptedLen := int(sr.header.DataLen) - box.Overhead
	if len(p) >= decryptedLen {
		_, ok := box.OpenAfterPrecomputation(p[:0], sr.box, &sr.header.Nonce, sr.sharedKey)
		if !ok {
			return 0, errors.New("could not decrypt message")
		}
		return decryptedLen, nil
	}

	// otherwise, decrypt into our other scratch slice
	decrypted, ok := box.OpenAfterPrecomputation(sr.openBox[:0], sr.box, &sr.header.Nonce, sr.sharedKey)
	if !ok {
		return 0, errors.New("could not decrypt message")
	}
	// fulfill the Read into p
	n := copy(p, decrypted)
	// keep any left-over decrypted bytes for future reads
	if n < len(decrypted) {
		sr.decryptedBuf.Write(decrypted[n:])
	}
	return n, nil
}

// NewSecureReader instantiates a new SecureReader.
func NewSecureReader(r io.Reader, priv, pub *[32]byte) io.Reader {
	var sharedKey [32]byte
	box.Precompute(&sharedKey, pub, priv)
	return &SecureReader{
		r:            r,
		sharedKey:    &sharedKey,
		box:          make([]byte, 0, MaxMessageSize+box.Overhead),
		openBox:      make([]byte, 0, MaxMessageSize),
		decryptedBuf: bytes.NewBuffer(make([]byte, 0, MaxMessageSize)),
	}
}

// A SecureWriter box-encrypts all data prior to writing it.
type SecureWriter struct {
	w         io.Writer
	sharedKey *[32]byte
	header    messageHeader // Reusable messageHeader
	box       []byte        // Scratch slice for writing encrypted messages

	corruptedStream bool // Sticky error flag
}

// Write implements the io.Writer interface, allowing seamless
// encrypted writes.
// Due to encryption overhead, the return value n can be greater than
// len(p), up to `box.Overhead` more.
//
// Note: Some errors such as a partial write will corrupt the
// data stream. Any subsequent calls to Write will return
// an error.
func (sw *SecureWriter) Write(p []byte) (n int, err error) {
	// check for any previous error that may have corrupted the write stream
	// in which case error out immediately
	if sw.corruptedStream {
		return 0, errors.New("data stream corrupted by previous write")
	}
	// create unique nonce with negligible risk of collision
	if _, err = rand.Read(sw.header.Nonce[:]); err != nil {
		return 0, errors.New("could not create nonce: " + err.Error())
	}
	// set encrypted data length in header
	sw.header.DataLen = uint16(len(p) + box.Overhead)

	// write message header: nonce + length of encrypted data
	if err = binary.Write(sw.w, binary.LittleEndian, sw.header); err != nil {
		sw.corruptedStream = true
		return 0, errors.New("could not write message header: " + err.Error())
	}
	// encrypt message into scratch slice
	sw.box = box.SealAfterPrecomputation(sw.box, p, &sw.header.Nonce, sw.sharedKey)
	// write encrypted data
	if n, err = sw.w.Write(sw.box); err != nil {
		sw.corruptedStream = true
	}
	return
}

// NewSecureWriter instantiates a new SecureWriter.
func NewSecureWriter(w io.Writer, priv, pub *[32]byte) io.Writer {
	var sharedKey [32]byte
	box.Precompute(&sharedKey, pub, priv)
	return &SecureWriter{
		w:         w,
		sharedKey: &sharedKey,
		box:       make([]byte, 0, MaxMessageSize+box.Overhead),
	}
}

// Dial generates a private/public key pair,
// connects to the server, performs the handshake
// and returns a reader/writer.
func Dial(addr string) (io.ReadWriteCloser, error) {
	// generate key pair
	pub, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	// connect to server
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	// perform handshake
	// client                   server
	//      client public key -->
	//      <-- server public key
	if _, err = conn.Write(pub[:]); err != nil {
		return nil, errors.New("could not write client public key: " + err.Error())
	}
	var serverPubKey [32]byte
	if _, err = io.ReadFull(conn, serverPubKey[:]); err != nil {
		return nil, errors.New("could not read server public key: " + err.Error())
	}

	// combine 3 required interfaces to implement io.ReadWriteCloser
	return struct {
		io.Reader
		io.Writer
		io.Closer
	}{
		NewSecureReader(conn, priv, &serverPubKey),
		NewSecureWriter(conn, priv, &serverPubKey),
		conn,
	}, nil
}

// Serve starts a secure echo server on the given listener.
func Serve(l net.Listener) error {
	for {
		// wait for a connection
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		// handle connection
		go func(c net.Conn) {
			defer c.Close()
			// generate a key pair per client
			pub, priv, err := box.GenerateKey(rand.Reader)
			if err != nil {
				return
			}

			// handshake
			if _, err = c.Write(pub[:]); err != nil {
				return
			}
			var clientPubKey [32]byte
			if _, err = io.ReadFull(c, clientPubKey[:]); err != nil {
				return
			}
			// echo all data back to client
			io.Copy(NewSecureWriter(c, priv, &clientPubKey), NewSecureReader(c, priv, &clientPubKey))
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
	if err != nil && err != io.EOF {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", buf[:n])
}
