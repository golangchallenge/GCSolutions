//https://groups.google.com/forum/#!topic/golang-nuts/H_DZ3mmtY4U

package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
)

var page Page

type Page struct {
	Title      string
	Src        string
	PUID       string
	TileSize   int
	ImageName  string
	BtnCaption string
	Redirected string
}

func main() {
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", InputHandler)
	http.HandleFunc("/input", InputHandler)
	http.HandleFunc("/upload", UploadHandler)
	http.HandleFunc("/index/", IndexHandler)
	http.HandleFunc("/mossaic", MossaicHandler)

	log.Println("\nServer Started On Port :8000...")
	http.ListenAndServe(":8000", nil)
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	//str := r.URL.String()
	str, _ := url.QueryUnescape(r.URL.String())

	Source := str[len("/index"):strings.LastIndex(str, "/")]
	PUID := Source[len("/static/"):strings.LastIndex(Source, "/")]
	ImageName := Source[strings.LastIndex(Source, "/")+1:]
	TileSize := str[strings.LastIndex(str, "/")+1:]
	ts, err := strconv.Atoi(TileSize)
	if err != nil {
		ts = 25
	}

	lp := path.Join("templates", "layout.html")
	fp := path.Join("templates", "/index.html")

	redirected := strings.Contains(r.Referer(), "/index/static/")
	Redirected := ""
	if redirected {
		Redirected = "Redirected"
	}

	p := Page{"Title", Source, PUID, ts, ImageName, "PROCESS", Redirected}

	RenderTemplates(w, lp, fp, p)
}

func MossaicHandler(w http.ResponseWriter, r *http.Request) {
	PUID := r.FormValue("PUID")
	ImageName := r.FormValue("ImageName")

	TileSize := r.FormValue("TileSize")
	_TileSize, err := strconv.Atoi(TileSize)
	if err != nil {
		_TileSize = 25
	}

	m := NewMossaicGenerator(PUID, ImageName, _TileSize)
	m.GenerateMossaic()

	http.Redirect(w, r, "/index/"+m.MossaicImageName()+"/"+TileSize, http.StatusFound)
}

func InputHandler(w http.ResponseWriter, r *http.Request) {
	lp := path.Join("templates", "layout.html")
	var fp string
	if r.URL.Path == "/" {
		fp = path.Join("templates", "/input.html")
	} else {
		fp = path.Join("templates", r.URL.Path+".html")
	}
	p := Page{}
	RenderTemplates(w, lp, fp, p)
}

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	file, handler, err := r.FormFile("file")
	if err != nil {
		fmt.Println(err)
	}
	data, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println(err)
	}

	DirName := UniueIdStr()
	FilePath := "./static/" + DirName
	CheckAndCreateTempDirectory(FilePath)
	//FileName := FilePath + "/" + handler.Filename
	//causes Issue if file uploaded from diff directory.
	fileName := handler.Filename[strings.LastIndex(handler.Filename, "/")+1:]
	FileName := FilePath + "/" + fileName
	err = ioutil.WriteFile(FileName, data, 0777)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	SearchText := r.FormValue("searchtext")
	if SearchText != "" {
		f := NewFlickr(DirName)
		f.SearchFlickrPhotos(SearchText)
	}

	SelectedTileSize := r.FormValue("SelectedTileSize")
	if SelectedTileSize == "" {
		SelectedTileSize = "25"
	}

	http.Redirect(w, r, "/index/"+FileName+"/"+SelectedTileSize, http.StatusFound)
}

func RenderTemplates(w http.ResponseWriter, t1 string, t2 string, page Page) {
	tmpl, err := template.ParseFiles(t1, t2)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, http.StatusText(500), 500)
		return
	}
	if err := tmpl.ExecuteTemplate(w, "layout", page); err != nil {
		log.Println(err.Error())
		http.Error(w, http.StatusText(500), 500)
	}
}
