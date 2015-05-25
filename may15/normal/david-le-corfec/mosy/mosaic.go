package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"math"
	"os"
	"path"
	"strings"
)

//
// MOSAIC
//

// Mosaic is the type allowing the generation of a mosaic from a target image
// and a set of tiles.
type Mosaic struct {
	target *target
	tiles  *tileManager
	out    image.Image
}

// NewMosaic initialize a Mosaic given:
// - a target image specified by its path and its desired tile dimensions
// - tiles specified by their directory path and their downscaled dimensions
func NewMosaic(target string, xtiles, ytiles int,
	tiledir string, twidth, theight int) *Mosaic {
	return &Mosaic{target: newTarget(target, xtiles, ytiles), tiles: newTileManager(tiledir, twidth, theight)}
}

// Render loads/preprocess the target and the tiles then generate the mosaic
// The resulting image will have the resolution W * H pixels where
// - W is the number of horizontal tiles times their downscaled horizontal resolution
// - H is the number of vertical times times their downscaled vertical resolution
func (m *Mosaic) Render() error {
	if err := m.target.Load(); err != nil {
		return err
	}
	if err := m.tiles.Load(); err != nil {
		return err
	}
	m.out = m.generate()
	return nil
}

// Save the mosaic to a file, using PNG if path ends in .png, else JPEG.
func (m *Mosaic) Save(path string) error {
	if m.out == nil {
		return fmt.Errorf("image not rendered")
	}
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("cant save picture: " + err.Error())
	}
	defer f.Close()

	if strings.HasSuffix(path, ".png") {
		return png.Encode(f, m.out)
	}
	return jpeg.Encode(f, m.out, nil)
}

// generate the mosaic
func (m *Mosaic) generate() image.Image {
	xt := m.target.xTiles
	yt := m.target.yTiles
	tw := m.tiles.width
	th := m.tiles.height
	dst := image.NewRGBA(image.Rect(0, 0, xt*tw, yt*th))

	target := m.target.img
	// iterate over the downscaled target image
	for y := target.Bounds().Min.Y; y < target.Bounds().Max.Y; y++ {
		for x := target.Bounds().Min.X; x < target.Bounds().Max.X; x++ {
			// select the tile with the closest average color then copy it to output
			c := target.At(x, y)
			t := m.tiles.closest(c)
			draw.Draw(dst, image.Rect(x*tw, y*th, (x+1)*tw, (y+1)*th), t.img, image.Pt(0, 0), draw.Src)
		}
	}
	return dst
}

// Image returns the rendered image (thus only valid after a call to Render)
func (m *Mosaic) Image() image.Image {
	return m.out
}

//
// TARGET
//

// Target is the image the mosaic will look like.
type target struct {
	path           string
	img            image.Image
	xTiles, yTiles int
}

// newTarget returns a new Target object, without loading the underlying image.
func newTarget(path string, xt, yt int) *target {
	return &target{path: path, xTiles: xt, yTiles: yt}
}

// Load will load the image and downscale it to xTiles*yTiles pixels.
func (t *target) Load() error {
	ti, err := loadImage(t.path)
	if err != nil {
		return fmt.Errorf("cant load target picture: " + err.Error())
	}

	t.img = downscaleImage(ti, t.xTiles, t.yTiles)
	return nil
}

//
// TILE
//

// tile is a downscaled image used in constructing a mosaic
type tile struct {
	path          string
	img           image.Image
	avg           color.Color
	width, height int
}

func newTile(path string, w, h int) *tile {
	return &tile{path: path, width: w, height: h}
}

// Load decodes an image from a file, downscales it, and computes
// its average color
func (t *tile) Load() error {
	img, err := loadImage(t.path)
	if err != nil {
		return err
	}
	t.img = downscaleImage(img, t.width, t.height)
	t.avg = averager(t.img, t.img.Bounds())
	return nil
}

//
// TILE MANAGER
//

// tileManager handles a collection of tiles.
type tileManager struct {
	path          string
	tiles         []tile
	width, height int
}

func newTileManager(path string, width, height int) *tileManager {
	return &tileManager{path: path, width: width, height: height}
}

