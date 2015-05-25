package main

import (
	"image"
	"image/color"
	"os"
	"testing"
)

var srcImg = func() image.Image {
	f, _ := os.Open("testdata/cat-tile.jpg")
	img, _, _ := image.Decode(f)
	return img
}
var tileSize = 30

func TestGenerate(t *testing.T) {
	m := NewMosaic(srcImg(), tileSet, tileSize)

	for x := m.SrcImg.Bounds().Min.X; x < m.SrcImg.Bounds().Max.X; x += m.TileSize {
		for y := m.SrcImg.Bounds().Min.Y; y < m.SrcImg.Bounds().Max.Y; y += m.TileSize {
			if (m.DestImg.At(x, y) != color.RGBA{0, 0, 0, 0}) {
				t.Errorf("Not equal: Pixel[%v, %v] %#s (expected). Pixel[%v, %v] %#s (actual)", x, y, m.DestImg.At(x, y), x, y, color.RGBA{0, 0, 0, 0})
			}
		}
	}

	m.Generate()

	for x := m.SrcImg.Bounds().Min.X; x < m.SrcImg.Bounds().Max.X; x += m.TileSize {
		for y := m.SrcImg.Bounds().Min.Y; y < m.SrcImg.Bounds().Max.Y; y += m.TileSize {
			if (m.DestImg.At(x, y) == color.RGBA{0, 0, 0, 0}) {
				t.Errorf("Not equal: Pixel[%v, %v] %#s (expected). Pixel[%v, %v] %#s (actual)", x, y, m.DestImg.At(x, y), x, y, color.RGBA{0, 0, 0, 0})
			}
		}
	}
}
