package main

import (
	"encoding/json"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// envelope for PhotoQueryResponse
type PhotoQueryEnv struct {
	Photos PhotoQueryResponse `json:"photos"`
}
type PhotoQueryResponse struct {
	Page    int             `json:"page"`
	Pages   int             `json:"pages"`
	PerPage int             `json:"perpage"`
	Total   string          `json:"total"`
	Photo   []PhotoResponse `json:"photo"`
}

type PhotoResponse struct {
	Id       string `json:"id"`
	Owner    string `json:"owner"`
	Secret   string `json:"secret"`
	Server   string `json:"server"`
	Farm     int    `json:"farm"`
	Title    string `json:"title"`
	IsPublic int    `json:"ispublic"`
	IsFriend int    `json:"isfriend"`
	IsFamily int    `json:"isfamily"`
}

type FlickrPhoto struct {
	URL    string
	AvgRGB map[string]uint32
}

// run a query against flickr, break it up into smaller chunks if needed since
// flickr will only give us 4k results per query
func flickrQueryNew(fg HttpGetter, startDate time.Time, endDate time.Time, pageCount int, query string) []PhotoResponse {
	resultCountQuery := "https://api.flickr.com/services/rest/?method=flickr.photos.search&api_key=488c1e7018f1ddf78b09d51a9604622a&media=photos&per_page=400&page=1&format=json&nojsoncallback=1&min_upload_date=" + strconv.FormatInt(startDate.Unix(), 10) + "&max_upload_date=" + strconv.FormatInt(endDate.Unix(), 10) + "&text=" + query + "&sort=relevance"
	body, err := fg.Get(resultCountQuery)
	if err != nil {
		logger.Println("error getting Flickr results: ", err)
	}
	var flickrResp PhotoQueryEnv
	err = json.Unmarshal(body, &flickrResp)
	if err != nil {
		logger.Println("error unmarshaling JSON: ", err)
	}

	allPhotos := make([]PhotoResponse, 0)
	endDate = startDate.Add((time.Hour * 168))
	photoChan := make(chan PhotoResponse, pageCount*400*2)
	for pageCount*400 > len(photoChan) {
		allPhotos = append(allPhotos, flickrGroupQuery(fg, startDate, endDate, query, pageCount*400, photoChan)...)
		startDate = startDate.Add((time.Hour * 168))
		endDate = endDate.Add((time.Hour * 168))
		// don't go into the future
		if endDate.Unix() > time.Now().Unix() {
			break
		}
	}
	close(photoChan)
	for tempPhoto := range photoChan {
		allPhotos = append(allPhotos, tempPhoto)
	}
	return allPhotos
}

// used to break up the query into 4k chunks
func flickrGroupQuery(fg HttpGetter, startDate time.Time, endDate time.Time, query string, totalNeed int, photoChan chan PhotoResponse) []PhotoResponse {
	if len(photoChan) >= totalNeed {
		return nil
	}

	allPhotos := make([]PhotoResponse, 0)
	resultCountQuery := "https://api.flickr.com/services/rest/?method=flickr.photos.search&api_key=488c1e7018f1ddf78b09d51a9604622a&media=photos&per_page=400&page=1&format=json&nojsoncallback=1&min_upload_date=" + strconv.FormatInt(startDate.Unix(), 10) + "&max_upload_date=" + strconv.FormatInt(endDate.Unix(), 10) + "&text=" + query + "&sort=relevance"
	body, err := fg.Get(resultCountQuery)
	if err != nil {
		logger.Println("error getting Flickr results: ", err)
	}
	var flickrResp PhotoQueryEnv
	err = json.Unmarshal(body, &flickrResp)
	if err != nil {
		logger.Println("error unmarshaling JSON: ", err)
	}

	resultCount, err := strconv.Atoi(flickrResp.Photos.Total)
	if err != nil {
		logger.Println("error converting result to int: ", err)
	}
	// if query has more than 4000 results we need to break it up
	// cycle through trying to find the largest date range that gets
	// us under 4000
	if resultCount > 4000 {
		if endDate.Sub(startDate).Hours() > 24 {
			logger.Println("doing daily")
			for i := 0; i < 7; i++ {
				allPhotos = append(allPhotos, flickrGroupQuery(fg, startDate, startDate.Add(time.Hour*24), query, totalNeed, photoChan)...)
				startDate = startDate.Add((time.Hour * 24))
				if len(allPhotos) >= totalNeed {
					break
				}
			}
		} else {
			if endDate.Sub(startDate).Minutes() > 60 {
				logger.Println("doing hourly")
				for i := 0; i < int(endDate.Sub(startDate).Hours()); i++ {
					allPhotos = append(allPhotos, flickrGroupQuery(fg, startDate, startDate.Add(time.Hour*1), query, totalNeed, photoChan)...)
					startDate = startDate.Add((time.Hour * 1))
					if len(allPhotos) >= totalNeed {
						break
					}
				}
			} else {
				if endDate.Sub(startDate).Seconds() > 60 {
					logger.Println("doing minutes :( - ", resultCount)
					for i := 0; i < int(endDate.Sub(startDate).Minutes()); i++ {
						allPhotos = append(allPhotos, flickrGroupQuery(fg, startDate, startDate.Add(time.Minute*1), query, totalNeed, photoChan)...)
						startDate = startDate.Add((time.Minute * 1))
						if len(allPhotos) >= totalNeed {
							break
						}
					}
				} else {
					logger.Println("doing seconds :( x2 - ", resultCount)
					// at this point we just give up and do the first 10 pages
					// since we're doing 400 images per page that gives us 4000
					for i := 1; i <= 10; i++ {
						if len(photoChan) >= totalNeed {
							break
						} else {
							flickrQueryPage(fg, startDate, startDate.Add(time.Second*1), strconv.Itoa(i), query, photoChan)
						}
						startDate = startDate.Add((time.Second * 1))
					}
				}
			}
		}
	} else {
		if flickrResp.Photos.Pages > 0 {
			var wg sync.WaitGroup
			logger.Printf("under 4k, %v pages\n", flickrResp.Photos.Pages)
			for i := 1; i <= flickrResp.Photos.Pages; i++ {
				if len(photoChan) >= totalNeed {
					break
				} else {
					wg.Add(1)
					go func(fg HttpGetter, startDate time.Time, endDate time.Time, page string, query string, photoChan chan PhotoResponse) {
						defer wg.Done()
						flickrQueryPage(fg, startDate, endDate, page, query, photoChan)
					}(fg, startDate, endDate, strconv.Itoa(i), query, photoChan)
				}
			}
			wg.Wait()
		}
	}
	return allPhotos
}

// loads a specific page from a query and parses out all the image details
func flickrQueryPage(fg HttpGetter, startDate time.Time, endDate time.Time, page string, query string, photoChan chan PhotoResponse) {
	resultCountQuery := "https://api.flickr.com/services/rest/?method=flickr.photos.search&api_key=488c1e7018f1ddf78b09d51a9604622a&media=photos&per_page=400&page=" + page + "&format=json&nojsoncallback=1&min_upload_date=" + strconv.FormatInt(startDate.Unix(), 10) + "&max_upload_date=" + strconv.FormatInt(endDate.Unix(), 10) + "&text=" + query + "&sort=relevance"
	body, err := fg.Get(resultCountQuery)
	if err != nil {
		logger.Println("Error getting Flickr results: ", err)
	}
	var flickrResp PhotoQueryEnv
	err = json.Unmarshal(body, &flickrResp)
	if err != nil {
		logger.Println("Error unmarshaling JSON: ", err)
	}
	for _, val := range flickrResp.Photos.Photo {
		select {
		case photoChan <- val:
		default:
			//		logger.Println("no message sent")
		}
	}
}

// gets a photo from Flickr and saves it to disk
func getFlickrPhoto(photo PhotoResponse, flickrPhotos *[]FlickrPhoto, fg HttpGetter) {
	getString := "https://farm" + strconv.Itoa(photo.Farm) + ".staticflickr.com/" + photo.Server + "/" + photo.Id + "_" + photo.Secret + "_s.jpg"
	fileName := strconv.Itoa(photo.Farm) + "_" + photo.Server + "_" + filepath.Base(getString)

	if _, err := os.Stat(fg.GetSaveDir() + "/" + fileName); os.IsNotExist(err) {
		body, err := fg.Get(getString)
		if err != nil {
			logger.Println("getFlickerPhoto getting image from Flickr: ", err)
			return
		}
		m, _, err := image.Decode(strings.NewReader(string(body)))
		if err != nil {
			logger.Println("error decoding image from Flickr: ", err)
			return
		}

		remoteResize := convertToNRGBA(m)
		avgRGB := avgRGB(remoteResize, true)

		*flickrPhotos = append(*flickrPhotos, FlickrPhoto{URL: fileName, AvgRGB: avgRGB})
		out, err := os.Create(fg.GetSaveDir() + "/" + fileName)
		if err != nil {
			logger.Println("error creating file: ", err)
		}
		defer out.Close()
		err = jpeg.Encode(out, m, nil)
		if err != nil {
			logger.Println("error encoding remote image: ", err)
		}
	} else {
		reader, err := os.Open(fg.GetSaveDir() + "/" + fileName)
		tempImage, _, err := image.Decode(reader)
		if err != nil {
			logger.Println(err)
		}
		defer reader.Close()
		remoteResize := convertToNRGBA(tempImage)
		if remoteResize != nil {
			avgRGB := avgRGB(remoteResize, true)
			*flickrPhotos = append(*flickrPhotos, FlickrPhoto{URL: fileName, AvgRGB: avgRGB})
		} else {
			logger.Printf("cannot convert %v to NRGBA\n", fileName)
		}
	}
}

// gets a photo from flickr but does NOT save to disk
func getFlickrPhotoNoSave(getString string, fg HttpGetter) (image.Image, error) {
	splits := strings.Split(getString, "_")
	splitGetString := "https://farm" + splits[0] + ".staticflickr.com/" + splits[1] + "/" + splits[2] + "_" + splits[3] + "_s.jpg"
	body, err := fg.Get(splitGetString)
	if err != nil {
		logger.Println("getFlickrPhotoNoSave error getting image from Flickr: ", err)
		return nil, err
	}
	tempImage, _, err := image.Decode(strings.NewReader(string(body)))
	if err != nil {
		logger.Println("error decoding image: ", err)
		return nil, err
	}
	return tempImage, nil
}

// query Flickr to detect a search range that will give us the amount of images we need for the mosaic
func queryFixer(fg HttpGetter, startDate time.Time, endDate time.Time, resultsNeeded int, query string, called int) (time.Time, string) {
	var goodStart time.Time
	resultCountQuery := "https://api.flickr.com/services/rest/?method=flickr.photos.search&api_key=488c1e7018f1ddf78b09d51a9604622a&media=photos&per_page=400&page=1&format=json&nojsoncallback=1&min_upload_date=" + strconv.FormatInt(startDate.Unix(), 10) + "&max_upload_date=" + strconv.FormatInt(endDate.Unix(), 10) + "&text=" + query + "&sort=relevance"
	body, err := fg.Get(resultCountQuery)
	if err != nil {
		logger.Println("error getting Flickr photo list: ", err)
	}

	var flickrResp PhotoQueryEnv
	err = json.Unmarshal(body, &flickrResp)
	if err != nil {
		logger.Println("error unmarshaling JSON: ", err)
	}
	resultCount, err := strconv.Atoi(flickrResp.Photos.Total)
	if err != nil {
		logger.Println("error converting result to int: ", err)
	}
	//TODO: lame, should really have a better way to stop this from looping forever
	if resultsNeeded > resultCount && called < 100 {
		newStart := startDate.Add(-(time.Hour * 168 * 2))
		return queryFixer(fg, newStart, endDate, resultsNeeded, query, called+1)
	}
	logger.Println("returning: ", startDate.String())
	goodStart = startDate
	return goodStart, flickrResp.Photos.Total

}
