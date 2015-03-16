package drum

import (
	"path"
	"testing"
)

func BenchmarkPattern_String_1(b *testing.B) {
	benchmarkPatternString(b, "1")
}

func BenchmarkPattern_String_2(b *testing.B) {
	benchmarkPatternString(b, "2")
}

func BenchmarkPattern_String_3(b *testing.B) {
	benchmarkPatternString(b, "3")
}

func BenchmarkPattern_String_4(b *testing.B) {
	benchmarkPatternString(b, "4")
}

func BenchmarkPattern_String_5(b *testing.B) {
	benchmarkPatternString(b, "5")
}

func benchmarkPatternString(b *testing.B, no string) {
	p, err := DecodeFile(path.Join("fixtures", "pattern_"+no+".splice"))
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.String()
	}
}
