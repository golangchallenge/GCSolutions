package mosaic

import (
	"image"
	"image/color"
	"math"
)

// Color defines a type for colors used for finding mosaic tiles.
type Color struct {
	R, G, B, A uint32
}

// RGBA implements the color.Color interface.
func (c Color) RGBA() (r, g, b, a uint32) {
	r = c.R
	g = c.G
	b = c.B
	a = 0xFFFF
	return
}

// AverageColor find the average color of a given image.
func AverageColor(img image.Image) Color {
	avg := AvgColor{}
	interval := 5
	pixelIterator(img, interval, interval, func(x, y int) {
		c := img.At(x, y)
		avg.Add(c)
	})

	return avg.Color
}

// AvgColor is used to find the average color of an image.
type AvgColor struct {
	count uint32
	Color
}

// RGBA implements the color.Color interface.
func (c *AvgColor) RGBA() (r, g, b, a uint32) {
	return c.Color.RGBA()
}

// Add add a color to average color and calculates a new average.
func (c *AvgColor) Add(cl color.Color) {
	c.count++
	r, g, b, _ := cl.RGBA()
	c.R = newAvg(c.R, r, c.count)
	c.G = newAvg(c.G, g, c.count)
	c.B = newAvg(c.B, b, c.count)
}

// newAvg is a helper function for the average color calculation.
func newAvg(baseAvg, newValue, count uint32) uint32 {
	if baseAvg == 0 {
		return newValue
	}
	return ((baseAvg * count) + newValue) / (count + 1)
}

// Histogram hold the data for a 16 bit Histogram.
type Histogram struct {
	Data [16][4]int
}

// NewHistogram creates a histogram ofr a given image.
func NewHistogram(img image.Image) Histogram {
	histogram := Histogram{}
	interval := 10
	pixelIterator(img, interval, interval, func(x, y int) {
		r, g, b, a := img.At(x, y).RGBA()
		// A color's RGBA method returns values in the range [0, 65535].
		// Shifting by 12 reduces this to the range [0, 15].
		histogram.Data[r>>12][0]++
		histogram.Data[g>>12][1]++
		histogram.Data[b>>12][2]++
		histogram.Data[a>>12][3]++
	})

	return histogram
}

// DistanceHistogram calculates the distance between two histograms.
func DistanceHistogram(h0, h1 *Histogram) float64 {
	eps := 1e-10

	d := float64(0)
	for i := 0; i < 16; i++ {
		for j := 0; j < 4; j++ {
			v0 := math.Pow((float64(h0.Data[i][j]/10) - float64(h1.Data[i][j]/75)), 2)
			v1 := ((h0.Data[i][j] / 10) + (h1.Data[i][j] / 75) + int(eps))
			r := v0 / float64(v1)
			if !math.IsNaN(r) {
				d += r
			}
			d += math.Abs(float64(h0.Data[i][0]) - float64(h1.Data[i][0]))
		}
	}
	return d
}

// DistanceRGB calculates the distance between two colors
func DistanceRGB(c0, c1 color.Color) float64 {
	r0, g0, b0, _ := c0.RGBA()
	r1, g1, b1, _ := c1.RGBA()
	return math.Sqrt(float64(sq(r0-r1) + sq(g0-g1) + sq(b0-b1)))
}

func sq(v uint32) uint32 {
	return v * v
}
