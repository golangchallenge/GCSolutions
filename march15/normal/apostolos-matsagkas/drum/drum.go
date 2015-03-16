package drum

import (
	"encoding/binary"
	"errors"
	"os"
	"path/filepath"
	"strconv"
)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	pattern := &Pattern{}
	var file *os.File
	file, err := validateAndOpenFile(path)
	if err != nil {
		return nil, err
	}
	err = getName(path, pattern)
	if err != nil {
		return nil, err
	}
	err = getFileSize(file, pattern)
	if err != nil {
		return nil, err
	}
	err = getHWVersion(file, pattern)
	if err != nil {
		return nil, err
	}
	err = getTempo(file, pattern)
	if err != nil {
		return nil, err
	}
	err = getTracks(file, pattern)
	if err != nil {
		return nil, err
	}
	return pattern, nil
}

// Get the name part from a file
func getName(path string, pattern *Pattern) error {
	fileWithExtension := filepath.Base(path)
	extension := filepath.Ext(path)
	name := fileWithExtension[:len(fileWithExtension)-len(extension)]
	if fileWithExtension == "." || fileWithExtension == "/" || name == "" {
		err := errors.New("Invalid file path")
		return err
	}
	pattern.name = name
	return nil
}

// Open a file and check whether it had a splice header
func validateAndOpenFile(path string) (*os.File, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	var header [6]byte
	err = binary.Read(f, binary.BigEndian, &header)
	if err != nil {
		return nil, err
	}
	if string(header[:6]) == "SPLICE" {
		return f, nil
	}
	err = errors.New("This isn't a valid file, does not have a SPLICE header")
	return nil, err
}

// Given a pattern and a file at the right point, get the size
func getFileSize(f *os.File, pattern *Pattern) error {
	var sizeInBytes int64
	err := binary.Read(f, binary.BigEndian, &sizeInBytes)
	if err != nil {
		return err
	}
	//Accounting for the bytes before this to get the total file size
	pattern.totalSize = sizeInBytes + 14
	return nil
}

// Given a pattern and a file at the right point, load the hardware version used
func getHWVersion(f *os.File, pattern *Pattern) error {
	version := ""
	var versionChar byte
	err := binary.Read(f, binary.BigEndian, &versionChar)
	for i := 0; i <= 30; err, i = binary.Read(f, binary.BigEndian, &versionChar), i+1 {
		if err != nil {
			return err
		}
		if int(versionChar) != 0 {
			version += string(versionChar)
		}
	}
	pattern.hwversion = version
	return nil
}

// Given a pattern and a file at the right point, load the tempo
func getTempo(f *os.File, pattern *Pattern) error {
	var tempo float32
	err := binary.Read(f, binary.LittleEndian, &tempo)
	if err != nil {
		return err
	}
	pattern.tempo = tempo
	return nil
}

// Given a pattern and a file at the right point, load the track information
func getTracks(f *os.File, pattern *Pattern) error {
	//The sizeInBytes is listed in the file
	//This doesn't account for the 14 bytes for the splice header and the size info itself
	//We consume 36 bytes after that for the additional data we read
	//This leaves us with these bytes to consume
	var numOfTracks int64
	//Start with 0 tracks and up the counter for each one encountered
	numOfTracks = 0
	tracks := make([][16]bool, 0)
	trackIds := make([]int32, 0)
	trackNames := make([]string, 0)
	bytesLeft := pattern.totalSize - 50
	//We'll be reading from the file until the end of the expected size we got earlier
	for bytesLeft > 0 {
		//First should be the track id
		var trackId int32
		err := binary.Read(f, binary.LittleEndian, &trackId)
		if err != nil {
			return err
		}
		bytesLeft = bytesLeft - 4
		numOfTracks++
		trackIds = append(trackIds, trackId)
		//Then come the length of the track name
		var trackNameLengthAsByte int8
		err = binary.Read(f, binary.LittleEndian, &trackNameLengthAsByte)
		if err != nil {
			return err
		}
		bytesLeft = bytesLeft - 1
		trackNameLength := int(trackNameLengthAsByte)
		trackName := ""
		var nextChar byte
		//And now we'll read characters for the track name, knowing the length we know how many to read
		for i := 0; i < trackNameLength; i++ {
			err = binary.Read(f, binary.LittleEndian, &nextChar)
			if err != nil {
				return err
			}
			bytesLeft = bytesLeft - 1
			trackName += string(nextChar)

		}
		trackNames = append(trackNames, trackName)
		//Now to get the track information, we'll save the track info as a boolean array to save space instead of using bytes
		var track [16]bool
		for i := 0; i < 16; i++ {
			err = binary.Read(f, binary.LittleEndian, &nextChar)
			if err != nil {
				return err
			}
			bytesLeft = bytesLeft - 1
			if int(nextChar) == 1 {
				track[i] = true
			} else if int(nextChar) == 0 {
				track[i] = false
			} else {
				err = errors.New("Invalid file path")
				return err
			}

		}
		tracks = append(tracks, track)
	}
	//Put all these on the pattern struct
	pattern.numOfTracks = numOfTracks
	pattern.trackNames = trackNames
	pattern.trackIds = trackIds
	pattern.tracks = tracks
	return nil
}

// This function is used to print out the pattern
func (pattern *Pattern) String() string {
	patternAsString := ""
	patternAsString += "Saved with HW Version: " + pattern.hwversion + "\n"
	patternAsString += "Tempo: " + strconv.FormatFloat(float64(pattern.tempo), 'f', -1, 32) + "\n"
	for i := 0; i < int(pattern.numOfTracks); i++ {
		patternAsString += "(" + strconv.Itoa(int(pattern.trackIds[i])) + ") " + pattern.trackNames[i] + "\t|"
		for j := 0; j < 16; j++ {
			if pattern.tracks[i][j] {
				patternAsString += "x"
			} else {
				patternAsString += "-"
			}
			if ((j + 1) % 4) == 0 {
				patternAsString += "|"
			}
		}
		patternAsString += "\n"
	}
	return patternAsString
}

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	name string
	//Total size for the file, including the SPLICE header and the 4 byte int that holds the size
	totalSize   int64
	hwversion   string
	tempo       float32
	numOfTracks int64
	trackIds    []int32
	trackNames  []string
	tracks      [][16]bool
}
