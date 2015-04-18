package main

import (
	"bytes"
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

// NonceLength is the number of bytes for the nonce.
const NonceLength = 24

// KeyLength is the number of bytes for the public or private key.
const KeyLength = 32

var (
	// ErrMsgOutOfOrder is the error returned for a message with a lower or equal sequence id
	// than the last consumed one.
	ErrMsgOutOfOrder = errors.New("message out of order")
	// ErrDecrypt is returned when a sealed message can not be decrypted.
	ErrDecrypt = errors.New("failed to decrypt message")
)

// SecureReader reads and decrypts messages from an underlying reader using NaCl.
// The expected format is defined at the SecureWriter.
// A decrypted message is buffered internally until it is read completely.
// Methods are not thread safe.
type SecureReader struct {
	in        io.Reader       // underlying reader
	sharedKey [KeyLength]byte // shared key between peersPublicKey and privateKey
	msgBuffer *bytes.Reader   // message buffer
	lastRecv  uint64          // seqId of last received message
}

// NewSecureReader returns a Reader that reads from r and handles message decryption under the hood.
// The underlying implementation is a *NewSecureReader.
func NewSecureReader(r io.Reader, priv, pub *[KeyLength]byte) io.Reader {
	var sharedKey [KeyLength]byte
	box.Precompute(&sharedKey, pub, priv)
	return &SecureReader{r, sharedKey, bytes.NewReader([]byte{}), 0}
}

// Read reads and decrypts a single message. The plain content data is copied into p which
// may be less than len(p). The number of bytes read is returned.
func (r *SecureReader) Read(p []byte) (int, error) {
	if r.msgBuffer.Len() > 0 {
		return r.msgBuffer.Read(p)
	}
	var nonce [NonceLength]byte
	if err := r.readNonce(nonce[:]); err != nil {
		return 0, err
	}
	decMsg, err := r.readMessage(&nonce)
	if err != nil {
		return 0, err
	}
	r.msgBuffer = bytes.NewReader(decMsg)
	return r.Read(p)
}

// readNonce reads the 24 bytes of the nonce from the underlying reader, checks for valid sequence number
// within the nonce and copies all bytes into p.
// The sequence number must be increasing with every nonce to prevent replay attacks.
// When seq ID is not greater than the last one received an ErrMsgOutOfOrder is returned.
func (r *SecureReader) readNonce(p []byte) error {
	if _, err := io.ReadFull(r.in, p); err != nil {
		return fmt.Errorf("read raw nonce: %v", err)
	}
	var nonce Nonce
	if err := binary.Read(bytes.NewReader(p), binary.BigEndian, &nonce); err != nil {
		return fmt.Errorf("unmarshal nonce: %v", err)
	}

	if nonce.SeqID <= r.lastRecv {
		return ErrMsgOutOfOrder
	}
	r.lastRecv = nonce.SeqID
	return nil
}

// readMessage reads 2 bytes for the message size and n bytes for the encrypted message
// from underlying reader.
func (r SecureReader) readMessage(nonce *[NonceLength]byte) ([]byte, error) {
	// read message size
	var messageLength uint16
	if err := binary.Read(r.in, binary.BigEndian, &messageLength); err != nil {
		return nil, fmt.Errorf("parse message length: %v", err)
	}

	// read encrypted message
	encMsg := make([]byte, int(messageLength))
	if _, err := io.ReadFull(r.in, encMsg); err != nil {
		return nil, fmt.Errorf("read encrypted message: %v", err)
	}
	// decrypt message
	decMsg, ok := box.OpenAfterPrecomputation(nil, encMsg, nonce, &r.sharedKey)
	if !ok {
		return nil, ErrDecrypt
	}
	return decMsg, nil
}

// A Nonce is an arbitrary number used only once in a cryptographic communication.
// It contains 16 random bytes and a 8 byte sequence id to prevent forgery in replay attacks.
// The number should be increasing and verified on the peers side.
type Nonce struct {
	Noise [NonceLength - 8]byte // random bytes
	SeqID uint64                // sequence number
}

// NewNonce creates a new Nonce with given sequence number and cryptographically secure
// pseudorandom bytes.
func NewNonce(seqID uint64) (*Nonce, error) {
	n := Nonce{SeqID: seqID}
	if _, err := rand.Read(n.Noise[:]); err != nil {
		return nil, err
	}
	return &n, nil
}

// Marshal Nonce into byte representation. The length is 24 bytes therefore any give slice must not be
// smaller. If Marshal method fails to write all bytes an io.ErrShortWrite error is returned.
func (n Nonce) Marshal(p []byte) error {
	buf := bytes.NewBuffer(make([]byte, 0, NonceLength))
	if err := binary.Write(buf, binary.BigEndian, n); err != nil {
		return fmt.Errorf("marshal nonce: %v", err)
	}
	if l := copy(p, buf.Bytes()); l < NonceLength {
		return io.ErrShortWrite
	}
	return nil
}

// SecureWriter encrypts and writes messages to an underlying writer using NaCl.
// A sequential message id is encoded into the nonce that a reader should use to
// protect against replay attacks.
// The nonce and message length is written unencrypted before the sealed content.
// Methods are not thread safe.
type SecureWriter struct {
	out       io.Writer       // underlying writer
	sharedKey [KeyLength]byte // precomputed shared key
	lastSent  uint64          // sequence ID of last message sent
}

// NewSecureWriter instantiates a new SecureWriter.
func NewSecureWriter(w io.Writer, priv, pub *[KeyLength]byte) io.Writer {
	var sharedKey [KeyLength]byte
	box.Precompute(&sharedKey, pub, priv)
	return &SecureWriter{w, sharedKey, 0}
}

// nextNonce generates a new nonce with the next sequence number set.
func (w *SecureWriter) nextNonce(p []byte) error {
	w.lastSent++
	nonce, err := NewNonce(w.lastSent)
	if err != nil {
		return err
	}
	return nonce.Marshal(p)
}

// Write encrypts p into a single sealed message and writes it into the underlying writer.
// First the 24 bytes of the nonce are written, then 2 bytes for the message length
// and the whole sealed message.
// The method returns len(p) or an error.
func (w *SecureWriter) Write(p []byte) (int, error) {
	var nonce [NonceLength]byte
	if err := w.nextNonce(nonce[:]); err != nil {
		return 0, fmt.Errorf("next nonce: %v", err)
	}
	// write nonce
	if n, err := io.Copy(w.out, bytes.NewReader(nonce[:])); err != nil {
		return int(n), fmt.Errorf("write nonce: %v", err)
	}

	// encrypt message
	b := box.SealAfterPrecomputation(nil, p, &nonce, &w.sharedKey)
	// write message length
	if err := binary.Write(w.out, binary.BigEndian, uint16(len(b))); err != nil {
		return 0, fmt.Errorf("write message length: %v", err)
	}
	// write message
	if n, err := io.Copy(w.out, bytes.NewReader(b)); err != nil {
		return int(n), fmt.Errorf("write message: %v", err)
	}
	return len(p), nil
}

// Close closes the underlying Writer and returns its Close return value, if the Writer
// is also an io.Closer. Otherwise it returns nil.
func (w *SecureWriter) Close() error {
	if c, ok := w.out.(io.Closer); ok {
		return c.Close()
	}
	return nil
}

// SecureSession combines a SecureReader and a SecureWriter.
type SecureSession struct {
	io.Reader
	io.Writer
}

// NewSecureSession exchanges the public key with a connected peer and initializes a secure Session.
func NewSecureSession(c io.ReadWriter, priv, pub *[KeyLength]byte) (*SecureSession, error) {
	peersPublicKey, err := exchangePubKeys(c, pub)
	if err != nil {
		return nil, err
	}
	return &SecureSession{NewSecureReader(c, priv, peersPublicKey),
		NewSecureWriter(c, priv, peersPublicKey)}, nil
}

// Close closes the underlying Writer and returns its Close return value, if the Writer
// is also an io.Closer. Otherwise it returns nil.
func (s SecureSession) Close() error {
	if c, ok := s.Writer.(io.Closer); ok {
		if err := c.Close(); err != nil {
			return err
		}
	}
	return nil
}

// Exchange public key with a connected peer. First the pub key is sent then the
// peers key is received and returned by the method.
func exchangePubKeys(c io.ReadWriter, pub *[KeyLength]byte) (*[KeyLength]byte, error) {
	if err := binary.Write(c, binary.BigEndian, pub); err != nil {
		return nil, fmt.Errorf("send public key: %v", err)
	}
	var peersPublicKey [KeyLength]byte
	if err := binary.Read(c, binary.BigEndian, &peersPublicKey); err != nil {
		return nil, fmt.Errorf("recieve peers public key: %v", err)
	}
	return &peersPublicKey, nil
}

// Dial generates a private/public key pair,
// connects to the server, perform the handshake
// and return a reader/writer.
func Dial(addr string) (io.ReadWriteCloser, error) {
	myPub, myPriv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	return NewSecureSession(conn, myPriv, myPub)
}

// Serve starts a secure echo server on the given listener.
func Serve(l net.Listener) error {
	for {
		conn, err := l.Accept()
		if err != nil {
			return fmt.Errorf("accept connection: %v", err)
		}
		myPub, myPriv, err := box.GenerateKey(rand.Reader)
		if err != nil {
			return fmt.Errorf("generate keys: %v", err)
		}
		secCon, err := NewSecureSession(conn, myPriv, myPub)
		if err != nil {
			return err
		}
		go func(s io.ReadWriteCloser) {
			io.Copy(s, s)
			s.Close()
		}(secCon)
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
