package main

import (
	"io/ioutil"
	"net/http"
	"time"
)

// This guy handles the actual http to/from Flickr
type FlickrGetter struct {
	SaveDir string
}

var flickrClient *http.Client

func init() {
	var myTransport http.RoundTripper = &http.Transport{
		TLSHandshakeTimeout: 10 * time.Second,
		DisableKeepAlives:   false,
	}
	flickrClient = &http.Client{Transport: myTransport}
}
func (FlickrGetter) Get(url string) ([]byte, error) {
	response, err := flickrClient.Get(url)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	return contents, nil
}

func (f FlickrGetter) GetSaveDir() string {
	return f.SaveDir
}
