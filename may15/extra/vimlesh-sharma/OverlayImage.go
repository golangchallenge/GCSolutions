package main

import (
	"image"
	"image/draw"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"log"
	"os"
	"time"
)

//const OFFSET = 20
const LABDISTANCE = 0.1

type MossaicGenerator struct {
	PUID                  string
	ImgName               string
	FlickrImageCollection []FlickrImage
	SearchResult          map[string]TiledImage
	OFFSET                int
	DefaultImageDir       bool
}

func NewMossaicGenerator(PUID string, ImgName string, OFFSET int) *MossaicGenerator {
	m := new(MossaicGenerator)
	m.PUID = PUID
	m.ImgName = ImgName
	m.OFFSET = OFFSET
	m.SearchResult = make(map[string]TiledImage)
	m.DefaultImageDir = false
	return m
}

func (m *MossaicGenerator) LoadTiledImagesFromDefaultDir() {
	m.DefaultImageDir = true
}

func (m MossaicGenerator) ImageName() string {
	return "./static/" + m.PUID + "/" + m.ImgName
}

//Only for Tiled Images..
func (m MossaicGenerator) ImageDirectory() string {
	/*if m.DefaultImageDir {
		return "SampleImages"
	}*/
	return "./static/" + m.PUID
}

func (m MossaicGenerator) MossaicImageName() string {
	return "./static/" + m.PUID + "/Mossaic" + m.ImgName
}

func (m MossaicGenerator) GenerateMossaic() {
	m.TestImageColl()

	start := time.Now()
	reader, err := os.Open(m.ImageName())

	if err != nil {
		panic(err)
	}
	defer reader.Close()

	pencil, _, err1 := image.Decode(reader)
	if err1 != nil {
		log.Fatal(err1)
	}
	boundsp := pencil.Bounds()

	newImage := image.NewRGBA(image.Rect(0, 0, boundsp.Dx(), boundsp.Dy()))
	draw.Draw(newImage, newImage.Bounds(), pencil, boundsp.Min, draw.Src)

	prevX, prevY := m.OFFSET, 0
	for y := boundsp.Min.Y; y < boundsp.Max.Y+m.OFFSET; y = y + m.OFFSET {
		if y == 0 {
			continue
		}

		prevX = 0
		for x := boundsp.Min.X; x < boundsp.Max.X+m.OFFSET; x = x + m.OFFSET {
			if x == 0 {
				continue
			}
			X1, Y1, X2, Y2 := prevX, prevY, x, y
			midX := (X2-X1)/2 + X1
			midY := (Y2-Y1)/2 + Y1
			_color := pencil.At(midX, midY)
			_Image, tiledImage, found := m.SearchColorInMatrics(_color)
			if found {
				r := image.Rectangle{image.Point{X1, Y1}, image.Point{X2, Y2}}
				draw.Draw(newImage, r, _Image, tiledImage.sp, draw.Src)
			}
			prevX = x
		}
		prevY = y
	}
	file, err := os.Create(m.MossaicImageName())
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	result := newImage.SubImage(newImage.Bounds())
	jpeg.Encode(file, result, &jpeg.Options{Quality: 100})

	elapsed := time.Since(start)
	log.Printf("Picture Mossaic took %s\n", elapsed)
}
