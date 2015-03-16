package drum

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"os"
	"strconv"
)

func MoreCowbell(filename string) (*Pattern, error) {
	pat := &Pattern{Header{}, ""}
	file, err := os.Open(filename)
	defer file.Close()
	if err != nil {
		return nil, err
	}

	err = binary.Read(file, binary.LittleEndian, &pat.Header)
	if err != nil {
		fmt.Println("Failed to read binary: ", err)
	}

	stat, _ := file.Stat()
	size := stat.Size()
	bytes, buffer := make([]byte, size), bufio.NewReader(file)
	buffer.Read(bytes)
	start, bytes_length := 0, len(bytes)-50

	for k, v := range bytes {
		if start == k {
			id := v
			instrument_length := int(bytes[start+4])
			instrument, next_track := bytes[start+5:start+5+instrument_length], start+5+instrument_length+16
			track_bytes, quarter_note, track_string := bytes[start+5+instrument_length:start+5+instrument_length+16], 1, "|"

			for _, vv := range track_bytes {
				if int(vv) == 1 {
					track_string += "x"
				} else {
					track_string_length := len(track_string)
					if string(instrument) == "cowbell" && track_string_length >= 2 && (track_string[track_string_length-2:] == "--" || track_string[track_string_length-2:] == "|-" || track_string[track_string_length-2:] == "-|") {
						track_string += "x"
					} else {
						track_string += "-"
					}
				}

				if quarter_note == 4 {
					track_string += "|"
					quarter_note = 0
				}
				quarter_note += 1
			}

			if bytes_length > next_track && string(bytes[next_track:next_track+6]) != "SPLICE" {
				start = next_track
				pat.Tracks += fmt.Sprint("(" + strconv.Itoa(int(id)) + ") " + string(instrument) + "\t" + track_string + "\n")
			} else {
				pat.Tracks += fmt.Sprint("(" + strconv.Itoa(int(id)) + ") " + string(instrument) + "\t" + track_string + "\n")
			}
		}
	}
	return pat, nil
}
