package drum

import (
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"os"
)

const (
	header_end_pos = 0xa
	magic_header   = "SPLICE\x00\x00\x00\x00"
)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (p *Pattern, err error) {
	p = &Pattern{}
	sp_file, err := os.Open(path)
	if err != nil {
		return
	}
	defer sp_file.Close()
	data, err := ioutil.ReadAll(sp_file)
	if err != nil {
		return p, err
	}
	//magic string
	if string(data[:header_end_pos]) != magic_header {
		fmt.Printf("wrong magic header string, actural:\"%s\", expect:\"%s\"\r\n",
			string(data[:header_end_pos]), magic_header)
		return p, WrongHeaderERR
	}
	data_len := int(binary.BigEndian.Uint32(data[header_end_pos : header_end_pos+4]))
	if data_len > len(data[len_end_pos:]) {
		fmt.Printf("wrong length, expect:%x, actural:%x\r\n",
			data_len, len(data[len_end_pos:]))
		return p, FileFormatERR
	}
	err = p.UnmarshalBinary(data[len(magic_header) : header_end_pos+data_len])
	return p, err
}
