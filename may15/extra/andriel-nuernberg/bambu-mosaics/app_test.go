package main

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

var testTmpDir = "testdata/tmp"
var tileSetTest, _ = NewTileSet("testdata/tiles")
var app = NewApp(tileSetTest)

func TestHomeHandler(t *testing.T) {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)

	app.HomeHandler(res, req)

	if res.Code != 200 {
		t.Errorf("Not equal: %#v (expected). %#v (actual)", 200, res.Code)
	}
}

func TestGenerateMosaicHandler(t *testing.T) {
	fp := "testdata/cat-tile.jpg"
	file, _ := os.Open(fp)
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", filepath.Base(fp))
	io.Copy(part, file)
	writer.Close()

	res := httptest.NewRecorder()

	req, _ := http.NewRequest("POST", "/m", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	mosaicFilename = func() string {
		return "12345"
	}

	os.MkdirAll(testTmpDir, 0777)
	uploadTo = func(filename string) string {
		return fmt.Sprintf("%v/%v.jpg", testTmpDir, filename)
	}

	app.PostMosaicHandler(res, req)

	if res.Code != 301 {
		t.Errorf("Not equal: %#v (expected). %#v (actual)", 301, res.Code)
	}

	id := mosaicFilename()
	_, err := os.Stat(uploadTo(id))
	if err != nil {
		t.Errorf("No error is expected but got %v", err)
	}
}
