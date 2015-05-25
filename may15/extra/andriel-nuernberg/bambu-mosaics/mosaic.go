package main

import (
	"image"

	"image/draw"
	_ "image/jpeg"
	"image/png"

	"os"
)

type Mosaic struct {
	SrcImg   image.Image
	DestImg  *image.RGBA
	TileSet  *TileSet
	TileSize int
}

// Create a new mosaic.
func NewMosaic(srcImg image.Image, tileSet *TileSet, size int) *Mosaic {
	destImg := image.NewRGBA(image.Rect(0, 0, srcImg.Bounds().Max.X, srcImg.Bounds().Max.Y))

	return &Mosaic{SrcImg: srcImg, DestImg: destImg, TileSet: tileSet, TileSize: size}
}

// Generate process the SrcImg and replaces each tile with the closest tile in
// the tile set, creating the DestImage.
func (m *Mosaic) Generate() {
	rgbImg, _ := m.SrcImg.(interface {
		SubImage(r image.Rectangle) image.Image
	})

	for x := m.SrcImg.Bounds().Min.X; x < m.SrcImg.Bounds().Max.X; x += m.TileSize {
		for y := m.SrcImg.Bounds().Min.Y; y < m.SrcImg.Bounds().Max.Y; y += m.TileSize {
			subImg := rgbImg.SubImage(image.Rect(x, y, x+m.TileSize, y+m.TileSize))

			avgColor := NewTile(subImg).AvgColor
			closestTile := m.TileSet.FindClosestByRGBA(avgColor)
			draw.Draw(m.DestImg, image.Rect(x, y, x+m.TileSize, y+m.TileSize), closestTile, image.ZP, draw.Src)
		}
	}
}

// Save saves the raw byte data of the DestImg in the file system.
func (m *Mosaic) Save(filepath string) error {
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	png.Encode(file, m.DestImg)

	return nil
}
