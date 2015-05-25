package main

import (
	"flag"
	"fmt"
	"image"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"image/draw"
	"image/jpeg"
	"image/png"
)

var (
	cconly      = flag.Bool("cconly", true, "use only pictures under a CreativeCommons license")
	dir         = flag.String("dir", "", "Directory with photos to use as tiles")
	maxpics     = flag.Int("maxpics", 10, "maximum number of pictures to consider")
	output      = flag.String("output", "mosaic.jpg", "filename for JPG output")
	outTileSize = flag.Int("outtilesize", 75, "size of the tiles in the output image")
	photo       = flag.String("photo", "https://farm6.staticflickr.com/5043/5352841545_a854e3b2da_o_d.jpg", "photo to create mosaic for")
	query       = flag.String("query", "beach", "search term for tiles")
	tileSize    = flag.Int("tilesize", 16, "size of tile used for matching")
	workers     = 4
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.Parse()
	log.SetFlags(log.Ltime | log.Lmicroseconds)
}

func getCachePath(fn, tag string) (string, error) {
	uu, err := url.Parse(fn)
	if err != nil {
		return "", err
	}

	sep := string(os.PathSeparator)
	path := strings.Join([]string{"cache", tag, uu.Host, uu.Path}, sep)

	s := strings.Split(path, sep)
	os.MkdirAll(strings.Join(s[:len(s)-1], sep), os.ModeDir|os.ModePerm)

	return path, nil
}

func getCachedImage(u, tag string) (string, *image.RGBA, error) {
	fn, err := getCachePath(u, tag)
	if err != nil {
		return fn, nil, err
	}
	if _, err := os.Stat(fn); err != nil {
		return fn, nil, err
	}
	img, err := loadLocalImage(fn)
	return fn, img, err
}

func loadRGBA(r io.Reader) (*image.RGBA, error) {
	img, _, err := image.Decode(r)
	if err != nil {
		return nil, err
	}

	b := img.Bounds()
	out := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(out, out.Bounds(), img, b.Min, draw.Src)

	return out, nil
}

func loadLocalImage(fn string) (*image.RGBA, error) {
	rc, err := os.Open(fn)
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	return loadRGBA(rc)
}

func getImage(u string) (*image.RGBA, error) {
	if uu, urlErr := url.Parse(u); urlErr != nil || uu.Scheme == "" {
		return loadLocalImage(u)
	}

	fn, img, err := getCachedImage(u, "orig")
	if err == nil {
		log.Println("Using cache for URL:", u)
		return img, nil
	}

	log.Println("Retrieving URL:", u)

	resp, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var fh io.Reader

	os.Remove(fn)
	if w, err := os.Create(fn); err == nil {
		defer w.Close()
		fh = io.TeeReader(resp.Body, w)
	} else {
		fh = resp.Body
	}

	return loadRGBA(fh)
}

type Tile struct {
	orig  *image.RGBA
	thumb *image.RGBA
}

func preparePhoto(url string) *Tile {
	tile := new(Tile)

	var err error

	tile.orig, err = getImage(url)

	if err != nil {
		log.Fatal(err)
	}

	// XXX: needs to handle upscaling
	if tile.orig.Bounds().Dx() != *tileSize || tile.orig.Bounds().Dy() != *tileSize {
		tag := fmt.Sprintf("%[1]dx%[1]d", *tileSize)
		fn, img, err := getCachedImage(url, tag)
		if err == nil {
			tile.thumb = img
		} else {
			log.Println("Scaling url:", url)
			tile.thumb = downscale(tile.orig, image.Rect(0, 0, *tileSize, *tileSize))
			log.Println("Scaled url:", url)

			// write scaled down image to fs
			os.Remove(fn)
			if w, err := os.Create(fn); err == nil {
				defer w.Close()
				if err := png.Encode(w, tile.thumb); err != nil {
					os.Remove(fn)
				}
			}
		}
	} else {
		tile.thumb = tile.orig
	}

	return tile
}

func dist2(a, b *image.RGBA) float32 {
	a0 := a.Bounds().Min
	b0 := b.Bounds().Min
	s := a.Bounds().Size()

	if s.Eq(b.Bounds().Size()) == false {
		return math.MaxFloat32
	}

	d := float32(0.)

	for j := 0; j < s.Y; j++ {
		for i := 0; i < s.X; i++ {
			ca := a.RGBAAt(a0.X+i, a0.Y+j)
			cb := b.RGBAAt(b0.X+i, b0.Y+j)

			dr := (float32(ca.R) - float32(cb.R)) / 255.
			dg := (float32(ca.G) - float32(cb.G)) / 255.
			db := (float32(ca.B) - float32(cb.B)) / 255.
			da := (float32(ca.A) - float32(cb.A)) / 255.

			d += dr*dr + db*db + dg*dg + da*da
		}
	}

	return d
}

func queryFlickr(query string, r chan string, done chan int) {
	defer close(r)

	flickrClient := NewFlickrClient("config.json")

	for page := 1; true; page++ {
		search := map[string]string{
			"text":           query,
			"content_type":   "1",
			"media":          "photos",
			"is_commons":     "1",
			"per_page":       "500",
			"page":           strconv.Itoa(page),
			"format":         "json",
			"nojsoncallback": "1",
		}

		if *cconly {
			search["license"] = "1,4,5,7,8"
		}

		photos, err := flickrClient.FlickrPhotosSearch(search)

		if err != nil || len(photos.Photos.Photo) == 0 {
			return
		}

		for _, photo := range photos.Photos.Photo {
			select {
			case r <- photo.ThumbnailUrl():
			case <-done:
				return
			}
		}
	}
}

