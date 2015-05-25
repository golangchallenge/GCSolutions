package main

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"os"
	"path/filepath"
	"sync"
)

type TiledImage struct {
	sp        image.Point
	tilecolor color.Color
	L         float64
	A         float64
	B         float64
}

var TilesMatrices map[string][]TiledImage

type FlickrImage struct {
	ImageName string
	Img       image.Image
}

type ColorDistance struct {
	Idx         int
	LabDistance float64
}

/*REQUIRED FOR SORTING CUSTOM TYPES.....*/
type ByColorDistance []*ColorDistance

func (a ByColorDistance) Len() int           { return len(a) }
func (a ByColorDistance) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByColorDistance) Less(i, j int) bool { return a[i].LabDistance < a[j].LabDistance }

func init() {
	TilesMatrices = make(map[string][]TiledImage)
	//SearchResult = make(map[string]TiledImage)
}

func (f FlickrImage) ConvertToMatrixs(wg *sync.WaitGroup, m *MossaicGenerator) {
	defer func() {
		wg.Done()
	}()

	_Image := f.Img
	TiledImages := []TiledImage{}

	bound := _Image.Bounds()
	prevX, prevY := m.OFFSET, 0
	for y := bound.Min.Y; y < bound.Max.Y+m.OFFSET; y = y + m.OFFSET {
		if y == 0 {
			continue
		}
		prevX = 0
		for x := bound.Min.X; x < bound.Max.X+m.OFFSET; x = x + m.OFFSET {
			if x == 0 {
				continue
			}
			X1 := prevX
			Y1 := prevY
			X2 := x
			Y2 := y
			prevX = x

			midX := (X2-X1)/2 + X1
			midY := (Y2-Y1)/2 + Y1
			_color := _Image.At(midX, midY)
			L, a, b := RgbToLab(_color)

			_TiledImage := TiledImage{image.Point{X1, Y1}, _color, L, a, b}
			TiledImages = append(TiledImages, _TiledImage)
		}
		prevY = y
	}

	TilesMatrices[f.ImageName] = TiledImages
}

func (f FlickrImage) SearchInMatrices(searchColor color.Color, wg *sync.WaitGroup, m *MossaicGenerator) {
	defer func() {
		wg.Done()
	}()

	closestIdx := -1
	TileImages := TilesMatrices[f.ImageName]
	l1, a1, b1 := RgbToLab(searchColor)

	colorDistance := make([]*ColorDistance, 0)
	for idx, tileImage := range TileImages {
		l2, a2, b2 := tileImage.L, tileImage.A, tileImage.B
		res := math.Sqrt((l1-l2)*(l1-l2) + (a1-a2)*(a1-a2) + (b1-b2)*(b1-b2))
		if res <= LABDISTANCE {
			closestIdx = idx
			break
		}
		colorDistance = append(colorDistance, &ColorDistance{idx, res})
	}
	if closestIdx > -1 {
		m.SearchResult[f.ImageName] = TileImages[closestIdx]
	} else {
		cIdx := -1
		cDistance := math.MaxFloat64
		//Iterating thro' searched result is much better than sorting and find [0]
		for _, c := range colorDistance {
			if c.LabDistance < cDistance {
				cIdx = c.Idx
				cDistance = c.LabDistance
			}
		}
		if cIdx > -1 {
			m.SearchResult[f.ImageName] = TileImages[cIdx]
		}
	}
}

func (m MossaicGenerator) SearchColorInMatrics(searchColor color.Color) (img image.Image, tiledImage TiledImage, found bool) {

	//To Initialize SearchResult
	m.SearchResult = make(map[string]TiledImage)

	var wg sync.WaitGroup
	NosofGoRoutines := len(m.FlickrImageCollection)
	wg.Add(NosofGoRoutines)

	for _, i := range m.FlickrImageCollection {
		go i.SearchInMatrices(searchColor, &wg, &m)
	}
	wg.Wait()

	/* WE HAVE ALL CLOSEST SEARCH INCLUDING 0.1 LAB DISTANCE TRY TO FIND MIN OF ALL.
	FIND MIN DISTANCE OF ALL SEARCHES.. EFFECTIVE WHEN EXACT
	COLORED TILED IS NOT FOUND	*/

	closestImageName := ""
	closestLabDistance := math.MaxFloat64
	l1, a1, b1 := RgbToLab(searchColor)
	for imageName, tileImage := range m.SearchResult {
		l2, a2, b2 := tileImage.L, tileImage.A, tileImage.B

		res := math.Sqrt((l1-l2)*(l1-l2) + (a1-a2)*(a1-a2) + (b1-b2)*(b1-b2))

		if res < closestLabDistance {
			closestLabDistance = res
			closestImageName = imageName
		}
	}
	if closestImageName != "" {
		return m.GetImageFromCollection(closestImageName), m.SearchResult[closestImageName], true
	}
	return nil, TiledImage{}, false
}

