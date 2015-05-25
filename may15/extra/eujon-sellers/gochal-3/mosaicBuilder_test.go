package main

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	"image/jpeg"
	"image/png"
	"os"
	"strings"
	"testing"
	"time"
)

type TestGoodsGetter struct {
}

func (TestGoodsGetter) Get(url string) ([]byte, error) {

	if strings.Contains(url, "https://api.flickr.com/services/rest/?method=flickr.photos.search&api_key=488c1e7018f1ddf78b09d51a9604622a&media=photos&per_page=400&page=1&format=json") {
		return mockFlickrResp("1", "5", "2000"), nil
	}
	if strings.Contains(url, "https://api.flickr.com/services/rest/?method=flickr.photos.search&api_key=488c1e7018f1ddf78b09d51a9604622a&media=photos&per_page=400&page=2&format=json") {
		return mockFlickrResp("2", "5", "2000"), nil
	}
	if strings.Contains(url, "https://api.flickr.com/services/rest/?method=flickr.photos.search&api_key=488c1e7018f1ddf78b09d51a9604622a&media=photos&per_page=400&page=3&format=json") {
		return mockFlickrResp("3", "5", "2000"), nil
	}
	if strings.Contains(url, "https://api.flickr.com/services/rest/?method=flickr.photos.search&api_key=488c1e7018f1ddf78b09d51a9604622a&media=photos&per_page=400&page=4&format=json") {
		return mockFlickrResp("4", "5", "2000"), nil
	}
	if strings.Contains(url, "https://api.flickr.com/services/rest/?method=flickr.photos.search&api_key=488c1e7018f1ddf78b09d51a9604622a&media=photos&per_page=400&page=5&format=json") {
		return mockFlickrResp("5", "5", "2000"), nil
	}

	if strings.Contains(url, "https://farm") {
		reader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(data))
		m, _, err := image.Decode(reader)
		if err != nil {
			fmt.Println("Error loading data")
		}
		buf := new(bytes.Buffer)
		err = png.Encode(buf, convertToNRGBA(m))
		if err != nil {
			fmt.Println("error encoding test image: ", err)
		}
		return []byte(buf.Bytes()), nil
	}

	return nil, errors.New("Don't recognize URL: " + url)
}
func (TestGoodsGetter) GetSaveDir() string {
	return "testData"
}
func TestGetFlickrPhotos(t *testing.T) {
	os.RemoveAll("testData")
	os.Mkdir("testData", 0777)
	var flickrPhotos []FlickrPhoto
	//	var wg sync.WaitGroup
	pageCount := 2
	testGetter := TestGoodsGetter{}
	getFlickrPhotos(testGetter, pageCount, &flickrPhotos, "cats", false)
	//wg.Wait()
	if len(flickrPhotos) != 1600 {
		t.Errorf("flickrPhotos size is %v, should be 1600", len(flickrPhotos))
	}
	os.RemoveAll("testData")
}

func TestInitMosaic(t *testing.T) {
	reader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(data))
	m, _, err := image.Decode(reader)
	if err != nil {
		t.Errorf("Error loading data")
	}
	out, err := os.Create("uploads/createTiles-test.jpg")
	if err != nil {
		t.Errorf("Error creating test jpg")
	}
	err = jpeg.Encode(out, m, nil)
	if err != nil {
		t.Errorf("Error writing test jpg")
	}

	testPic, err := initMosaic("createTiles-test.jpg")
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	if len(testPic.Tiles) != 165 {
		t.Errorf("Tile count should be 165, is %v", len(testPic.Tiles))
	}
	os.Remove("uploads/createTiles-test.jpg")
}

/**
func BenchmarkInitMosaic(b *testing.B) {
	reader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(data))
	m, _, err := image.Decode(reader)
	if err != nil {
		b.Errorf("Error loading data")
	}
	out, err := os.Create("createTiles-test.jpg")
	if err != nil {
		b.Errorf("Error creating test jpg")
	}
	err = jpeg.Encode(out, m, nil)
	if err != nil {
		b.Errorf("Error writing test jpg")
	}
	testPic := ImageFile{Name: "createTiles-test.jpg"}

	for i := 0; i < b.N; i++ {
		testPic.CreateTiles()
	}
	os.Remove("createTiles-test.jpg")
}
**/
func TestBuildMosaic(t *testing.T) {
	os.RemoveAll("testData")
	os.Mkdir("testData", 0777)
	testGetter := TestGoodsGetter{}
	reader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(data))
	m, _, err := image.Decode(reader)
	if err != nil {
		t.Errorf("Error loading data")
	}
	out, err := os.Create("uploads/createTiles-test.jpg")
	if err != nil {
		t.Errorf("Error creating test jpg")
	}
	err = jpeg.Encode(out, m, nil)
	if err != nil {
		t.Errorf("Error writing test jpg")
	}

	testPic, err := initMosaic("createTiles-test.jpg")
	if err != nil {
		t.Errorf("error: %v", err)
	}
	if len(testPic.Tiles) != 165 {
		t.Errorf("Tile count should be 165, is %v", len(testPic.Tiles))
	}
	buildChan := make(chan bool, 1)
	buildSuccess := false
	go func() {
		buildMosaic(testGetter, testPic, "", buildChan)
	}()
	select {
	case res := <-buildChan:
		buildSuccess = res
	case <-time.After(time.Minute * 5):
		t.Errorf("timeout on mosaic build")
	}
	if !buildSuccess {
		t.Errorf("mosaic build failure")
	}
	matchedTiles := 0
	for _, val := range testPic.Tiles {
		if val.MatchURL != "" {
			matchedTiles++
		}
	}
	if matchedTiles != len(testPic.Tiles) {
		t.Errorf("Matched tile count should be 165, is %v", len(testPic.Tiles))
	}

	//os.Remove("uploads/createTiles-test.jpg")
	//os.RemoveAll("testData")
}

