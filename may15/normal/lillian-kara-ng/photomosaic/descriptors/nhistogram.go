// N-Histogram is a descriptor built from many Histograms
// For example, a N-Histogram Builder with 2 x 2, width x height, will create
// descriptors equivalent to 4 Histograms, each corresponding to one quadrant
// of the image.

package descriptors

import (
	"bitbucket.org/lillian_ng/photomosaic/tile"
	"errors"
	"fmt"
	"image"
	"math"
)

type NHistogram struct {
	Normalized []float64
}

// Width and Height define how many Histograms to create from a bounding region
type NHistogramBuilder struct {
	Width, Height int
}

// Implements the DescriptionBuilder interface
func (b NHistogramBuilder) GetDescription(m image.Image, bounds image.Rectangle) (Description, error) {
	var h NHistogram

	grid, _ := tile.ByCount(bounds, b.Width, b.Height)
	histograms := make([]Histogram, len(grid))
	var max int

	// Get each histogram
	for i := range grid {
		histograms[i].Data = partialHistogram(m, grid[i])
		thisMax := histogramMax(histograms[i].Data)
		if thisMax > max {
			max = thisMax
		}
	}

	// Normalize the histograms based on global max
	for i := range histograms {
		histograms[i].Normalized = normalizeFlatten(histograms[i].Data, max)
	}

	// Flatten into a 1 dimensional slice
	h.Normalized = make([]float64, len(grid)*64)
	count := 0
	for i := range histograms {
		for _, v := range histograms[i].Normalized {
			h.Normalized[count] = v
			count++
		}
	}

	return h, nil
}

// Implements the Description interface
func (h NHistogram) MatchScore(other Description) (float64, error) {
	otherHistogram, ok := other.(NHistogram)
	if !ok {
		fmt.Println("Description does not match QuadHistogram")
		return math.MaxFloat64, errors.New("description type error")
	}
	var score float64
	for i := range h.Normalized {
		score += math.Abs(h.Normalized[i] - otherHistogram.Normalized[i])
	}
	return score, nil
}
