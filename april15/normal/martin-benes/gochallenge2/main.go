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

const maxUint64 = 1<<64 - 1
const nonceLength = 24
const messageMaxLength = 32*1024 + nonceLength + box.Overhead

type secureReader struct {
	r              io.Reader
	sharedKey      [32]byte
	nonceGen       *nonceGenerator
	bufEnc, bufDec []byte
}

type secureWriter struct {
	w              io.Writer
	sharedKey      [32]byte
	nonceGen       *nonceGenerator
	bufEnc, bufDec []byte
}

type secureReadWriteCloser struct {
	io.Reader
	io.Writer
	conn io.ReadWriteCloser
}

// nonceGenerator generates successive nonces
// nonce consists of 16-byte randrom prefix and 8-byte increment part
type nonceGenerator struct {
	prefix [16]byte
	inc    uint64
}

func (sr *secureReader) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}

	// if the nonce generator is nil, no data has been transfered yet
	// and initial nonce should be waiting for us on the reader
	nonceIn := make([]byte, 24)
	if sr.nonceGen == nil {
		if _, err := sr.r.Read(nonceIn); err != nil {
			return 0, err
		}
		sr.nonceGen = newNonceGeneratorFromNonce(nonceIn)
	}

	n, err := sr.r.Read(sr.bufEnc)
	if err != nil {
		return n, err
	}

	enc := sr.bufEnc[:n]

	// current decoding nonce
	nonce := sr.nonceGen.Get()

	dec, success := box.OpenAfterPrecomputation(sr.bufDec[:0], enc, nonce, &sr.sharedKey)
	if !success {
		return 0, errors.New("Filed to decode encrypted message")
	}

	// if we retrieved last nonce of a series, we need to get a new nonce from
	// the end of message
	if sr.nonceGen.LastNonceInSeries() {
		// extract new nonce from the end of the message
		newNonce := dec[len(dec)-nonceLength:]
		// set our NonceGenerator to the same state as sender's
		sr.nonceGen = newNonceGeneratorFromNonce(newNonce)
		// shorten the output - caller doesn't want the nonce
		dec = dec[:len(dec)-nonceLength]
	} else {
		// advance the nonce counter
		sr.nonceGen.Next()
	}

	if len(dec) > len(p) {
		return 0, errors.New("Buffer is too small for retrieved message")
	}

	return copy(p, dec), nil
}

func (sw *secureWriter) Write(p []byte) (int, error) {
	// if the nonce generator is nil, it means this is the first write.
	// we need to create new generator and send unencrypted initial nonce
	// to the peer
	if sw.nonceGen == nil {
		// new nonce generator with random prefix
		sw.nonceGen = newNonceGenerator()
		// in order to have a brand new nonce sent in the first encrypted
		// message(this one), we set the increment to highest possible value
		sw.nonceGen.inc = maxUint64
		// send the nonce
		if _, err := sw.w.Write(sw.nonceGen.Get()[:]); err != nil {
			return 0, err
		}
	}

	// current encoding nonce
	nonce := sw.nonceGen.Get()

	// if the nonce increment is at its max limit, we send new nonce
	// attached to the end of encrypted message (new nonce is also encrypted)
	if sw.nonceGen.LastNonceInSeries() {
		sw.nonceGen = newNonceGenerator()
		newNonce := sw.nonceGen.Get()
		// construct the messge with appended nonce
		sw.bufDec = append(sw.bufDec[:0], p...)
		sw.bufDec = append(sw.bufDec, newNonce[:]...)
		p = sw.bufDec
	} else {
		// if no new nonce is to be distributed, we just increase the counter
		sw.nonceGen.Next()
	}

	enc := box.SealAfterPrecomputation(sw.bufEnc[:0], p, nonce, &sw.sharedKey)

	n, err := sw.w.Write(enc)
	if err != nil {
		return n, err
	}

	return len(p), nil
}

// NewReadWriteCloser creates new io.ReadWriteCloser
func NewReadWriteCloser(conn io.ReadWriteCloser, sr io.Reader, sw io.Writer) io.ReadWriteCloser {
	srwc := secureReadWriteCloser{conn: conn, Reader: sr, Writer: sw}
	return &srwc
}

