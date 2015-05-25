package fetcher

import (
	"errors"
	"fmt"
	"image"
	"net/http"

	_ "image/jpeg"
)

// Obtains an image from a url
func GetImage(url string) (image.Image, error) {
	response, err := http.Get(url)
	if err != nil {
		fmt.Printf("HTTP GET FAILED %s", err)
		return nil, errors.New("image http get failed")
	}

	defer response.Body.Close()

	// Decode JPG data
	m, _, err := image.Decode(response.Body)
	if err != nil {
		fmt.Printf("DECODE FAIL %s", err)
		return nil, errors.New("image decode failed")
	}

	return m, nil
}
