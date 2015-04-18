// This package uses Nacl/Box to create an implementation of io.Reader and io.Writer
// that allows to write and read encoded data seamlessly
// See http://nacl.cr.yp.to/

package main

import (
	"io"
	"bytes"
	"encoding/binary"
	"errors"
	"golang.org/x/crypto/nacl/box"	
	"crypto/rand"
	"log"
)

const NONCE_LENGTH = 24 // The lenght of the nonce that Nacl/box will use

// NewSecureReader instantiates a new SecureReader
func NewSecureReader(r io.Reader, priv, pub *[32]byte) io.Reader {
	return SecureReader{pub: pub, priv: priv, r: r, read: false}
}
// An io.Reader implementation that decodes data coming from r.Read
type SecureReader struct {
	r         io.Reader // The io.Reader of where the data will be read
	tmp       []byte    
	pub, priv *[32]byte 
	read      bool      // A flag to see wheter the data has been read from r
	buf 	bytes.Buffer 
}



// Secure reader instantiates io.Reader. It reads data from
// s.rdecodes it using binary.Read and uses box.Open to decrypt it
func (s SecureReader) Read(p []byte) (n int, err error) {

	// Check if this is the first time that we enter this method
	if !s.read {		
		// Get the first 2 bytes with the length of the encoded data
		lenght := make([]byte, 2)
		n, err := s.r.Read(lenght)
		if !(err == nil || err == io.EOF){
			return 0, err
		}
		if n < 2 {			
			return 0,errors.New("Error when reading data length")
		}		
		buf := bytes.NewBuffer(lenght)
		var size uint16
		binary.Read(buf, binary.LittleEndian, &size)

		// Read the following bytes corresponding to the binary encoded data	
		raw := make([]byte,size)
		n, err = s.r.Read(raw)		
		buf = bytes.NewBuffer(raw)		
		binary.Read(buf,binary.LittleEndian,raw)

		if !(err == nil || err == io.EOF) {
			return 0, err
		}

		// The nonce is the first NONCE_LENGTH bytes of the input text.
		// The encoding data are the following bytes until size
		if len(raw) < NONCE_LENGTH {
			return 0, errors.New("The data to be decoded is not valid")
		}

		var nonce [NONCE_LENGTH]byte
		copy(nonce[:], raw[:NONCE_LENGTH])
		raw = raw[NONCE_LENGTH:]
		var out []byte
		var ok bool
		s.tmp, ok = box.Open(out, raw, &nonce, s.pub, s.priv)

		if !ok {
			return 0, errors.New("Error when decoding text")
		}
		s.buf = *bytes.NewBuffer(s.tmp)		
		s.read = true
	}
	return s.buf.Read(p)
}



// NewSecureWriter instantiates a new SecureWriter
func NewSecureWriter(w io.Writer, priv, pub *[32]byte) io.Writer {
	sw := SecureWriter{pub: pub, priv: priv, w: w}
	return sw
}

// A io.Writer that encodes incoming data and writes it w
type SecureWriter struct {
	w         io.Writer // An io.Writer where the encoded data will be written to
	pub, priv *[32]byte
}

// Writer implements io.Writer, this function encodes a message  using
// crypto.nacl.box.Seal, then prepends a random generated nonce to it,
// and writes it to s.w encoded using binary.Write
func (s SecureWriter) Write(message []byte) (n int, err error) {
	if len(message) > MAX_BYTE_LENGTH {
		return 0, errors.New("The data to be encoded is to big")
	}

	nonce := GenerateNonce()

	// Use box.Seal to encode the data
	var out []byte
	encText := box.Seal(out, message, &nonce, s.pub, s.priv)

	// The nonce will be preceding the binary data of the message		
	binaryData := append(nonce[:], encText...)

	// Encode the nonce + encodedMsg using binary.Write
	buf := new(bytes.Buffer)	
	err = binary.Write(buf,binary.LittleEndian,binaryData)

	// Get the size of encoded data and prepend it in format [2]byte
	size := uint16(buf.Len())
  	dataSize := make([]byte,2) 
    binary.LittleEndian.PutUint16(dataSize, size)            

    data := make([]byte,size)
  	n,err = buf.Read(data)

  	data = append(dataSize,data...)

  	// Data will be:
  	// |length [2]byte| binary encoded data [length]byte
    n,err = s.w.Write(data)   
	return n, err
}

// A wrapper for SecureWriter and SecureReader, and its connection.Close()
type secureRWC struct {
	io.ReadWriteCloser
	Reader io.Reader
	Writer io.Writer
	Closer io.Closer
}

func (rwc secureRWC) Read(p []byte) (n int, err error) {
	return rwc.Reader.Read(p)
}

func (rwc secureRWC) Write(message []byte) (n int, err error) {
	return rwc.Writer.Write(message)
}

func (rwc secureRWC) Close() error {
	return rwc.Closer.Close()
}

// Generates a Public and a Private keys using crypto.nacl.box.GenerateKey
func GenerateKeys() (publicKey, privateKey *[32]byte, err error) {
	publicKey, privateKey, err = box.GenerateKey(rand.Reader)
	return publicKey, privateKey, err
}


// Generates a nonce with random data
// The nonce will be the first NONCE_LENGHT bytes on every message sent
func GenerateNonce() (nonce [NONCE_LENGTH]byte) {
	b := make([]byte, NONCE_LENGTH)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatal("error:", err)
	}
	copy(nonce[:], b[0:NONCE_LENGTH])
	return nonce
}
