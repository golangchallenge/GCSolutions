package image

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

const (
	maxConcurrentDownloads int           = 10
	httpClientTimeout      time.Duration = 10 * time.Second
)

// Discover is responsible for searching and downloading images.
type Discover struct {
	provider Provider
	client   *http.Client
}

// Search searches images for a given query with the current Provider.
func (i Discover) Search(q string, page int) (*SearchResult, error) {
	return i.provider.Search(q, page)
}

// DownloadFile downloads a file from given sourde and saves it at a given destination.
func (i Discover) DownloadFile(src, dest string) (string, error) {

	if fileExists(dest) {
		return dest, nil
	}

	file, err := os.Create(dest)
	defer file.Close()
	if err != nil {
		return "", err
	}

	resp, err := i.client.Get(src)
	defer resp.Body.Close()
	if err != nil {
		return "", err
	}

	_, errCopy := io.Copy(file, resp.Body)
	if errCopy != nil {
		return "", err
	}

	return dest, nil
}

// Fetch downloads all images of a given slice of sources to a destination dir.
func (i Discover) Fetch(srcs []string, destDir string) error {
	tasks := make(chan string, 64)

	// spawn four worker goroutines
	var wg sync.WaitGroup
	for j := 0; j < maxConcurrentDownloads; j++ {
		wg.Add(1)
		go func() {
			for src := range tasks {
				dest := filepath.Join(destDir, GenerateFilename(src))
				i.DownloadFile(src, dest)
			}
			wg.Done()
		}()
	}

	// generate some tasks
	for _, src := range srcs {
		tasks <- src
	}
	close(tasks)

	// wait for the workers to finish
	wg.Wait()

	return nil
}

// Provider provides the actual implementation of searching images with different services.
type Provider interface {
	Search(q string, page int) (*SearchResult, error)
}

// SearchResult represents the result of an image search.
type SearchResult struct {
	Page   int
	Pages  int
	Images []string
}

// FlickrProvider implements an image provider for the Flickr API.
type FlickrProvider struct {
	key, secretKey, baseURL string
}

// FlickrSearchResult represents the result of an image search with the FlickrProvider.
type FlickrSearchResult struct {
	Photos struct {
		Page  int `json:"page"`
		Pages int `json:"pages"`
		Photo []struct {
			ID     string `json:"id"`
			Farm   int    `json:"farm"`
			Server string `json:"server"`
			Secret string `json:"secret"`
		} `json:"photo"`
	} `json:"photos"`
}

// Search searches images with the Flickr API.
func (p FlickrProvider) Search(q string, page int) (*SearchResult, error) {
	params := p.getSearchParams(q, page)
	sURL, err := formatURL(p.baseURL, params)
	if err != nil {
		return nil, err
	}
	data, err := callAPI(sURL)
	if err != nil {
		return nil, err
	}
	result := p.parseSearchResult(data)

	return result, nil
}

// parseSearchResult genrates the generic SearchResult from FlickrSearchResult.
func (p FlickrProvider) parseSearchResult(data []byte) *SearchResult {
	result := &SearchResult{}
	jsn := new(FlickrSearchResult)
	errJsn := json.Unmarshal(data, &jsn)
	if errJsn != nil {
		panic(errJsn.Error())
	}

	result.Page = jsn.Photos.Page
	result.Pages = jsn.Photos.Pages

	size := "s"
	for _, photo := range jsn.Photos.Photo {
		result.Images = append(result.Images, p.generateImageURL(
			photo.ID,
			photo.Farm,
			photo.Server,
			photo.Secret,
			size,
		))
	}

	return result
}

func (p FlickrProvider) generateImageURL(id string, farm int, server string, secret string, size string) string {
	return fmt.Sprintf("https://farm%d.staticflickr.com/%s/%s_%s_%s.jpg", farm, server, id, secret, size)
}

func (p FlickrProvider) getSearchParams(q string, page int) map[string]string {
	return map[string]string{
		"method":         "flickr.photos.search",
		"api_key":        p.key,
		"format":         "json",
		"media":          "photos",
		"nojsoncallback": "1",
		"text":           q,
		"page":           strconv.Itoa(page),
	}
}

func formatURL(srcURL string, params map[string]string) (string, error) {
	baseURL, err := url.Parse(srcURL)
	if err != nil {
		return "", err
	}

	urlParams := url.Values{}
	for key, val := range params {
		urlParams.Add(key, val)
	}
	baseURL.RawQuery = urlParams.Encode()

	return baseURL.String(), nil
}

func callAPI(reqURL string) ([]byte, error) {
	var body []byte
	res, err := http.Get(reqURL)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	body, errRead := ioutil.ReadAll(res.Body)
	if errRead != nil {
		return nil, errRead
	}

	return body, nil
}

// GenerateFilename generates a valid filename from a given string using a md5 hash.
func GenerateFilename(src string) string {
	hasher := md5.New()
	hasher.Write([]byte(src))
	name := hex.EncodeToString(hasher.Sum(nil))
	filename := name + path.Ext(src)
	return filename
}

// GenerateDirname generates a valid directory from a given string using a md5 hash.
func GenerateDirname(src string) string {
	hasher := md5.New()
	hasher.Write([]byte(src))
	filename := hex.EncodeToString(hasher.Sum(nil))
	return filename
}

func fileExists(filename string) bool {
	if _, err := os.Stat(filename); !os.IsNotExist(err) {
		return true
	}
	return false
}

// NewDiscover creates a new Discover pointer.
func NewDiscover(provider Provider) *Discover {
	client := http.Client{
		Timeout: httpClientTimeout,
	}
	return &Discover{
		provider: provider,
		client:   &client,
	}
}

// NewFlickrProvider creates a new FlickrProvider pointer.
func NewFlickrProvider(key, secretKey string) *FlickrProvider {
	baseURL := "https://api.flickr.com/services/rest/"
	return &FlickrProvider{
		key:       key,
		secretKey: secretKey,
		baseURL:   baseURL,
	}
}

// Open a file at the given source, decodes the image and returns a pointer.
func Open(src string) (image.Image, error) {
	if !fileExists(src) {
		return nil, fmt.Errorf("Unable to open image, file does not exist: %s", src)
	}
	file, err := os.Open(src)
	defer file.Close()
	if err != nil {
		return nil, err
	}

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	return img, nil
}
