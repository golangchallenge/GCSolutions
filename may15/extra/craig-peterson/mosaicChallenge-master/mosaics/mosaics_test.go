package mosaics

import (
	"math/rand"
	"testing"
)

const testTileSize = 90
const testMaxDimension = 1000

func checkDim(t *testing.T, dim *mosaicDimensions, width, height, pptx, ppty, tilesX, tilesY, origX, origY, tileSize int) {
	if width != dim.width {
		t.Errorf("Width wrong. %d != %d\n", dim.width, width)
	}
	if height != dim.height {
		t.Errorf("Height wrong. %d != %d\n", dim.height, height)
	}
	if pptx != dim.sourcePixelsPerTileX {
		t.Errorf("PixelsPerTileX wrong. %d != %d\n", dim.sourcePixelsPerTileX, pptx)
	}
	if ppty != dim.sourcePixelsPerTileY {
		t.Errorf("PixelsPerTileY wrong. %d != %d\n", dim.sourcePixelsPerTileY, ppty)
	}
	if tilesX != dim.tilesX {
		t.Errorf("TilesX wrong. %d != %d\n", dim.tilesX, tilesX)
	}
	if tilesY != dim.tilesY {
		t.Errorf("TilesY wrong. %d != %d\n", dim.tilesY, tilesY)
	}
	if dim.tilesX*dim.sourcePixelsPerTileX > origX {
		t.Errorf("tilesX * ppt overflows original image! %d > %d \n", dim.tilesX*dim.sourcePixelsPerTileX, origX)
	}
	if dim.tilesY*dim.sourcePixelsPerTileY > origY {
		t.Errorf("tilesY * ppt overflows original image! %d > %d \n", dim.tilesY*dim.sourcePixelsPerTileY, origY)
	}
	if dim.height%tileSize != 0 {
		t.Errorf("Height %d not a multiple of tileSize\n", dim.height)
	}
	if dim.width%tileSize != 0 {
		t.Errorf("Width %d not a multiple of tileSize\n", dim.width)
	}
}

func TestMosaicDimensions_1To1(t *testing.T) {
	dim := getMosaicDimensions(990, 990, testMaxDimension, testTileSize)
	checkDim(t, dim, 990, 990, 90, 90, 11, 11, 1000, 1000, testTileSize)
}

func TestMosaicDimensions_Landscape(t *testing.T) {
	dim := getMosaicDimensions(2000, 1000, testMaxDimension, testTileSize)
	checkDim(t, dim, 990, 450, 181, 200, 11, 5, 2000, 1000, testTileSize)
}

func TestMosaicDimensions_Portrait(t *testing.T) {
	dim := getMosaicDimensions(100, 5000, testMaxDimension, testTileSize)
	checkDim(t, dim, 90, 990, 100, 454, 1, 11, 100, 5000, testTileSize)
}

func TestMosaicDimensionFuzz(t *testing.T) {
	//try some random sizes to see if we can break the bounds checking or ppt checks.
	rand.Seed(42)
	for i := 0; i < 1000; i++ {
		w, h := rand.Intn(10000), rand.Intn(10000)
		maxSize := rand.Intn(9000) + 50
		tileSize := rand.Intn(200) + 20
		dim := getMosaicDimensions(w, h, maxSize, tileSize)
		checkDim(t, dim, dim.width, dim.height, dim.sourcePixelsPerTileX, dim.sourcePixelsPerTileY, dim.tilesX, dim.tilesY, w, h, tileSize)
	}
}
