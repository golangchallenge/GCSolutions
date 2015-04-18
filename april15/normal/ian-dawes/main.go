package main

/****
*
* This software is a solution to the 2nd Golang challenge, detailed at http://golang-challenge.com/go-challenge2/
*
* It is provided under the BSD-3-clause license as detailed below:

-----------------------------------------------------------------

    Copyright (c) 2015, Ian Dawes
    All rights reserved.

    Redistribution and use in source and binary forms, with or without modification, are permitted provided that the following conditions are met:

    1. Redistributions of source code must retain the above copyright notice, this list of conditions and the following disclaimer.

    2. Redistributions in binary form must reproduce the above copyright notice, this list of conditions and the following disclaimer in the documentation and/or other materials provided with the distribution.

    3. Neither the name of the copyright holder nor the names of its contributors may be used to endorse or promote products derived from this software without specific prior written permission.

    THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO,
    THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS
    BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE
    GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT
    LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

*
*/

import (
	crand "crypto/rand"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"os"

	"golang.org/x/crypto/nacl/box"
)

// As specified in the challenge details, it is safe to assume that no message will ever be more than 32KB.
// That being accepted, it's always better to check. :)
const MaxMessageBytes = 32 * 1024

var (
	// ErrMessageTooBig is returned when a an attempt is made to write a message that is bigger than the maximum size (32KB), or
	// a message length is read that indicates that an incoming message is too big.
	ErrMessageTooBig = fmt.Errorf("Max message size is %d bytes", MaxMessageBytes)

	// ErrTooManyMessages is returned when an attempt is made to write any message to a secure writer after 2^64 messages have been written.
	ErrTooManyMessages = fmt.Errorf("Maximum number of messages (%d) has been sent on this connection", 2^64)

	// ErrUndecryptable is returned when an incoming message cannot be decrypted.
	ErrUndecryptable = fmt.Errorf("Couldn't decrypt message")
)

// SecureReader wraps an io.Reader, adding NaCl public-key "box" decryption to every read.
type SecureReader struct {
	r io.Reader
	// we're precomputing a shared
	shared [32]byte
	// we're using reusable temporary buffers to hold the nonce and the encrypted message
	nonce [24]byte
	buf   [MaxMessageBytes + box.Overhead]byte
}

// Read satisfies the io.Reader interface. It expects to find the following sequence:
//    1) the fixed length nonce used during encryption of the message
//    2) the length of the original message
//    3) the encrytped message
func (r *SecureReader) Read(p []byte) (n int, err error) {
	// read the nonce
	_, err = io.ReadFull(r.r, r.nonce[:])
	if err != nil {
		return 0, err
	}

	// read the message length
	var mlen uint32
	err = binary.Read(r.r, binary.BigEndian, &mlen)
	if err != nil {
		return 0, err
	}
	if mlen > MaxMessageBytes {
		return 0, ErrMessageTooBig
	}

	// read the encrypted message
	em := r.buf[0 : mlen+box.Overhead]
	n, err = io.ReadFull(r.r, em)
	if err != nil {
		return 0, err
	}

	// decrypt the message
	_, ok := box.OpenAfterPrecomputation(p[0:0], em, &r.nonce, &r.shared)
	if !ok {
		return 0, ErrUndecryptable
	}
	return int(mlen), nil
}

// NewSecureReader instantiates a new SecureReader
func NewSecureReader(r io.Reader, priv, pub *[32]byte) io.Reader {
	sr := &SecureReader{r: r}
	box.Precompute(&sr.shared, pub, priv)
	return sr
}

// SecureWriter wraps an io.Writer, adding NaCl public-key "box" encryption to every write.
type SecureWriter struct {
	w      io.Writer
	shared [32]byte
	// We're using a 64 bit counter as a nonce, as it it will be guaranteed to be unique for the session.
	// Even if the session were to generate 1 message every millisecond, it would take 292 million years for the counter to wrap around,
	// and being forced to restart the session once every 292 million years seems like a reasonable trade off for simplicity.
	iNonce   uint64
	rawNonce uint64
	nonce    [24]byte
	buf      [8 + MaxMessageBytes + box.Overhead]byte
}

