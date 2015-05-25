// Fetches image metadata from the Instagram Search API
// https://api.instagram.com/v1/tags/hashtag/media/recent

package fetcher

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type ImageMetaData struct {
	IgId            string
	FullSizeUrl     string
	SmallSizeUrl    string
	SmallDimensions struct {
		width  int
		height int
	}
}

type InstagramSearchResponse struct {
	Data       []InstagramMedia `json:"data"`
	Pagination struct {
		NextUrl string `json:"next_url"`
	} `json:"pagination"`
}

type InstagramMedia struct {
	Type   string `json:"type"`
	Images struct {
		Thumbnail struct {
			Url    string `json:"url"`
			Width  int    `json:"width"`
			Height int    `json:"height"`
		} `json:"thumbnail"`
		StandardResolution struct {
			Url string `json:"url"`
		} `json:"standard_resolution"`
	} `json:"images"`
	Id string `json:"id"`
}

type InstagramClient struct {
	Id string
}

func NewInstagramClient(clientID string) InstagramClient {
	client := InstagramClient{clientID}
	return client
}

// Fetches a given number of items from Instagram
func (client InstagramClient) Search(totalImageCount int, hashtag string, c chan *ImageMetaData) {
	remainingImageCount := totalImageCount

	recentMedia := "https://api.instagram.com/v1/tags/%s/media/recent?client_id=%s&count=100"

	url := fmt.Sprintf(recentMedia, hashtag, client.Id)
	var count int
	for remainingImageCount > 0 {
		url, count = InstagramQuery(url, remainingImageCount, c)
		remainingImageCount -= count
	}
}

// Fetches one page of results from recently tagged media
func InstagramQuery(url string, remainingImageCount int, c chan *ImageMetaData) (string, int) {
	resp, err := http.Get(url)
	if err != nil {
		log.Println("Failed to search instagram", err)
	}

	// Read the whole response from instagram, ignore the error
	body, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	var searchResponse InstagramSearchResponse
	err = json.Unmarshal(body, &searchResponse)
	if err != nil {
		log.Println("Couldn't decode instagram search response", err)
	}

	count := 0
	for _, im := range searchResponse.Data {
		if im.Type == "image" {
			var metadata ImageMetaData
			metadata.IgId = im.Id
			metadata.FullSizeUrl = im.Images.StandardResolution.Url
			metadata.SmallSizeUrl = im.Images.Thumbnail.Url
			metadata.SmallDimensions.width = im.Images.Thumbnail.Width
			metadata.SmallDimensions.height = im.Images.Thumbnail.Height
			c <- &metadata
			count++
		}
		if remainingImageCount == count {
			break
		}
	}

	return searchResponse.Pagination.NextUrl, count
}
