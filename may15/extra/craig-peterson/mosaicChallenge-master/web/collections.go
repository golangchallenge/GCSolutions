package main

import (
	"bytes"
	"encoding/base64"
	"github.com/captncraig/mosaicChallenge/mosaics"
	"html/template"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

type imageCollection struct {
	Desc    string
	path    string
	library *mosaics.ThumbnailLibrary
	Count   int
}

// A cool helper to encode a selection of sample images to be directly embedded into <img> tags
func (i *imageCollection) GetSampleImages() []template.HTMLAttr {
	imgs := []template.HTMLAttr{}
	for _, img := range i.library.GetSampleImages(9) {
		buf := &bytes.Buffer{}
		err := jpeg.Encode(buf, img, &jpeg.Options{30})
		if err != nil {
			return imgs
		}
		data := "src=data:image/jpg;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())
		imgs = append(imgs, template.HTMLAttr(data))
	}
	return imgs
}

var collections map[string]*imageCollection

var collectionRoot string

func init() {
	if collectionRoot = os.Getenv("CollectionsDir"); collectionRoot == "" {
		collectionRoot = "../collections"
	}
}

func loadBuiltInCollections() {
	collections = map[string]*imageCollection{
		// these would also normally be dynamically loaded from a config file or something but focus moved away from web app.
		"Design-seeds": {"Thumbnails scraped from design-seeds.com. Great for color variety.", "designseeds", mosaics.NewLibrary(mosaics.AveragingEvaluator()), 0},
		"Aww":          {"top images from /r/aww", "reddits/aww", mosaics.NewLibrary(mosaics.AveragingEvaluator()), 0},
		"Cats":         {"top images from /r/cats", "reddits/cats", mosaics.NewLibrary(mosaics.AveragingEvaluator()), 0},
		"Food":         {"top images from /r/food", "reddits/food", mosaics.NewLibrary(mosaics.AveragingEvaluator()), 0},
		"Flowers":      {"top images from /r/flowers", "reddits/flowers", mosaics.NewLibrary(mosaics.AveragingEvaluator()), 0},
		"Humans":       {"top images from /r/humanPorn", "reddits/humanPorn", mosaics.NewLibrary(mosaics.AveragingEvaluator()), 0},
		"Earth":        {"top images from /r/earthPorn", "reddits/earthPorn", mosaics.NewLibrary(mosaics.AveragingEvaluator()), 0},
		"bttf":         {"Stills from the best movie ever made", "bttf", mosaics.NewLibrary(mosaics.AveragingEvaluator()), 0},
		"nemo":         {"Stills from a colorful movie", "nemo", mosaics.NewLibrary(mosaics.AveragingEvaluator()), 0},
	}
	all := &imageCollection{"All built-in collections combined for maximum variety. May take a long time to build.", "", mosaics.NewLibrary(mosaics.AveragingEvaluator()), 0}
	log.Printf("Loading %d built-in collections\n", len(collections))
	for name, collection := range collections {
		dir := filepath.Join(collectionRoot, collection.path)
		log.Printf("\tLoading %s from %s\n", name, dir)
		files, err := ioutil.ReadDir(dir)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("\t\tLoading %d images", len(files))
		targetPercent := .1
		for i, file := range files {
			f, err := os.Open(filepath.Join(dir, file.Name()))
			if err != nil {
				log.Fatal(err)
			}

			img, _, err := image.Decode(f)
			if err != nil {
				log.Fatal(err)
			}
			f.Close()
			collection.library.AddImage(img)
			collection.Count++
			all.library.AddImage(img)
			all.Count++
			progress := float64(i) / float64(len(files))
			if progress > targetPercent {
				log.Printf("%.0f%%\n", progress*100)
				targetPercent += .1
			}
		}
	}
	collections["all"] = all
	log.Println("Done loading built-in collections")
}

// some example images you can choose to make a mosaic from.
// plan was to make this dynamically loaded from an imgur gallery,
// but requirements changed away from a web app.
var sampleImages = []string{
	"http://i.imgur.com/qaHCKuM.jpg",
	"http://i.imgur.com/9MGZuJU.jpg",
	"http://i.imgur.com/SMEDFqW.jpg",
	"http://www.empireonline.com/images/uploaded/marty-guitar.jpg",
	"https://unsplash.imgix.net/photo-1421986527537-888d998adb74?fit=crop&fm=jpg&h=700&q=75&w=1050",
}
