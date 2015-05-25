package imgtools

import (
	"bufio"
	"github.com/aishraj/gopherlisa/common"
	"image"
	"image/jpeg"
	"os"
)

func SameToDisk(context *common.AppContext, imagePath string, imageToSave *image.Image) error {
	if imgFilePng, err := os.Create(imagePath); err != nil {
		context.Log.Printf("Error saving PNG image: %s\n", err)
		return err
	} else {
		defer imgFilePng.Close()
		buffer := bufio.NewWriter(imgFilePng)
		var opt jpeg.Options
		opt.Quality = 95
		err := jpeg.Encode(buffer, *imageToSave, &opt)
		if err != nil {
			context.Log.Printf("Error encoding image:%s", err)
			return err
		}
		buffer.Flush()
		return nil
	}
}
func LoadFromDisk(context *common.AppContext, imagePath string) (image.Image, error) {
	file, err := os.Open(imagePath)
	if err != nil {
		context.Log.Printf("Cannot Load Image %s", err)
		return nil, err
	}
	defer file.Close()
	loadedImage, _, err := image.Decode(file)
	return loadedImage, err
}
