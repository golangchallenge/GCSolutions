package drum_test

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/rogpeppe/misc/drum"
)

var decodeTests = []struct {
	name         string
	data         string
	expectOutput string
	expectError  string
}{{
	name: "pattern_1.splice",
	data: `
00000000  53 50 4c 49 43 45 00 00  00 00 00 00 00 c5 30 2e  |SPLICE........0.|
00000010  38 30 38 2d 61 6c 70 68  61 00 00 00 00 00 00 00  |808-alpha.......|
00000020  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |................|
00000030  f0 42 00 00 00 00 04 6b  69 63 6b 01 00 00 00 01  |.B.....kick.....|
00000040  00 00 00 01 00 00 00 01  00 00 00 01 00 00 00 05  |................|
00000050  73 6e 61 72 65 00 00 00  00 01 00 00 00 00 00 00  |snare...........|
00000060  00 01 00 00 00 02 00 00  00 04 63 6c 61 70 00 00  |..........clap..|
00000070  00 00 01 00 01 00 00 00  00 00 00 00 00 00 03 00  |................|
00000080  00 00 07 68 68 2d 6f 70  65 6e 00 00 01 00 00 00  |...hh-open......|
00000090  01 00 01 00 01 00 00 00  01 00 04 00 00 00 08 68  |...............h|
000000a0  68 2d 63 6c 6f 73 65 01  00 00 00 01 00 00 00 00  |h-close.........|
000000b0  00 00 00 01 00 00 01 05  00 00 00 07 63 6f 77 62  |............cowb|
000000c0  65 6c 6c 00 00 00 00 00  00 00 00 00 00 01 00 00  |ell.............|
000000d0  00 00 00                                          |...|
`,
	expectOutput: `Saved with HW Version: 0.808-alpha
Tempo: 120
(0) kick	|x---|x---|x---|x---|
(1) snare	|----|x---|----|x---|
(2) clap	|----|x-x-|----|----|
(3) hh-open	|--x-|--x-|x-x-|--x-|
(4) hh-close	|x---|x---|----|x--x|
(5) cowbell	|----|----|--x-|----|
`,
}, {
	name: "pattern_2.splice",
	data: `
00000000  53 50 4c 49 43 45 00 00  00 00 00 00 00 8f 30 2e  |SPLICE........0.|
00000010  38 30 38 2d 61 6c 70 68  61 00 00 00 00 00 00 00  |808-alpha.......|
00000020  00 00 00 00 00 00 00 00  00 00 00 00 00 00 cd cc  |................|
00000030  c4 42 00 00 00 00 04 6b  69 63 6b 01 00 00 00 00  |.B.....kick.....|
00000040  00 00 00 01 00 00 00 00  00 00 00 01 00 00 00 05  |................|
00000050  73 6e 61 72 65 00 00 00  00 01 00 00 00 00 00 00  |snare...........|
00000060  00 01 00 00 00 03 00 00  00 07 68 68 2d 6f 70 65  |..........hh-ope|
00000070  6e 00 00 01 00 00 00 01  00 01 00 01 00 00 00 01  |n...............|
00000080  00 05 00 00 00 07 63 6f  77 62 65 6c 6c 00 00 00  |......cowbell...|
00000090  00 00 00 00 00 01 00 00  00 00 00 00 00           |.............|
`,
	expectOutput: `Saved with HW Version: 0.808-alpha
Tempo: 98.4
(0) kick	|x---|----|x---|----|
(1) snare	|----|x---|----|x---|
(3) hh-open	|--x-|--x-|x-x-|--x-|
(5) cowbell	|----|----|x---|----|
`,
}, {
	name: "pattern_3.splice",
	data: `
00000000  53 50 4c 49 43 45 00 00  00 00 00 00 00 c5 30 2e  |SPLICE........0.|
00000010  38 30 38 2d 61 6c 70 68  61 00 00 00 00 00 00 00  |808-alpha.......|
00000020  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |................|
00000030  ec 42 28 00 00 00 04 6b  69 63 6b 01 00 00 00 00  |.B(....kick.....|
00000040  00 00 00 01 00 00 00 00  00 00 00 01 00 00 00 04  |................|
00000050  63 6c 61 70 00 00 00 00  01 00 00 00 00 00 00 00  |clap............|
00000060  01 00 00 00 03 00 00 00  07 68 68 2d 6f 70 65 6e  |.........hh-open|
00000070  00 00 01 00 00 00 01 00  01 00 01 00 00 00 01 00  |................|
00000080  05 00 00 00 07 6c 6f 77  2d 74 6f 6d 00 00 00 00  |.....low-tom....|
00000090  00 00 00 01 00 00 00 00  00 00 00 00 0c 00 00 00  |................|
000000a0  07 6d 69 64 2d 74 6f 6d  00 00 00 00 00 00 00 00  |.mid-tom........|
000000b0  01 00 00 00 00 00 00 00  09 00 00 00 06 68 69 2d  |.............hi-|
000000c0  74 6f 6d 00 00 00 00 00  00 00 00 00 01 00 00 00  |tom.............|
000000d0  00 00 00                                          |...|
`,
	expectOutput: `Saved with HW Version: 0.808-alpha
Tempo: 118
(40) kick	|x---|----|x---|----|
(1) clap	|----|x---|----|x---|
(3) hh-open	|--x-|--x-|x-x-|--x-|
(5) low-tom	|----|---x|----|----|
(12) mid-tom	|----|----|x---|----|
(9) hi-tom	|----|----|-x--|----|
`,
}, {
	name: "pattern_4.splice",
	data: `
00000000  53 50 4c 49 43 45 00 00  00 00 00 00 00 93 30 2e  |SPLICE........0.|
00000010  39 30 39 00 00 00 00 00  00 00 00 00 00 00 00 00  |909.............|
00000020  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |................|
00000030  70 43 00 00 00 00 07 53  75 62 4b 69 63 6b 00 00  |pC.....SubKick..|
00000040  00 00 00 00 00 00 00 00  00 00 00 00 00 00 01 00  |................|
00000050  00 00 04 4b 69 63 6b 01  00 00 00 00 00 00 00 01  |...Kick.........|
00000060  00 00 00 00 00 00 00 63  00 00 00 07 4d 61 72 61  |.......c....Mara|
00000070  63 61 73 01 00 01 00 01  00 01 00 01 00 01 00 01  |cas.............|
00000080  00 01 00 ff 00 00 00 09  4c 6f 77 20 43 6f 6e 67  |........Low Cong|
00000090  61 00 00 00 00 01 00 00  00 00 00 00 00 01 00 00  |a...............|
000000a0  00                                                |.|
`,
	expectOutput: `Saved with HW Version: 0.909
Tempo: 240
(0) SubKick	|----|----|----|----|
(1) Kick	|x---|----|x---|----|
(99) Maracas	|x-x-|x-x-|x-x-|x-x-|
(255) Low Conga	|----|x---|----|x---|
`,
}, {
	name: "pattern_5.splice",
	data: `
00000000  53 50 4c 49 43 45 00 00  00 00 00 00 00 57 30 2e  |SPLICE.......W0.|
00000010  37 30 38 2d 61 6c 70 68  61 00 00 00 00 00 00 00  |708-alpha.......|
00000020  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 c0  |................|
00000030  79 44 01 00 00 00 04 4b  69 63 6b 01 00 00 00 00  |yD.....Kick.....|
00000040  00 00 00 01 00 00 00 00  00 00 00 02 00 00 00 05  |................|
00000050  48 69 48 61 74 01 00 01  00 01 00 01 00 01 00 01  |HiHat...........|
00000060  00 01 00 01 00 53 50 4c  49 43 45 00 00 00 05 48  |.....SPLICE....H|
00000070  69 48 61 74 01 00 01 00  01 00 01 00 01 00 01 00  |iHat............|
00000080  01 00 01 00                                       |....|
`,
	expectOutput: `Saved with HW Version: 0.708-alpha
Tempo: 999
(1) Kick	|x---|----|x---|----|
(2) HiHat	|x-x-|x-x-|x-x-|x-x-|
`,
}, {
	name:        "empty file",
	data:        ``,
	expectError: `cannot read header: EOF`,
}, {
	name: "short header",
	data: `
000000 53 50 |SP|
`,
	expectError: `cannot read header: unexpected EOF`,
}, {
	name: "unexpected signature",
	data: `
00000000  46 41 52 43 45 00 00 00  00 00 00 00 00 57 30 2e  |FARCE........W0.|
00000010  37 30 38 2d 61 6c 70 68  61 00 00 00 00 00 00 00  |708-alpha.......|
00000020  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 c0  |................|
00000030  79 44 01 00 00 00 04 4b  69 63 6b 01 00 00 00 00  |yD.....Kick.....|
00000040  00 00 00 01 00 00 00 00  00 00 00 02 00 00 00 05  |................|
00000050  48 69 48 61 74 01 00 01  00 01 00 01 00 01 00 01  |HiHat...........|
00000060  00 01 00 01 00 53 50 4c  49 43 45 00 00 00 05 48  |.....SPLICE....H|
00000070  69 48 61 74 01 00 01 00  01 00 01 00 01 00 01 00  |iHat............|
00000080  01 00 01 00                                       |....|
`,
	expectError: `unexpected header, got "FARCE", want "SPLICE"`,
}, {
	name: "no version",
	data: `
00000000  53 50 4c 49 43 45 00 00  00 00 00 00 00 57 00 00  |SPLICE.......W..|
00000010  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |................|
00000020  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 c0  |................|
00000030  79 44 01 00 00 00 04 4b  69 63 6b 01 00 00 00 00  |yD.....Kick.....|
00000040  00 00 00 01 00 00 00 00  00 00 00 02 00 00 00 05  |................|
00000050  48 69 48 61 74 01 00 01  00 01 00 01 00 01 00 01  |HiHat...........|
00000060  00 01 00 01 00 53 50 4c  49 43 45 00 00 00 05 48  |.....SPLICE....H|
00000070  69 48 61 74 01 00 01 00  01 00 01 00 01 00 01 00  |iHat............|
00000080  01 00 01 00                                       |....|
`,
	expectError: `no version found`,
}, {
	name: "truncated track header",
	data: `
00000000  53 50 4c 49 43 45 00 00  00 00 00 00 00 57 30 2e  |SPLICE.......W0.|
00000010  37 30 38 2d 61 6c 70 68  61 00 00 00 00 00 00 00  |708-alpha.......|
00000020  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 c0  |................|
00000030  79 44 01 00 00 00  |yD....|
`,
	expectError: `cannot read channel header: unexpected EOF`,
}, {
	name: "spurious data with non-0.708 version",
	data: `
00000000  53 50 4c 49 43 45 00 00  00 00 00 00 00 57 30 2e  |SPLICE.......W0.|
00000010  37 30 39 2d 61 6c 70 68  61 00 00 00 00 00 00 00  |709-alpha.......|
00000020  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 c0  |................|
00000030  79 44 01 00 00 00 04 4b  69 63 6b 01 00 00 00 00  |yD.....Kick.....|
00000040  00 00 00 01 00 00 00 00  00 00 00 02 00 00 00 05  |................|
00000050  48 69 48 61 74 01 00 01  00 01 00 01 00 01 00 01  |HiHat...........|
00000060  00 01 00 01 00 53 50 4c  49 43 45 00 00 00 05 48  |.....SPLICE....H|
00000070  69 48 61 74 01 00 01 00  01 00 01 00 01 00 01 00  |iHat............|
00000080  01 00 01 00                                       |....|
`,
	expectError: `cannot read channel name, size 67: unexpected EOF`,
}, {
	name: "truncated channel beats",
	data: `
00000000  53 50 4c 49 43 45 00 00  00 00 00 00 00 57 30 2e  |SPLICE.......W0.|
00000010  37 30 39 2d 61 6c 70 68  61 00 00 00 00 00 00 00  |709-alpha.......|
00000020  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 c0  |................|
00000030  79 44 01 00 00 00 04 4b  69 63 6b 01 00 00 00 00  |yD.....Kick.....|
00000040  00 00 00 01 00 00 00 00  00 00 00 02 00 00 00 05  |................|
00000050  48 69 48 61 74 01 00 01  00 01 00 01 00 01 00 01  |HiHat...........|
00000060  00 |.|
`,
	expectError: `cannot read channel beats: unexpected EOF`,
}, {
	name: "unexpected beat value",
	data: `
00000000  53 50 4c 49 43 45 00 00  00 00 00 00 00 57 30 2e  |SPLICE.......W0.|
00000010  37 30 38 2d 61 6c 70 68  61 00 00 00 00 00 00 00  |708-alpha.......|
00000020  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 c0  |................|
00000030  79 44 01 00 00 00 04 4b  69 63 6b 01 00 00 00 00  |yD.....Kick.....|
00000040  00 00 00 01 00 00 00 00  00 00 00 02 00 00 00 05  |................|
00000050  48 69 48 61 74 01 00 02  00 01 00 01 00 01 00 01  |HiHat...........|
00000060  00 01 00 01 00                            |....|
`,
	expectError: `unexpected beat value 2 in channel "HiHat"`,
}}

