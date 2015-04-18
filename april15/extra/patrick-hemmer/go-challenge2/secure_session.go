package main

// This file provides all types and functions relevant to 2-way communication
// with NaCl.

import (
	"crypto/rand"
	"io"
	"net"

	"golang.org/x/crypto/nacl/box"
)

// KeyPair contains an NaCl public & private key pair.
type KeyPair struct {
	PublicKey  *[32]byte
	PrivateKey *[32]byte
}

// NewKeyPair generates a new NaCl public & private key pair.
func NewKeyPair() (*KeyPair, error) {
	kp := &KeyPair{}
	var err error
	kp.PublicKey, kp.PrivateKey, err = box.GenerateKey(rand.Reader)
	return kp, err
}

// SecureSession implements io.ReadWriteCloser and provides NaCl
// encryption & decryption.
type SecureSession struct {
	kp     *KeyPair
	rwc    io.ReadWriteCloser
	reader *SecureReader
	writer *SecureWriter
}

// NewSecureSession creates a new SecureSession around an io.ReadWriteCloser.
// NewSecureSession will generate a public & private key, and then send the
// private key to the peer. The peer is required to send its key as well.
func NewSecureSession(rwc io.ReadWriteCloser, kp *KeyPair) (*SecureSession, error) {
	ss := &SecureSession{
		kp:  kp,
		rwc: rwc,
	}

	// Need to perform a handshake. However we only write our key, and don't
	// wait for theirs.
	// This is so that we don't block, and give the peer's key time to come
	// across the wire.
	if _, err := rwc.Write(kp.PublicKey[:]); err != nil {
		return nil, err
	}

	return ss, nil
}

// handshake reads the peer's public key off the wire, and initializes the
// secure reader & writer.
func (ss *SecureSession) handshake() error {
	if ss.reader != nil {
		// handshake not necessary
		return nil
	}

	// Our key was already sent in NewSecureSession. We just need the peer key.
	var peerPublicKey [32]byte
	if _, err := io.ReadFull(ss.rwc, peerPublicKey[:]); err != nil {
		return err
	}
	ss.reader = NewSecureReader(ss.rwc, ss.kp.PrivateKey, &peerPublicKey)
	ss.writer = NewSecureWriter(ss.rwc, ss.kp.PrivateKey, &peerPublicKey)

	return nil
}

// Read reads data off the underlying reader and performs NaCl decryption.
func (ss *SecureSession) Read(p []byte) (int, error) {
	if err := ss.handshake(); err != nil {
		return 0, err
	}
	return ss.reader.Read(p)
}

// Write performs NaCl encryption and writes to the underlying writer.
func (ss *SecureSession) Write(p []byte) (int, error) {
	if err := ss.handshake(); err != nil {
		return 0, err
	}
	return ss.writer.Write(p)
}

// Close closes the underlying io.ReadWriteCloser.
func (ss *SecureSession) Close() error {
	return ss.rwc.Close()
}

// SecureListener implements the net.Listener interface and provides NaCl
// encryption & deryption on top of a net.Listener.
type SecureListener struct {
	net.Listener
	kp *KeyPair
}

// NewSecureListener creates a new SecureListener which wraps a net.Listener,
// and performs NaCl encryption & decryption.
func NewSecureListener(l net.Listener, kp *KeyPair) *SecureListener {
	return &SecureListener{
		Listener: l,
		kp:       kp,
	}
}

// Listen generates a private/public key pair, and returns a SecureListener.
func Listen(addr string) (*SecureListener, error) {
	kp, err := NewKeyPair()
	if err != nil {
		return nil, err
	}

	l, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	return NewSecureListener(l, kp), nil
}

// Accept waits for an incoming connection, and wraps it with NewSecureConn.
// This results in sending the local public key, and requires that the remote
// does the same.
func (sl *SecureListener) Accept() (net.Conn, error) {
	conn, err := sl.Listener.Accept()
	if err != nil {
		return nil, err
	}

	return NewSecureConn(conn, sl.kp)
}

// SecureConn implements the net.Conn interface and provide NaCl encryption &
// decryption on top of a net.Conn.
type SecureConn struct {
	net.Conn
	session *SecureSession
}

// NewSecureConn creates a SecureConn which wraps a net.Conn to provide NaCl
// encryption & decryption.
// NewSecureConn will generate a public & private key, and then send the public
// key to the peer. The peer is required to send its key as well.
func NewSecureConn(conn net.Conn, kp *KeyPair) (*SecureConn, error) {
	ss, err := NewSecureSession(conn, kp)
	if err != nil {
		return nil, err
	}

	return &SecureConn{
		Conn:    conn,
		session: ss,
	}, nil
}

// Dial generates a private/public key pair, connects to the server, and
// returns a new SecureConn which performs encryption & decryption.
// The SecureConn will send the local public key to the peer, and expects
// the peer to do the same.
func Dial(addr string) (*SecureConn, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	kp, err := NewKeyPair()
	if err != nil {
		return nil, err
	}

	return NewSecureConn(conn, kp)
}

// Read reads data off the underlying connection and performs NaCl decryption.
func (sc *SecureConn) Read(b []byte) (int, error) {
	return sc.session.Read(b)
}

// Write performs NaCl encryption and writes to the underlying connection.
func (sc *SecureConn) Write(b []byte) (int, error) {
	return sc.session.Write(b)
}
