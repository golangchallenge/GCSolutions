package main

import (
	"image"
	"image/color"
	"os"
	"testing"
)

var tileTest = "testdata/cat-tile.jpg"

func TestNewTile(t *testing.T) {
	file, err := os.Open(tileTest)
	if err != nil {
		t.Error(err)
	}

	img, _, err := image.Decode(file)
	if err != nil {
		t.Error(err)
	}

	tile := NewTile(img)

	if tile.Image.Bounds().Max.X != 150 {
		t.Error("Tile bound X got %v. Expected %v.", tile.Image.Bounds().Max.X, 150)
	}

	if tile.Image.Bounds().Max.Y != 150 {
		t.Error("Tile bound Y got %v. Expected %v.", tile.Image.Bounds().Max.Y, 150)
	}

	c := color.RGBA{175, 159, 138, 255}
	if tile.AvgColor != c {
		t.Errorf("RGB AvgColor got %v. Expected %v.", tile.AvgColor, c)
	}
}