func TestDecode(t *testing.T) {
	for i, test := range decodeTests {
		t.Logf("\ntest %d: %s", i, test.name)
		data := undump(test.data)
		p, err := drum.Decode(bytes.NewReader(data))
		if test.expectError != "" {
			if err == nil {
				t.Fatalf("got no error; expected error matching %q", test.expectError)
			}
			ok, err1 := regexp.MatchString("^("+test.expectError+")$", err.Error())
			if err1 != nil {
				t.Fatalf("bad error pattern in test: %v", err1)
			}
			if !ok {
				t.Fatalf("got unexpected error %q; want %q", err.Error(), test.expectError)
			}
			if p != nil {
				t.Fatalf("non-nil return from error-returning Decode")
			}
			continue
		}
		if err != nil {
			t.Fatalf("unexpected error when decoding: %v", err)
		}
		output := p.String()
		if output != test.expectOutput {
			t.Fatalf("unexpected output\nGot\n%s\nWant\n%s\n", output, test.expectOutput)
		}
		// Test that it round-trips OK through MarshalBinary
		data, err = p.MarshalBinary()
		if err != nil {
			t.Fatalf("cannot marshal binary: %v", err)
		}
		p1, err := drum.Decode(bytes.NewReader(data))
		if err != nil {
			t.Fatalf("cannot decode round-tripped binary: %v", err)
		}
		output = p1.String()
		if output != test.expectOutput {
			t.Fatalf("round trip gave unexpected output\nGot\n%s\nWant\n%s\n", output, test.expectOutput)
		}
	}
}

