package mosaic

import (
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"path/filepath"
	"testing"
)

func TestMosaicGenerate(t *testing.T) {
	pool := []image.Image{
		openTestImage("tile0.jpg"),
		openTestImage("tile1.jpg"),
		openTestImage("tile2.jpg"),
		openTestImage("tile3.jpg"),
	}
	mosaic := New(openTestImage("target.jpg"))
	img, _ := mosaic.Generate(pool)

	f, _ := os.Create("../test_fixtures/mosaic.jpg")
	defer f.Close()

	jpeg.Encode(f, img, &jpeg.Options{Quality: jpeg.DefaultQuality})
}

func TestRGBMatcherMatch(t *testing.T) {
	pool := []image.Image{
		openTestImage("tile0.jpg"),
		openTestImage("tile1.jpg"),
		openTestImage("tile2.jpg"),
		openTestImage("tile3.jpg"),
	}
	target := openTestImage("target.jpg")
	tile := target.(mosaicImage).SubImage(image.Rect(10, 0, 20, 10))
	m := RGBMatcher{}
	match := m.Match(tile, pool)
	if match != pool[3] {
		got := ""
		for i, t := range pool {
			if t == match {
				got = fmt.Sprintf("tile%d.jpg", i)
				break
			}
		}
		t.Errorf("RGBMatcher.Match: Expected tile3.jpg, got %s", got)
	}
}

func TestHistogramMatcherMatch(t *testing.T) {
	pool := []image.Image{
		openTestImage("tile0.jpg"),
		openTestImage("tile1.jpg"),
		openTestImage("tile2.jpg"),
		openTestImage("tile3.jpg"),
	}
	target := openTestImage("target.jpg")
	tile := target.(mosaicImage).SubImage(image.Rect(10, 0, 20, 10))
	m := HistogramMatcher{}
	match := m.Match(tile, pool)
	if match != pool[2] {
		got := ""
		for i, t := range pool {
			if t == match {
				got = fmt.Sprintf("tile%d.jpg", i)
				break
			}
		}
		t.Errorf("HistogramMatcher.Match: Expected tile3.jpg, got %s", got)
	}
}

func openTestImage(filename string) image.Image {
	file, err := os.Open(filepath.Join("../test_fixtures", filename))
	defer file.Close()
	if err != nil {
		panic(err)
	}

	img, _, err := image.Decode(file)
	if err != nil {
		panic(err)
	}
	return img
}
