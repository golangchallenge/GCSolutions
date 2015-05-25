// Tile package includes some utilities to divide image bounds into a slice of
// image bounds by some given dimension.

package tile

import (
	"image"
)

// Input Width and Height are count
func ByCount(b image.Rectangle, w, h int) ([]image.Rectangle, int) {
	heightpx := b.Max.Y - b.Min.Y
	widthpx := b.Max.X - b.Min.X
	tileHeight := heightpx / h
	tileWidth := widthpx / w

	return grid(b, w, h, tileWidth, tileHeight)
}

// Input Width and Height are pixels
func ByPixel(b image.Rectangle, w, h int) ([]image.Rectangle, int) {
	heightpx := b.Max.Y - b.Min.Y
	widthpx := b.Max.X - b.Min.X
	tilesHigh := heightpx / h
	tilesWide := widthpx / w

	return grid(b, tilesWide, tilesHigh, w, h)
}

func grid(b image.Rectangle, tilesWide, tilesHigh, tileWidth, tileHeight int) ([]image.Rectangle, int) {
	grid := make([]image.Rectangle, tilesHigh*tilesWide)
	count := 0

	for y := 0; y < tilesHigh; y++ {
		for x := 0; x < tilesWide; x++ {
			tl := image.Point{x*tileWidth + b.Min.X, y*tileHeight + b.Min.Y}
			br := image.Point{tl.X + tileWidth, tl.Y + tileHeight}
			grid[count] = image.Rectangle{tl, br}
			count++
		}
	}
	return grid, tilesWide
}
