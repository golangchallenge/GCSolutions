package main

import (
	"flag"
	"fmt"
	"os"
)

const helptext = `
Usage: %s -target=TARGET -tiles=TILEDIR -out=OUTPUT [OPTION]...

Generate a mosaic image representing TARGET using images from TILEDIR and save it to OUTPUT
You can adjust the number of tiles in the mosaic (xt, yt) and their resolution (tw, th)

Example: %[1]s -target=example/tiles/gopher-squishy-3.jpg -tiles=example/tiles -out=mosaic.png

`

func main() {
	// flags
	target := flag.String("target", "", "path of the target picture")
	tiledir := flag.String("tiles", "", "directory containing tile pictures")
	out := flag.String("out", "", "path of the generated mosaic image")
	xt := flag.Int("xt", 120, "number of horizontal tiles in target rendering")
	yt := flag.Int("yt", 80, "number of vertical tiles in target rendering")
	tw := flag.Int("tw", 50, "tile width in pixels")
	th := flag.Int("th", 30, "tile height in pixels")
	help := flag.Bool("h", false, "print usage")

	flag.Parse()
	if *help || *target == "" || *tiledir == "" || *out == "" {
		fmt.Fprintf(os.Stderr, helptext, os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}

	if *xt < 1 || *yt < 1 || *tw < 1 || *th < 1 {
		fmt.Fprintf(os.Stderr, "all dimensions must be greater than 0!")
		os.Exit(2)
	}

	m := NewMosaic(*target, *xt, *yt, *tiledir, *tw, *th)
	if err := m.Render(); err != nil {
		fmt.Fprintf(os.Stderr, "cant render mosaic: "+err.Error())
		os.Exit(3)
	}
	if err := m.Save(*out); err != nil {
		fmt.Fprintf(os.Stderr, "cant save result: "+err.Error())
		os.Exit(4)
	}
}
