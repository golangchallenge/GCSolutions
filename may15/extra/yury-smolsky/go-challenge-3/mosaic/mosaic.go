package mosaic

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"
	"runtime"
)

const ignoreDelta int = 5
const maxTileSize int = 30
const targetNumTiles int = 150
const allowedCrop float64 = 0.02

type avgColorImage struct {
	img   image.Image
	color color.RGBA
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// RequiredImages returns minimum suggested amount of tile images
func RequiredImages() int {
	return ignoreDelta * ignoreDelta * ignoreDelta
}

// Returns averaged color for a rectangular region (x0, y0)-(x1, y1) on img
func averageColorBound(img image.Image, x0, y0, x1, y1 int) color.RGBA {
	var r, g, b uint64
	for x := x0; x < x1; x++ {
		for y := y0; y < y1; y++ {
			dr, dg, db, _ := img.At(x, y).RGBA()
			r += uint64(dr)
			g += uint64(dg)
			b += uint64(db)
		}
	}
	pixels := (x1 - x0) * (y1 - y0)
	r = r / uint64(pixels)
	g = g / uint64(pixels)
	b = b / uint64(pixels)
	return color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), 255}
}

// Returns average color for image m
func averageColor(m image.Image) color.RGBA {
	bounds := m.Bounds()
	x0, x1 := bounds.Min.X, bounds.Max.X
	y0, y1 := bounds.Min.Y, bounds.Max.Y
	return averageColorBound(m, x0, y0, x1, y1)
}

// Converts m image into RGBA image
func intoRGBA(m image.Image) image.Image {
	b := m.Bounds()
	m2 := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(m2, b, m, b.Min, draw.Src)
	return m2
}

// Returns the minimum side of all images imgs
func minDimension(imgs []image.Image) int {
	a := math.MaxInt32
	for _, img := range imgs {
		a = min(a, img.Bounds().Dx())
		a = min(a, img.Bounds().Dy())
	}
	return a
}

// Crops image to resolution cx, cy from center
func cropCenter(src image.Image, cx int, cy int) image.Image {
	b := src.Bounds()
	deltaX := int((b.Dx() - cx) / 2)
	deltaY := int((b.Dy() - cy) / 2)
	sp := image.Pt(deltaX, deltaY)
	r := image.Rect(0, 0, cx, cy)
	m := image.NewRGBA(r)
	draw.Draw(m, r, src, sp, draw.Src)
	return m
}

// Crops image into square image
func cropIntoSquare(src image.Image) image.Image {
	side := min(src.Bounds().Dx(), src.Bounds().Dy())
	return cropCenter(src, side, side)
}

// Resizes square image into new size newSide
func resizeSquareImage(src image.Image, newSide int) image.Image {
	b := src.Bounds()
	ratio := (uint32(b.Dx())<<16)/uint32(newSide) + 1

	r := image.Rect(0, 0, newSide, newSide)
	m := image.NewRGBA(r)
	for x := 0; x < newSide; x++ {
		for y := 0; y < newSide; y++ {
			x2 := int((ratio * uint32(x)) >> 16)
			y2 := int((ratio * uint32(y)) >> 16)
			m.Set(x, y, src.At(x2, y2))
		}
	}
	return m
}

type imageJob struct {
	img  image.Image
	size int
}

// Crops images into square image and resizes to size.
// Also calculates average color packs everything in averageColor struct
func prepareLibImage(jobs <-chan imageJob, results chan<- avgColorImage) {
	for j := range jobs {
		img := j.img
		img = cropIntoSquare(img)
		img = resizeSquareImage(img, j.size)
		results <- avgColorImage{img, averageColor(img)}
	}
	return
}

// Processes all the images into square images of a size along with average colors
// using multiple CPUs.
func prepareLibImages(imgs []image.Image, size int) (avgColorImages []avgColorImage) {
	jobs := make(chan imageJob, 20)
	results := make(chan avgColorImage, 20)
	for i := 0; i < runtime.NumCPU(); i++ {
		go prepareLibImage(jobs, results)
	}
	go func() {
		for _, img := range imgs {
			jobs <- imageJob{img, size}
		}
		close(jobs)
	}()
	for i := 0; i < len(imgs); i++ {
		r := <-results
		avgColorImages = append(avgColorImages, r)
	}
	return
}

