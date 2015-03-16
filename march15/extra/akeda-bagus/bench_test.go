package drum

import (
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func BenchmarkDecodePattern1(b *testing.B) {
	bencmarkDecode(b, "pattern_1.splice")
}

func BenchmarkDecodePattern2(b *testing.B) {
	bencmarkDecode(b, "pattern_2.splice")
}

func BenchmarkDecodePattern3(b *testing.B) {
	bencmarkDecode(b, "pattern_3.splice")
}

func BenchmarkDecodePattern4(b *testing.B) {
	bencmarkDecode(b, "pattern_4.splice")
}

func BenchmarkDecodePattern5(b *testing.B) {
	bencmarkDecode(b, "pattern_5.splice")
}

func bencmarkDecode(b *testing.B, f string) {
	b.StopTimer()
	for n := 0; n < b.N; n++ {
		f, _ := os.Open(path.Join("fixtures", f))
		d := NewDecoder(f)
		b.StartTimer()
		d.Decode()
		b.StopTimer()
	}
}

func BenchmarkEncodePattern1(b *testing.B) {
	bencmarkEncode(b, "pattern_1.splice")
}

func BenchmarkEncodePattern2(b *testing.B) {
	bencmarkEncode(b, "pattern_2.splice")
}

func BenchmarkEncodePattern3(b *testing.B) {
	bencmarkEncode(b, "pattern_3.splice")
}

func BenchmarkEncodePattern4(b *testing.B) {
	bencmarkEncode(b, "pattern_4.splice")
}

func BenchmarkEncodePattern5(b *testing.B) {
	bencmarkEncode(b, "pattern_5.splice")
}

func bencmarkEncode(b *testing.B, f string) {
	b.StopTimer()
	for n := 0; n < b.N; n++ {
		p, _ := DecodeFile(path.Join("fixtures", f))
		b.StartTimer()
		NewEncoder(ioutil.Discard).Encode(p)
		b.StopTimer()
	}

}
