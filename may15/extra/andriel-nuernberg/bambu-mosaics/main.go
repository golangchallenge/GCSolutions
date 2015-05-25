package main

import (
	"fmt"
	"log"
	"os"

	"net/http"
)

func main() {
	log.Println("Loading tile set...")
	ts, err := NewTileSet("tiles")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	log.Println("Loading app...")
	app := NewApp(ts)

	log.Println("Loading router...")
	r := NewRouter()

	r.Get("/", http.HandlerFunc(app.HomeHandler))
	r.Post("/m", http.HandlerFunc(app.PostMosaicHandler))
	r.Get("/m/:id", http.HandlerFunc(app.GetMosaicHandler))
	http.Handle("/mosaics/", http.StripPrefix("/mosaics/", http.FileServer(http.Dir("mosaics"))))
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))

	log.Println("App up and running on port 4000!")
	r.Listen(4000)
}