// Write satisfies the io.Writer interface. It encrypts the given message and writes the following sequence to the underlying writer:
//    1) the fixed length nonce used during encryption
//    2) the length of the original message
//    3) the encrypted message
func (w *SecureWriter) Write(p []byte) (n int, err error) {
	// make sure we're not trying to send an oversize message
	if len(p) > MaxMessageBytes {
		return 0, ErrMessageTooBig
	}
	// make sure that the nonce hasn't wrapped around
	if w.rawNonce == w.iNonce {
		return 0, ErrTooManyMessages
	}

	// write the nonce
	binary.PutUvarint(w.nonce[:], w.rawNonce)
	n, err = w.w.Write(w.nonce[:])
	if n != len(w.nonce) || err != nil {
		return 0, err
	}
	w.rawNonce++ // and increment it to ensure uniqueness

	// write the message length
	err = binary.Write(w.w, binary.BigEndian, uint32(len(p)))
	if err != nil {
		return 0, err
	}

	// encrypt the message
	buf := w.buf[0:0]
	em := box.SealAfterPrecomputation(buf, p, &w.nonce, &w.shared)

	// and write it
	n, err = w.w.Write(em)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

// Close satisfies io.Closer. It closes the underlying writer if it is closeable.
func (w *SecureWriter) Close() error {
	if c, ok := w.w.(io.Closer); ok {
		return c.Close()
	}
	return nil
}

// NewSecureWriter creates a new SecureWriter which will write encrypted messages to the given writer.
// The caller must supply it's own 32 byte private key, and the 32 byte public key of the entity that will be receiving the encrypted messages.
func NewSecureWriter(w io.Writer, myPriv, peerPub *[32]byte) io.Writer {
	iNonce := uint64(rand.Int63())
	sw := &SecureWriter{w: w, iNonce: iNonce, rawNonce: iNonce + 1}
	box.Precompute(&sw.shared, peerPub, myPriv)
	return sw
}

// SecureConn implements a bidirectional connection on which encrypted messages can be sent and received. It makes use of the SecureReader and SecureWriter above.
type SecureConn struct {
	SecureReader
	SecureWriter
}

// NewSecureConn instantiates a new secure connection given the caller's private key and the public key of the intended receiver.
func NewSecureConn(rw io.ReadWriter, myPriv, peerPub *[32]byte) io.ReadWriteCloser {
	sr := NewSecureReader(rw, myPriv, peerPub).(*SecureReader)
	sw := NewSecureWriter(rw, myPriv, peerPub).(*SecureWriter)
	return &SecureConn{SecureReader: *sr, SecureWriter: *sw}

}

// Dial generates a private/public key pair, connects to the server, performs the handshake and returns a closeable ReadWriter.
func Dial(addr string) (io.ReadWriteCloser, error) {
	myPub, myPriv, err := box.GenerateKey(crand.Reader)
	if err != nil {
		return nil, fmt.Errorf("Couldn't generate key pair, killing connection")
	}
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	_, err = conn.Write(myPub[:])
	if err != nil {
		return nil, fmt.Errorf("Couldn't send public key to peer, killing connection")
	}
	peerPub := [32]byte{}
	_, err = io.ReadFull(conn, peerPub[:])
	if err != nil {
		return nil, fmt.Errorf("Couldn't read peer's public key, killing connection")
	}
	return NewSecureConn(conn, myPriv, &peerPub), nil
}

// Serve starts a secure echo server on the given listener.
func Serve(l net.Listener) error {
	// It's a server... loop forever until OS kills the application.
	for {
		// Start listening for incoming connections.
		conn, err := l.Accept()
		if err != nil {
			return err
		}

		// spawn a responder for each new connection.
		go func(c net.Conn) {

			// generate a new key pair.
			myPub, myPriv, err := box.GenerateKey(crand.Reader)
			if err != nil {
				log.Println("Couldn't generate key pair, killing connection")
				c.Close()
				return
			}
			// Send our public key to the peer
			_, err = c.Write(myPub[:])
			if err != nil {
				log.Println("Couldn't send public key to peer, killing connection")
				c.Close()
				return
			}
			// Read peer's public key
			peerPub := [32]byte{}
			_, err = io.ReadFull(c, peerPub[:])
			if err != nil {
				log.Println("Couldn't read peer's public key, killing connection")
				c.Close()
				return
			}

			// Fire up a new secure reader/writer
			sconn := NewSecureConn(conn, myPriv, &peerPub)
			defer sconn.Close() // make sure the resources get cleaned up when the responder shuts down.
			buf := new([MaxMessageBytes]byte)
			for {
				// Every message read from the connection gets written back to the connection.
				n, err := sconn.Read(buf[:])
				if err != nil {
					return
				}
				_, err = sconn.Write(buf[:n])
				if err != nil {
					return
				}
			}
		}(conn)
	}
}

// This provides a small command line utility which will either create a secure echo server (if the -l flag is provided) on a given port, or
// will send messages to a secure echo server which is listening on the given port. It is only useful for testing, as it only communicates on "localhost"
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
