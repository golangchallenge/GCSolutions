package imgur

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

// http client to make all non-user-specific imgur queries.
var anonClient *http.Client

type anonRoundTripper struct{}

func (a *anonRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	r.Header.Add("Authorization", fmt.Sprintf("Client-ID %s", client_id))
	resp, err := http.DefaultTransport.RoundTrip(r)
	return resp, err
}

func init() {
	anonClient = &http.Client{}
	anonClient.Transport = &anonRoundTripper{}
}

type ImgurClient interface {
	GetTopSubredditImages(r string, n int) ([]string, error)
	UploadImage(data []byte) (string, error)
}

func NewClient(tok *ImgurAccessToken) ImgurClient {
	return &client{tok}
}

type client struct {
	token *ImgurAccessToken
}

const (
	urlGetSubredditImages = "https://api.imgur.com/3/gallery/r/%s/top/all/%d"
	urlUpload             = "https://api.imgur.com/3/image"
)

type galleryImageList struct {
	Data []struct {
		ID          string      `json:"id"`
		Title       string      `json:"title"`
		Description interface{} `json:"description"`
		Datetime    int         `json:"datetime"`
		Type        string      `json:"type"`
		Width       int         `json:"width"`
		Height      int         `json:"height"`
		Size        int         `json:"size"`
		Bandwidth   int64       `json:"bandwidth"`
		Nsfw        bool        `json:"nsfw"`
		Link        string      `json:"link"`
		IsAlbum     bool        `json:"is_album"`
	} `json:"data"`
	Success bool `json:"success"`
	Status  int  `json:"status"`
}

func (i *client) GetTopSubredditImages(r string, n int) ([]string, error) {
	ids := []string{}
	page := 0
	for len(ids) < n {
		url := fmt.Sprintf(urlGetSubredditImages, r, page)
		page++
		resp, err := anonClient.Get(url)
		if err != nil {
			return nil, err
		}
		decoder := json.NewDecoder(resp.Body)
		data := &galleryImageList{}
		err = decoder.Decode(data)
		if err != nil {
			return nil, err
		}
		fmt.Println(page, len(data.Data))
		for _, img := range data.Data {
			ids = append(ids, img.ID)
		}
		if len(data.Data) == 0 {
			break
		}
	}
	return ids, nil
}

func (i *client) UploadImage(data []byte) (string, error) {
	b64 := base64.StdEncoding.EncodeToString(data)
	resp, err := anonClient.PostForm(urlUpload, url.Values{"image": []string{b64}})
	if err != nil {
		return "", err
	}
	bod, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	fmt.Println(string(bod))
	return "", nil
}
