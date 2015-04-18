package secure

import (
	"bufio"
	"bytes"
	cryptrand "crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"golang.org/x/crypto/nacl/box"
	"golang.org/x/crypto/nacl/secretbox"
	"io"
	"math/rand"
)

// Binary message format (three parts, no delimiter):
// message length not including nonce as 2 bytes
//     (guaranteed not over 32k, so fits into two bytes just fine but binary.PutUvarint will use an extra bit)
// nonce as 24 bytes
// encrypted message contents (guaranteed not over 32k)

const (
	messageLengthSize = binary.MaxVarintLen16                                 // Number of bytes for message length info
	nonceSize         = 24                                                    // Number of bytes for the nonce
	chunkHeaderSize   = messageLengthSize + nonceSize                         // The total header size
	maxMessageSize    = 32000                                                 // The maximum size the message will be
	maxChunkSize      = chunkHeaderSize + maxMessageSize + secretbox.Overhead // The maximum size for an entire chunk
)

var (
	// ErrVerifyFailed occurs when decryption fails
	ErrVerifyFailed = errors.New("Verification failure")
	// ErrTooLarge occurs when a write is too large
	ErrTooLarge = fmt.Errorf("Write too large, only %v bytes allowed", maxMessageSize)
)

func newSecureChunkScanner(r io.Reader) *bufio.Scanner {
	scanner := bufio.NewScanner(r)
	scanner.Split(scanSecureChunks)
	return scanner
}

// Callback for bufio.Scanner.Split
func scanSecureChunks(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if len(data) < chunkHeaderSize+1 {
		return 0, nil, nil
	}

	// Get message length and confirm we have enough information
	msgLen, n := binary.Uvarint(data[:messageLengthSize])
	if n < 0 {
		panic("secure.scanSecureChunks: Unable to read message length")
	}
	fullMsgLen := int(chunkHeaderSize + msgLen)
	if len(data) < fullMsgLen {
		return 0, nil, nil
	}

	// Return the entire message, header and all, as the token
	return fullMsgLen, data[:fullMsgLen], nil
}

// Reader that sits on top of another reader and decrypts upon read
type Reader struct {
	sharedKey          *[32]byte      // Precomputed shared key
	secureChunkScanner *bufio.Scanner // Buffered scanner to get message chunks
	decryptedBuf       *bytes.Buffer  // Buffer to store decrypted contents
	nonceBuf           *[24]byte      // Byte array used each Read to copy the nonce into for decryption
}

// NewReader creates a new secure reader from the given reader, private key, and peer public key
func NewReader(r io.Reader, priv, peerPub *[32]byte) *Reader {
	var shared [32]byte
	box.Precompute(&shared, peerPub, priv)
	return NewReaderPrecomputed(r, &shared)
}

// NewReaderPrecomputed is like NewReader but uses a precomputed shared key
func NewReaderPrecomputed(r io.Reader, shared *[32]byte) *Reader {
	// The decrypted buffer never needs to be more than two chunks in size. This is because
	// Read is guaranteed to fetch no more than one chunk at a time and is then drained
	// every call.
	return &Reader{
		sharedKey:          shared,
		secureChunkScanner: newSecureChunkScanner(r),
		decryptedBuf:       bytes.NewBuffer(make([]byte, 0, maxChunkSize+maxChunkSize)),
		nonceBuf:           &[24]byte{},
	}
}

// Read encrypted data up to len(p). If the decrypted
func (r *Reader) Read(p []byte) (n int, err error) {
	// To keep this fast, we are only going to read one chunk instead of several to fill up
	// the buffer and fill up p. This is akin to bufio.Reader. This means we put the onus
	// on the caller to keep calling Read. Effort is made here to do no allocations.
	if r.decryptedBuf.Len() < len(p) {

		// Get next chunk, but only accept it if the scan was ok or we reached expected EOF.
		// Per the Scan docs, the result will be false but the Err is nil on EOF. We expect
		// this to block on the internal Read call.
		if !r.secureChunkScanner.Scan() {
			if r.secureChunkScanner.Err() != nil {
				return 0, r.secureChunkScanner.Err()
			}
			err = io.EOF
		}

		// Decrypt the chunk and store it in the buffer. Per the Bytes docs, no allocation
		// is done to fetch this.
		bytes := r.secureChunkScanner.Bytes()

		// Not thread safe of course reusing a nonce buffer, but it's cheaper and the reader
		// should not be considered thread safe anyways.
		copy(r.nonceBuf[:], bytes[messageLengthSize:chunkHeaderSize])

		// We reuse the buffer's byte array after growing it as necessary. This is similar to
		// what Write would do, but saves us the allocation. Unlike Seal, we must use the result
		// part here due to how the Nacl code modifies the underlying slice.
		r.decryptedBuf.Grow(len(bytes) - chunkHeaderSize - secretbox.Overhead)
		res, success := secretbox.Open(r.decryptedBuf.Bytes()[r.decryptedBuf.Len():],
			bytes[chunkHeaderSize:], r.nonceBuf, r.sharedKey)
		if !success {
			return 0, ErrVerifyFailed
		}

		// Unlike Seal which writes properly to the slice passed as the first param, Open does not
		r.decryptedBuf.Write(res)
	}

	// We know the buffer is not empty unless err is already io.EOF, so we can ignore the error
	n, _ = r.decryptedBuf.Read(p)
	return n, err
}

