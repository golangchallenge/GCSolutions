package drum

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"testing"
	"testing/quick"
)

// fixturePath yields the location of the named test fixture.
func fixturePath(name string) string { return path.Join("fixtures", name) }

// loadFixture loads the fixture file as a closeable stream for reading.
func loadFixture(name string) (io.ReadCloser, error) {
	return os.Open(fixturePath(name))
}

func TestDecodeFile(t *testing.T) {
	tData := []struct {
		path string
		want string
		err  error
	}{
		{"pattern_1.splice",
			`Saved with HW Version: 0.808-alpha
Tempo: 120
(0) kick	|x---|x---|x---|x---|
(1) snare	|----|x---|----|x---|
(2) clap	|----|x-x-|----|----|
(3) hh-open	|--x-|--x-|x-x-|--x-|
(4) hh-close	|x---|x---|----|x--x|
(5) cowbell	|----|----|--x-|----|
`,
			nil,
		},
		{"pattern_2.splice",
			`Saved with HW Version: 0.808-alpha
Tempo: 98.4
(0) kick	|x---|----|x---|----|
(1) snare	|----|x---|----|x---|
(3) hh-open	|--x-|--x-|x-x-|--x-|
(5) cowbell	|----|----|x---|----|
`,
			nil,
		},
		{"pattern_3.splice",
			`Saved with HW Version: 0.808-alpha
Tempo: 118
(40) kick	|x---|----|x---|----|
(1) clap	|----|x---|----|x---|
(3) hh-open	|--x-|--x-|x-x-|--x-|
(5) low-tom	|----|---x|----|----|
(12) mid-tom	|----|----|x---|----|
(9) hi-tom	|----|----|-x--|----|
`,
			nil,
		},
		{"pattern_4.splice",
			`Saved with HW Version: 0.909
Tempo: 240
(0) SubKick	|----|----|----|----|
(1) Kick	|x---|----|x---|----|
(99) Maracas	|x-x-|x-x-|x-x-|x-x-|
(255) Low Conga	|----|x---|----|x---|
`,
			nil,
		},
		{"pattern_5.splice",
			`Saved with HW Version: 0.708-alpha
Tempo: 999
(1) Kick	|x---|----|x---|----|
(2) HiHat	|x-x-|x-x-|x-x-|x-x-|
`,
			nil,
		},
	}

	for _, data := range tData {
		decoded, err := DecodeFile(fixturePath(data.path))
		if got, want := err, data.err; got != want {
			t.Fatalf("DecodeFile(%#v) = _, %#v; want _, %#v", data.path, got, want)
		}
		if got, want := fmt.Sprint(decoded), data.want; got != want {
			t.Fatalf("DecodeFile(%#v) =\n%s, _\nwant\n%s, _", data.path, got, want)
		}
	}
}

// bmDecode benchmarks the decoding time for a single .drum stream.
func bmDecode(name string, b *testing.B) {
	file, err := loadFixture(name)
	if err != nil {
		b.Fatal(err)
	}
	payload, err := ioutil.ReadAll(file)
	if err != nil {
		file.Close()
		b.Fatal(err)
	}
	file.Close()
	readers := make([]io.Reader, b.N)
	for i := 0; i < b.N; i++ {
		readers[i] = bytes.NewReader(payload)
	}
	// Report how many bytes each operation processes in terms of whitebox data.
	b.SetBytes(int64(len(payload)))
	// Report how many allocations and memory copies, etc.
	b.ReportAllocs()
	// Reset the benchmark to a clean-slate.  Note: No defers or anything beyond
	// here.  We attempt to measure strictly the system-under-test.
	b.ResetTimer()
	for _, r := range readers { // len(readers) == b.N
		Decode(r)
	}
}

func BenchmarkDecodePattern1(b *testing.B) {
	bmDecode("pattern_1.splice", b)
}

func BenchmarkDecodePattern2(b *testing.B) {
	bmDecode("pattern_2.splice", b)
}

func BenchmarkDecodePattern3(b *testing.B) {
	bmDecode("pattern_3.splice", b)
}

func BenchmarkDecodePattern4(b *testing.B) {
	bmDecode("pattern_4.splice", b)
}

func BenchmarkDecodePattern5(b *testing.B) {
	bmDecode("pattern_5.splice", b)
}

// The fuzz tests that follow are not really representative of real-world data
// cases but just OK for facial validations of API completeness.

// TestDecodeFuzz throws Haskell QuickCheck-style fuzz at the function in the
// hopes of getting it to crash or perform an unexpected behavior.  It is not
// representative of anything outside of pure corruption.
func TestDecodeFuzz(t *testing.T) {
	f := func(stream []byte) (success bool) {
		defer func() {
			if err := recover(); err != nil && success {
				t.Logf("Decode(%#v) panicked: %#v", stream, err)
				success = false
			}
		}()
		_, err := Decode(bytes.NewReader(stream))
		switch err {
		case nil, ErrMagic, ErrVersion, io.ErrUnexpectedEOF, io.EOF:
			return true
		default:
			t.Logf("Decode(%#v) = _, %#v", stream, err)
			return false
		}
	}
	if err := quick.Check(f, nil); err != nil {
		t.Fatal(err)
	}
}

// TestDecodeFuzzLegitimate throws Haskell QuickCheck-style fuzz at the function
// in the hopes of getting it to crash or perform an unexpected behavior.  The
// data is semi-legitimate in the sense that the stream has a valid magic
// header.
func TestDecodeFuzzLegitimate(t *testing.T) {
	f := func(stream []byte) (success bool) {
		defer func() {
			if err := recover(); err != nil && success {
				t.Logf("Decode(%#v) panicked: %#v", stream, err)
				success = false
			}
		}()
		var buf bytes.Buffer
		if _, err := buf.Write(magic); err != nil {
			t.Fatal(err)
		}
		if _, err := buf.Write(stream); err != nil {
			t.Fatal(err)
		}
		_, err := Decode(&buf)
		switch err {
		case ErrVersion, io.ErrUnexpectedEOF, io.EOF:
			return true
		default:
			t.Logf("Decode(%#v) = _, %#v", stream, err)
			return false
		}
	}
	if err := quick.Check(f, nil); err != nil {
		t.Fatal(err)
	}
}

// TestDecodeTruncation is another blackbox fuzz variant that ensure that
// partially truncated streams do not crash the decoding.
func TestDecodeTruncation(t *testing.T) {
	r, err := loadFixture("pattern_1.splice")
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	payload, err := ioutil.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := recover(); err != nil {
			t.Fatalf("encountered unexpected panic: %#v", err)
		}
	}()
	for i := len(payload) - 1; i > len(magic); i-- {
		buf := bytes.NewReader(payload[0:i])
		_, err := Decode(buf)
		switch err {
		case io.EOF, io.ErrUnexpectedEOF:
			continue
		default:
			t.Fatalf("unexpected behavior on Decode(buf[0:%d]): %#v", i, err)
			break
		}
	}
}
