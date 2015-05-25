package instagram

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strings"
)

const (
	instagramURL      = "https://api.instagram.com/v1%s"
	instagramClientID = "6b2ea4cc0093441fb38990045a855e2a"
	instagramSecuret  = "ea785b48dd014eaeb4fd97c0a23d6ae5"
)

const (
	// ThumbnailSize is the width/height of the ThumbnailImage representation.
	ThumbnailSize = 150
)

// Client is what talks to the Instagram API.
type Client interface {
	// Popular calls the Instagram Popular API and returns the data.
	Popular() (*MediaList, error)

	// Search calls the Instagram Search API and returns the data.
	Search(lat, lng string) (*MediaList, error)

	// Tagged calls the Instagram Tagged API and returns the data.
	Tagged(tag, maxTagID string) (*MediaList, error)
}

// Client makes requests to Instagram.
type apiClient struct {
	BaseURL string
	URLSigner
}

// NewClient creates an initialized Client.
func NewClient() Client {
	return &apiClient{
		BaseURL:   instagramURL,
		URLSigner: clientSecretSigner{instagramClientID, instagramSecuret},
	}
}

// MediaList is a result set containing media.
type MediaList struct {
	Media      []Media `json:"data"`
	Pagination `json:"pagination"`
}

// Pagination provides details for getting the next set of records.
type Pagination struct {
	NextURL  string `json:"next_url"`
	MaxTagID string `json:"next_max_tag_id"`
}

// Media is either a photo or video. If it's a video, it has both Images and
// Videos representations. If it's a photo, it only has Images representations.
type Media struct {
	Type   string          `json:"type"`
	Images map[string]*Rep `json:"images"`
	Videos map[string]*Rep `json:"videos"`
}

// IsPhoto tells you if this is a photo. If it's not, it's a video.
func (m Media) IsPhoto() bool {
	return m.Type == "image"
}

// StandardImage returns the standard resolution image representation.
func (m Media) StandardImage() *Rep {
	return m.Images["standard_resolution"]
}

// ThumbnailImage returns the thumbnail resolution image representation.
func (m Media) ThumbnailImage() *Rep {
	return m.Images["thumbnail"]
}

// Rep is a JPG representation of an image or video, located at a URL and at a
// specific width and height.
type Rep struct {
	URL     string `json:"url"`
	Width   uint   `json:"width"`
	Height  uint   `json:"height"`
	fetched bool
	body    io.ReadCloser
	code    int
}

type nopCloser struct {
	io.Reader
}

func (nopCloser) Close() error { return nil }

// NewFakeRep initializes a Rep with a URL and a fake image. This allows you to
// fake a Rep without hitting the network.
func NewFakeRep(url string) *Rep {
	m := image.NewRGBA(image.Rect(0, 0, 100, 100))
	imageData := bytes.NewBuffer([]byte{})
	jpeg.Encode(imageData, m, nil)
	return NewFakeRepWithImage(url, imageData)
}

// NewFakeRepWithImage initializes a Rep with a URL and image data. This allows
// you to fake a Rep without hitting the network, while using a real image.
func NewFakeRepWithImage(url string, imageData io.Reader) *Rep {
	return &Rep{
		URL:     url,
		fetched: true,
		code:    http.StatusOK,
		body:    nopCloser{imageData},
	}
}

// Read implements io.Reader to fetch the JPG data.
func (r *Rep) Read(p []byte) (int, error) {
	if !r.fetched {
		r.fetched = true
		res, err := http.Get(r.URL)
		if err != nil {
			return 0, err
		}
		r.code = res.StatusCode
		if r.code == http.StatusOK {
			r.body = res.Body
		}
	}
	if r.body != nil {
		return r.body.Read(p)
	}
	return 0, fmt.Errorf("failed to fetch data, response was code %d", r.code)
}

// Image returns an image object from the JPG.
func (r *Rep) Image() (image.Image, error) {
	var buf bytes.Buffer
	if c, err := buf.ReadFrom(r); err != nil || c == 0 {
		return nil, err
	}
	return jpeg.Decode(&buf)
}

func (c apiClient) Popular() (*MediaList, error) {
	var m MediaList
	params := map[string]string{
		"count": "100",
	}
	url := c.formatURL("/media/popular", params)
	err := c.getJSON(url, &m)
	return &m, err
}

func (c apiClient) Search(lat, lng string) (*MediaList, error) {
	var m MediaList
	params := map[string]string{
		"lat":   lat,
		"lng":   lng,
		"count": "100",
	}
	url := c.formatURL("/media/search", params)
	err := c.getJSON(url, &m)
	return &m, err
}

func (c apiClient) Tagged(tag, maxTagID string) (*MediaList, error) {
	var m MediaList
	params := map[string]string{
		"count":      "100",
		"max_tag_id": maxTagID,
	}
	endpoint := fmt.Sprintf("/tags/%s/media/recent", tag)
	url := c.formatURL(endpoint, params)
	err := c.getJSON(url, &m)
	return &m, err
}

// getJSON calls a URL and marshals the resulting JSON into the data struct. If
// the response is anything but 200 an error is returned.
func (c apiClient) getJSON(url string, data interface{}) error {
	res, err := http.Get(url)
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("Call to %s failed, status %d", url, res.StatusCode)
	}
	// TODO use streaming unmarshal
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, &data)
	if err != nil {
		return err
	}
	return nil
}

// formatURL combines query parameters to an endpoint, then signs the URL.
func (c apiClient) formatURL(endpoint string, params map[string]string) string {
	u, err := url.Parse(fmt.Sprintf(c.BaseURL, endpoint))
	if err != nil {
		panic("failed to parse instagram base url")
	}
	// Add custom params to the query string.
	q := u.Query()
	for k, v := range params {
		q.Set(k, v)
	}

	// Sign the URL.
	c.URLSigner.Sign(endpoint, &q)

	// Set new query string and stringify.
	u.RawQuery = q.Encode()
	return u.String()
}

// URLSigner is the interface for things that modify a query string for
// security purposes.
type URLSigner interface {
	Sign(endpoint string, queryString *url.Values)
}

// clientSecretSigner implements URLSigner, it adds the client_id and sig
// params to a query string.
type clientSecretSigner struct {
	ClientID string
	Secret   string
}

// Sign implements URLSigner.
func (s clientSecretSigner) Sign(endpoint string, q *url.Values) {
	q.Set("client_id", s.ClientID)
	q.Set("sig", sig(s.Secret, endpoint, *q))
}

// sig calculates the signature for an Instagram URL.
// https://instagram.com/developer/secure-api-requests/
func sig(secret, endpoint string, params url.Values) string {
	// Parts is made up of the endpoint and params sorted by key.
	parts := make([]string, 0, len(params)+1)
	parts = append(parts, endpoint)

	// Get the sorted keys.
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Sort(sort.StringSlice(keys))

	// Accumulate sorted key/value pairs.
	for _, k := range keys {
		// TODO: use url.Values#Get?
		join := fmt.Sprintf("%s=%s", k, params[k][0])
		parts = append(parts, join)
	}

	// Join parts and sign it.
	sig := strings.Join(parts, "|")

	// Calculate the sha256 hexdigest.
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(sig))
	sum := mac.Sum(nil)
	return hex.EncodeToString(sum)
}