// Writer that encrypts to the underlying writer on Write
type Writer struct {
	sharedKey    *[32]byte           // The precomputed shared key
	writer       io.Writer           // The underlying writer to write encrypted values to
	nonceCounter uint64              // The incremented counter per-Write used to compute nonce
	writeBuf     *[maxChunkSize]byte // A byte array we can overwrite every Write call
	nonceBuf     *[24]byte           // Byte array used each Write to copy the nonce into for encryption
}

// NewWriter creates a new secure writer on top of the underlying writer with given private key
// and peer public key.
func NewWriter(w io.Writer, priv, peerPub *[32]byte) *Writer {
	var shared [32]byte
	box.Precompute(&shared, peerPub, priv)
	return NewWriterPrecomputed(w, &shared)
}

// NewWriterPrecomputed is the same as NewWriter but uses a precomputed shared key
func NewWriterPrecomputed(w io.Writer, shared *[32]byte) *Writer {
	// Start with a random u32 bit number to start nonce at so we know we have
	// at least 32 more bits to work with when counting. Technically the nonce
	// doesn't need to be random or unpredictable, just different on each Write.
	return &Writer{
		sharedKey:    shared,
		writer:       w,
		nonceCounter: uint64(rand.Uint32()),
		writeBuf:     &[maxChunkSize]byte{},
		nonceBuf:     &[24]byte{},
	}
}

// Write an encrypted chunk. The given byte array/slice cannot be over 32k.
func (w *Writer) Write(p []byte) (n int, err error) {
	if len(p) > maxMessageSize {
		return 0, ErrTooLarge
	}

	// Fill the header up with 0's
	for i := 0; i < chunkHeaderSize; i++ {
		w.writeBuf[i] = 0
	}

	// Write the first part of the message: the length
	binary.PutUvarint(w.writeBuf[:messageLengthSize], uint64(len(p)+secretbox.Overhead))

	// Write the second part: the nonce. Also copy to nonce array for encryption.
	w.nonceCounter++
	binary.PutUvarint(w.writeBuf[messageLengthSize:chunkHeaderSize], w.nonceCounter)
	copy(w.nonceBuf[:], w.writeBuf[messageLengthSize:chunkHeaderSize])

	// The Write method makes me promise I am not altering p in anyway but I have no
	// similar guarantee from the nacl code but we assume based on a quick inspection
	// that it is not mutated.
	secretbox.Seal(w.writeBuf[chunkHeaderSize:chunkHeaderSize:maxChunkSize], p, w.nonceBuf, w.sharedKey)
	return w.writer.Write(w.writeBuf[:chunkHeaderSize+len(p)+secretbox.Overhead])
}

// An io.ReadWriteCloser implementation on top of secure reader and writer
type secureReadWriteCloser struct {
	closer io.Closer // The underlying closer to delegate to on Close
	reader *Reader   // Secured reader
	writer *Writer   // Secured writer
}

// NewReadWriteCloser creates an io.ReadWriteCloser that is secured. This will perform a
// handshake to do a plaintext key exchange. If this is the listening side, pass true in
// for server. This will ensure it waits on the client's key first.
func NewReadWriteCloser(rwc io.ReadWriteCloser, server bool) (sec io.ReadWriteCloser, err error) {
	// Shake hands. Servers receive peer public key first, clients send first.
	var peerPub, priv *[32]byte
	if server {
		peerPub, err = readHandshake(rwc)
	}
	if err == nil {
		// We don't care about our public key...it is sent across.
		_, priv, err = generateKeyAndWriteHandshake(rwc)
	}
	if err == nil && !server {
		peerPub, err = readHandshake(rwc)
	}

	// Now create the reader and writer
	if err == nil {
		sec = secureReadWriteCloser{
			closer: rwc,
			reader: NewReader(rwc, priv, peerPub),
			writer: NewWriter(rwc, priv, peerPub),
		}
	}
	return sec, err
}

func generateKeyAndWriteHandshake(w io.Writer) (pub, priv *[32]byte, err error) {
	// Generate key and write the pub side
	pub, priv, err = box.GenerateKey(cryptrand.Reader)
	if err == nil {
		_, err = w.Write(pub[:])
	}
	return
}

func readHandshake(r io.Reader) (peerPub *[32]byte, err error) {
	peerPub = &[32]byte{}
	_, err = r.Read(peerPub[:])
	return
}

func (s secureReadWriteCloser) Read(p []byte) (n int, err error) {
	n, err = s.reader.Read(p)
	return
}

func (s secureReadWriteCloser) Write(p []byte) (n int, err error) {
	n, err = s.writer.Write(p)
	return
}

func (s secureReadWriteCloser) Close() error {
	return s.closer.Close()
}
