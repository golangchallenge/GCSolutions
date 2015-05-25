package service

import (
	"fmt"
	"image"
	"log"
	"sync"

	"github.com/rcarver/golang-challenge-3-mosaic/instagram"
	"github.com/rcarver/golang-challenge-3-mosaic/mosaic"
)

const (
	// MosaicStatusNew is a mosaic that has been stored but not worked on.
	MosaicStatusNew = "new"
	// MosaicStatusWorking is a mosaic that is being generated.
	MosaicStatusWorking = "working"
	// MosaicStatusFailed is a mosaic that failed to be generated.
	MosaicStatusFailed = "failed"
	// MosaicStatusCreated is a mosaic that was generated.
	MosaicStatusCreated = "created"
)

var (
	// ImagesPerTag is how many images to download when populating a tag.
	ImagesPerTag    = 1000
	mosaicIDCounter = 0
)

// Inventory of mosaics that have been created.
type mosaicInventory struct {
	cache   mosaic.ImageCache
	mosaics []*mosaicRecord
}

type mosaicID string

type mosaicRecord struct {
	ID     mosaicID
	Tag    string
	Status string
}

func (i *mosaicInventory) Create(tag string) (*mosaicRecord, error) {
	mosaicIDCounter++
	id := mosaicID(fmt.Sprintf("%d", mosaicIDCounter))
	d := &mosaicRecord{
		ID:     id,
		Tag:    tag,
		Status: MosaicStatusNew,
	}
	i.mosaics = append(i.mosaics, d)
	return d, nil
}

func (i *mosaicInventory) Get(id mosaicID) (*mosaicRecord, error) {
	for _, m := range i.mosaics {
		if m.ID == id {
			return m, nil
		}
	}
	return nil, nil
}

func (i *mosaicInventory) SetStatus(id mosaicID, status string) error {
	for _, d := range i.mosaics {
		if d.ID == id {
			d.Status = status
			break
		}
	}
	return nil
}

func (i *mosaicInventory) StoreImage(id mosaicID, m image.Image) error {
	key := i.cache.Key(string(id))
	if err := i.cache.Put(key, m); err != nil {
		return err
	}
	return nil
}

func (i *mosaicInventory) GetImage(id mosaicID) (image.Image, error) {
	key := i.cache.Key(string(id))
	return i.cache.Get(key)
}

func (i *mosaicInventory) Size() int {
	return len(i.mosaics)
}

func (i *mosaicInventory) List() []*mosaicRecord {
	return i.mosaics
}

type tagCacheFunc func(string) mosaic.ImageCache

// thumbInventory tracks the thumbnails that have been acquired.
type thumbInventory struct {
	tagCacheFunc
	api instagram.Client

	mu     sync.Mutex
	images map[string]*mosaic.ImageInventory
	states map[string]chan bool
}

func (i *thumbInventory) AddTag(tag string) chan bool {
	i.mu.Lock()
	defer i.mu.Unlock()

	// Initialize the inventory.
	if _, ok := i.images[tag]; !ok {
		cache := i.tagCacheFunc(tag)
		i.images[tag] = mosaic.NewImageInventory(cache)

	}
	inv := i.images[tag]

	// Initialize the state.
	if ch, ok := i.states[tag]; ok {
		log.Printf("AddTag(%s) already has it\n", tag)
		return ch
	}
	i.states[tag] = make(chan bool)

	log.Printf("AddTag(%s) beginning fetch\n", tag)
	go func() {
		fetcher := instagram.NewTagFetcher(i.api, tag)
		if err := inv.Fetch(fetcher, ImagesPerTag); err != nil {
			log.Printf("Failed to fetch tag %s: %s", tag, err)
		}
		close(i.states[tag])
	}()

	return i.states[tag]
}

func (i *thumbInventory) PopulatePalette(tag string, p *mosaic.ImagePalette) error {
	inventory, ok := i.images[tag]
	if !ok {
		return nil
	}
	return inventory.PopulatePalette(p)

}

func (i *thumbInventory) Contents() map[string]int {
	res := make(map[string]int)
	for tag, inv := range i.images {
		res[tag] = inv.Size()
	}
	return res
}
