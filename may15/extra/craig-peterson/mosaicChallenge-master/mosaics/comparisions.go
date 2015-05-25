package mosaics

import (
	"image"
	"image/color"
	"math"
)

// Evaluator is a common interface to use for evaluating images, and comparing them.
// Functions return interface{} because we do not know what data an evaluator may want
// to emit about an image.
//
// It is guaranteed that the Compare function will only be called with values previously emitted by the
// Evaluate function.
type Evaluator interface {
	// Evaluate an image within the specified bounds and return whatever metadata is representative of that section.
	Evaluate(img image.Image, x, y, width, height int) interface{}
	// Compare two peices of image metadata and score their similarity. Lower numbers mean the images are a better match.
	Compare(a, b interface{}) float64
}

type average struct{}

func AveragingEvaluator() Evaluator {
	return &average{}
}

func (v *average) Evaluate(img image.Image, x, y, width, height int) interface{} {
	var numPixels, rTotal, gTotal, bTotal uint64
	for curY := y; curY < y+height; curY++ {
		for curX := x; curX < x+width; curX++ {
			numPixels++
			r, g, b, _ := img.At(curX, curY).RGBA()

			rTotal += uint64(r >> 8)
			gTotal += uint64(g >> 8)
			bTotal += uint64(b >> 8)
		}
	}
	return color.RGBA{uint8(rTotal / numPixels), uint8(gTotal / numPixels), uint8(bTotal / numPixels), 255}
}

func (v *average) Compare(a, b interface{}) float64 {
	ca, cb := a.(color.Color), b.(color.Color)
	ar, ag, ab, _ := ca.RGBA()
	br, bg, bb, _ := cb.RGBA()
	dr := ar - br
	dg := ag - bg
	db := ab - bb
	return math.Sqrt(float64(dr*dr) + float64(dg*dg) + float64(db*db))
}

type gridEvaluator struct {
	segments int // number of divisions in any dimension. segments = 3 iwll evaluate a 3x3 grid
	ev       average
}

// A Grid evaluator breaks each thumbnail and target image segment into an NxN grid.
// Each segment is evaluated independently, and the candidate with the lowest total average difference
// across all segments is chosen. This gives better matching at a small performance cost.
func GridEvaluator(size int) Evaluator {
	return &gridEvaluator{size, average{}}
}

func (v *gridEvaluator) Evaluate(img image.Image, x, y, width, height int) interface{} {

	var numPixels uint64
	numBuckets := v.segments * v.segments
	var rSums = make([]uint64, numBuckets)
	var gSums = make([]uint64, numBuckets)
	var bSums = make([]uint64, numBuckets)
	for relY := 0; relY < height; relY++ {
		for relX := 0; relX < width; relX++ {
			numPixels++
			bucket := relY/height + relX/width
			curX, curY := relX+x, relY+y
			r, g, b, _ := img.At(curX, curY).RGBA()

			rSums[bucket] += uint64(r >> 8)
			gSums[bucket] += uint64(g >> 8)
			bSums[bucket] += uint64(b >> 8)
		}
	}
	result := make([]color.RGBA, numBuckets)
	for i := 0; i < numBuckets; i++ {
		result[i] = color.RGBA{uint8(rSums[i] / numPixels), uint8(gSums[i] / numPixels), uint8(bSums[i] / numPixels), 255}
	}
	return result
}

func (v *gridEvaluator) Compare(a, b interface{}) float64 {
	aList, bList := a.([]color.RGBA), b.([]color.RGBA)
	if len(aList) != len(bList) {
		panic("mismatched list sizes")
	}
	var diffSum float64
	for i, aRes := range aList {
		bRes := bList[i]
		diffSum += v.ev.Compare(aRes, bRes)
	}
	return diffSum
}
