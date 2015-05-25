package main

import (
	"flag"
	"fmt"
	"image"
	"image/color/palette"
	"image/jpeg"
	"log"
	"os"
	"path"
	"time"

	"github.com/rcarver/golang-challenge-3-mosaic/instagram"
	"github.com/rcarver/golang-challenge-3-mosaic/mosaic"
	"github.com/rcarver/golang-challenge-3-mosaic/service"
)

var (
	fetch         *flag.FlagSet
	gen           *flag.FlagSet
	serve         *flag.FlagSet
	tag           string
	baseDirName   string
	imgDirName    string
	inName        string
	outName       string
	outDownsample float64
	units         int
	unitSize      int
	numImages     int
	solid         bool
	port          int
)

var help = `
# Download images by tag
mosaic -run download -dir images -tag balloon 

# Generate mosaic with tag
mosaic -run generate -dir images -tag balloon -in balloon.jpg -out balloon-mosaic.jpg
`

func init() {
	fetch = flag.NewFlagSet("fetch", flag.ExitOnError)
	fetch.StringVar(&baseDirName, "dir", "./cache/thumbs", "dir to store images by tag")
	fetch.StringVar(&tag, "tag", "cat", "image tag to use")
	fetch.IntVar(&numImages, "num", 1000, "number of images to download")

	gen = flag.NewFlagSet("gen", flag.ExitOnError)
	gen.StringVar(&baseDirName, "dir", "./cache/thumbs", "dir to store images by tag")
	gen.StringVar(&tag, "tag", "cat", "image tag to use")
	gen.StringVar(&inName, "in", "", "image file to read")
	gen.StringVar(&outName, "out", "./mosaic.jpg", "image file to write")
	gen.Float64Var(&outDownsample, "shrink", 0.5, "perentage to shrink the output image as a percentage 0-1")
	gen.StringVar(&imgDirName, "imgdir", "", "dir to find images (uses $dir/thumbs/$tag by default)")
	gen.IntVar(&units, "units", 40, "number of units wide to generate the mosaic")
	gen.IntVar(&unitSize, "unitSize", instagram.ThumbnailSize, "pixels w/h of the thumbnail images")
	gen.BoolVar(&solid, "solid", false, "generate a mosaic with solid colors, not images")

	serve = flag.NewFlagSet("serve", flag.ExitOnError)
	serve.StringVar(&baseDirName, "dir", "./cache", "dir to store thumbs and mosaics")
	serve.IntVar(&numImages, "num", 1000, "number of images to download")
	serve.IntVar(&units, "units", 40, "number of units wide to generate the mosaic")
	serve.IntVar(&unitSize, "unitSize", instagram.ThumbnailSize, "pixels w/h of the thumbnail images")
	serve.IntVar(&port, "port", 8080, "port number of the server")
}

func main() {
	usage := fmt.Sprintf("Usage: %s <command> <args>", path.Base(os.Args[0]))
	if len(os.Args) == 1 {
		fmt.Println(usage)
		os.Exit(2)
	}
	command := os.Args[1]
	switch command {
	case "fetch":
		fetch.Parse(os.Args[2:])
	case "gen":
		gen.Parse(os.Args[2:])
	case "serve":
		serve.Parse(os.Args[2:])
	default:
		fmt.Println(usage)
		fmt.Printf("fetch:\n")
		fetch.PrintDefaults()
		fmt.Printf("gen:\n")
		gen.PrintDefaults()
		fmt.Printf("serve:\n")
		serve.PrintDefaults()
		os.Exit(2)
	}

	// thumbsDir is imgDirName if set, or join(baseDirName, "thumbs", tag)
	var thumbsDir string
	if imgDirName != "" {
		thumbsDir = imgDirName
	} else {
		thumbsDir = path.Join(baseDirName, "thumbs", tag)
	}
	if err := os.MkdirAll(thumbsDir, 0755); err != nil {
		fmt.Printf("Error initializing: %s\n", err)
	}

	// inventory reads and writes from the dir.
	inventory := newInventory(thumbsDir)

	switch command {
	case "fetch":
		if err := downloadImages(tag, numImages, inventory); err != nil {
			fmt.Printf("Download error: %s\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	case "gen":
		if inName == "" {
			fmt.Printf("Missing -in file\n")
			os.Exit(1)
		}

		// Read and decode input image.
		in, err := os.Open(inName)
		if err != nil {
			fmt.Printf("Error initializing: %s\n", err)
			os.Exit(1)
		}
		defer in.Close()
		src, _, err := image.Decode(in)
		if err != nil {
			fmt.Printf("Error initializing: %s\n", err)
			os.Exit(1)
		}

		// Generate the mosaic.
		img, err := generateMosaic(src, tag, units, solid, inventory)
		if err != nil {
			fmt.Printf("Error generating: %s\n", err)
			os.Exit(1)
		}

		// Encode and write the output image.
		out, err := os.Create(outName)
		if err != nil {
			fmt.Printf("Error outputting: %s\n", err)
			os.Exit(1)
		}
		defer out.Close()
		err = jpeg.Encode(out, img, nil)
		if err != nil {
			fmt.Printf("Error outputting: %s\n", err)
			os.Exit(1)
		}

		os.Exit(0)
	case "serve":
		service.HostPort = fmt.Sprintf(":%d", port)
		service.MosaicsDir = path.Join(baseDirName, "mosaics")
		service.ThumbsDir = path.Join(baseDirName, "thumbs")
		service.ImagesPerTag = numImages
		service.Units = units
		service.UnitSize = unitSize
		service.Serve()
		os.Exit(0)
	default:
		flag.PrintDefaults()
		os.Exit(1)
	}
}

func newInventory(dir string) *mosaic.ImageInventory {
	cache := mosaic.NewFileImageCache(dir)
	return mosaic.NewImageInventory(cache)
}

func downloadImages(tag string, numImages int, inv *mosaic.ImageInventory) error {
	api := instagram.NewClient()
	fetcher := instagram.NewTagFetcher(api, tag)
	if err := inv.Fetch(fetcher, numImages); err != nil {
		return err
	}
	time.Sleep(1 * time.Second)
	return nil
}

var (
	// Number of colors in the mosaic color palette.
	paletteSize = 256
)

func generateMosaic(src image.Image, tag string, units int, solid bool, inv *mosaic.ImageInventory) (image.Image, error) {
	var p *mosaic.ImagePalette
	if solid {
		p = mosaic.NewSolidPalette(palette.WebSafe)
		log.Printf("Generating %dx%d solid mosaic with %d colors", units, units, p.NumColors())
	} else {
		p = mosaic.NewImagePalette(paletteSize)
		if err := inv.PopulatePalette(p); err != nil {
			return nil, err
		}
		if p.NumColors() == 0 {
			return nil, fmt.Errorf("No images are available")
		}
		log.Printf("Generating %dx%d %s mosaic with %d colors and %d images\n", units, units, tag, p.NumColors(), p.NumImages())
	}
	sq := mosaic.ComposeSquare(src, units, unitSize, p)
	return mosaic.Shrink(sq, outDownsample), nil
}