func (m MossaicGenerator) GetImageFromCollection(name string) image.Image {
	for _, i := range m.FlickrImageCollection {
		if i.ImageName == name {
			return i.Img
		}
	}
	return nil
}

func linearize(v float64) float64 {
	if v <= 0.04045 {
		return v / 12.92
	}
	return math.Pow((v+0.055)/1.055, 2.4)
}

func LinearRgbToXyz(r, g, b float64) (x, y, z float64) {
	x = 0.4124564*r + 0.3575761*g + 0.1804375*b
	y = 0.2126729*r + 0.7151522*g + 0.0721750*b
	z = 0.0193339*r + 0.1191920*g + 0.9503041*b
	return
}

// This is the default reference white point.
var D65 = [3]float64{0.95047, 1.00000, 1.08883}

func lab_f(t float64) float64 {
	if t > 6.0/29.0*6.0/29.0*6.0/29.0 {
		return math.Cbrt(t)
	}
	return t/3.0*29.0/6.0*29.0/6.0 + 4.0/29.0
}

func xyzToLab(x, y, z float64, wref [3]float64) (l, a, b float64) {
	fy := lab_f(y / wref[1])
	l = 1.16*fy - 0.16
	a = 5.0 * (lab_f(x/wref[0]) - fy)
	b = 2.0 * (fy - lab_f(z/wref[2]))
	return
}

func rgbToXyx(S color.Color) (float64, float64, float64) {
	r, g, b, _ := S.RGBA()
	r1, g1, b1 := float64(r)/65535.0, float64(g)/65535.0, float64(b)/65535.0
	r_ := linearize(r1)
	g_ := linearize(g1)
	b_ := linearize(b1)
	return LinearRgbToXyz(r_, g_, b_)
}

func RgbToLab(S color.Color) (float64, float64, float64) {
	x, y, z := rgbToXyx(S)
	return xyzToLab(x, y, z, D65)
}

func (m *MossaicGenerator) TestImageColl() {
	m.FillCollection()
	var wg sync.WaitGroup
	NosofGoRoutines := len(m.FlickrImageCollection)
	wg.Add(NosofGoRoutines)

	for _, i := range m.FlickrImageCollection {
		go i.ConvertToMatrixs(&wg, m)
	}
	wg.Wait()
}

func (m MossaicGenerator) GetTileImageNames(ImgDirName string) []string {
	file, err := os.Open(ImgDirName)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	var fi os.FileInfo
	fi, err = file.Stat()
	if err != nil && !fi.IsDir() {
		panic(err)
	}
	var names []string
	names, err = file.Readdirnames(-1)
	if err != nil {
		panic(err)
	}
	return names
}

func (m *MossaicGenerator) FillCollection() {
	// Walk through images subfolder and populate
	names := m.GetTileImageNames(m.ImageDirectory())
	/*
		if len(names) <= 2 {
			m.LoadTiledImagesFromDefaultDir()
			names = m.GetTileImageNames(m.ImageDirectory())
		}

	*/

	var f FlickrImage
	for _, k := range names {
		f = FlickrImage{k, ReadImage(filepath.Join(m.ImageDirectory(), k))}
		m.FlickrImageCollection = append(m.FlickrImageCollection, f)
	}
}

func ReadImage(fpath string) image.Image {
	reader, err := os.Open(fpath)
	if err != nil {
		fmt.Println(err.Error())
	}
	defer reader.Close()

	img, _, err := image.Decode(reader)
	if err != nil {
		fmt.Println(err.Error())
	}
	return img
}
