package main

import (
	"flag"
	"io"
	"net/http"
	"path/filepath"

	"image"
	"image/gif"
	"image/jpeg"
	_ "image/png"
	"io/ioutil"
	"log"
	"os"
	"path"

	"github.com/captncraig/mosaicChallenge/mosaics"
)

var img = flag.String("img", "input.jpg", "filenamme or url of image to build mosaic from")
var thumbDir = flag.String("dir", "thumbnails", "Directory containing 90x90 thumbnails to build mosaic from")
var grid = flag.Int("grid", 0, "If set, will evaluate thumbnails using a nxn grid. If not specified, will simply average the entire thumbnail color vs the target image segment.")
var doGif = flag.Bool("gif", false, "Build gifzaic. This could take a long time.")
var outputFile = flag.String("out", "out.jpg", "Name of output file.")

func main() {
	flag.Parse()
	var comparision mosaics.Evaluator
	if *grid == 0 {
		comparision = mosaics.AveragingEvaluator()
	} else {
		comparision = mosaics.GridEvaluator(*grid)
	}
	lib := mosaics.NewLibrary(comparision)
	files, err := ioutil.ReadDir(*thumbDir)
	if err != nil {
		check(err)
	}
	log.Println("Loading thumbnails...")
	for _, file := range files {
		img := parseFile(*thumbDir, file.Name())
		lib.AddImage(img)
	}
	log.Println("Loading target image...")
	var r io.Reader
	if (*img)[0:4] == "http" {
		resp, err := http.Get(*img)
		check(err)
		r = resp.Body
	} else {
		f, err := os.Open(*img)
		check(err)
		defer f.Close()
		r = f
	}

	progress := make(chan float64)
	go func() {
		for pct := range progress {
			log.Printf("%.2f percent done\n", pct)
		}
	}()

	if !*doGif {
		target, _, err := image.Decode(r)
		check(err)
		log.Println("Building mosaic...")
		mos := mosaics.BuildMosaicFromLibrary(target, lib, progress)
		output, err := os.Create(*outputFile)
		check(err)
		defer output.Close()
		err = jpeg.Encode(output, mos, &jpeg.Options{20})
		check(err)
	} else {
		g, err := gif.DecodeAll(r)
		check(err)
		log.Println("Building mosaic...")
		mos, err := mosaics.BuildGifzaic(g, lib, progress)
		check(err)
		output, err := os.Create(*outputFile)
		check(err)
		defer output.Close()
		err = gif.EncodeAll(output, mos)
		check(err)
	}
}
func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
func parseFile(dir, file string) image.Image {
	f, err := os.Open(path.Join(dir, file))
	check(err)
	defer f.Close()
	img, _, err := image.Decode(f)
	check(err)
	// For most of the challenge I assumed I was going to use imgur generated 90x90 thumbnails.
	// Once requirements changed, I never got to implement thumbnail generation properly.
	if img.Bounds().Dx() != 90 || img.Bounds().Dy() != 90 {
		log.Printf("WARNING: This program works best with 90x90 thumbnails. %s is %dx%d", file, img.Bounds().Dx(), img.Bounds().Dy())
	}
	return img
}
