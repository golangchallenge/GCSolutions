package service

import (
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	// To encode/decode png
	_ "image/png"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/rcarver/golang-challenge-3-mosaic/instagram"
	"github.com/rcarver/golang-challenge-3-mosaic/mosaic"
)

func init() {
	http.HandleFunc("/inventory", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			handleAddTagToInventory(w, r)
			return
		}
		handleGetInventory(w, r)
	})
	http.HandleFunc("/mosaics/img", handleRenderMosaic)
	http.HandleFunc("/mosaics", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			handleCreateMosaic(w, r)
			return
		}
		if r.FormValue("id") != "" {
			handleGetMosaic(w, r)
			return
		}
		handleListMosaics(w, r)

	})
}

// thumbs is the database of thumbnail images available.
var thumbs *thumbInventory

// mosaics is the database of mosaics generated.
var mosaics *mosaicInventory

var (
	// HostPort is where the server runs.
	HostPort = ":8080"
	// MosaicsDir is where the server stores mosaics.
	MosaicsDir = "./images/mosaics"
	// ThumbsDir is where the server stores thumbs.
	ThumbsDir = "./images/thumbs"
)

// Serve starts up a server. It initializes thumb and mosaic on-disk storage
// and blocks waiting for requests.
func Serve() {
	if err := os.MkdirAll(MosaicsDir, 0755); err != nil {
		log.Fatalf("Failed to create mosaics dir: %s\n", err)
	}
	if err := os.MkdirAll(ThumbsDir, 0755); err != nil {
		log.Fatalf("Failed to create thumbs dir: %s\n", err)
	}
	mosaics = &mosaicInventory{
		cache: mosaic.NewFileImageCache(MosaicsDir),
	}
	thumbs = &thumbInventory{
		tagCacheFunc: func(tag string) mosaic.ImageCache {
			path := path.Join(ThumbsDir, tag)
			if err := os.MkdirAll(path, 0755); err != nil {
				log.Fatalf("Failed to create cache dir: %s\n", err)
			}
			return mosaic.NewFileImageCache(path)
		},
		api:    instagram.NewClient(),
		images: make(map[string]*mosaic.ImageInventory),
		states: make(map[string]chan bool),
	}

	log.Fatal(http.ListenAndServe(HostPort, nil))
}

// GET /mosaics
// List all mosaics that have been created.

func newMosaicRes(m *mosaicRecord) *mosaicRes {
	return &mosaicRes{
		OK:     true,
		ID:     string(m.ID),
		Tag:    m.Tag,
		Status: m.Status,
		URL:    fmt.Sprintf("/mosaics?id=%s", m.ID),
		ImgURL: fmt.Sprintf("/mosaics/img?id=%s", m.ID),
	}
}

type mosaicsListRes struct {
	OK      bool         `json:"ok"`
	Mosaics []*mosaicRes `json:"mosaics"`
}

type mosaicRes struct {
	OK     bool   `json:"ok"`
	ID     string `json:"id"`
	Tag    string `json:"tag"`
	Status string `json:"status"`
	URL    string `json:"url"`
	ImgURL string `json:"img"`
}

func handleListMosaics(w http.ResponseWriter, r *http.Request) {
	res := &mosaicsListRes{
		true,
		make([]*mosaicRes, mosaics.Size()),
	}
	for i, m := range mosaics.List() {
		res.Mosaics[i] = newMosaicRes(m)
	}
	respondOK(w, res)
}

// POST /mosaics?tag=<tag> img=<FILE>
// Create a new mosaic.

var (
	// Units is how many mosaic units to use for width and height.
	Units = 40
	// UnitSize is how big the thumbnail images are, width and height.
	UnitSize    = 150
	paletteSize = 256
)

func handleCreateMosaic(w http.ResponseWriter, r *http.Request) {
	// Read tag.
	tag := r.FormValue("tag")
	if tag == "" {
		respondErr(w, http.StatusBadRequest, "missing 'tag' param")
		return
	}
	ch := thumbs.AddTag(tag)

	// Read image upload.
	fi, _, err := r.FormFile("img")
	if err != nil {
		respondErr(w, http.StatusBadRequest, "upload failed")
	}
	defer fi.Close()
	in, _, err := image.Decode(fi)
	if err != nil {
		respondErr(w, http.StatusBadRequest, "image parsing failed")
	}

	// Create a record to track the mosaic.
	m, err := mosaics.Create(tag)
	if err != nil {
		respondErr(w, http.StatusBadRequest, "failed to create record")
		return
	}

	// Generate mosaic offline.
	go func() {
		log.Printf("Waiting for tags...\n")
		<-ch
		log.Printf("Tags are ready...\n")
		generateMosaic(tag, in, m)
	}()

	// Respond immediately.
	res := newMosaicRes(m)
	respondOK(w, res)
}

