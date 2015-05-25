package main

//import "fmt"
import "image"
import "image/color"
import "testing"

func TestAverageColor(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		// 0x80 = 0.5 * 0xFF.
		img.SetRGBA(0, y, color.RGBA{0xFF, 0xFF, 0x00, 0x00})
	}

	out := averageColor(img)

	if out.colorR != 32767 && out.colorG != 32767 && out.colorB != 0 {
		t.Fail()
	}
}

func TestColorDistance(t *testing.T) {
	dist := measureColorDist(20000, 20000, 20000, 40000, 40000, 40000)

	if int(dist) != 34641 {
		t.Fail()
	}
}
