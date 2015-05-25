package imgtools

import (
	"github.com/aishraj/gopherlisa/common"
	"image"
	"image/jpeg"
	"os"
)

func init() {
	image.RegisterFormat("jpeg", "jpeg", jpeg.Decode, jpeg.DecodeConfig)
}

func AddImagesToIndex(context *common.AppContext, directoryName string) (inexedCount int, err error) {
	dirDescriptor, err := os.Open(common.DownloadBasePath + directoryName)
	context.Log.Println("The directory name is :", directoryName)
	if err != nil {
		context.Log.Fatal("Unable to read directory.", err)
		return 0, err
	}
	defer dirDescriptor.Close()
	files, err := dirDescriptor.Readdir(-1)
	if err != nil {
		context.Log.Fatal("Unable to read files in the directrory")
		return 0, err
	}

	maxGoRoutines := 8 //Did this based on the number of cores in my system. Also didn't want too many db connections.
	fileNames := make(chan string, len(files))
	results := make(chan string, len(files))
	errorsChannel := make(chan error, 1)

	for _, fileObj := range files {
		if fileObj.Mode().IsRegular() {
			fileName := fileObj.Name()
			fileNames <- fileName
		}
	}

	close(fileNames)

	for j := 1; j <= maxGoRoutines; j++ {
		go extractAndCalculate(context, fileNames, results, errorsChannel, directoryName)
	}

	count := 0
	for _, fileObj := range files {
		if fileObj.Mode().IsRegular() {
			select {
			case result := <-results:
				context.Log.Printf("The result for file: %v processing was %v \n", fileObj.Name(), result)
				count++
			case errMsg := <-errorsChannel:
				context.Log.Fatal("Sadly, something went wrong. Here's the error : ", errMsg)
			}
		}
	}

	return count, nil
}

func extractAndCalculate(context *common.AppContext, fileNames <-chan string, results chan<- string, errChan chan<- error, directoryName string) {
	db := context.Db
	statement, err := db.Prepare("INSERT Images SET imgtype = ?, img=?,red=?,green=?,blue=?")
	if err != nil {
		context.Log.Println("Error in inserting to db", err)
		errChan <- err
	}
	defer statement.Close()
	for fileName := range fileNames {
		imageFile, err := os.Open(common.DownloadBasePath + directoryName + "/" + fileName)
		if err != nil {
			context.Log.Printf("Unable to open the image file %v Error is %v \n", fileName, err)
			errChan <- err
		}
		img, _, err := image.Decode(imageFile)
		if err != nil {
			context.Log.Println("ERROR: Not able to decode the image file. Error is: ", err)
			errChan <- err
		}
		context.Log.Println("Finding prominent color for image", fileName)
		prominentColor := FindProminentColour(img)
		_, err = statement.Exec(directoryName, fileName, prominentColor.R, prominentColor.G, prominentColor.B)
		if err != nil {
			context.Log.Println("Unable to insert into db. Error is", err)
			errChan <- err
		}
		results <- fileName
	}
}
