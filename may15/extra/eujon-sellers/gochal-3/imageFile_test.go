package main

import (
	"encoding/base64"
	"image"
	"image/color"
	"image/draw"
	_ "image/gif"
	"image/jpeg"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"strings"
	"testing"
)

func TestCreateTiles(t *testing.T) {
	reader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(data))
	m, _, err := image.Decode(reader)
	if err != nil {
		t.Errorf("Error loading data")
	}
	out, err := os.Create("createTiles-test.jpg")
	if err != nil {
		t.Errorf("Error creating test jpg")
	}
	err = jpeg.Encode(out, m, nil)
	if err != nil {
		t.Errorf("Error writing test jpg")
	}
	testPic := ImageFile{Name: "createTiles-test.jpg"}
	testPic.createTiles()
	if testPic.YEnd != 103 {
		t.Errorf("YEnd is %v, should be 103", testPic.YEnd)
	}
	if testPic.XEnd != 150 {
		t.Errorf("XEnd is %v, should be 150", testPic.XEnd)
	}
	if len(testPic.Tiles) != 165 {
		t.Errorf("Tile count is %v, should be 165", len(testPic.Tiles))
	}
	os.Remove("createTiles-test.jpg")
}

func BenchmarkCreateTiles(b *testing.B) {
	reader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(data))
	m, _, err := image.Decode(reader)
	if err != nil {
		b.Errorf("Error loading data")
	}
	out, err := os.Create("createTiles-test.jpg")
	if err != nil {
		b.Errorf("Error creating test jpg")
	}
	err = jpeg.Encode(out, m, nil)
	if err != nil {
		b.Errorf("Error writing test jpg")
	}
	testPic := ImageFile{Name: "createTiles-test.jpg"}

	for i := 0; i < b.N; i++ {
		testPic.createTiles()
	}
	os.Remove("createTiles-test.jpg")
}