func TestMarshalBinaryWithNoVersion(t *testing.T) {
	p := &drum.Pattern{
		Tempo: 120,
	}
	data, err := p.MarshalBinary()
	if err != nil {
		t.Fatalf("cannot marshal: %v", err)
	}
	p1, err := drum.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("error decoding: %v", err)
	}
	want := "1.0"
	if p1.Version != want {
		t.Fatalf("bad version, got %q; want %q", p1.Version, want)
	}
}

func TestMarshalBinaryWithTrackNameTooLong(t *testing.T) {
	longName := strings.Repeat("a", 300)
	p := &drum.Pattern{
		Tempo: 120,
		Tracks: []drum.Track{{
			Channel: 1,
			Name:    longName,
		},
		}}
	_, err := p.MarshalBinary()
	if err == nil {
		t.Fatalf("got no error; expected name-too-long error")
	}
	want := fmt.Sprintf("track 1 has name too long (%q)", longName)
	if err.Error() != want {
		t.Fatalf("unexpected error %q; want %q", err.Error(), want)
	}
}

func TestDecodeFile(t *testing.T) {
	f, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatalf("cannot make temp file: %v", err)
	}
	defer f.Close()
	defer os.Remove(f.Name())
	test := decodeTests[0]
	_, err = f.Write(undump(test.data))
	if err != nil {
		t.Fatalf("cannot make temp file: %v", err)
	}
	p, err := drum.DecodeFile(f.Name())
	if err != nil {
		t.Fatalf("cannot decode file: %v", err)
	}
	output := p.String()
	if output != test.expectOutput {
		t.Fatalf("unexpected output\nGot\n%s\nWant\n%s\n", output, test.expectOutput)
	}
}

