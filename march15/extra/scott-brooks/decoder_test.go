package drum

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path"
	"testing"
)

func TestDecodeFile(t *testing.T) {
	tData := []struct {
		path   string
		output string
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
		},
		{"pattern_2.splice",
			`Saved with HW Version: 0.808-alpha
Tempo: 98.4
(0) kick	|x---|----|x---|----|
(1) snare	|----|x---|----|x---|
(3) hh-open	|--x-|--x-|x-x-|--x-|
(5) cowbell	|----|----|x---|----|
`,
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
		},
		{"pattern_4.splice",
			`Saved with HW Version: 0.909
Tempo: 240
(0) SubKick	|----|----|----|----|
(1) Kick	|x---|----|x---|----|
(99) Maracas	|x-x-|x-x-|x-x-|x-x-|
(255) Low Conga	|----|x---|----|x---|
`,
		},
		{"pattern_5.splice",
			`Saved with HW Version: 0.708-alpha
Tempo: 999
(1) Kick	|x---|----|x---|----|
(2) HiHat	|x-x-|x-x-|x-x-|x-x-|
`,
		},
	}

	for _, exp := range tData {
		decoded, err := DecodeFile(path.Join("fixtures", exp.path))
		if err != nil {
			t.Fatalf("something went wrong decoding %s - %v", exp.path, err)
		}
		if fmt.Sprint(decoded) != exp.output {
			t.Logf("decoded:\n%#v\n", fmt.Sprint(decoded))
			t.Logf("expected:\n%#v\n", exp.output)
			t.Fatalf("%s wasn't decoded as expect.\nGot:\n%s\nExpected:\n%s",
				exp.path, decoded, exp.output)
		}
	}
}

func BenchmarkDecodeFile(b *testing.B) {
	for i := 0; i < b.N; i++ {
		DecodeFile(path.Join("fixtures", "pattern_3.splice"))
	}

}

func benchmarkDecode(i int, b *testing.B) {
	fileName := fmt.Sprintf("pattern_%d.splice", i)

	data, err := ioutil.ReadFile(path.Join("fixtures", fileName))
	if err != nil {
		b.Fatalf("Error opening fixture: %+v", err)
	}

	for n := 0; n < b.N; n++ {
		_, err := decode(bytes.NewBuffer(data))
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDecode1(b *testing.B) {
	benchmarkDecode(1, b)
}

func BenchmarkDecode2(b *testing.B) {
	benchmarkDecode(2, b)
}

func BenchmarkDecode3(b *testing.B) {
	benchmarkDecode(3, b)
}

func BenchmarkDecode4(b *testing.B) {
	benchmarkDecode(4, b)
}

func BenchmarkDecode5(b *testing.B) {
	benchmarkDecode(5, b)
}

func TestCleanBytesToString(t *testing.T) {
	// Byte array
	if cleanBytesToString([]byte{'h', 'e', 'l', 'l', 'o'}) != "hello" {
		t.Error("byte string did not match string")
	}
	// Byte array with nulls
	if cleanBytesToString([]byte{'h', 'e', 'l', 'l', 'o', 0, 0}) != "hello" {
		t.Error("byte string did not match string")
	}

}
