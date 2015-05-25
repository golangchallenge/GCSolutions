package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"
	"text/template"
	"time"
)

type Page struct {
	Title     string
	Body      []byte
	ImageName string
}

type CreateResponse struct {
	Success bool
	Name    string
}

var templates = template.Must(template.ParseFiles("index.html", "done.html"))
var logger *log.Logger

func init() {
	logOut, err := os.OpenFile("mosaic.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Println("error creating log file: ", err)
	}
	logger = log.New(logOut, "", log.LstdFlags)
}

func main() {
	runtime.GOMAXPROCS(2)
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("static/css"))))
	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("static/js"))))
	http.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("static/images"))))
	http.Handle("/fonts/", http.StripPrefix("/fonts/", http.FileServer(http.Dir("static/fonts"))))
	http.Handle("/created/", http.StripPrefix("/created/", http.FileServer(http.Dir("created"))))

	http.HandleFunc("/bm", createHandler)
	http.HandleFunc("/d/", doneHandler)
	http.HandleFunc("/", indexHandler)
	http.ListenAndServe(":80", nil)
}

// load the index page
func indexHandler(w http.ResponseWriter, r *http.Request) {
	p := Page{Title: "Mo!"}
	renderTemplate(w, "index", &p)
}

// handle form submission, kick of creation process
func createHandler(w http.ResponseWriter, r *http.Request) {
	file, header, err := r.FormFile("file")
	if !strings.Contains(header.Header["Content-Type"][0], "image/") {
		failedResponse(w, "This does not appear to be a valid image file")
		return
	}
	if err != nil {
		logger.Println("error reading form file: ", err)
		failedResponse(w, "Mo is having issues with your file :(")
		return
	}

	defer file.Close()
	hash := md5.New()
	fileName := time.Now().Format(time.RFC3339) + "_" + header.Filename
	out, err := os.Create("uploads/" + fileName)
	if err != nil {
		failedResponse(w, "Unable to create the file for writing")
		return
	}

	defer out.Close()

	// write the content from POST to the file and md5sum
	_, err = io.Copy(out, io.TeeReader(file, hash))
	if err != nil {
		fmt.Fprintln(w, err)
	}
	md5sum := hex.EncodeToString(hash.Sum(nil))

	submitPic, err := initMosaic(fileName)
	if err != nil {
		failedResponse(w, "Sadly your image is too big, please shrink it down to a max of 1200px by 1200px and resubmit :(")
		return
	}
	flickrGetter := FlickrGetter{SaveDir: "fromFlickr"}
	buildChan := make(chan bool, 1)
	buildSuccess := false
	if r.FormValue("searchTerms") != "" {
		go func() {
			buildMosaic(flickrGetter, submitPic, r.FormValue("searchTerms"), buildChan)
		}()
	} else {
		go func() {
			buildMosaic(flickrGetter, submitPic, "", buildChan)
		}()
	}

	select {
	case res := <-buildChan:
		buildSuccess = res
	case <-time.After(time.Minute * 5):
		logger.Println("timeout on mosaic build")
		failedResponse(w, "Timed out looking for images :(")
		return
	}
	if !buildSuccess {
		failedResponse(w, "There were errors creating your mosaic :(")
		return
	}
	success, path := writeMosaic(flickrGetter, submitPic, md5sum)
	if !success {
		failedResponse(w, "Unable to write your mosaic to disk :(")
	}
	goodResp := CreateResponse{Success: true, Name: path}
	b, err := json.Marshal(goodResp)
	if err != nil {
		logger.Println("unable to marshal success response: ", err)
	}
	logger.Printf("%v\n", string(b))
	fmt.Fprintf(w, string(b))

}

// handler for displaying completed mosaics
func doneHandler(w http.ResponseWriter, r *http.Request) {
	imgName := r.URL.Path[len("/d/"):]
	p := Page{ImageName: imgName}
	renderTemplate(w, "done", &p)
}

// send back failures
func failedResponse(w http.ResponseWriter, resp string) {
	badResp := CreateResponse{Success: false, Name: resp}
	b, err := json.Marshal(badResp)
	if err != nil {
		logger.Println("unable to marshal failed response: ", err)
	}
	logger.Printf("%v\n", string(b))
	fmt.Fprintf(w, string(b))
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
