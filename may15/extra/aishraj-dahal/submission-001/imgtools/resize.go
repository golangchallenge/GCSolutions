package imgtools

import (
	"fmt"
	"github.com/aishraj/gopherlisa/common"
	"image"
	"image/color"
	"image/jpeg"
	"math"
	"os"
)

func init() {
	image.RegisterFormat("jpeg", "jpeg", jpeg.Decode, jpeg.DecodeConfig)
}

func ResizeImages(context *common.AppContext, directoryName string) (int, bool) {
	dirDescriptor, err := os.Open(common.DownloadBasePath + directoryName) //TODO change this
	context.Log.Println("The directory name is :", directoryName)
	if err != nil {
		context.Log.Fatal("Unable to read directory.", err)
		return 0, false
	}
	defer dirDescriptor.Close()
	files, err := dirDescriptor.Readdir(-1)
	if err != nil {
		context.Log.Fatal("Unable to read files in the directrory")
		return 0, false
	}

	maxGoRoutines := 8
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
		go extractAndProcess(context, fileNames, results, errorsChannel, directoryName)
	}

	successCount := 0

	for _, fileObj := range files {
		if fileObj.Mode().IsRegular() {
			select {
			case result := <-results:
				successCount++
				context.Log.Printf("The result for file: %v processing was %v \n", fileObj.Name(), result)
			case errMsg := <-errorsChannel:
				context.Log.Fatal("Sadly, something went wrong. Here's the error : ", errMsg)
			}
		}
	}

	return successCount, true
}

func extractAndProcess(context *common.AppContext, fileNames <-chan string, results chan<- string, errChan chan<- error, directoryName string) {
	for fileName := range fileNames {
		filePath := common.DownloadBasePath + directoryName + "/" + fileName
		imageFile, err := os.Open(filePath)
		if err != nil {
			context.Log.Printf("Unable to open the image file %v Error is %v \n", fileName, err)
			errChan <- err
		}
		img, _, err := image.Decode(imageFile)
		if err != nil {
			context.Log.Println("ERROR: Not able to decode the image file. Error is: ", err)
			errChan <- err
		}

		nrgbaImage := ToNRGBA(img)
		bounds := nrgbaImage.Bounds()
		boundsString := fmt.Sprintln(bounds)
		context.Log.Println("The bounds BEFORE the resize are: ", boundsString)

		nrgbaImage = Resize(context, nrgbaImage, tileSize, tileSize)

		bounds = nrgbaImage.Bounds()
		boundsString = fmt.Sprintln(bounds)
		context.Log.Println("The bounds AFTER the resize are: ", boundsString)

		var opt jpeg.Options

		opt.Quality = 80

		imageFile.Close()
		writeFile, err := os.Create(filePath)
		if err != nil {
			context.Log.Println("Unable to open the file for writing. Error is:", err)
		}
		err = jpeg.Encode(writeFile, nrgbaImage, &opt)
		if err != nil {
			context.Log.Println("ERROR: Not able to write to file with JPEG encoding", err)
			errChan <- err
		}

		results <- boundsString
		writeFile.Close()
	}
}

func ToNRGBA(img image.Image) *image.NRGBA {
	sourceBounds := img.Bounds()
	if sourceBounds.Min.X == 0 && sourceBounds.Min.Y == 0 {
		if source0, ok := img.(*image.NRGBA); ok {
			return source0
		}
	}
	return CloneImage(img)
}

func Resize(context *common.AppContext, source *image.NRGBA, width, height int) *image.NRGBA {
	//Naive nn resize.
	destinationW, destinationH := width, height

	sourceBounds := source.Bounds()
	sourceW := sourceBounds.Max.X
	sourceH := sourceBounds.Max.Y

	destination := image.NewNRGBA(image.Rect(0, 0, destinationW, destinationH))

	dx := float64(sourceW) / float64(destinationW)
	dy := float64(sourceH) / float64(destinationH)

	for destinationY := 0; destinationY < destinationH; destinationY++ {
		fy := (float64(destinationY)+0.5)*dy - 0.5

		for destinationX := 0; destinationX < destinationW; destinationX++ {
			fx := (float64(destinationX)+0.5)*dx - 0.5

			sourceX := int(math.Min(math.Max(math.Floor(fx+0.5), 0.0), float64(sourceW)))
			sourceY := int(math.Min(math.Max(math.Floor(fy+0.5), 0.0), float64(sourceH)))

			sourceOff := sourceY*source.Stride + sourceX*4
			destinationOff := destinationY*destination.Stride + destinationX*4

			copy(destination.Pix[destinationOff:destinationOff+4], source.Pix[sourceOff:sourceOff+4])
		}
	}

	return destination
}

func CloneImage(img image.Image) *image.NRGBA {
	sourceBounds := img.Bounds()
	sourceMinX := sourceBounds.Min.X
	sourceMinY := sourceBounds.Min.Y

	destinationBounds := sourceBounds.Sub(sourceBounds.Min)
	destinationW := destinationBounds.Dx()
	destinationH := destinationBounds.Dy()
	destination := image.NewNRGBA(destinationBounds)

	for destinationY := 0; destinationY < destinationH; destinationY++ {
		index := destination.PixOffset(0, destinationY)
		for destinationX := 0; destinationX < destinationW; destinationX++ {

			c := color.NRGBAModel.Convert(img.At(sourceMinX+destinationX, sourceMinY+destinationY)).(color.NRGBA)
			destination.Pix[index+0] = c.R
			destination.Pix[index+1] = c.G
			destination.Pix[index+2] = c.B
			destination.Pix[index+3] = c.A

			index += 4

		}
	}

	return destination
}