func TestDecodeFileWithError(t *testing.T) {
	_, err := drum.DecodeFile("no-such-file")
	if err == nil {
		t.Fatalf("no error from Decode: expected no-such-file error")
	}
	if !os.IsNotExist(err) {
		t.Fatalf("unexpected error; got %q; expected not-found error", err)
	}
}

func undump(hexStr string) []byte {
	var data []byte
	for _, line := range strings.Split(hexStr, "\n") {
		if line == "" {
			continue
		}
		i0 := strings.Index(line, " ")
		i1 := strings.Index(line, "|")
		if i0 == -1 || i1 == -1 {
			panic(fmt.Errorf("bad dump line %q", line))
		}
		line = strings.Replace(line[i0:i1], " ", "", -1)
		binData, err := hex.DecodeString(line)
		if err != nil {
			panic(fmt.Errorf("bad dump line %q", line))
		}
		data = append(data, binData...)
	}
	return data
}

// We define this here so it won't be included in the example.
var cowbellExampleData = undump(`
00000000  53 50 4c 49 43 45 00 00  00 00 00 00 00 c5 30 2e  |SPLICE........0.|
00000010  38 30 38 2d 61 6c 70 68  61 00 00 00 00 00 00 00  |808-alpha.......|
00000020  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |................|
00000030  f0 42 00 00 00 00 04 6b  69 63 6b 01 00 00 00 01  |.B.....kick.....|
00000040  00 00 00 01 00 00 00 01  00 00 00 01 00 00 00 05  |................|
00000050  73 6e 61 72 65 00 00 00  00 01 00 00 00 00 00 00  |snare...........|
00000060  00 01 00 00 00 02 00 00  00 04 63 6c 61 70 00 00  |..........clap..|
00000070  00 00 01 00 01 00 00 00  00 00 00 00 00 00 03 00  |................|
00000080  00 00 07 68 68 2d 6f 70  65 6e 00 00 01 00 00 00  |...hh-open......|
00000090  01 00 01 00 01 00 00 00  01 00 04 00 00 00 08 68  |...............h|
000000a0  68 2d 63 6c 6f 73 65 01  00 00 00 01 00 00 00 00  |h-close.........|
000000b0  00 00 00 01 00 00 01 05  00 00 00 07 63 6f 77 62  |............cowb|
000000c0  65 6c 6c 00 00 00 00 00  00 00 00 00 00 01 00 00  |ell.............|
000000d0  00 00 00                                          |...|
`)
