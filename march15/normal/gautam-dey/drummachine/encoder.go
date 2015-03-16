package drum

var zeros = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

func zeroPad(b []byte, l int) {
	if len(b) >= l {
		return
	}
	b = append(b, zeros[:l-len(b)]...)
}
