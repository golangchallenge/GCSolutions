package mosaics

import (
	"image"
	_ "image/jpeg"
	"log"
	"os"
	"testing"
)

var testImage image.Image

func init() {
	f, err := os.Open("test.jpg")
	if err != nil {
		log.Fatal(err)
	}
	testImage, _, err = image.Decode(f)
	if err != nil {
		log.Fatal(err)
	}
}

func BenchmarkAveragingEvaluator(b *testing.B) {
	ev := average{}
	for i := 0; i < b.N; i++ {
		ev.Evaluate(testImage, 0, 0, 90, 90)
	}
}

func BenchmarkAveragingCompare(b *testing.B) {
	ev := average{}
	val := ev.Evaluate(testImage, 0, 0, 90, 90)
	for i := 0; i < b.N; i++ {
		ev.Compare(val, val)
	}
}

func Benchmark2x2Evaluator(b *testing.B) {
	ev := GridEvaluator(2)
	for i := 0; i < b.N; i++ {
		ev.Evaluate(testImage, 0, 0, 90, 90)
	}
}

func Benchmark2x2Compare(b *testing.B) {
	ev := GridEvaluator(5)
	val := ev.Evaluate(testImage, 0, 0, 90, 90)
	for i := 0; i < b.N; i++ {
		ev.Compare(val, val)
	}
}

func Benchmark5x5Evaluator(b *testing.B) {
	ev := GridEvaluator(5)
	for i := 0; i < b.N; i++ {
		ev.Evaluate(testImage, 0, 0, 90, 90)
	}
}

func Benchmark5x5Compare(b *testing.B) {
	ev := GridEvaluator(5)
	val := ev.Evaluate(testImage, 0, 0, 90, 90)
	for i := 0; i < b.N; i++ {
		ev.Compare(val, val)
	}
}
