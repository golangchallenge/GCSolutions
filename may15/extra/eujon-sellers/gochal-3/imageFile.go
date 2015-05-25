package main

import (
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
)

type ImageFile struct {
	Name  string
	Image image.Image
	Tiles []*OrigTile
	YEnd  int
	XEnd  int
}

type OrigTile struct {
	Filename string
	YStart   int
	YEnd     int
	XStart   int
	XEnd     int
	MatchURL string
	AvgRGB   map[string]uint32
}

// create an array of 10x10 tiles for img
func (img *ImageFile) createTiles() {
	rowCount := 0
	columnCount := 0
	reader, err := os.Open(img.Name)
	if err != nil {
		logger.Println("error reading original image: ", err)
	}
	defer reader.Close()

	srcImg, _, err := image.Decode(reader)
	if err != nil {
		logger.Println("error decoding original image: ", err)
	}
	img.Image = srcImg
	bounds := srcImg.Bounds()
	img.YEnd = bounds.Dy()
	img.XEnd = bounds.Dx()
	if bounds.Max.Y%10 > 0 {
		rowCount = (bounds.Max.Y / 10) + 1
	} else {
		rowCount = bounds.Max.Y / 10
	}

	if bounds.Max.X%10 > 0 {
		columnCount = (bounds.Max.X / 10) + 1
	} else {
		columnCount = bounds.Max.X / 10
	}

	var origTiles []*OrigTile
	yStart := bounds.Min.Y
	yEnd := yStart + 10
	xStart := bounds.Min.X
	xEnd := xStart + 10
	for i := 0; i < rowCount; i++ {
		if i != 0 {
			yStart = i * 10
			yEnd = yStart + 10
			if yEnd >= rowCount*10 {
				yEnd = bounds.Max.Y
			}
			xStart = bounds.Min.X
			xEnd = xStart + 10
		}
		for j := 0; j < columnCount; j++ {
			tempTile := OrigTile{Filename: img.Name, YStart: yStart, YEnd: yEnd, XStart: xStart, XEnd: xEnd}
			origTiles = append(origTiles, &tempTile)
			xStart += 10
			xEnd += 10
			if xEnd >= columnCount*10 {
				xEnd = bounds.Max.X
			}
		}
	}
	img.Tiles = origTiles
}

// get the avg RGB value for the 10x10 tile
func (ot *OrigTile) buildColorMap(srcImg ImageFile) {
	var dstImg image.Image
	switch srcImg.Image.(type) {
	case *image.RGBA:
		dstImg = srcImg.Image.(*image.RGBA).SubImage(image.Rect(ot.XStart, ot.YStart, ot.XEnd, ot.YEnd))
	case *image.RGBA64:
		dstImg = srcImg.Image.(*image.RGBA64).SubImage(image.Rect(ot.XStart, ot.YStart, ot.XEnd, ot.YEnd))
	case *image.NRGBA:
		dstImg = srcImg.Image.(*image.NRGBA).SubImage(image.Rect(ot.XStart, ot.YStart, ot.XEnd, ot.YEnd))
	case *image.NRGBA64:
		dstImg = srcImg.Image.(*image.NRGBA64).SubImage(image.Rect(ot.XStart, ot.YStart, ot.XEnd, ot.YEnd))
	case *image.YCbCr:
		dstImg = srcImg.Image.(*image.YCbCr).SubImage(image.Rect(ot.XStart, ot.YStart, ot.XEnd, ot.YEnd))
	}
	avgRGB := avgRGB(dstImg, false)
	ot.AvgRGB = avgRGB
}
