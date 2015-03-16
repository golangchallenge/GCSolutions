package drum

import (
	"bytes"
	"testing"
)

func TestHeaderParsing(t *testing.T) {
	buffer := bytes.NewBuffer([]byte{'S', 'P', 'L', 'I', 'C', 'E',
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xEF,
		'0', '.', '8', '0', '8', '-', 'a', 'l', 'p', 'h', 'a', 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0xCD, 0xCC, 0xC4, 0x42, 0x00, 0x00, 0x00, 0x00})
	header, _ := parseHeader(buffer)

	if string(header.signature[:]) != "SPLICE" {
		t.Errorf("Header signature was not parsed correctly. Got: %s, Expected: SPLICE", string(header.signature[:]))
	}
	if header.contentLength != uint64(239) {
		t.Errorf("Header contentLength was not parsed correctly. Got: %d, Expected: 239", header.contentLength)
	}
	if string(header.version[:11]) != "0.808-alpha" {
		t.Errorf("Header version was not parsed correctly. Got: %s, Expected: 0.808-alpha", string(header.version[:11]))
	}
	if header.tempo != 98.4 {
		t.Errorf("Header tempo was not parsed correctly. Got: %v, Expected: 98.4", header.tempo)
	}
}

func TestHeaderParserErrorHandling(t *testing.T) {
	headerBytes := []byte{'S', 'P', 'L', 'I', 'C', 'E',
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xEF,
		'0', '.', '8', '0', '8', '-', 'a', 'l', 'p', 'h', 'a', 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0xCD, 0xCC, 0xC4, 0x42, 0x00, 0x00, 0x00, 0x00}

	testCases := []struct {
		name         string
		sliceStart   int
		sliceEnd     int
		errorMessage string
	}{{"Bad Signature", 1, 52, "error while parsing header: signature mismatch"},
		{"EOF while parsing a field", 0, 4, "error while parsing header signature: unexpected EOF"},
		{"EOF before beginning to parse a field", 1, 15, "error while parsing header version: EOF"}}

	for _, test := range testCases {
		t.Logf("Test case: %s\n", test.name)
		buffer := bytes.NewBuffer(headerBytes[test.sliceStart:test.sliceEnd])
		header, err := parseHeader(buffer)
		if header != nil {
			t.Fatalf("Expected header parsing to fail. Got:\t%s\n", header)
		}
		if err.Error() != test.errorMessage {
			t.Errorf("Received error:\n\t%v\n\tError should have been:\n\t%v", err.Error(), test.errorMessage)
		}
	}
}

func TestHeaderVersionString(t *testing.T) {
	header := Header{version: [32]byte{'0', '.', '9', '0', '9', '-', 'a', 'l', 'p', 'h', 'a'}}

	if header.versionString() != "0.909-alpha" {
		t.Errorf("Header versionString() is incorrect. Got: %v, Expected: 0.909-alpha", header.versionString())
	}
}

func TestHeaderContentLengthExcludesHeaderSize(t *testing.T) {
	header := Header{signature: [6]byte{'S', 'P', 'L', 'I', 'C', 'E'},
		contentLength: 100,
		version:       [32]byte{'0', '.', '9', '0', '9', '-', 'a', 'l', 'p', 'h', 'a'},
		tempo:         78.5}

	if header.contentSize() != 60 {
		t.Errorf("Header contentSize() is incorrect. Got: %v, Expected: 60", header.contentSize())
	}
}

func TestHeaderStringRepresentation(t *testing.T) {
	header := Header{signature: [6]byte{'S', 'P', 'L', 'I', 'C', 'E'},
		contentLength: 100,
		version:       [32]byte{'0', '.', '9', '0', '9', '-', 'a', 'l', 'p', 'h', 'a'},
		tempo:         78.5}

	expectedStringRepresentation := `Saved with HW Version: 0.909-alpha
Tempo: 78.5`
	if header.String() != expectedStringRepresentation {
		t.Errorf("Header string representation is incorrect. \n\tGot:\n\t%v\n\n\tExpected:\n\t%v\n", header.String(), expectedStringRepresentation)
	}
}

