package mosaic

import (
	"testing"

	"github.com/ptrost/mosaic2go/test"
)

func TestAverageColor(t *testing.T) {
	c := AverageColor(openTestImage("target.jpg"))
	test.Assert("AverageColor R", uint32(34676), c.R, t)
	test.Assert("AverageColor G", uint32(57807), c.G, t)
	test.Assert("AverageColor B", uint32(15416), c.B, t)
}

func TestDistanceHistogram(t *testing.T) {
	h0 := NewHistogram(openTestImage("img.jpg"))
	h1 := NewHistogram(openTestImage("img.jpg"))
	result := DistanceHistogram(&h0, &h1)
	expected := 749.6217574983835
	test.Assert("HistogramDistance", expected, result, t)
}

func TestHistogram(t *testing.T) {
	NewHistogram(openTestImage("img.jpg"))
}
