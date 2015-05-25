package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/captncraig/mosaicChallenge/imgur"
)

func main() {
	flag.Parse()
	loadBuiltInCollections()
	go runWorkQueue()
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/oauth", imgur.HandleCallback)
	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, imgur.ImgurLoginUrl(), 302) })
	http.HandleFunc("/logout", logout)
	http.HandleFunc("/options", imgurHandler(options))
	http.HandleFunc("/submit", imgurHandler(submit))
	http.HandleFunc("/status", imgurHandler(status))
	http.HandleFunc("/", imgurHandler(home))
	log.Println("Listening on port 7777.")
	http.ListenAndServe(":7777", nil)
}

func home(w http.ResponseWriter, r *http.Request, token *imgur.ImgurAccessToken) {
	renderTemplate(w, "home", struct{ Token *imgur.ImgurAccessToken }{token})
}

func logout(w http.ResponseWriter, r *http.Request) {
	imgur.ClearImgurCookie(w)
	http.Redirect(w, r, "/", 302)
}

func options(w http.ResponseWriter, r *http.Request, token *imgur.ImgurAccessToken) {
	renderTemplate(w, "options",
		struct {
			Token       *imgur.ImgurAccessToken
			Collections map[string]*imageCollection
			Imgs        []string
		}{token, collections, sampleImages})
}

func submit(w http.ResponseWriter, r *http.Request, token *imgur.ImgurAccessToken) {
	url := r.FormValue("img")
	gallery := r.FormValue("lib")
	if url == "" || gallery == "" {
		http.Redirect(w, r, "/options", 302)
		return
	}
	var id string
	for i := 0; i < 1; i++ {
		id = createJob(url, gallery, token)
	}
	renderTemplate(w, "watch", struct {
		Token *imgur.ImgurAccessToken
		Id    string
	}{token, id})
}

func status(w http.ResponseWriter, r *http.Request, token *imgur.ImgurAccessToken) {
	id := r.FormValue("id")
	last := r.FormValue("last")
	jobMutex.RLock()
	job, ok := allJobs[id]
	jobMutex.RUnlock()
	w.Header().Add("Content-Type", "application/json")
	marshaller := json.NewEncoder(w)
	if !ok {
		marshaller.Encode(struct{ Error string }{"Job not found"})
		return
	}
	lastVersion, err := strconv.ParseInt(last, 10, 32)
	if err != nil {
		lastVersion = 0
	}

	if int(lastVersion) < job.status.Version {
		marshaller.Encode(job.status)
		return
	}
	ch := make(chan jobStatus, 1) // buffer of one so publisher never blocks, even if I timeout and forget about it.
	job.subscribe(ch)
	select {
	case st := <-ch:
		marshaller.Encode(st)
	case <-time.After(30 * time.Second):
		marshaller.Encode(struct{ Error string }{"Timeout"})
	}
}

// Special handler type that accepts an access token prepopulated by the authentication middleware
type credentialHandler func(w http.ResponseWriter, r *http.Request, token *imgur.ImgurAccessToken)

func imgurHandler(handler credentialHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tok := imgur.TokenForRequest(w, r)
		handler(w, r, tok)
	}
}

// Turns "/" into a strict match.
func rootOnly(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		handler(w, r)
	}
}
