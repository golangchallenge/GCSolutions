package app

import (
	"github.com/aishraj/gopherlisa/common"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

func DownloadImages(context *common.AppContext, images []string, searchTerm string) (int, bool) {
	//TODO: use a combination of generator pattern and select to create n go routines and synchronize them using select.
	//The core download method would just take a string url, and download it to $baseDownloadpath/tag
	//the advantage is that when index creation is done it can directly use the tag.
	//
	var maxGoRoutines int
	if len(images) > 50 {
		maxGoRoutines = 50
	} else {
		maxGoRoutines = len(images)
	}
	downloadCount := 0
	tasks := make(chan string, maxGoRoutines)
	results := make(chan int64, len(images))
	for j := 1; j <= maxGoRoutines; j++ {
		go downloader(tasks, results, searchTerm)
	}

	for _, link := range images {
		tasks <- link
	}

	for i := 0; i < len(images); i++ {
		select {
		case result := <-results:
			context.Log.Println("Downloaded image sequence number", i)
			downloadCount++
			if result == 0 {
				context.Log.Println("There was a problem downloading image sequence", i)
			}
		}
	}
	return downloadCount, true
}

func downloader(links <-chan string, results chan<- int64, downloadDir string) {
	for link := range links {
		//process and put the result in results
		//TODO: Send a get request and download the file. See what built in options go has for image downloading.
		basePath := common.DownloadBasePath + downloadDir + "/"
		lastIndex := strings.LastIndex(link, "/")
		name := link[lastIndex+1:]
		log.Println("Processing filename", name)
		filePath := basePath + name
		log.Println("****** creating file at *****", filePath)
		output, err := os.Create(filePath)
		defer output.Close()
		response, err := http.Get(link)
		if err != nil {
			log.Println("Error while downloading", link, "-", err)
			results <- 0
		}
		defer response.Body.Close()
		n, err := io.Copy(output, response.Body)
		log.Println(n, "bytes downloaded")
		results <- n
	}
}
