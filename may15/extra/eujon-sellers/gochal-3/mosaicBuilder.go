package main

import (
	"errors"
	"image"
	"image/draw"
	"image/png"
	"math"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"
)

var semGet = make(chan int, 500)
var sem = make(chan int, 20)

// break up the original image into tiles and get the
// avg RGB for each tile
func initMosaic(ulPhoto string) (ImageFile, error) {
	submitPic := ImageFile{Name: "uploads/" + ulPhoto}
	logger.Println("creating tiles...")
	submitPic.createTiles()
	if len(submitPic.Tiles) > 14400 {
		return ImageFile{}, errors.New("more than 14400 tiles")
	}
	logger.Println("tiles size: ", len(submitPic.Tiles))
	for _, val := range submitPic.Tiles {
		val.buildColorMap(submitPic)
	}
	return submitPic, nil
}

// gather all the data and images needed to make the actual mosaic
func buildMosaic(flickrGetter HttpGetter, submitPic ImageFile, query string, buildChan chan bool) {
	var flickrPhotos []FlickrPhoto
	// get twice as many tiles to try and improve the mosaic
	pageCount := (len(submitPic.Tiles)/400 + 1) * 2
	logger.Println("getting flickr photos...")
	getFlickrPhotos(flickrGetter, pageCount, &flickrPhotos, query, true)
	logger.Println("trying to match tiles...")
	for _, tile := range submitPic.Tiles {
		if tile.MatchURL == "" {
			bestIndex := 0
			bestDiff := 0
			for i := 0; i < len(flickrPhotos); i++ {
				diff := compareRGB(tile.AvgRGB, flickrPhotos[i].AvgRGB)
				if i == 0 {
					bestDiff = diff
					bestIndex = i
				} else {
					if diff < bestDiff {
						bestIndex = i
						bestDiff = diff
					}
				}
			} // no match, break out
			tile.MatchURL = flickrPhotos[bestIndex].URL
		}
	}
	flickrPhotos = nil
	logger.Println("origTiles: ", len(submitPic.Tiles))
	buildChan <- true
}

// write out the actual mosaic to disk
func writeMosaic(flickrGetter HttpGetter, submitPic ImageFile, md5sum string) (bool, string) {
	mosaicImage := image.NewNRGBA(image.Rect(0, 0, submitPic.XEnd*4, submitPic.YEnd*4))
	for _, tile := range submitPic.Tiles {
		var tempImage image.Image
		reader, err := os.Open(flickrGetter.GetSaveDir() + "/" + tile.MatchURL)
		if err != nil {
			logger.Printf("unable to use %v image as tile: %v\n", tile.MatchURL, err)
			tempImage, err = getFlickrPhotoNoSave(tile.MatchURL, flickrGetter)
			if err != nil {
				logger.Println("error getting image from flickr: ", err)
			}
		} else {
			tempImage, _, err = image.Decode(reader)
			if err != nil {
				logger.Println("error decoding tile image: ", err)
			}
		}
		if err == nil {
			tempImageNRGBA := convertToNRGBA(tempImage)
			tempImageNRGBA = resizeNearest(tempImageNRGBA, 40, 40)
			point := image.Point{}
			if tile.XStart == 0 && tile.YStart == 0 {

			} else {
				point = image.Point{Y: tile.YStart * 4, X: tile.XStart * 4}
			}
			rect := tempImageNRGBA.Bounds()
			r := image.Rectangle{point, point.Add(rect.Size())}
			draw.FloydSteinberg.Draw(mosaicImage, r, tempImageNRGBA, rect.Min)
		}
		reader.Close()
	}
	// resize a bit
	mosaicImage = resizeNearest(mosaicImage, mosaicImage.Bounds().Max.X/2, mosaicImage.Bounds().Max.Y/2)
	// write file
	outFileName := "mosaic_" + md5sum + strconv.Itoa(int(time.Now().Unix())) + ".png"
	out, err := os.Create("created/" + outFileName)
	if err != nil {
		logger.Println("error creating file: ", err)
	}
	err = png.Encode(out, mosaicImage)
	if err != nil {
		logger.Println("error encoding mosaic image: ", err)
	}
	return true, outFileName
}

// call all of the various Flickr operations in order to build out all of our tile images
// this is where most of the heavy work is done
func getFlickrPhotos(fg HttpGetter, pageCount int, flickrPhotos *[]FlickrPhoto, query string, enableCache bool) {
	var wg sync.WaitGroup
	var cacheWait sync.WaitGroup
	var cachedFiles []FlickrPhoto
	if enableCache {
		cacheWait.Add(1)
		go func(cache []FlickrPhoto) {
			defer cacheWait.Done()
			cache = cacheFromDisk(fg.GetSaveDir())
		}(cachedFiles)
	}
	today := time.Now()
	lastMonth := today.Add(-(time.Hour * 168))
	logger.Println(lastMonth)
	rand.Seed(time.Now().Unix())
	startDate, resCount := queryFixer(fg, lastMonth, today, pageCount*400, query, 1)
	totalResCount, err := strconv.Atoi(resCount)
	var flickrResp []PhotoResponse
	if err != nil {
		logger.Println("error converting")
	}
	logger.Printf("will be getting %v pages...\n", pageCount)
	cacheWait.Wait()
	if totalResCount < pageCount*400 {
		logger.Println("we didn't return enough images...")
		smallCount := math.Floor(float64(totalResCount / 400))
		flickrResp = flickrQueryNew(fg, startDate, today, int(smallCount), query)
		// make up the rest
		if len(cachedFiles) >= (pageCount-int(smallCount))*400 {
			*flickrPhotos = append(*flickrPhotos, cachedFiles...)
		} else {
			flickrResp = append(flickrResp, flickrQueryNew(fg, lastMonth, today, pageCount-int(smallCount), query)...)
		}
	} else {
		flickrResp = flickrQueryNew(fg, startDate, today, pageCount, query)
	}
	if query == "" {
		*flickrPhotos = append(*flickrPhotos, cachedFiles...)
	}
	logger.Println("we got photos: ", len(flickrResp))
	// attempt to gate our gets out to flickr so we don't make them mad
	for _, photo := range flickrResp {
		photo := photo
		sem <- 1
		wg.Add(1)
		go func() {
			defer wg.Done()
			getFlickrPhoto(photo, flickrPhotos, fg)
			<-sem
		}()
	}
	wg.Wait()
}

// load some images from disk that were previously saved
func cacheFromDisk(dirName string) []FlickrPhoto {
	d, err := os.Open(dirName)
	if err != nil {
		logger.Println("unable to open cache directory: ", err)
	}
	defer d.Close()

	files, err := d.Readdir(-1)
	if err != nil {
		logger.Println(err)
	}
	logger.Println("Reading ", dirName)
	cachedFiles := make([]FlickrPhoto, 0)
	counter := 0
	for _, file := range files {
		if counter > 5000 {
			return cachedFiles
		}
		reader, err := os.Open(dirName + "/" + file.Name())
		var tempImage image.Image
		if err != nil {
			logger.Println("unable to open cached file: ", err)
		} else {
			tempImage, _, err = image.Decode(reader)
			if err != nil {
				logger.Println("unable to decode cached file: ", err)
			}
		}
		remoteResize := convertToNRGBA(tempImage)
		avgRGB := avgRGB(remoteResize, true)
		cachedFiles = append(cachedFiles, FlickrPhoto{URL: file.Name(), AvgRGB: avgRGB})
		reader.Close()
		counter++
	}
	return cachedFiles
}