func TestWriteMosaic(t *testing.T) {
	//os.RemoveAll("testData")
	//os.Mkdir("testData", 0777)
	testGetter := TestGoodsGetter{}
	reader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(data))
	m, _, err := image.Decode(reader)
	if err != nil {
		t.Errorf("Error loading data")
	}
	out, err := os.Create("uploads/createTiles-test.jpg")
	if err != nil {
		t.Errorf("Error creating test jpg")
	}
	err = jpeg.Encode(out, m, nil)
	if err != nil {
		t.Errorf("Error writing test jpg")
	}

	testPic, err := initMosaic("createTiles-test.jpg")
	if err != nil {
		t.Errorf("error: %v", err)
	}
	if len(testPic.Tiles) != 165 {
		t.Errorf("Tile count should be 165, is %v", len(testPic.Tiles))
	}
	buildChan := make(chan bool, 1)
	buildSuccess := false
	go func() {
		buildMosaic(testGetter, testPic, "", buildChan)
	}()
	select {
	case res := <-buildChan:
		buildSuccess = res
	case <-time.After(time.Minute * 5):
		t.Errorf("timeout on mosaic build")
	}
	if !buildSuccess {
		t.Errorf("mosaic build failure")
	}
	matchedTiles := 0
	for _, val := range testPic.Tiles {
		if val.MatchURL != "" {
			matchedTiles++
		}
	}
	if matchedTiles != len(testPic.Tiles) {
		t.Errorf("Matched tile count should be 165, is %v", len(testPic.Tiles))
	}

	success, path := writeMosaic(testGetter, testPic, "testingblah")
	if !success {
		t.Errorf("writing mosaic failed")
	}

	moReader, err := os.Open("created/" + path)
	if err != nil {
		t.Errorf("unable to open mosiac image")
	}
	moImage, _, err := image.Decode(moReader)
	if err != nil {
		t.Errorf("unable to decode mosaic image")
	}
	moBounds := moImage.Bounds()
	if (testPic.XEnd * 2) != moBounds.Max.X {
		t.Errorf("max X is %v, should be %v", moBounds.Max.X, (testPic.XEnd * 2))
	}
	if (testPic.YEnd * 2) != moBounds.Max.Y {
		t.Errorf("max Y is %v, should be %v", moBounds.Max.Y, (testPic.YEnd * 2))
	}

	os.Remove("uploads/createTiles-test.jpg")
	os.Remove("created/testingblah.png")
	os.RemoveAll("testData")
}

func TestWriteMosaicNoSave(t *testing.T) {
	os.RemoveAll("testData")
	os.Mkdir("testData", 0777)
	testGetter := TestGoodsGetter{}
	reader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(data))
	m, _, err := image.Decode(reader)
	if err != nil {
		t.Errorf("Error loading data")
	}
	out, err := os.Create("uploads/createTiles-test.jpg")
	if err != nil {
		t.Errorf("Error creating test jpg")
	}
	err = jpeg.Encode(out, m, nil)
	if err != nil {
		t.Errorf("Error writing test jpg")
	}

	testPic, err := initMosaic("createTiles-test.jpg")
	if err != nil {
		t.Errorf("error: %v", err)
	}
	if len(testPic.Tiles) != 165 {
		t.Errorf("Tile count should be 165, is %v", len(testPic.Tiles))
	}
	buildChan := make(chan bool, 1)
	buildSuccess := false
	go func() {
		buildMosaic(testGetter, testPic, "", buildChan)
	}()
	select {
	case res := <-buildChan:
		buildSuccess = res
	case <-time.After(time.Minute * 5):
		t.Errorf("timeout on mosaic build")
	}
	if !buildSuccess {
		t.Errorf("mosaic build failure")
	}
	matchedTiles := 0
	for _, val := range testPic.Tiles {
		if val.MatchURL != "" {
			matchedTiles++
		}
	}
	if matchedTiles != len(testPic.Tiles) {
		t.Errorf("Matched tile count should be 165, is %v", len(testPic.Tiles))
	}
	os.RemoveAll("testData")
	success, path := writeMosaic(testGetter, testPic, "testingblah")
	if !success {
		t.Errorf("writing mosaic failed")
	}

	moReader, err := os.Open("created/" + path)
	if err != nil {
		t.Errorf("unable to open mosiac image")
	}
	moImage, _, err := image.Decode(moReader)
	if err != nil {
		t.Errorf("unable to decode mosaic image")
	}
	moBounds := moImage.Bounds()
	if (testPic.XEnd * 2) != moBounds.Max.X {
		t.Errorf("max X is %v, should be %v", moBounds.Max.X, (testPic.XEnd * 2))
	}
	if (testPic.YEnd * 2) != moBounds.Max.Y {
		t.Errorf("max Y is %v, should be %v", moBounds.Max.Y, (testPic.YEnd * 2))
	}
}
