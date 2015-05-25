package main

import (
	"bytes"
	"flag"
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"io/ioutil"
	"log"
	"math"
	"os"
)

//Photo struct provides photo as tile and average RGB of this phto
type Photo struct {
	img image.Image
	rgb *RGB
}

//TileImage is
type TileImage struct {
	//x,y - position of the target image
	x, y int
	//Tile image
	photo image.Image
}

//RGB struct is needed for more compact representation of result from averageRGB methods
type RGB struct {
	R, G, B uint32
}

//GridImages type provides grid of tile images
type GridImages []TileImage

//PrepareTargetImage provides read and decoding target image
func PrepareTargetImage(path string) (image.Image, error) {
	reader, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	prepared, err := decodingImage(reader)
	if err != nil {
		return nil, err
	}
	return prepared, nil
}

//This private method provides decoding image
//Returns decoded image. If during decoding process occurred problem, return error
func decodingImage(data []byte) (image.Image, error) {
	rimg, err := jpeg.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	return rimg, nil
}

//ReadFromDir method provides reading from dir with tile pictures
//Returns slice of Photo objects
func ReadFromDir(path string) []Photo {
	photos := []Photo{}
	log.Print(fmt.Sprintf("Try to read tile pictures from dir %s", path))
	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}
	if path[len(path)-1] != '/' {
		path = path + "/"
	}
	for _, data := range files {
		img, _ := os.Open(path + data.Name())
		decoded, _, _ := image.Decode(img)
		bounds := decoded.Bounds()
		photos = append(photos,
			getAverageRGBWithPicture(decoded, uint(bounds.Max.X), uint(bounds.Max.Y)))
		img.Close()
	}
	return photos
}

//This method provides average RGB from tile image
func getAverageRGBWithPicture(img image.Image, newwwidth, newheight uint) Photo {
	return Photo{img, getAverageRGBA(img, int(newwwidth), int(newheight))}
}

//This public method provides average RGBA value from full photo
func getAverageRGBA(img image.Image, tilewidth, tileheight int) *RGB {
	bounds := img.Bounds()
	result := getAverageRGBARegion(img, bounds.Min.X, bounds.Max.X, bounds.Min.Y, bounds.Max.Y, tilewidth, tileheight)
	return result
}

//GetNearestPicturesToRegion method provides nearest picture to each region of target picture
//Return pictures associated with regions or error message
func GetNearestPicturesToRegion(img image.Image, photos []Photo, tilewidth, tileheight int) (GridImages, string) {
	log.Print("Start to splitting user image to regions")
	bounds := img.Bounds()
	//Getting total number of regions
	width := bounds.Max.X
	height := bounds.Max.Y
	if tilewidth == 0 {
		return nil, "Tile width must be greater than zero"
	}

	if tileheight == 0 {
		return nil, "Tile height must be greater than zero"
	}

	if tilewidth > width {
		return nil, "Width of tile image can't be greather then original"
	}
	if tileheight > height {
		return nil, "Height of tile image can't be greather then original"
	}
	newwidth := width / tilewidth
	newheight := height / tileheight
	grid := []TileImage{}
	log.Print("Split target image on the blocks and getting nearest picture")
	stepX := 0
	stepY := 0
	currentX := 1
	currentY := 1
	for i := 0; i < newwidth; i++ {
		currentY = 1
		stepY = 1
		for j := 0; j < newheight; j++ {
			if currentX > width {
				break
			}

			if currentY > height {
				break
			}
			avgrgb := subImageAVG(img, currentX, currentY, currentX+tilewidth, currentY+tileheight, tilewidth, tileheight)
			targetimg := computeDistance(avgrgb, photos)
			grid = append(grid, TileImage{stepX, stepY, targetimg})
			stepY += tileheight
			currentY += tileheight
		}
		stepX += tilewidth
		currentX += tilewidth
	}
	log.Print("Finished to splitting user image to regions")
	return grid, ""

}

//This private method provides getting average RGB from region of the target image
func subImageAVG(img image.Image, currentX, currentY, currentXM, currentYM, tilewidth, tileheight int) *RGB {
	return getAverageRGBA(
		img.(interface {
			SubImage(r image.Rectangle) image.Image
		}).SubImage(image.Rect(currentX, currentY, currentXM, currentYM)), tilewidth, tileheight)
}

