package main

import (
	"fmt"
	"io"
)

// sendPublic delivers a public key to the provided io.Writer.
func sendPublic(w io.Writer, pub *[32]byte) error {
	if n, err := w.Write(pub[:]); n != 32 {
		return fmt.Errorf("error delivering public key, too short: len() == %v", n)
	} else if err != nil {
		return fmt.Errorf("error delivering public key: %v", err)
	}
	return nil
}

// receivePublic retrieves a public key from the provided io.Reader.
func receivePublic(r io.Reader) (*[32]byte, error) {
	var peersPub [32]byte
	if n, err := r.Read(peersPub[:]); n != 32 {
		return nil, fmt.Errorf("error receiving peers public key, too short: len() == %v", n)
	} else if err != nil {
		return nil, fmt.Errorf("error receiving peers public key: %v", err)
	}
	return &peersPub, nil
}