func TestTrackParsing(t *testing.T) {
	buffer := bytes.NewBuffer([]byte{0x63, 0x00, 0x00, 0x00,
		0x09, 'L', 'o', 'w', ' ', 'C', 'o', 'n', 'g', 'a',
		0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00})
	track, _ := parseTrack(buffer)

	if track.id != uint32(99) {
		t.Errorf("Track id was not parsed correctly. Got: %v, Expected: 99", track.id)
	}
	if track.name.String() != "Low Conga" {
		t.Errorf("Track name was not parsed correctly. Got: %v, Expected: Low Conga", track.name)
	}
	expectedSteps := [16]uint8{0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00}
	if track.steps != expectedSteps {
		t.Errorf("Track steps were not parsed correctly. \n\tGot:\n\t%v\n\nExpected:\n\t%v\n", track.steps, expectedSteps)
	}
}

func TestTrackParserErrorHandling(t *testing.T) {
	trackBytes := []byte{0x63, 0x00, 0x00, 0x00,
		0x09, 'L', 'o', 'w', ' ', 'C', 'o', 'n', 'g', 'a',
		0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00}

	testCases := []struct {
		name         string
		sliceStart   int
		sliceEnd     int
		errorMessage string
	}{{"EOF while parsing track id", 0, 3, "error while parsing track id: unexpected EOF"},
		{"EOF while parsing track steps", 0, 16, "error while parsing track steps: unexpected EOF"},
		{"EOF before beginning to parse a field", 0, 14, "error while parsing track steps: EOF"}}

	for _, test := range testCases {
		t.Logf("Test case: %s\n", test.name)
		buffer := bytes.NewBuffer(trackBytes[test.sliceStart:test.sliceEnd])
		track, err := parseTrack(buffer)
		if track != nil {
			t.Fatalf("Expected track parsing to fail. Got:\t%s\n", track)
		}
		if err.Error() != test.errorMessage {
			t.Errorf("Received error:\n\t%v\n\tError should have been:\n\t%v", err.Error(), test.errorMessage)
		}
	}
}

func TestTrackCollectionParsing(t *testing.T) {
	content := []byte{0xFF, 0x00, 0x00, 0x00,
		0x09, 'L', 'o', 'w', ' ', 'C', 'o', 'n', 'g', 'a',
		0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00,
		0x63, 0x00, 0x00, 0x00,
		0x07, 'M', 'a', 'r', 'a', 'c', 'a', 's',
		0x01, 0x00, 0x01, 0x00, 0x01, 0x00, 0x01, 0x00, 0x01, 0x00, 0x01, 0x00, 0x01, 0x00, 0x01, 0x00}
	buffer := bytes.NewBuffer(content)
	tracks, _ := parseTrackCollection(buffer, uint64(len(content)))

	if len(tracks) != 2 {
		t.Errorf("Incorrect number of tracks were parsed. Got: %v, Expected: 2", len(tracks))
	}

	testTracks := []struct {
		id    uint32
		name  string
		steps [16]uint8
	}{{uint32(255), "Low Conga", [16]uint8{0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00}},
		{uint32(99), "Maracas", [16]uint8{0x01, 0x00, 0x01, 0x00, 0x01, 0x00, 0x01, 0x00, 0x01, 0x00, 0x01, 0x00, 0x01, 0x00, 0x01, 0x00}}}

	for i, testTrack := range testTracks {
		if tracks[i].id != testTrack.id {
			t.Errorf("Track %v was parsed with incorrect id. Got: %v, Expected: %v", i, tracks[i].id, testTrack.id)
		}
		if tracks[i].name.String() != testTrack.name {
			t.Errorf("Track %v was parsed with incorrect name. Got: %v, Expected: %v", i, tracks[i].name.String(), testTrack.name)
		}
		if tracks[i].steps != testTrack.steps {
			t.Errorf("Track %v was parsed with incorrect steps. \n\tGot:\n\t%v\n\nExpected:\n\t%v\n", i, tracks[i].steps, testTrack.steps)
		}
	}
}

