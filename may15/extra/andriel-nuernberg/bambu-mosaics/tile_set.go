package main

import (
	"image"
	"image/color"
	"math"
	"os"
	"path/filepath"
	"strings"
)

type TileSet struct {
	Tiles []*Tile
}

// NewTileSet iterate over the given path and reads all ".jpg" images.
// A new tile is created for each image, then a TileSet struct is returned
// containing all tiles.
func NewTileSet(path string) (*TileSet, error) {
	fd, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fd.Close()

	list, err := fd.Readdir(-1)
	if err != nil {
		return nil, err
	}

	ts := &TileSet{}

	for _, f := range list {
		if strings.HasSuffix(f.Name(), ".jpg") {
			filename := filepath.Join(path, f.Name())
			fo, err := os.Open(filename)
			if err != nil {
				fo.Close()
				continue
			}

			img, _, err := image.Decode(fo)
			if err != nil {
				fo.Close()
				continue
			}

			ts.Tiles = append(ts.Tiles, NewTile(img))
			fo.Close()
		}
	}

	return ts, nil
}

// FindClosestByRGBA looks for a tile that has the closest average color compared
// to the given RGBA.
//
// Once the human eye is more sensitive for green colors than blues, different
// weights was used for each color. 0.3 for red, 0.59 for green and 011 for blue.
//
// The color distance is calculated between the given color and each tile
// average color. The lowest distance means better aproximation.
//
// Once iterated over all the tile set, the tile that got the lowest color
// distanced is returned.
func (ts *TileSet) FindClosestByRGBA(color color.RGBA) *Tile {
	min := 1000.0
	result := ts.Tiles[0]

	for _, t := range ts.Tiles {
		dr := (float64(int(t.AvgColor.R)-int(color.R)) * float64(0.3)) * (float64(int(t.AvgColor.R)-int(color.R)) * float64(0.3))
		dg := (float64(int(t.AvgColor.G)-int(color.G)) * float64(0.59)) * (float64(int(t.AvgColor.G)-int(color.G)) * float64(0.59))
		db := (float64(int(t.AvgColor.B)-int(color.B)) * float64(0.11)) * (float64(int(t.AvgColor.B)-int(color.B)) * float64(0.11))
		d := math.Sqrt(dr + dg + db)
		if d < min {
			min = d
			result = t
		}
	}

	return result
}
