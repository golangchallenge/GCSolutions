package image

import (
	"path"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/ptrost/mosaic2go/config"
	"github.com/ptrost/mosaic2go/test"
)

func TestFormatURL(t *testing.T) {
	params := map[string]string{
		"param1": "val1",
		"param2": "val2",
	}
	result, err := formatURL("http://example.com/path", params)
	expected := "http://example.com/path?param1=val1&param2=val2"

	test.AssertNotErr("formatURL", err, t)
	test.Assert("formatURL", expected, result, t)
}

func TestFlickrProviderSearch(t *testing.T) {
	cfg := getConfig()
	p := NewFlickrProvider(cfg.Get("flickr_key"), cfg.Get("flickr_secret_key"))
	result, err := p.Search("nature", 1)

	test.AssertNotErr("FlickrProvider.Search()", err, t)
	test.AssertNot("FlickrProvider.Search() SearchResult.Pages", 0, result.Pages, t)
}

func TestDiscoverDownloadFile(t *testing.T) {
	cfg := getConfig()
	dest := path.Join(getRootDir(), cfg.Get("tmp_dir"), "downloadTest.jpg")
	p := NewFlickrProvider(cfg.Get("flickr_key"), cfg.Get("flickr_secret_key"))
	img := NewDiscover(p)
	filename, err := img.DownloadFile("https://farm6.staticflickr.com/5337/17095007133_61efedd70b_z.jpg", dest)

	test.AssertNotErr("Discover.DownloadFile", err, t)
	test.AssertNot("Discover.DownloadFile", "", filename, t)
}
func TestDiscoverFetch(t *testing.T) {
	cfg := getConfig()
	p := NewFlickrProvider(cfg.Get("flickr_key"), cfg.Get("flickr_secret_key"))
	img := NewDiscover(p)
	input := []string{
		"https://farm9.staticflickr.com/8835/17528122588_21ba7d7e81_z.jpg",
		"https://farm6.staticflickr.com/5339/17528119008_6884c9c7d6_z.jpg",
		"https://farm8.staticflickr.com/7698/17093363964_66691a273c_z.jpg",
		"https://farm6.staticflickr.com/5468/17713312732_111dfc51f3_z.jpg",
		"https://farm8.staticflickr.com/7656/17529543009_bcc794970a_z.jpg",
		"https://farm6.staticflickr.com/5350/17528765239_95c6cc1ef1_z.jpg",
		"https://farm9.staticflickr.com/8826/17527976638_afe3363dba_z.jpg",
		"https://farm6.staticflickr.com/5457/17529458929_295f03b6bb_z.jpg",
	}

	destDir := path.Join(getRootDir(), cfg.Get("tmp_dir"), "test")
	img.Fetch(input, destDir)
}

func getConfig() *config.Config {
	cfg := config.New(filepath.Join(getRootDir(), "config.json"))
	return cfg
}

func getRootDir() string {
	_, currentfile, _, _ := runtime.Caller(1)
	return path.Dir(filepath.Join(currentfile, "../"))
}
