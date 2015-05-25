package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"sync"
)

//var client *Flickr

const (
	endpoint                      = "https://api.flickr.com/services/rest/?"
	uploadEndpoint                = "https://api.flickr.com/services/upload/"
	replaceEndpoint               = "https://api.flickr.com/services/replace/"
	apiHost                       = "api.flickr.com"
	flickr_groups_search          = "flickr.groups.search"
	flickr_groups_pools_getPhotos = "flickr.groups.pools.getPhotos"
)

type Flickr struct {
	HttpClient     *http.Client
	PhotoDirectory string
}

func NewFlickr(Dir string) *Flickr {
	tr := &http.Transport{
		TLSClientConfig:    &tls.Config{},
		DisableCompression: true,
	}
	httpClient := &http.Client{Transport: tr}
	return &Flickr{httpClient, Dir}
}

func (f *Flickr) PrepareQuery(parameters map[string]string) string {
	//patach parameters
	parameters["api_key"] = Config[API_KEY_ID]
	parameters["format"] = "json"
	parameters["nojsoncallback"] = "1"

	i := 0
	s := bytes.NewBuffer(nil)
	for k, v := range parameters {
		if i != 0 {
			s.WriteString("&")
		}
		i++
		s.WriteString(k + "=" + url.QueryEscape(v))
	}
	return s.String()
}

func (f *Flickr) SearchFlickrPhotos(searchtext string) error {
	//const CntsearchImagesInGroups = "method=%s&api_key=%s&tags=pattern&format=json&nojsoncallback=1"

	Paramerter := make(map[string]string)
	Paramerter["method"] = "flickr.photos.search"
	Paramerter["tags"] = searchtext
	Paramerter["page"] = "1"
	Paramerter["per_page"] = "15"

	queryString := f.PrepareQuery(Paramerter)
	urlstring := fmt.Sprintf("%s%s", endpoint, queryString)

	res, err := f.HttpClient.Get(urlstring)
	if err != nil {
		fmt.Println("Flickr failed to return images..:" + err.Error())
		return err
	}
	defer res.Body.Close()

	poolPhotos := &AllFlickrPool{}
	err = json.NewDecoder(res.Body).Decode(poolPhotos)
	if err != nil {
		fmt.Println(err.Error())
	}

	var wg sync.WaitGroup
	NosofGoRoutines := len(poolPhotos.Groups.PoolPhotos)
	wg.Add(NosofGoRoutines)

	for _, p := range poolPhotos.Groups.PoolPhotos {
		go f.GetFlickrPhotosWg(p, &wg)
	}
	wg.Wait()
	return nil
}

func (f *Flickr) GetFlickrPhotosWg(photo FlickPoolPhoto, wg *sync.WaitGroup) {
	defer wg.Done()

	//const getPhotos = "https://farm%d.staticflickr.com/%s/%s_%s_s.jpg"
	const getPhotos = "https://farm%d.staticflickr.com/%s/%s_%s.jpg"
	urlstring := fmt.Sprintf(getPhotos, photo.Farm, photo.Server, photo.Id, photo.Secret)

	res, err := f.HttpClient.Get(urlstring)
	if err != nil {
		return
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}

	photopath := "./static/" + f.PhotoDirectory + "/" + photo.Id + ".jpg"
	fp, _ := os.Create(photopath)
	fp.Write(body)
	fp.Close()
}
