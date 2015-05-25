package main

import (
	"image/color"
	"testing"
)

var tileSet, _ = NewTileSet("testdata/tiles")
var redTile = color.RGBA{254, 0, 0, 255}
var greenTile = color.RGBA{0, 255, 1, 255}
var blueTile = color.RGBA{0, 0, 254, 255}

func TestFindClosestByRGBA(t *testing.T) {
	redish := color.RGBA{190, 10, 18, 255}
	found := tileSet.FindClosestByRGBA(redish)
	if found.AvgColor != redTile {
		t.Errorf("Not equal: %#s (expected). %#s (actual)", redTile, found.AvgColor)
	}

	greenish := color.RGBA{18, 213, 165, 255}
	found = tileSet.FindClosestByRGBA(greenish)
	if found.AvgColor != greenTile {
		t.Errorf("Not equal: %#s (expected). %#s (actual)", greenTile, found.AvgColor)
	}

	blueish := color.RGBA{18, 124, 192, 255}
	found = tileSet.FindClosestByRGBA(blueish)
	if found.AvgColor != blueTile {
		t.Errorf("Not equal: %#s (expected). %#s (actual)", blueTile, found.AvgColor)
	}
}
