package main

import (
	"challenge2/secureio"
	"crypto/rand"
	"fmt"
	"io"
	"net"

	"golang.org/x/crypto/nacl/box"
)

// Serve generates a private/public key pair and starts a secure echo server
// on the given listener.
func Serve(l net.Listener) error {
	pub, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return fmt.Errorf("error generating public and private key: %v", err)
	}

	errc := make(chan error)
	go func() {
		for {
			c, err := l.Accept()
			if err != nil {
				errc <- err
				return
			}
			go echo(c, priv, pub, errc)
		}
	}()
	return <-errc
}

// echo performs a public-key handshake with a provided network connection and
// uses a secure reader and writer to unencrypt and re-encrypt any message sent
// on the connection, echoing the message back.
func echo(c net.Conn, priv, pub *[32]byte, errc chan error) {
	peersPub, err := receivePublic(c)
	if err != nil {
		c.Write([]byte("error receiving public key"))
		return
	}
	if err := sendPublic(c, pub); err != nil {
		errc <- err
	}

	secureR := secureio.NewSecureReader(c, priv, peersPub)
	secureW := secureio.NewSecureWriter(c, priv, peersPub)
	io.Copy(secureW, secureR)
}
