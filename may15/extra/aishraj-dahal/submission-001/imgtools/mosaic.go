package imgtools

import (
	"github.com/aishraj/gopherlisa/common"
	"image"
	"image/color"
	"image/draw"
)

const outputWidth = 1024
const tileSize = 8

func CreateMosaic(context *common.AppContext, srcName, destDirName string) image.Image {
	srcImg, err := LoadFromDisk(context, "/tmp/"+srcName+".jpg")
	if err != nil {
		context.Log.Fatal("Unable to open the input file. Error is ", err)
		return nil
	}
	sourceImage := ToNRGBA(srcImg)
	outputImageWidth := outputWidth
	outputImageHeight := findImageHeight(sourceImage.Bounds().Max.X, sourceImage.Bounds().Max.Y, outputImageWidth)
	resizedImage := Resize(context, sourceImage, outputImageWidth, outputImageHeight)
	imageTiles := createTiles(context, outputImageWidth, outputImageHeight)
	processedTiles := processColors(context, resizedImage, imageTiles)
	preparedTiles := updateSimilarColourImages(context, processedTiles, destDirName)
	photoImage := drawPhotoTiles(context, resizedImage, &preparedTiles, 8, destDirName)
	outputImagePath := "/tmp/output_" + srcName + ".jpeg"
	context.Log.Println("Generating output file now.......")
	err = SameToDisk(context, outputImagePath, &photoImage)
	return photoImage
}

func findImageHeight(originalWidth int, originalHeight int, targetWidth int) int {
	floatWidth := float64(originalWidth)
	floatHeight := float64(originalHeight)
	aspectRatio := float64(targetWidth) / floatWidth
	adjustedHeight := floatHeight * aspectRatio
	targetHeight := int(adjustedHeight)
	return targetHeight
}

func createTiles(context *common.AppContext, targetWidth int, targetHeight int) [][]common.Tile {
	horzTiles := targetWidth / tileSize
	if targetWidth%tileSize > 0 {
		horzTiles++
	}
	vertTiles := targetHeight / tileSize
	if targetHeight%tileSize > 0 {
		vertTiles++
	}
	imageTiles := make([][]common.Tile, horzTiles)
	for i := range imageTiles {
		imageTiles[i] = make([]common.Tile, vertTiles)
	}
	for x := 0; x < horzTiles; x++ {
		for y := 0; y < vertTiles; y++ {
			currentTile := &imageTiles[x][y]
			currentTile.X = x
			currentTile.Y = y
			tileStartX := x * tileSize
			tileStartY := y * tileSize
			tileEndX := tileStartX + tileSize
			tileEndY := tileStartY + tileSize
			if tileEndX >= targetWidth {
				tileEndX = targetWidth
			}
			if tileEndY >= targetHeight {
				tileEndY = targetHeight
			}
			currentTile.Rect = image.Rectangle{
				image.Point{tileStartX, tileStartY},
				image.Point{tileEndX, tileEndY},
			}
		}
	}
	return imageTiles
}

func processColors(context *common.AppContext, sourceImage image.Image, imageTiles [][]common.Tile) [][]common.Tile {
	for _, tiles := range imageTiles {
		for _, tile := range tiles {
			tile.AverageColor = findAverageColor(context, sourceImage, tile.Rect)
			imageTiles[tile.X][tile.Y].AverageColor = tile.AverageColor
		}
	}

	return imageTiles
}

func findAverageColor(context *common.AppContext, sourceImage image.Image, targetRect image.Rectangle) color.RGBA {
	croppedImage := Crop(context, sourceImage, targetRect)
	return FindProminentColour(croppedImage)

}

func updateSimilarColourImages(context *common.AppContext, imageTiles [][]common.Tile, indexName string) [][]common.Tile {
	for _, tiles := range imageTiles {
		for _, tile := range tiles {
			imageTiles[tile.X][tile.Y].MatchedImage = findClosestMatch(context, tile.AverageColor, indexName)
		}
	}
	return imageTiles
}

func drawPhotoTiles(context *common.AppContext, sourceImage image.Image, imageTiles *[][]common.Tile, tileWidth int, indexName string) image.Image {
	bounds := sourceImage.Bounds()
	photoImage := image.NewRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	draw.Draw(photoImage, photoImage.Bounds(), sourceImage, bounds.Min, draw.Src)
	for _, tiles := range *imageTiles {
		for _, tile := range tiles {
			if tile.MatchedImage != "" {
				tileImage, err := LoadFromDisk(context, common.DownloadBasePath+indexName+"/"+tile.MatchedImage)
				if err != nil {
					panic("Error loading image")
				}
				tileImageNRGBA := ToNRGBA(tileImage)
				resizedImage := Resize(context, tileImageNRGBA, tileWidth, tileWidth)
				draw.Draw(photoImage, tile.Rect, resizedImage, tileImage.Bounds().Min, draw.Src)
			}
		}
	}
	return photoImage
}
