package drum

import "errors"

// ErrNotSplice indicates that a decoded file is not in a valid SPLICE format.
var ErrNotSplice = errors.New("decode: not a SPLICE file")

var spliceMagic = [8]byte{'S', 'P', 'L', 'I', 'C', 'E', 0, 0}

// header holds the decoded header data from a SPLICE file.
type header struct {
	Magic   [8]byte
	_       [5]byte // padding
	Size    uint8
	Version [32]byte
	Tempo   float32
}