func (srwc *secureReadWriteCloser) Close() error {
	return srwc.conn.Close()
}

// LastNonceInSeries returns true, if increment is the highest possible
// that means that new nonce is about to arrive
func (ng *nonceGenerator) LastNonceInSeries() bool {
	return ng.inc == maxUint64
}

func newNonceGenerator() *nonceGenerator {
	ng := nonceGenerator{}
	ng.randomizePrefix()
	return &ng
}

// NewNonceGeneratorFromNonce creates new generator with
// prefix and increment obtained from nonceIn
func newNonceGeneratorFromNonce(nonceIn []byte) *nonceGenerator {
	ng := nonceGenerator{}
	copy(ng.prefix[:], nonceIn[:])
	err := binary.Read(bytes.NewReader(nonceIn[16:24]), binary.BigEndian, &ng.inc)
	if err != nil {
		panic(err)
	}

	return &ng
}

// advance the generator to the next nonce
func (ng *nonceGenerator) Next() {
	if ng.LastNonceInSeries() {
		panic("Cannot increment nonce beyond uint64 limit")
	}
	ng.inc++
}

func (ng *nonceGenerator) Get() *[24]byte {
	var rtn [24]byte
	copy(rtn[:], ng.prefix[:])
	buf := bytes.NewBuffer(rtn[:16])
	if err := binary.Write(buf, binary.BigEndian, ng.inc); err != nil {
		panic(err)
	}

	return &rtn
}

// randomizePrefix randomizes the first 16 bytes of nonce
func (ng *nonceGenerator) randomizePrefix() {
	m, err := rand.Read(ng.prefix[:])
	if err != nil || m != len(ng.prefix) {
		panic("Failed to fill nonce with random data")
	}
}

// NewSecureReader instantiates a new SecureReader
func NewSecureReader(r io.Reader, priv, pub *[32]byte) io.Reader {
	sr := &secureReader{r: r}
	sr.bufEnc = make([]byte, messageMaxLength)
	sr.bufDec = make([]byte, messageMaxLength)
	box.Precompute(&sr.sharedKey, pub, priv)
	return sr
}

// NewSecureWriter instantiates a new SecureWriter
func NewSecureWriter(w io.Writer, priv, pub *[32]byte) io.Writer {
	sw := secureWriter{w: w}
	sw.bufEnc = make([]byte, messageMaxLength)
	sw.bufDec = make([]byte, messageMaxLength)
	box.Precompute(&sw.sharedKey, pub, priv)
	return &sw
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

	sr, sw, err := performHandshake(conn, priv, pub)
	if err != nil {
		return nil, err
	}
	srwc := NewReadWriteCloser(conn, sr, sw)

	return srwc, nil
}

// performHandshake sends the public key, receives peer's public key and
// creates appropriate SecureWriter and SecureReader
func performHandshake(c net.Conn, priv, pub *[32]byte) (io.Reader, io.Writer, error) {
	// first we send our public key
	if _, err := c.Write(pub[:]); err != nil {
		return nil, nil, errors.New(fmt.Sprint("Failed to send public key to peer: ", err))
	}

	// read the peers public key
	peerPub := [32]byte{}
	_, err := c.Read(peerPub[:])
	if err != nil {
		return nil, nil, errors.New(fmt.Sprint("Filed to obtain peer's public key: ", err))
	}

	sr := NewSecureReader(c, priv, &peerPub)
	sw := NewSecureWriter(c, priv, &peerPub)

	return sr, sw, nil
}

// Serve starts a secure echo server on the given listener.
func Serve(l net.Listener) error {
	// need to generate server keys
	pub, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}

		go func(c net.Conn) {
			// the handshake first
			sr, sw, err := performHandshake(c, priv, pub)
			if err != nil {
				log.Println("Failed to shake hands with peer: ", err)
				return
			}
			srwc := NewReadWriteCloser(c, sr, sw)
			io.Copy(srwc, srwc)
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
	defer conn.Close()

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
