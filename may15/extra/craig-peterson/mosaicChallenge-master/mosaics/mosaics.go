// Package moasics provides the main logic for generating a photomosaic from a set of smaller images
package mosaics

import (
	"bytes"
	"image"
	"image/draw"
	"image/gif"
	"log"
)

// TileSize is the expected size of each subImage. If images are larger than this value, only the top-left corner up to TileSize will be used.
// If any dimension is smaller than this, they will be backfilled with black. If possible, subImages should be prescaled to a square of this size.
const DefaultTileSize = 90

//Maximum dimension in an finalized mosaic. Target image will be scaled up or down such that its largest side is this length.
const DefaultMaxDimension = DefaultTileSize * 70

func BuildMosaicFromLibrary(master image.Image, tiles *ThumbnailLibrary, reporter chan<- float64) image.Image {
	tileSize := DefaultTileSize
	//calculate final dimensions of resultant image
	dim := getMosaicDimensions(master.Bounds().Dx(), master.Bounds().Dy(), DefaultMaxDimension, tileSize)

	output := image.NewRGBA(image.Rect(0, 0, dim.width, dim.height))

	for tileY := 0; tileY < dim.tilesY; tileY++ {
		if reporter != nil {
			reporter <- float64(tileY) / float64(dim.tilesY) * 100
		}
		for tileX := 0; tileX < dim.tilesX; tileX++ {
			c := tiles.evaluator.Evaluate(master, tileX*dim.sourcePixelsPerTileX, tileY*dim.sourcePixelsPerTileY, dim.sourcePixelsPerTileX, dim.sourcePixelsPerTileY)
			tile := tiles.getBestMatch(c)
			rect := tile.Bounds().Add(image.Point{tileX * tileSize, tileY * tileSize})
			draw.Draw(output, rect, tile, image.ZP, draw.Over)
		}
	}
	return output
}

// Calclulates scaling factor for final mosaic so we can map original image tiles onto the final mosaic.
// Scale the largest side to the maxDimension, and scale the smaller side to (roughly) match, while still being a multiple of tile size in all dimensions.
// Calculate the number of source pixels in each dimension that correspond to each output tile.
func getMosaicDimensions(originalX, originalY int, maxDimension, tileSize int) *mosaicDimensions {
	dim := mosaicDimensions{}
	if originalX >= originalY {
		dim.width = maxDimension
		dim.height = int(float64(originalY) * (float64(maxDimension) / float64(originalX)))
	} else {
		dim.height = maxDimension
		dim.width = int(float64(originalX) * (float64(maxDimension) / float64(originalY)))
	}
	// Make sure we are a multiple of tile size in both directions.
	sanitize := func(s int) int {
		if s < tileSize {
			return tileSize
		}
		return s - (s % tileSize)
	}
	dim.height = sanitize(dim.height)
	dim.width = sanitize(dim.width)

	//count tiles and source pixels per resultant tile
	dim.tilesX = dim.width / tileSize
	dim.tilesY = dim.height / tileSize
	dim.sourcePixelsPerTileX = originalX / dim.tilesX
	dim.sourcePixelsPerTileY = originalY / dim.tilesY
	return &dim
}

type mosaicDimensions struct {
	width                int
	height               int
	sourcePixelsPerTileX int
	sourcePixelsPerTileY int
	tilesX, tilesY       int
}

func BuildGifzaic(g *gif.GIF, tiles *ThumbnailLibrary, reporter chan<- float64) (*gif.GIF, error) {
	for i, img := range g.Image {
		log.Printf("Building frame %d of %d\n------------------\n", i+1, len(g.Image))
		mozImg := BuildMosaicFromLibrary(img, tiles, reporter)
		log.Printf("Encoding frame %d of %d\n", i+1, len(g.Image))
		buf := &bytes.Buffer{}
		err := gif.Encode(buf, mozImg, &gif.Options{50, nil, nil})
		if err != nil {
			return nil, err
		}
		log.Printf("Encoding frame %d of %d\n", i+1, len(g.Image))
		newGif, err := gif.DecodeAll(buf)
		if err != nil {
			return nil, err
		}
		g.Image[i] = newGif.Image[0]
	}
	return g, nil
}