// Load will load, downscale and compute the average color of each tile image
// located in the tile directory.
func (tm *tileManager) Load() error {
	dir := tm.path
	d, err := os.Open(dir)
	if err != nil {
		return fmt.Errorf("cant open tiledir: " + err.Error())
	}
	defer d.Close()
	for {
		names, err := d.Readdirnames(5)
		if err != nil && err != io.EOF {
			return fmt.Errorf("error reading tiledir: " + err.Error())
		}
		for _, name := range names {
			if !strings.HasSuffix(name, ".jpeg") && !strings.HasSuffix(name, ".jpg") && !strings.HasSuffix(name, ".png") {
				log.Printf("file %s is not a supported image type, ignoring", name)
				continue
			}
			tile := newTile(path.Join(dir, name), tm.width, tm.height)
			if err := tile.Load(); err != nil {
				log.Printf("error loading tile %s: %s", name, err.Error())
				continue
			}
			tm.tiles = append(tm.tiles, *tile)
		}
		if err == io.EOF {
			break
		}
	}
	return nil
}

// closest returns the tile best matching the given color amongst a set of tiles,
// currently using euclidean distance
func (tm *tileManager) closest(target color.Color) tile {
	var dmin float64 = 0xffffffffff
	var closest tile

	for _, t := range tm.tiles {
		d := distance(t.avg, target)
		if d < dmin {
			closest = t
			dmin = d
		}
	}
	//log.Printf("closest for %v is tile %v\n", target, closest.avg)
	return closest
}

// distance computes the euclidean distance between 2 colors
func distance(c1, c2 color.Color) float64 {
	r1, b1, g1, _ := c1.RGBA()
	r2, b2, g2, _ := c2.RGBA()
	dr := int(r1) - int(r2)
	dg := int(g1) - int(g2)
	db := int(b1) - int(b2)
	return math.Sqrt(float64(dr*dr) + float64(dg*dg) + float64(db*db))
}

//
// IMAGE UTILITIES
//

// loadImage decodes an image from the given file
func loadImage(path string) (image.Image, error) {
	f, err := os.Open(path)
	defer f.Close()
	if err != nil {
		return nil, err
	}
	img, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}
	return img, nil
}

// averager computes the average image color within the given bounds.
func averager(src image.Image, r image.Rectangle) color.Color {
	npix := (r.Max.Y - r.Min.Y) * (r.Max.X - r.Min.X)
	var sum [3]uint64
	for j := r.Min.Y; j < r.Max.Y; j++ {
		for i := r.Min.X; i < r.Max.X; i++ {
			r, g, b, _ := src.At(i, j).RGBA()
			sum[0] += uint64(r >> 8)
			sum[1] += uint64(g >> 8)
			sum[2] += uint64(b >> 8)
		}
	}
	var avg [3]uint64
	avg[0] = sum[0] / uint64(npix)
	avg[1] = sum[1] / uint64(npix)
	avg[2] = sum[2] / uint64(npix)

	return color.RGBA{R: uint8(avg[0]), G: uint8(avg[1]), B: uint8(avg[2]), A: 0xff}
}

// downscaleImage generate a downscaled image from the given using an average
// interpolation. w and h are the desired output dimensions.
func downscaleImage(src image.Image, w, h int) image.Image {
	dst := image.NewRGBA(image.Rect(0, 0, w, h))

	srcwidth := src.Bounds().Dx()
	srcheight := src.Bounds().Dy()
	for y := dst.Bounds().Min.Y; y < dst.Bounds().Max.Y; y++ {
		for x := dst.Bounds().Min.X; x < dst.Bounds().Max.X; x++ {
			//c := src.At(x*srcwidth/w, y*srcheight/h) // nearest-neighbour interpolation
			c := avgInterpolateAt(src, x*srcwidth/w, y*srcheight/h, srcwidth/w, srcheight/h)
			dst.Set(x, y, c)
		}
	}
	return dst
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

// avgInterpolate gets the pixel color at x,y obtained by averaging
// the pixels in a kw*kh rectangle around the given pixel.
// This rectangle may be reduced if it crosses image boundaries.
func avgInterpolateAt(src image.Image, x, y, kw, kh int) color.Color {
	ymin := max(y-kh/2, src.Bounds().Min.Y)
	ymax := min(y+kh/2+1, src.Bounds().Max.Y)
	xmin := max(x-kw/2, src.Bounds().Min.X)
	xmax := min(x+kw/2+1, src.Bounds().Max.X)

	return averager(src, image.Rect(xmin, ymin, xmax, ymax))
}
