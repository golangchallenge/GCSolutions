// This file contains handling for data packets sent/read using a Reader/Writer.
// Data packets are sent in the following format:
// [ 0:24]  nonce (bytes)
// [24:32]  length of data (int64)
// [32:  ]  data (bytes)
package main

import (
	"encoding/binary"
	"io"
)

// dataPacket represents the data that can be setnt
type dataPacket struct {
	nonce [nonceSize]byte
	data  []byte
}

// Read will read a data packet from the given reader. It will try to use the readBuffer for data
// unless the data packet size is too large for the buffer.
func (p *dataPacket) Read(r io.Reader, readBuffer []byte) error {
	if _, err := io.ReadFull(r, p.nonce[:]); err != nil {
		return err
	}

	var dataLen int64
	if err := binary.Read(r, binary.LittleEndian, &dataLen); err != nil {
		return err
	}

	// Try to use the readBuffer where possible to avoid extra memory allocation.
	if int64(len(readBuffer)) >= dataLen {
		p.data = readBuffer[:dataLen]
	} else {
		p.data = make([]byte, dataLen)
	}
	_, err := io.ReadFull(r, p.data)
	return err
}

// Write writes out the data packet to the given writer.
func (p *dataPacket) Write(w io.Writer) (int, error) {
	if _, err := w.Write(p.nonce[:]); err != nil {
		return 0, err
	}
	if err := binary.Write(w, binary.LittleEndian, int64(len(p.data))); err != nil {
		return 0, err
	}
	return w.Write(p.data)
}
