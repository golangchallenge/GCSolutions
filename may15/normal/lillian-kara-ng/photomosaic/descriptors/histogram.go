// Histogram is a descriptor of 16-bin histograms from the RGBA components of
// an image.

package descriptors

import (
	"errors"
	"fmt"
	"image"
	"math"
)

type Histogram struct {
	Data       [16][4]int
	Normalized [64]float64
}

type HistogramBuilder struct {
}

// Implements the DescriptionBuilder interface
func (b HistogramBuilder) GetDescription(m image.Image, bounds image.Rectangle) (Description, error) {
	var h Histogram

	h.Data = partialHistogram(m, bounds)
	h.Normalized = normalizeFlatten(h.Data, histogramMax(h.Data))

	return h, nil
}

// Given an image and bounding box, this computes the histogram
func partialHistogram(m image.Image, bounds image.Rectangle) [16][4]int {
	var histogram [16][4]int
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := m.At(x, y).RGBA()
			// A color's RGBA method returns values in the range [0, 65535].
			// Shifting by 12 reduces this to the range [0, 15].
			histogram[r>>12][0]++
			histogram[g>>12][1]++
			histogram[b>>12][2]++
			histogram[a>>12][3]++
		}
	}

	return histogram
}

// Flattens a histogram into a 1 dimensional array
func normalizeFlatten(h [16][4]int, max int) [64]float64 {
	var flattened [64]float64
	for y := 0; y < 16; y++ {
		for x := 0; x < 3; x++ { // only goes to 3 to ignore alpha channel
			flattened[y*4+x] = float64(h[y][x]) / float64(max)
		}
	}
	return flattened
}

// Finds the maximum value in a histogram
func histogramMax(h [16][4]int) int {
	var max int
	for y := 0; y < 16; y++ {
		for x := 0; x < 3; x++ { // only goes to 3 to ignore alpha channel
			if h[y][x] > max {
				max = h[y][x]
			}
		}
	}
	return max
}

// Implements the Description interface
func (h Histogram) MatchScore(other Description) (float64, error) {
	otherHistogram, ok := other.(Histogram)
	if !ok {
		fmt.Println("Description does not match Histogram")
		return math.MaxFloat64, errors.New("description type error")
	}
	var score float64
	for i := range h.Normalized {
		score += math.Abs(h.Normalized[i] - otherHistogram.Normalized[i])
	}
	return score, nil
}