func generateMosaic(tag string, in image.Image, m *mosaicRecord) {

	// First build a color palette.
	log.Printf("Mosaic[%s] Create Palette...", m.ID)
	p := mosaic.NewImagePalette(paletteSize)
	if err := thumbs.PopulatePalette(tag, p); err != nil {
		log.Printf("Failed to populate palette: %s", err)
		if err := mosaics.SetStatus(m.ID, MosaicStatusFailed); err != nil {
			log.Printf("Failed to set mosaic failed: %s", err)
		}
		return
	}
	log.Printf("Mosaic[%s] Create Palette Done with %d colors, %d images.", m.ID, p.NumColors(), p.NumImages())

	// Update that the mosaic is being worked on.
	if err := mosaics.SetStatus(m.ID, MosaicStatusWorking); err != nil {
		log.Printf("Failed to set working status: %s", err)
		return
	}

	// Generate the mosaic.
	log.Printf("Mosaic[%s] Compose...", m.ID)
	out := mosaic.ComposeSquare(in, Units, UnitSize, p)
	log.Printf("Mosaic[%s] Compose Done.", m.ID)

	// Store the image and update the the mosaic is done.
	if err := mosaics.StoreImage(m.ID, out); err != nil {
		log.Printf("Failed to store mosaic image: %s", err)
		return
	}
	if err := mosaics.SetStatus(m.ID, MosaicStatusCreated); err != nil {
		log.Printf("Failed to set mosaic created: %s", err)
		return
	}
}

// GET /mosaics?id=<id>
// Get information about a mosaic.

func handleGetMosaic(w http.ResponseWriter, r *http.Request) {
	// Read id.
	id := r.FormValue("id")
	if id == "" {
		respondErr(w, http.StatusBadRequest, "missing 'id' param")
		return
	}
	m, err := mosaics.Get(mosaicID(id))
	if err != nil {
		respondErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	if m == nil {
		http.NotFound(w, r)
		return
	}
	res := newMosaicRes(m)
	respondOK(w, res)
}

// GET /mosaics/img?id=<id>
// Get a mosaic image that was created.

func handleRenderMosaic(w http.ResponseWriter, r *http.Request) {
	// Read id.
	id := r.FormValue("id")
	if id == "" {
		respondErr(w, http.StatusBadRequest, "missing 'id' param")
		return
	}
	img, err := mosaics.GetImage(mosaicID(id))
	if err != nil {
		respondErr(w, http.StatusInternalServerError, "getting image")
	}
	if img == nil {
		http.NotFound(w, r)
		return
	}
	err = jpeg.Encode(w, img, nil)
	if err != nil {
		respondErr(w, http.StatusInternalServerError, "writing jpeg")
	}
	w.Header().Set("Content-Type", "image/jpg")
}

// POST /inventory?tag=<tag>
// Add images to the thumbnails inventory.

var (
	fetchImages = 100
)

type inventoryReq struct {
	OK bool `json:"ok"`
}

func handleAddTagToInventory(w http.ResponseWriter, r *http.Request) {
	tag := r.FormValue("tag")
	if tag == "" {
		respondErr(w, http.StatusBadRequest, "missing 'tag' param")
		return
	}
	thumbs.AddTag(tag)
	res := &inventoryReq{true}
	respondOK(w, res)
}

// GET /inventory
// Get information about the thumbnails inventory.

type inventoryRes struct {
	OK bool `json:"ok"`
	// Map of tag to images count.
	Images []inventoryImageRes `json:"images"`
}

type inventoryImageRes struct {
	Tag   string `json:"tag"`
	Count int    `json:"count"`
}

func handleGetInventory(w http.ResponseWriter, r *http.Request) {
	res := &inventoryRes{
		true,
		make([]inventoryImageRes, 0),
	}
	for tag, num := range thumbs.Contents() {
		res.Images = append(res.Images, inventoryImageRes{
			Tag:   tag,
			Count: num,
		})
	}
	respondOK(w, res)
}

// Utils

type errorRes struct {
	OK     bool   `json:"ok"`
	Reason string `json:"reason"`
}

func respondOK(w http.ResponseWriter, data interface{}) {
	js, err := json.Marshal(data)
	if err != nil {
		respondErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func respondErr(w http.ResponseWriter, code int, msg string) {
	data := errorRes{false, msg}
	js, err := json.Marshal(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(js)
}
