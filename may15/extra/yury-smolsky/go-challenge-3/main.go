package main

import (
	"fmt"
	"go-challenge-3/mosaic"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"io/ioutil"
	"os"
	"path"
	"runtime"
	"strings"
)

// Tries to read image from disk by name path
func readImage(path string) (img image.Image, err error) {
	reader, e := os.Open(path)
	if e != nil {
		err = e
		return
	}
	defer reader.Close()
	img, _, e = image.Decode(reader)
	if e != nil {
		err = e
		return
	}
	return
}

func readImageWorker(jobs chan string, results chan image.Image) {
	for j := range jobs {
		m, _ := readImage(j)
		results <- m
	}
}

// Reads concurrently all images in directory dir and returns them is slice
func readDirConc(dir string) (imgs []image.Image, err error) {
	files, e := ioutil.ReadDir(dir)
	if e != nil {
		err = e
		return
	}
	jobs := make(chan string, 10)
	results := make(chan image.Image, 10)
	for i := 0; i < runtime.NumCPU(); i++ {
		go readImageWorker(jobs, results)
	}
	go func() {
		for i := 0; i < len(files); i++ {
			fname := path.Join(dir, files[i].Name())
			jobs <- fname
		}
		close(jobs)
	}()

	for i := 0; i < len(files); i++ {
		img := <-results
		if img != nil {
			imgs = append(imgs, img)
		}
	}

	return
}

// Saves m image to disk by name as jpeg file
func saveJpeg(name string, m image.Image) (e error) {
	e = nil
	file, err := os.Create(name)
	if err != nil {
		e = err
		return
	}
	defer file.Close()
	jpeg.Encode(file, m, &jpeg.Options{Quality: 80})
	return
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	if len(os.Args) < 4 {
		fmt.Println("Mosaic maker. 2015 (c) Yury Smolsky")
		fmt.Println("Usage: ./go-challenge-3 <target_image> <path_to_tile_images> <output_image>[.jpg]")
		return
	}
	targetFile := os.Args[1]
	tilesDir := os.Args[2]
	outFile := os.Args[3]

	targetImg, e := readImage(targetFile)
	if e != nil {
		fmt.Println("Cannot read target image:", e)
		return
	}

	fmt.Print("reading tile imgs... ")
	imgs, e := readDirConc(tilesDir)
	if e != nil {
		fmt.Println(e)
		return
	}
	if len(imgs) < mosaic.RequiredImages() {
		fmt.Printf("Less than %d source images were found! (%d)\n", mosaic.RequiredImages, len(imgs))
		return
	}
	fmt.Println("images found:", len(imgs))

	m := mosaic.Create(targetImg, imgs)

	fmt.Println("saving image...")
	if !strings.HasSuffix(strings.ToLower(outFile), ".jpg") {
		outFile = outFile + ".jpg"
	}
	e = saveJpeg(outFile, m)
	if e != nil {
		fmt.Println("Cannot save to", outFile)
	}
}
