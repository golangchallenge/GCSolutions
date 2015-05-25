package main

import (
	"fmt"
	"html/template"
	"image"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"time"
)

// An App represents the web application.
type App struct {
	TileSet *TileSet
}

// Create a new application.
func NewApp(ts *TileSet) *App {
	return &App{TileSet: ts}
}

// Serve the home page of the application.
func (a *App) HomeHandler(res http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadFile("templates/index.html")
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	tmpl, err := template.New("Home").Parse(string(body))
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	tmpl.Execute(res, req.URL.Query().Get("err"))
}

// Serve the mosaic page. A parameter `id` is required to fetch the proper mosaic.
// Returns a 404 when the given `id` does not match with an existent mosaic.
func (a *App) GetMosaicHandler(res http.ResponseWriter, req *http.Request) {
	body, _ := ioutil.ReadFile("templates/mosaic.html")
	tmpl, _ := template.New("mosaic").Parse(string(body))

	res.Header().Set("Content-Type", "text/html")

	id := req.URL.Query().Get("id")

	if _, err := os.Stat(uploadTo(id)); err != nil {
		http.NotFound(res, req)
		return
	}

	tmpl.Execute(res, id)
}

// Receives a POST request to generate a new mosaic from the submited image.
// The generated mosaic is store in the file system and the page is redirected
// to the GetMosaic handler.
func (a *App) PostMosaicHandler(res http.ResponseWriter, req *http.Request) {
	file, _, err := req.FormFile("file")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	img, _, err := image.Decode(file)
	if err != nil {
		http.Redirect(res, req, "/?err=unknown-format", 301)
		return
	}

	m := NewMosaic(img, a.TileSet, 25)
	m.Generate()

	id := mosaicFilename()
	err = m.Save(uploadTo(id))
	if err != nil {
		http.Redirect(res, req, "/?err=failed", 301)
		return
	}

	http.Redirect(res, req, "/m/"+id, 301)
}

var mosaicFilename = func() string {
	rand.Seed(time.Now().Unix())

	return fmt.Sprintf("%d", rand.Int())
}

var uploadTo = func(filename string) string {
	return fmt.Sprintf("mosaics/%v.jpg", filename)
}