// PixelSize returns the size of subpixel and
// new Width and Height for the target image if the cropping will take an effect.
// It seeks after smaller side to be about targetNumTiles or less.
func pixelSize(w, h int) (size, w2, h2 int) {
	w2, h2 = w, h
	if min(w, h) <= targetNumTiles {
		size = 1
		return
	}
	// 5% is allowed crop
	eps := int(math.Ceil(float64(min(w, h)) * allowedCrop))
	start := min(w, h) / targetNumTiles
	for size = start; size < max(w, h); size++ {
		if w%size == 0 && h%size == 0 {
			return
		}
		if w%size == 0 {
			if h%size <= eps {
				h2 = h - h%size
				return
			}
		}
		if h%size == 0 {
			if w%size <= eps {
				w2 = w - w%size
				return
			}
		}
		if w%size <= eps && h%size <= eps {
			h2 = h - h%size
			w2 = w - w%size
			return
		}
	}
	size = 1
	return
}

// pixelate processes image m into array of colors. Each element of array is averaged colors of
// that particular subarea of image.
func pixelate(m image.Image) [][]color.RGBA {
	size, dx2, dy2 := pixelSize(m.Bounds().Dx(), m.Bounds().Dy())
	dx := dx2 / size
	dy := dy2 / size
	fmt.Println("pix size:", size, "frame:", dx, dy)
	frame := make([][]color.RGBA, dx)
	for i := range frame {
		frame[i] = make([]color.RGBA, dy)
	}
	r := image.Rect(0, 0, dx, dy)
	m2 := image.NewRGBA(r)
	for x := 0; x < dx; x++ {
		for y := 0; y < dy; y++ {
			x0 := m.Bounds().Min.X + x*size
			y0 := m.Bounds().Min.Y + y*size
			frame[x][y] = averageColorBound(m, x0, y0, x0+size, y0+size)
			m2.Set(x, y, frame[x][y])
		}
	}
	return frame
}

// Returns the distance between two RGB colors
func colorDistance(c, v color.RGBA) float64 {
	return math.Sqrt(
		math.Pow(float64(c.R)-float64(v.R), 2) +
			math.Pow(float64(c.G)-float64(v.G), 2) +
			math.Pow(float64(c.B)-float64(v.B), 2))
}

// Chooses the closest match to color c from imgs slice, ignoring colors with indexes found in ignore map.
// Returns image and it's index in imgs slice.
func closestImage(c color.RGBA, imgs []avgColorImage, ignore map[int]bool) (image.Image, int) {
	minImage := 0
	minDistance := colorDistance(c, imgs[0].color)
	for i, m := range imgs {
		if ignore[i] {
			continue
		}
		d := colorDistance(c, m.color)
		if d < minDistance {
			minImage = i
			minDistance = d
		}
	}
	return imgs[minImage].img, minImage
}

// Compose makes an image from 2d slices of colors by finding color-closest images from imgs slice.
// It used neighbour image check when picking a new tile.
// Result image will be of a (x*tileSize, y*tileSize) resolution
func compose(target [][]color.RGBA, imgs []avgColorImage) image.Image {
	dx := len(target)
	dy := len(target[0])
	tileSize := imgs[0].img.Bounds().Dx()
	r := image.Rect(0, 0, dx*tileSize, dy*tileSize)
	m := image.NewRGBA(r)
	idxs := make([][]int, dx)
	for i := range idxs {
		idxs[i] = make([]int, dy)
	}
	used := make(map[int]bool)
	for x := 0; x < dx; x++ {
		for y := 0; y < dy; y++ {
			ignore := make(map[int]bool)
			for i := max(0, x-ignoreDelta); i <= min(dx-1, x+ignoreDelta); i++ {
				for j := max(0, y-ignoreDelta); j <= min(dy-1, y+ignoreDelta); j++ {
					if i == x && j == y {
						continue
					}
					ignore[idxs[i][j]] = true
				}
			}
			img, imgIdx := closestImage(target[x][y], imgs, ignore)
			idxs[x][y] = imgIdx
			tr := image.Rect(x*tileSize, y*tileSize, (x+1)*tileSize, (y+1)*tileSize)
			draw.Draw(m, tr, img, img.Bounds().Min, draw.Src)
			used[imgIdx] = true
		}
	}
	fmt.Println(len(used), "images used")
	return m
}

// Create makes a mosaic image for the target image using imgs as tile images.
func Create(target image.Image, imgs []image.Image) image.Image {

	fmt.Print("pixelating... ")
	target = intoRGBA(target)
	pixels := pixelate(target)

	tileSize := minDimension(imgs)
	fitImgsInto := min(tileSize, maxTileSize)

	fmt.Println("resizing...")
	cimgs := prepareLibImages(imgs, fitImgsInto)

	fmt.Print("tiling... ")
	return compose(pixels, cimgs)
}