func getFlickrPhotos(query string, max int) chan string {
	urlChannel := make(chan string, 10)

	go func() {
		defer close(urlChannel)

		done := make(chan int)
		defer close(done)

		r := make(chan string)

		go queryFlickr(query, r, done)

		for n := 0; n < max; {
			url, ok := <-r
			if !ok {
				break
			}
			urlChannel <- url
			n++
		}
	}()

	return urlChannel
}

func getDirectoryPhotos(dir string, max int) chan string {
	fnChannel := make(chan string, 10)

	go func() {
		defer close(fnChannel)
		filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if info.IsDir() {
				return nil
			}
			fnChannel <- path
			return nil
		})
	}()

	return fnChannel
}

func getTiles(urlChannel chan string) chan *Tile {
	tiles := make(chan *Tile)
	wg := &sync.WaitGroup{}

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for url := range urlChannel {
				if tile := preparePhoto(url); tile != nil {
					tiles <- tile
				}
			}
			log.Print("Done processing images")
		}()
	}

	go func() {
		wg.Wait()
		close(tiles)
	}()

	return tiles
}

type TileInfo struct {
	subImage *image.RGBA
	bestTile *Tile
	dist2    float32
}

type TiledImage struct {
	data   []TileInfo // the set of tiles
	nx, ny int        // the number of tiles in the x and y directions
	sx, sy int        // the size of the tiles in the x and y directions
}

func NewTiledImage(img *image.RGBA, tileSize image.Point) *TiledImage {
	b := img.Bounds()
	nx := b.Dx() / tileSize.X
	ny := b.Dy() / tileSize.Y
	tiles := make([]TileInfo, nx*ny)

	for j := 0; j < ny; j++ {
		y := b.Min.Y + j*tileSize.Y
		for i := 0; i < nx; i++ {
			x := b.Min.X + i*tileSize.X
			rect := image.Rect(x, y, x+tileSize.X, y+tileSize.Y)
			k := i + j*nx
			tiles[k].subImage = img.SubImage(rect).(*image.RGBA)
			tiles[k].dist2 = math.MaxFloat32
		}
	}

	return &TiledImage{
		data: tiles,
		nx:   nx,
		ny:   ny,
		sx:   tileSize.X,
		sy:   tileSize.Y,
	}
}

func (t *TiledImage) Get(i, j int) *TileInfo {
	return &t.data[i+j*t.nx]
}

func MakeTileInfo(img *image.RGBA, tileSize image.Point) [][]TileInfo {
	b := img.Bounds()
	nx := b.Dx() / tileSize.X
	ny := b.Dy() / tileSize.Y

	tiles := make([][]TileInfo, nx)
	for i, _ := range tiles {
		tiles[i] = make([]TileInfo, ny)
	}

	for j, y := 0, b.Min.Y; j < ny; j, y = j+1, y+tileSize.Y {
		for i, x := 0, b.Min.X; i < nx; i, x = i+1, x+tileSize.X {
			rect := image.Rect(x, y, x+tileSize.X, y+tileSize.Y)
			tiles[i][j].subImage = img.SubImage(rect).(*image.RGBA)
			tiles[i][j].dist2 = math.MaxFloat32
		}
	}

	return tiles
}

func FindBestTiles(photos chan string, tiles *TiledImage) {
	// Horrible O(p*m) algorithm ahead, where p~1/n² is the number
	// of tiles in the image and m is the number of input tiles
	// considered for building the final image.
	//
	// This might work as an alternative:
	//
	// 1. Compute mean value for each RGB component in the tiles
	// 2. Insert tile in octree
	// 3. Compute mean value for each RGB component in the image tile
	// 4. For each tile in 3. look for the k closest tiles in octree
	// 5. Compare image tile to k closest tiles pixel-wise
	for tile := range getTiles(photos) {
		for i, _ := range tiles.data {
			info := &tiles.data[i]
			d2 := dist2(info.subImage, tile.thumb)
			if d2 < info.dist2 {
				info.bestTile = tile
				info.dist2 = d2
			}
		}
	}
}

func main() {
	if *tileSize < 1 || *tileSize > 75 {
		log.Fatal("'tilesize' must be between 1 and 75, inclusive")
	}

	if *outTileSize < 1 || *outTileSize > 75 {
		log.Fatal("'outtilesize' must be between 1 and 75, inclusive")
	}

	var photos chan string

	switch {
	case *dir != "":
		photos = getDirectoryPhotos(*dir, *maxpics)
	case *query != "":
		photos = getFlickrPhotos(*query, *maxpics)
	default:
		log.Fatal("Must provide either -dir or -query")
	}

	img, err := getImage(*photo)
	if err != nil {
		log.Fatalf("E: Failed to get %s: %s\n", *photo, err)
	}

	tiles := NewTiledImage(img, image.Point{*tileSize, *tileSize})

	FindBestTiles(photos, tiles)

	log.Println("Composing final image")

	s := *outTileSize
	dst := image.NewRGBA(image.Rect(0, 0, tiles.nx*s, tiles.ny*s))

	for j := 0; j < tiles.ny; j++ {
		for i := 0; i < tiles.nx; i++ {
			info := tiles.Get(i, j)
			r := image.Rect(i*s, j*s, (i+1)*s, (j+1)*s)
			t := info.bestTile.orig
			b := t.Bounds()
			if b.Dx() > s || b.Dy() > s {
				// XXX: needs cropping, upscaling
				t = downscale(t, image.Rect(0, 0, s, s))
				// replace orig with scaled down version
				// for reuse
				info.bestTile.orig = t
			}
			draw.Draw(dst, r, t, b.Min, draw.Over)
		}
	}

	w, err := os.Create(*output)
	if err != nil {
		log.Fatal(err)
	}
	if err := jpeg.Encode(w, dst, nil); err != nil {
		log.Fatal(err)
	}
}
