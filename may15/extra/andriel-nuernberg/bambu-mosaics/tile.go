package main

import (
	"image"
	"image/color"
	_ "image/jpeg"
)

// Tile is a image.Image but with the concept of a tile.
type Tile struct {
	image.Image
	AvgColor color.RGBA
}

// NewTile returns a new tile with its average color already processed.
func NewTile(image image.Image) *Tile {
	return &Tile{image, avgColor(image)}
}

// avgColor receives a tile and process its average color.
// Each pixel has its RGB color added and divided to the total number of
// pixels.
func avgColor(image image.Image) color.RGBA {
	sumRed := uint32(0)
	sumGreen := uint32(0)
	sumBlue := uint32(0)
	totalPx := uint32(0)

	for x := image.Bounds().Min.X; x < image.Bounds().Max.X; x++ {
		for y := image.Bounds().Min.Y; y < image.Bounds().Max.Y; y++ {
			c := image.At(x, y)
			r, g, b, _ := c.RGBA()

			sumRed += r
			sumGreen += g
			sumBlue += b
			totalPx += 1
		}
	}

	avgRed := sumRed / uint32(totalPx)
	avgGreen := sumGreen / uint32(totalPx)
	avgBlue := sumBlue / uint32(totalPx)

	return color.RGBA{uint8(avgRed / 256), uint8(avgGreen / 256), uint8(avgBlue / 256), 255}
}