//This method provides standard Euclidean distance between two RGB regions
func euclideanDistance(color1 *RGB, color2 *RGB) float64 {
	compute := func(param1, param2 uint32) float64 {
		return math.Pow(float64(param1-param2), 2)
	}
	result := compute(color1.R, color2.R) + compute(color1.G, color2.G) + compute(color1.B, color2.B)
	return math.Sqrt(result)
}

//This method provides computation distance between region from target image and all tile images
func computeDistance(region *RGB, photos []Photo) image.Image {
	if len(photos) == 0 {
		log.Panic("Photos to compute distance not found")
	}
	nearestphoto := photos[0].img
	newregion := RGB{region.R >> 10, region.G >> 10, region.B >> 10}
	dist := euclideanDistance(region, &RGB{photos[0].rgb.R >> 10, photos[0].rgb.G >> 10, photos[0].rgb.B >> 10})
	for _, photo := range photos {
		newrgb := RGB{photo.rgb.R >> 10, photo.rgb.G >> 10, photo.rgb.B >> 10}

		//fmt.Println(region, newrgb)
		canddist := euclideanDistance(&newregion, &newrgb)
		if canddist < dist {
			dist = canddist
			nearestphoto = photo.img
		}
	}
	return nearestphoto
}

//This method provides average region from image
func getAverageRGBARegion(img image.Image, Xmin, Xmax, Ymin, Ymax, tilewidth, tileheight int) *RGB {
	var avgr, avgg, avgb uint64
	avgr, avgg, avgb = 0, 0, 0
	for i := Xmin; i < Xmax; i++ {
		for j := Ymin; j < Ymax; j++ {
			r, g, b, _ := img.At(i, j).RGBA()
			avgr += uint64(r)
			avgg += uint64(g)
			avgb += uint64(b)
		}
	}
	var rect uint64
	Xmaxfin := uint64(tilewidth)
	Ymaxfin := uint64(tileheight)
	rect = Xmaxfin * Ymaxfin
	return &RGB{uint32(avgr / rect), uint32(avgg / rect), uint32(avgb / rect)}
}

//ConstructFullImageFromTiles provides result output image from tile images
func ConstructFullImageFromTiles(grid GridImages, outpath string, totalwidth, totalheight, tilewidth, tileheight int) {
	log.Println("Construct grid from tile images")
	img := image.NewRGBA(image.Rect(0, 0, totalwidth, totalheight))
	for _, pic := range grid {
		draw.Draw(img, img.Bounds(), pic.photo, image.Point{-pic.x, -pic.y}, draw.Src)
	}
	toing, _ := os.Create(outpath)
	defer toing.Close()
	jpeg.Encode(toing, img, nil)
	log.Println("Finished to construct grid from tile images")
}

func main() {
	targetpath := flag.String("target", "", "Path to target image")
	dir := flag.String("dir", "", "Directory with tile pictures")
	outpath := flag.String("outpath", "output.jpg", "Output dir to result image")
	//Note: Tile pictures must be same size. I.E, tilewidth x tileheight
	tilewidth := flag.Int("tilewidth", 0, "Width of tile image")
	tileheight := flag.Int("tileheight", 0, "Height of tile image")
	flag.Parse()
	if *targetpath == "" {
		log.Fatal("Target picture is not selected")
	}

	if *dir == "" {
		log.Fatal("Directory contained tile pictures is not selected")
	}

	//First, read and decode image which will be split on tile images
	targetimg, err := PrepareTargetImage(*targetpath)
	if err != nil {
		log.Fatal(err)
	}

	//Read from directory which contain tile images for mosaic
	photos := ReadFromDir(*dir)

	//associate each region (tilewidth x tileheight) with tile image
	grid, msg := GetNearestPicturesToRegion(targetimg, photos, *tilewidth, *tileheight)
	if msg != "" {
		log.Fatal(msg)
	}
	bounds := targetimg.Bounds()
	//Construct gird and write it to output image
	ConstructFullImageFromTiles(grid, *outpath, bounds.Max.X, bounds.Max.Y, *tilewidth, *tileheight)
}
