package mosaic

import (
	"errors"
	"image"
	"image/draw"
	_ "image/jpeg" // To be able to deal with jpeg images
)

// Mosaic creates a mosaic from a target image
type Mosaic struct {
	image      image.Image
	matcher    Matcher
	tileWidth  int
	tileHeight int
}

// Matcher defines an interface for different matching implementations that can be
// used to generate the mosaic
type Matcher interface {
	Match(tile image.Image, pool []image.Image) image.Image
}

// RGBMatcher uses avarage RGB color to match images
type RGBMatcher struct {
	colors map[image.Image]Color
}

// HistogramMatcher uses histograms to match images
type HistogramMatcher struct {
	histograms map[image.Image]Histogram
}

// Generate creates a mosaic form the provided tile images.
func (m *Mosaic) Generate(pool []image.Image) (image.Image, error) {
	if len(pool) == 0 {
		return nil, errors.New("No tile images in pool")
	}

	b := m.image.Bounds()
	targetWidth := b.Dx()
	targetHeight := b.Dy()

	// Get the tile width an height from the frist tile, we expect them all to be of equal size
	imagesWidth := pool[0].Bounds().Dx()
	imagesHeight := pool[0].Bounds().Dx()

	// The ratio from the target tile to the mosaic images
	xRatio := (imagesWidth / m.tileWidth)
	yRatio := (imagesHeight / m.tileHeight)

	mosaic := image.NewRGBA(image.Rect(0, 0, (xRatio * targetWidth), (yRatio * targetHeight)))

	pixelIterator(m.image, m.tileHeight, m.tileWidth, func(x, y int) {
		rect := newTileRectangle(x, y, targetWidth, targetHeight, m.tileWidth, m.tileHeight)
		tile := m.image.(mosaicImage).SubImage(rect)
		match := m.matcher.Match(tile, pool)
		srcRect := tile.Bounds()
		destRect := image.Rect((srcRect.Min.X * xRatio), (srcRect.Min.Y * yRatio), (srcRect.Max.X * xRatio), (srcRect.Max.Y * yRatio))
		draw.Draw(mosaic, destRect, match, image.Point{1, 1}, draw.Src)
	})

	return mosaic, nil
}

// Match finds similar images by comparing RGB colors
func (m *RGBMatcher) Match(tile image.Image, pool []image.Image) image.Image {
	if m.colors == nil {
		m.colors = make(map[image.Image]Color)
		for _, img := range pool {
			m.colors[img] = AverageColor(img)
		}
	}

	c0 := AverageColor(tile)

	var match image.Image
	var bestDist float64

	for _, img := range pool {
		c1 := m.colors[img]
		dist := DistanceRGB(c0, c1)
		if (match == nil) || (dist < bestDist) {
			match = img
			bestDist = dist
		}
	}
	return match
}

// Match finds similar images by comparing histograms
func (m *HistogramMatcher) Match(tile image.Image, pool []image.Image) image.Image {
	if m.histograms == nil {
		m.histograms = make(map[image.Image]Histogram)
		for _, img := range pool {
			m.histograms[img] = NewHistogram(img)
		}
	}

	h0 := NewHistogram(tile)

	var match image.Image
	var bestDist float64

	for _, img := range pool {
		h1 := m.histograms[img]
		dist := DistanceHistogram(&h0, &h1)
		if (match == nil) || (dist < bestDist) {
			match = img
			bestDist = dist
		}
	}
	return match
}

// mosaicImage defines an interface that implements the SubImage function to create
// tiles from the target image.
type mosaicImage interface {
	image.Image
	SubImage(image.Rectangle) image.Image
}

// newTileRectangle creates a ractange from given arguments.
func newTileRectangle(x, y, imgWidth, imgHeight, tileHeight, tileWidth int) image.Rectangle {
	x0 := x
	y0 := y
	x1 := x0 + tileWidth
	y1 := y0 + tileHeight
	// handle cases where the tile would overlap the image and cause a panic
	if x1 > imgWidth {
		x1 = imgWidth
	}
	if y1 > imgHeight {
		y1 = imgHeight
	}
	rect := image.Rect(x0, y0, x1, y1)
	return rect
}

// pixelIterator iterates over all pixels in an image with a given interval and
// executes a callback with the X and Y coordinates.
func pixelIterator(img image.Image, xInterval, yInterval int, cb func(int, int)) {
	b := img.Bounds()
	for x := b.Min.X; x < b.Max.X; x += xInterval {
		for y := b.Min.Y; y < b.Max.Y; y += yInterval {
			cb(x, y)
		}
	}
}

// New creates a new Mosaic pointer.
func New(img image.Image) *Mosaic {
	return &Mosaic{
		image:      img,
		matcher:    &RGBMatcher{},
		tileWidth:  10,
		tileHeight: 10,
	}
}
