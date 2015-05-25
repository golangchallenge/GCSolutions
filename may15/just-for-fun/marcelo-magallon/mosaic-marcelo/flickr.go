package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
)

type FlickrPhoto struct {
	Farm     int    `json:"farm"`
	Id       string `json:"id"`
	IsFamily int    `json:"isfamily"`
	IsFriend int    `json:"isfriend"`
	IsPublic int    `json:"ispublic"`
	Owner    string `json:"owner"`
	Secret   string `json:"secret"`
	Server   string `json:"server"`
	Title    string `json:"title"`
}

type FlickrPhotosSearchResult struct {
	Photos struct {
		Page    int           `json:"page"`
		Pages   int           `json:"pages"`
		PerPage int           `json:"perpage"`
		Total   string        `json:"total"`
		Photo   []FlickrPhoto `json:"photo"`
	} `json:"photos"`
	Stat string `json:"stat"`
}

type FlickrClient struct {
	ApiKey    string
	ApiSecret string
}

const (
	flickrEndpoint = "https://api.flickr.com/services/rest/"
)

func (c *FlickrClient) NewSignedGetRequest(method string, args map[string]string) string {

	args["api_key"] = c.ApiKey
	args["method"] = method

	v := make(url.Values)

	for key, value := range args {
		v.Add(key, value)
	}

	base := "GET" + "&" + flickrEndpoint + "&" + v.Encode()

	mac := hmac.New(sha1.New, []byte(c.ApiSecret))
	mac.Write([]byte(base))
	log.Println(hex.EncodeToString(mac.Sum(nil)))

	return base
}

func (c *FlickrClient) NewGetRequest(method string, args map[string]string) string {
	args["api_key"] = c.ApiKey
	args["method"] = method

	v := make(url.Values)

	for key, value := range args {
		v.Add(key, value)
	}

	url, _ := url.Parse(flickrEndpoint)
	url.RawQuery = v.Encode()

	return url.String()
}

func (p FlickrPhoto) SmallUrl() string {
	const urlFmt = "https://farm%d.staticflickr.com/%s/%s_%s_%s.jpg"
	return fmt.Sprintf(urlFmt, p.Farm, p.Server, p.Id, p.Secret, "s")
}

func (p FlickrPhoto) ThumbnailUrl() string {
	const urlFmt = "https://farm%d.staticflickr.com/%s/%s_%s_%s.jpg"
	return fmt.Sprintf(urlFmt, p.Farm, p.Server, p.Id, p.Secret, "s")
}

func NewFlickrClient(filename string) *FlickrClient {
	config, err := os.Open(filename)
	if err != nil {
		log.Fatal("Error:", err)
	}
	defer config.Close()

	dec := json.NewDecoder(config)
	var m struct {
		FlickrApi struct {
			Secret string
			Key    string
		}
	}
	if err := dec.Decode(&m); err != nil {
		log.Fatal("Error:", err)
	}

	return &FlickrClient{
		ApiKey:    m.FlickrApi.Key,
		ApiSecret: m.FlickrApi.Secret,
	}
}

func (c *FlickrClient) FlickrPhotosSearch(args map[string]string) (FlickrPhotosSearchResult, error) {
	photosUrl := c.NewGetRequest("flickr.photos.search", args)

	resp, err := http.Get(photosUrl)
	if err != nil {
		log.Fatal("Error:", err)
	}
	defer resp.Body.Close()

	photosJson, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var photos FlickrPhotosSearchResult

	if err := json.Unmarshal(photosJson, &photos); err != nil {
		log.Fatal("E: Unmarshal error in getPhotos:", err)
	}

	if photos.Stat != "ok" {
		log.Fatal("E: Unexpected reply getPhotos:", photos.Stat)
	}

	return photos, nil
}