func TestBuildColorMap(t *testing.T) {
	red := color.NRGBA{255, 0, 0, 255}
	green := color.NRGBA{0, 255, 0, 255}
	blue := color.NRGBA{0, 0, 255, 255}
	white := color.NRGBA{0, 0, 0, 255}

	redSquare := image.NewNRGBA(image.Rect(0, 0, 20, 20))
	draw.Draw(redSquare, redSquare.Bounds(), &image.Uniform{red}, image.ZP, draw.Src)
	greenSquare := image.NewNRGBA(image.Rect(0, 0, 20, 20))
	draw.Draw(greenSquare, greenSquare.Bounds(), &image.Uniform{green}, image.ZP, draw.Src)
	blueSquare := image.NewNRGBA(image.Rect(0, 0, 20, 20))
	draw.Draw(blueSquare, blueSquare.Bounds(), &image.Uniform{blue}, image.ZP, draw.Src)
	whiteSquare := image.NewNRGBA(image.Rect(0, 0, 20, 20))
	draw.Draw(whiteSquare, whiteSquare.Bounds(), &image.Uniform{white}, image.ZP, draw.Src)

	testImage := image.NewNRGBA(image.Rect(0, 0, 40, 40))
	point := image.Point{}
	rect := redSquare.Bounds()
	r := image.Rectangle{point, point.Add(rect.Size())}
	draw.Draw(testImage, r, redSquare, rect.Min, draw.Src)

	point = image.Point{Y: 0, X: 20}
	rect = greenSquare.Bounds()
	r = image.Rectangle{point, point.Add(rect.Size())}
	draw.Draw(testImage, r, greenSquare, rect.Min, draw.Src)

	point = image.Point{Y: 20, X: 0}
	rect = blueSquare.Bounds()
	r = image.Rectangle{point, point.Add(rect.Size())}
	draw.Draw(testImage, r, blueSquare, rect.Min, draw.Src)

	point = image.Point{Y: 20, X: 20}
	rect = whiteSquare.Bounds()
	r = image.Rectangle{point, point.Add(rect.Size())}
	draw.Draw(testImage, r, whiteSquare, rect.Min, draw.Src)

	testPic := ImageFile{Name: "createTiles-test.jpg", Image: testImage}

	testTile := OrigTile{Filename: "createTiles-test.jpg", YStart: 0, YEnd: 19, XStart: 0, XEnd: 19}
	testTile.buildColorMap(testPic)
	if testTile.AvgRGB == nil {
		t.Errorf("NRGBA AvgRGB is nil")
	}
	if testTile.AvgRGB["red"] != 65535 || testTile.AvgRGB["green"] != 0 || testTile.AvgRGB["blue"] != 0 || testTile.AvgRGB["alpha"] != 65535 {
		t.Errorf("NRGBA red is %v, should be 25700", testTile.AvgRGB["red"])
		t.Errorf("NRGBA green is %v, should be 38550", testTile.AvgRGB["green"])
		t.Errorf("NRGBA blue is %v, should be 65535", testTile.AvgRGB["blue"])
		t.Errorf("NRGBA alpha is %v, should be 65535", testTile.AvgRGB["alpha"])
	}

	testTile = OrigTile{Filename: "createTiles-test.jpg", YStart: 0, YEnd: 19, XStart: 20, XEnd: 39}
	testTile.buildColorMap(testPic)
	if testTile.AvgRGB == nil {
		t.Errorf("NRGBA AvgRGB is nil")
	}
	if testTile.AvgRGB["red"] != 0 || testTile.AvgRGB["green"] != 65535 || testTile.AvgRGB["blue"] != 0 || testTile.AvgRGB["alpha"] != 65535 {
		t.Errorf("NRGBA red is %v, should be 25700", testTile.AvgRGB["red"])
		t.Errorf("NRGBA green is %v, should be 38550", testTile.AvgRGB["green"])
		t.Errorf("NRGBA blue is %v, should be 65535", testTile.AvgRGB["blue"])
		t.Errorf("NRGBA alpha is %v, should be 65535", testTile.AvgRGB["alpha"])
	}

	testTile = OrigTile{Filename: "createTiles-test.jpg", YStart: 19, YEnd: 39, XStart: 0, XEnd: 39}
	testTile.buildColorMap(testPic)
	if testTile.AvgRGB == nil {
		t.Errorf("NRGBA AvgRGB is nil")
	}
	if testTile.AvgRGB["red"] != 1680 || testTile.AvgRGB["green"] != 1596 || testTile.AvgRGB["blue"] != 31927 || testTile.AvgRGB["alpha"] != 65535 {
		t.Errorf("NRGBA red is %v, should be 25700", testTile.AvgRGB["red"])
		t.Errorf("NRGBA green is %v, should be 38550", testTile.AvgRGB["green"])
		t.Errorf("NRGBA blue is %v, should be 65535", testTile.AvgRGB["blue"])
		t.Errorf("NRGBA alpha is %v, should be 65535", testTile.AvgRGB["alpha"])
	}

	// simple test for checking our switch case...
	//image.YCbCr
	//m2 := image.NewYCbCr(image.Rect(0, 0, 200, 200), image.YCbCrSubsampleRatio444)
	m2 := image.NewRGBA(image.Rect(0, 0, 200, 200))
	draw.Draw(m2, m2.Bounds(), &image.Uniform{red}, image.ZP, draw.Src)
	testPic.Image = m2
	testTile.buildColorMap(testPic)
	if testTile.AvgRGB == nil {
		t.Errorf("RGBA AvgRGB is nil")
	}
	if testTile.AvgRGB["red"] != 65535 || testTile.AvgRGB["green"] != 0 || testTile.AvgRGB["blue"] != 0 || testTile.AvgRGB["alpha"] != 65535 {
		t.Errorf("RGBA red is %v, should be 32296", testTile.AvgRGB["red"])
		t.Errorf("RGBA green is %v, should be 4683", testTile.AvgRGB["green"])
		t.Errorf("RGBA blue is %v, should be 691", testTile.AvgRGB["blue"])
		t.Errorf("RGBA alpha is %v, should be 65535", testTile.AvgRGB["alpha"])
	}

	m3 := image.NewRGBA64(image.Rect(0, 0, 200, 200))
	draw.Draw(m3, m3.Bounds(), &image.Uniform{red}, image.ZP, draw.Src)
	testPic.Image = m3
	testTile.buildColorMap(testPic)
	if testTile.AvgRGB == nil {
		t.Errorf("RGBA64 AvgRGB is nil")
	}
	if testTile.AvgRGB["red"] != 65535 || testTile.AvgRGB["green"] != 0 || testTile.AvgRGB["blue"] != 0 || testTile.AvgRGB["alpha"] != 65535 {
		t.Errorf("RGBA64 red is %v, should be 32296", testTile.AvgRGB["red"])
		t.Errorf("RGBA64 green is %v, should be 4683", testTile.AvgRGB["green"])
		t.Errorf("RGBA64 blue is %v, should be 691", testTile.AvgRGB["blue"])
		t.Errorf("RGBA64 alpha is %v, should be 65535", testTile.AvgRGB["alpha"])
	}

	m4 := image.NewNRGBA64(image.Rect(0, 0, 200, 200))
	draw.Draw(m4, m4.Bounds(), &image.Uniform{red}, image.ZP, draw.Src)
	testPic.Image = m4
	testTile.buildColorMap(testPic)
	if testTile.AvgRGB == nil {
		t.Errorf("NRGBA64 AvgRGB is nil")
	}
	if testTile.AvgRGB["red"] != 65535 || testTile.AvgRGB["green"] != 0 || testTile.AvgRGB["blue"] != 0 || testTile.AvgRGB["alpha"] != 65535 {
		t.Errorf("NRGBA64 red is %v, should be 32296", testTile.AvgRGB["red"])
		t.Errorf("NRGBA64 green is %v, should be 4683", testTile.AvgRGB["green"])
		t.Errorf("NRGBA64 blue is %v, should be 691", testTile.AvgRGB["blue"])
		t.Errorf("NRGBA64 alpha is %v, should be 65535", testTile.AvgRGB["alpha"])
	}
}
