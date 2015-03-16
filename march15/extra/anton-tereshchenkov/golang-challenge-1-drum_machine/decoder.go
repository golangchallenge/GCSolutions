package drum

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
)

// isNumber checks that kind is a valid number: float, int or uint of any size.
func isNumber(kind reflect.Kind) bool {
	isInt := kind >= reflect.Int8 && kind <= reflect.Int64
	isUint := kind >= reflect.Uint8 && kind <= reflect.Uint64
	isFloat := kind >= reflect.Float32 && kind <= reflect.Float64
	return isInt || isUint || isFloat
}

// A OffsetError is a description of a splice encode or decode error.
type OffsetError struct {
	msg    string
	Offset int
}

func (d OffsetError) Error() string {
	return fmt.Sprintf("error at byte %d: %v", d.Offset, d.msg)
}

// A Decoder reads and decodes binary data from an input stream.
type Decoder struct {
	r  io.Reader
	bo binary.ByteOrder
}

// NewDecoder returns a new Decoder that reads from r.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		r:  r,
		bo: binary.LittleEndian,
	}
}

// Decode decodes next s bytes of binary data into a value pointed by v.
// If s is zero, the size of v is determined depending on it's type.
func (d *Decoder) Decode(s int, v interface{}) error {
	// Decode data into v
	if n, err := d.decodeValue(s, reflect.Indirect(reflect.ValueOf(v))); err != nil {
		return &OffsetError{err.Error(), n}
	}
	return nil
}

// decodeValue decodes next s bytes into a v.
// The return value n indicates the number of bytes decoded.
// If s is zero, the size of v is determined depending on it's type.
func (d *Decoder) decodeValue(s int, v reflect.Value) (n int, err error) {
	switch kind := v.Kind(); {
	case kind == reflect.Bool:
		n, err = d.decodeBool(v)
	case isNumber(kind):
		n, err = d.decodeNumber(v)
	case kind == reflect.Ptr:
		n, err = d.decodePtr(s, v)
	case kind == reflect.Slice:
		n, err = d.decodeSlice(s, v)
	case kind == reflect.String:
		n, err = d.decodeString(s, v)
	case kind == reflect.Struct:
		n, err = d.decodeStruct(s, v)
	default:
		err = fmt.Errorf("unknown type '%s'", kind)
	}
	return n, err
}

// readSize reads the size of next chunk of data in bytes.
func (d *Decoder) readSize() (int, error) {
	var s byte
	if err := binary.Read(d.r, d.bo, &s); err != nil {
		return 0, err
	}
	return int(s), nil
}

// decodeBool decodes a bool into a v.
// The return value n indicates the number of bytes decoded.
func (d *Decoder) decodeBool(v reflect.Value) (n int, err error) {
	var b byte
	if err := binary.Read(d.r, d.bo, &b); err != nil {
		return 0, err
	}
	v.SetBool(b != 0)
	return binary.Size(b), nil
}

// decodeNumber decodes any number into a v.
// The return value n indicates the number of bytes decoded.
func (d *Decoder) decodeNumber(v reflect.Value) (n int, err error) {
	// Get address of a v
	pv := v.Addr()
	if err := binary.Read(d.r, d.bo, pv.Interface()); err != nil {
		return 0, err
	}
	return binary.Size(v.Interface()), nil
}

// decodePtr decodes value with size s behind a pointer v.
// The return value n indicates the number of bytes decoded.
func (d *Decoder) decodePtr(s int, v reflect.Value) (n int, err error) {
	t := v.Type()
	value := reflect.New(t.Elem())
	n, err = d.decodeValue(s, reflect.Indirect(value))
	if err != nil {
		return 0, err
	}
	v.Set(value)
	return n, nil
}

// decodeSlice decodes a slice of size s into a v.
// If s is zero, the size of a slice is read from a stream.
// The return value n indicates the number of bytes decoded.
func (d *Decoder) decodeSlice(s int, v reflect.Value) (n int, err error) {
	// Read the size if it is a zero
	if s == 0 {
		s, err = d.readSize()
		if err != nil {
			return 0, err
		}
		// We have successfully read one byte of size already
		n++
		if s == 0 {
			return n, nil
		}
	}

	// Get type of slice element
	elType := v.Type().Elem()
	// Create a new []elType slice
	slice := reflect.MakeSlice(reflect.SliceOf(elType), 0, 0)

	// Decode until we reach the needed size
	for n < s {
		// Create a new element of type elType
		el := reflect.Indirect(reflect.New(elType))
		i, err := d.decodeValue(0, el)
		if err != nil {
			return n, err
		}

		// Add up the number of bytes read
		n += i
		slice = reflect.Append(slice, el)
	}
	// Error if we read more that we had to
	if n > s {
		return n, fmt.Errorf("decoded %d bytes, want %d", n, s)
	}
	// Update input slice
	v.Set(slice)
	return n, nil
}

// decodeString decodes a string of size s into a v.
// If s is zero, the size of a string is read from a stream.
// The return value n indicates the number of bytes decoded.
func (d *Decoder) decodeString(s int, v reflect.Value) (n int, err error) {
	// Read the size if it is a zero
	if s == 0 {
		s, err = d.readSize()
		if err != nil {
			return 0, err
		}
		// We have successfully read one byte of size already
		n++
		if s == 0 {
			return n, nil
		}
	}

	// Decode data of size s
	data := make([]byte, s)
	if err := binary.Read(d.r, d.bo, data); err != nil {
		return n, err
	}
	n += s
	// Trim null bytes
	data = bytes.TrimRight(data, "\x00")
	// Set the string
	v.SetString(string(data))
	return n, nil
}

// decodeStruct decodes a struct of size s into a v.
// The fields are decoded in the same order they appear in a struct.
// If s is zero, the size of struct is determined by the size of its fields.
// ",end" tag option indicates that the size of that field equals to
// size of struct - number of bytes decoded. It has no effect if used with
// a dynamic size struct.
// The return value n indicates the number of bytes decoded.
func (d *Decoder) decodeStruct(s int, v reflect.Value) (n int, err error) {
	vt := v.Type()

	for i := 0; i < v.NumField(); i++ {
		// Get field value and type
		fv := v.Field(i)
		ft := vt.Field(i)

		// Skip if we can't set that field, e.g. unexported
		if !fv.CanSet() {
			continue
		}
		// Or if it is set to be skipped
		if ft.Tag.Get("splice") == "-" {
			continue
		}

		// Parse a tag
		tag, err := parseSpliceTag(ft.Tag.Get("splice"))
		if err != nil {
			return n, err
		}
		// Get the size of a field
		fs := tag.Size
		if tag.HasOption("end") && s != 0 {
			fs = s - n
		}

		// Decode the field
		fn, err := d.decodeValue(fs, fv)
		if err != nil {
			return n, err
		}
		n += fn

		// Continue decoding for dynamic-size struct
		if s == 0 {
			continue
		}
		// Fixed-size struct
		// Break the loop if read enough bytes
		if n == s {
			break
		}
		// Error if we read more that we had to
		if n > s {
			return n, fmt.Errorf("decoded %d bytes, want %d", n, s)
		}
	}
	return n, nil
}

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Create a decoder
	dec := NewDecoder(file)
	// First, read file header
	header := &spliceHeader{}
	if err := dec.Decode(0, header); err != nil {
		return nil, err
	}
	// Check if signature is correct
	if !header.SignatureValid() {
		return nil, errors.New("bad file format")
	}

	// Decode data into pattern object
	p := &Pattern{}
	if err := dec.Decode(int(header.DataSize), p); err != nil {
		return nil, err
	}
	return p, nil
}
