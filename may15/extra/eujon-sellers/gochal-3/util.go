package main

import (
	"image"
	"image/draw"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"math/rand"
)

// create an average RGB value for an image, resize if asked
func avgRGB(img image.Image, resize bool) map[string]uint32 {
	toAvg := convertToNRGBA(img)
	if resize {
		toAvg = resizeNearest(toAvg, 40, 40)
	}
	bounds := toAvg.Bounds()

	r := uint32(0)
	g := uint32(0)
	b := uint32(0)
	a := uint32(0)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			red, grn, blu, alpha := toAvg.At(x, y).RGBA()
			r += red
			g += grn
			b += blu
			a += alpha
		}
	}
	r = r / (uint32(bounds.Max.Y) * uint32(bounds.Max.X))
	g = g / (uint32(bounds.Max.Y) * uint32(bounds.Max.X))
	b = b / (uint32(bounds.Max.Y) * uint32(bounds.Max.X))
	a = a / (uint32(bounds.Max.Y) * uint32(bounds.Max.X))
	var avgRGB = map[string]uint32{
		"red":   r,
		"green": g,
		"blue":  b,
		"alpha": a,
	}
	return avgRGB
}

// compare to avg RGB maps and return the diff/distance
func compareRGB(img1, img2 map[string]uint32) int {
	diff := math.Pow((float64(img1["red"])-float64(img2["red"])), 2.0) + math.Pow((float64(img1["green"])-float64(img2["green"])), 2.0) + math.Pow((float64(img1["blue"])-float64(img2["blue"])), 2.0)
	return int(diff)
}

// use the nearest neighbor algo to reszie an image
// http://tech-algorithm.com/articles/nearest-neighbor-image-scaling/
// https://github.com/disintegration/imaging/blob/master/resize.go
func resizeNearest(src *image.NRGBA, width, height int) *image.NRGBA {
	dstW, dstH := width, height

	srcBounds := src.Bounds()
	srcW := srcBounds.Max.X
	srcH := srcBounds.Max.Y

	dst := image.NewNRGBA(image.Rect(0, 0, dstW, dstH))

	dx := float64(srcW) / float64(dstW)
	dy := float64(srcH) / float64(dstH)

	for dstY := 0; dstY < dstH; dstY++ {
		fy := (float64(dstY)+0.5)*dy - 0.5

		for dstX := 0; dstX < dstW; dstX++ {
			fx := (float64(dstX)+0.5)*dx - 0.5

			srcX := int(math.Min(math.Max(math.Floor(fx+0.5), 0.0), float64(srcW)))
			srcY := int(math.Min(math.Max(math.Floor(fy+0.5), 0.0), float64(srcH)))

			srcOff := srcY*src.Stride + srcX*4
			dstOff := dstY*dst.Stride + dstX*4

			copy(dst.Pix[dstOff:dstOff+4], src.Pix[srcOff:srcOff+4])
		}
	}
	return dst
}

// convert known image types to NRGBA
func convertToNRGBA(src image.Image) *image.NRGBA {
	switch t := src.(type) {
	case *image.Alpha:
	case *image.Alpha16:
	case *image.Gray:
	case *image.Gray16:
	case *image.NRGBA:
	case *image.NRGBA64:
	case *image.Paletted:
	case *image.RGBA:
	case *image.RGBA64:
	case *image.YCbCr:
	default:
		logger.Printf("unknown image format: %v\n", t)
		return nil
	}
	b := src.Bounds()
	m := image.NewNRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(m, m.Bounds(), src, b.Min, draw.Src)
	return m
}

func contains(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func containsString(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func random(min, max int) int {
	return rand.Intn(max-min) + min
}