func TestTrackCollectionParserErrorHandling(t *testing.T) {
	content := []byte{0xFF, 0x00, 0x00, 0x00,
		0x09, 'L', 'o', 'w', ' ', 'C', 'o', 'n', 'g', 'a',
		0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00,
		0x63, 0x00, 0x00, 0x00,
		0x07, 'M', 'a', 'r', 'a', 'c', 'a', 's',
		0x01, 0x00, 0x01, 0x00, 0x01, 0x00, 0x01}

	buffer := bytes.NewBuffer(content)
	tracks, err := parseTrackCollection(buffer, uint64(len(content)))

	if len(tracks) != 1 {
		t.Errorf("Exactly one track should have been parsed correctly. Got: %d", len(tracks))
	}
	expectedErrorMessage := "error while parsing track collection: error while parsing track steps: unexpected EOF"
	if err.Error() != expectedErrorMessage {
		t.Errorf("Received error:\n\t%v\n\tError should have been:\n\t%v", err.Error(), expectedErrorMessage)
	}
}

func TestTrackSize(t *testing.T) {
	track := Track{id: 220,
		name:  &PascalString{length: 9, text: []byte{'L', 'o', 'w', ' ', 'C', 'o', 'n', 'g', 'a'}},
		steps: [16]uint8{0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00}}

	if track.size() != uint64(30) {
		t.Errorf("Track size() is incorrect. Got: %v, Expected: 30", track.size())
	}
}

func TestTrackStringRepresentation(t *testing.T) {
	track := Track{id: 220,
		name:  &PascalString{length: 9, text: []byte{'L', 'o', 'w', ' ', 'C', 'o', 'n', 'g', 'a'}},
		steps: [16]uint8{0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00}}

	expectedStringRepresentation := "(220) Low Conga\t|---x|----|---x|----|"
	if track.String() != expectedStringRepresentation {
		t.Errorf("Track string representation is incorrect. \n\tGot:\n\t%v\n\n\tExpected:\n\t%v", track.String(), expectedStringRepresentation)
	}
}

func TestPascalStringParsing(t *testing.T) {
	buffer := bytes.NewBuffer([]byte{11, 'T', 'e', 's', 't', ' ', 'S', 't', 'r', 'i', 'n', 'g'})
	pascalString, _ := parsePascalString(buffer)

	if pascalString.length != 11 {
		t.Errorf("PascalString length was not parsed correctly. Got: %v, Expected: 11", pascalString.length)
	}
	if string(pascalString.text) != "Test String" {
		t.Errorf("PascalString text was not parsed correctly. Got: %v, Expected: Test String", string(pascalString.text))
	}
}

func TestPascalStringParserErrorHandling(t *testing.T) {
	stringBytes := []byte{11, 'T', 'e', 's', 't', ' ', 'S', 't', 'r', 'i', 'n', 'g'}

	testCases := []struct {
		name         string
		sliceStart   int
		sliceEnd     int
		errorMessage string
	}{{"EOF while parsing pascal string length", 0, 0, "error while parsing pascal string length: EOF"},
		{"EOF while parsing pascal string text", 0, 4, "error while parsing pascal string text: unexpected EOF"},
		{"EOF before beginning to parse a field", 0, 1, "error while parsing pascal string text: EOF"}}

	for _, test := range testCases {
		t.Logf("Test case: %s\n", test.name)
		buffer := bytes.NewBuffer(stringBytes[test.sliceStart:test.sliceEnd])
		pstring, err := parsePascalString(buffer)
		if pstring != nil {
			t.Fatalf("Expected pascal string parsing to fail. Got:%s\n", pstring)
		}
		if err.Error() != test.errorMessage {
			t.Errorf("Received error:\n\t%v\n\tError should have been:\n\t%v", err.Error(), test.errorMessage)
		}
	}
}

func TestPascalStringSize(t *testing.T) {
	buffer := bytes.NewBuffer([]byte{11, 'T', 'e', 's', 't', ' ', 'S', 't', 'r', 'i', 'n', 'g'})
	pascalString, _ := parsePascalString(buffer)

	if pascalString.size() != 12 {
		t.Errorf("PascalString size() is incorrect. Got: %v, Expected: 12", pascalString.size())
	}
}

func TestPascalStringStringRepresentation(t *testing.T) {
	pascalString := PascalString{length: 11, text: []byte{'T', 'e', 's', 't', ' ', 'S', 't', 'r', 'i', 'n', 'g'}}

	if pascalString.String() != "Test String" {
		t.Errorf("PascalString string representation is incorrect. Got: %v, Expected: Test String", pascalString.String())
	}
}
