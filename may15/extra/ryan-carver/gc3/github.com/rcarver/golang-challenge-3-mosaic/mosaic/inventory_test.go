package mosaic

import (
	"fmt"
	"image"
	"testing"

	"github.com/rcarver/golang-challenge-3-mosaic/instagram"
)

type fakeCache struct {
	store      map[ImageCacheKey]image.Image
	errorOnGet ImageCacheKey
}

func (c *fakeCache) Key(name string) ImageCacheKey {
	return ImageCacheKey(name)
}
func (c *fakeCache) Put(k ImageCacheKey, m image.Image) error {
	c.store[k] = m
	return nil
}
func (c *fakeCache) Get(k ImageCacheKey) (image.Image, error) {
	if c.errorOnGet == k {
		return nil, fmt.Errorf("error by errorOnGet %s", c.errorOnGet)
	}
	if m, ok := c.store[k]; ok {
		return m, nil
	}
	return nil, fmt.Errorf("missing %s", k)
}
func (c *fakeCache) Has(k ImageCacheKey) bool {
	_, ok := c.store[k]
	return ok
}
func (c *fakeCache) Keys() ([]ImageCacheKey, error) {
	keys := make([]ImageCacheKey, 0, len(c.store))
	for k := range c.store {
		keys = append(keys, k)
	}
	return keys, nil
}
func (c *fakeCache) Size() int {
	return len(c.store)
}

type fakeFetcher struct {
	media []*instagram.Media
}

func (f *fakeFetcher) Fetch() (chan *instagram.Media, chan struct{}) {
	ch := make(chan *instagram.Media)
	done := make(chan struct{})
	go func() {
		for _, m := range f.media {
			ch <- m
		}
	}()
	return ch, done
}

func fakeThumbnailMedia(url string) *instagram.Media {
	m := &instagram.Media{
		Type:   "image",
		Images: make(map[string]*instagram.Rep),
	}
	m.Images["thumbnail"] = instagram.NewFakeRep(url)
	return m
}

func TestImageInventory_Fetch(t *testing.T) {
	c := &fakeCache{
		store: make(map[ImageCacheKey]image.Image),
	}
	i := &ImageInventory{c}
	f := &fakeFetcher{
		media: []*instagram.Media{
			fakeThumbnailMedia("/1"),
			fakeThumbnailMedia("/2"),
			fakeThumbnailMedia("/3"),
		},
	}
	if err := i.Fetch(f, 2); err != nil {
		t.Fatalf("Fetch got error %s", err)
	}
	if got, want := c.Size(), 2; got != want {
		t.Errorf("cache.Size() got %d, want %d", got, want)
	}
	if !c.Has("/1") {
		t.Errorf("want /1")
	}
	if !c.Has("/2") {
		t.Errorf("want /2")
	}
	if c.Has("/3") {
		t.Errorf("don't want /3")
	}
}

func TestImageInventory_PopulatePalette(t *testing.T) {
	c := &fakeCache{
		store: make(map[ImageCacheKey]image.Image),
	}
	i := &ImageInventory{c}
	c.Put(ImageCacheKey("a"), image.NewRGBA(image.Rect(0, 0, 100, 100)))
	c.Put(ImageCacheKey("b"), image.NewRGBA(image.Rect(0, 0, 100, 100)))
	c.Put(ImageCacheKey("c"), image.NewRGBA(image.Rect(0, 0, 100, 100)))
	// Error on "b" to show that we keep going on error.
	c.errorOnGet = ImageCacheKey("b")
	p := NewImagePalette(5)
	if err := i.PopulatePalette(p); err != nil {
		t.Fatalf("PopulatePalette got error %s", err)
	}
	if got, want := p.NumImages(), 2; got != want {
		t.Errorf("NumImages got %d, want %d", got, want)
	}
}

func Test_fileImageCache_Key(t *testing.T) {
	c := fileImageCache{"./images"}
	key := c.Key("foo.jpg")
	want := ImageCacheKey("c1b5cbd47aa3c44f029d1140cdf1b65a591bdb2c")
	if key != want {
		t.Errorf("Key got %s, want %s", key, want)
	}
}

func Test_fileImageCache_pathsAndKeys(t *testing.T) {
	c := fileImageCache{"./images"}
	fooKey, fooPath := "foo", "./images/foo.jpg"
	path := c.keyToPath(ImageCacheKey("foo"))
	if path != fooPath {
		t.Errorf("keyToPath got %s, want %s", path, fooPath)
	}
	key := c.pathToKey(fooPath)
	if key != ImageCacheKey(fooKey) {
		t.Errorf("pathToKey got %s, want %s", key, fooKey)
	}
}
