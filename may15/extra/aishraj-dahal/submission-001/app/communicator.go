package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aishraj/gopherlisa/common"
	"io/ioutil"
	"net/http"
	"net/url"
)

//TODO: This could be made concurrent.

func LoadImages(context *common.AppContext, searchTerm, authToken string) ([]string, error) {
	context.Log.Println("Trying to load images from instagram now.")
	serverURI := "https://api.instagram.com/v1/tags/" + searchTerm + "/media/recent/"

	uri, err := url.Parse(serverURI)
	if err != nil {
		return nil, errors.New("Unable to parse the URL.")
	}
	data := uri.Query()
	data.Set("access_token", authToken)

	uri.RawQuery = data.Encode()

	urlStr := fmt.Sprintf("%v", uri)
	context.Log.Println("The server URI is: ", urlStr)
	return fetchImages(context, urlStr, authToken)
}

func fetchImages(context *common.AppContext, serverURI, authToken string) ([]string, error) {
	items := make([]string, 0, 500)
	urlQueue := make([]string, 500, 500)
	firstURL := serverURI
	urlQueue = append(urlQueue, firstURL)

	for len(urlQueue) > 0 && len(items) <= 500 {
		fetchURL, urlQueue := urlQueue[len(urlQueue)-1], urlQueue[0:len(urlQueue)-1]
		context.Log.Println("Tyring to fetch from the URL: ", fetchURL)
		responseMap, err := fetchServerResponse(context, fetchURL)
		if err != nil {
			errorMessage := "Oops, there was an error geting the server response"
			context.Log.Println(errorMessage)
			return nil, errors.New(errorMessage)
		}
		responseData := responseMap.Data
		for _, responseMeta := range responseData {
			context.Log.Println("Iterating over the response metadata.")
			mediaType := responseMeta.MediaType
			context.Log.Println("The mediatype we got is: ", mediaType)
			if mediaType == "image" {
				thumbNailURL := responseMeta.Images.Thumbnail.URL
				context.Log.Println("*** Parsed the Response for Image URL:", thumbNailURL, "******")
				items = append(items, thumbNailURL)
			}
		}
		nextURL := responseMap.Pagination.NextURL
		context.Log.Println("**Count of values till this cycle are**", len(items))
		context.Log.Println("The next URL is:", nextURL)
		urlQueue = append(urlQueue, nextURL)
	}
	return items, nil
}

func fetchServerResponse(context *common.AppContext, serverURI string) (APIResponse, error) {
	//TODO: Add error handelling here.
	var responseMap APIResponse
	context.Log.Println("Trying to GET from the server on URI: ", serverURI)
	response, err := http.Get(serverURI)
	if err != nil {
		context.Log.Println("Unable to get the images from instagram. Error is: ", err)
		return responseMap, err
	}
	if response.StatusCode != http.StatusOK {
		context.Log.Println("The response was : ", response)
		context.Log.Println("Unable to get a valid response while trying to laod images")
		context.Log.Println("The error received was: ", response.StatusCode)
		return responseMap, errors.New("Can't get images from instagram.")
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		context.Log.Println("Unable to parse HTTP response to body")
		return responseMap, errors.New("Did not get a success while posting on instagram")
	}

	json.Unmarshal(body, &responseMap)
	context.Log.Println("OK got a valid response ")
	return responseMap, nil
}
