package main

import (
	"image"
	"image/color"
	"log"
	"testing"
)

func imageEqual(a, b image.Image) bool {
	if a.Bounds() != b.Bounds() {
		return false
	}
	for y := a.Bounds().Min.Y; y < a.Bounds().Max.Y; y++ {
		for x := a.Bounds().Min.X; x < a.Bounds().Max.X; x++ {
			if a.At(x, y).(color.RGBA) != b.At(x, y).(color.RGBA) {
				return false
			}
		}
	}
	return true
}

func TestMosaic(t *testing.T) {
	var tests = []struct {
		target         string
		tiles          string
		good           string
		xt, yt, tw, th int
	}{
		{"test/target.png", "test/tiles", "test/good1.png", 4, 3, 4, 3},
		{"test/target.png", "test/tiles", "test/good2.png", 4, 3, 1, 1},
	}
	for _, test := range tests {
		m := NewMosaic(test.target, test.xt, test.yt, test.tiles, test.tw, test.th)
		if err := m.Render(); err != nil {
			log.Fatal("cant render mosaic: " + err.Error())
		}
		g, err := loadImage(test.good)
		if err != nil {
			log.Fatal("cant load test image: " + err.Error())
		}
		if !imageEqual(m.Image(), g) {
			t.Fail()
		}
	}
}
