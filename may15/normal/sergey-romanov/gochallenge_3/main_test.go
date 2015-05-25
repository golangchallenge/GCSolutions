package main

import (
	"os"
	"testing"
)

const (
	ImgdirNumberOfFiles = 200
	TileWidth           = 20
	TileHeight          = 20
	inimg               = "img.jpg"
	imgdir              = "./imgdir"
	outimg              = "./outimage.jpg"
)

func TestPrepareImage(t *testing.T) {
	img, err := PrepareTargetImage(inimg)
	if err != nil {
		t.Fatal(err)
	}
	bounds := img.Bounds()
	if bounds.Max.X == 0 || bounds.Max.Y == 0 {
		t.Fatal("Unexpected error: Target image contains errors")
	}
}

func TestPrepareMissingImage(t *testing.T) {
	//image miss.jpg not exist, if after PrepareTargetImage
	//err != nil, something wen wrong
	_, err := PrepareTargetImage("miss.jpg")
	if err == nil {
		t.Fatal("This image actually not exist")
	}
}

func TestReadFromDir(t *testing.T) {
	photos := ReadFromDir(imgdir)
	if len(photos) != ImgdirNumberOfFiles {
		t.Fatal("Some of images in imgdir contain errors")
	}

}

//Another case to defined directory with tile pictures
func TestReadFromDir2(t *testing.T) {
	photos := ReadFromDir("./imgdir/")
	if len(photos) != ImgdirNumberOfFiles {
		t.Fatal("Some of images in imgdir contain errors")
	}

}

func TestGettingNearestPicturesToRegion(t *testing.T) {
	img, _ := PrepareTargetImage(inimg)
	photos := ReadFromDir(imgdir)
	grid, _ := GetNearestPicturesToRegion(img, photos, TileWidth, TileHeight)
	if len(grid) == 0 {
		t.Fatal("Tile pictures was not associated to regions")
	}
}

func TestGettingNearestPicturesWithLargeTiles(t *testing.T) {
	img, _ := PrepareTargetImage(inimg)
	photos := ReadFromDir(imgdir)
	_, msg := GetNearestPicturesToRegion(img, photos, 20, 10000)
	if msg == "" {
		t.Fatal(msg)
	}

	_, msg2 := GetNearestPicturesToRegion(img, photos, 10000, 20)
	if msg2 == "" {
		t.Fatal(msg2)
	}
}

func TestGettingNearestPicturesWithZeroTiles(t *testing.T) {
	img, _ := PrepareTargetImage(inimg)
	photos := ReadFromDir(imgdir)
	_, msg := GetNearestPicturesToRegion(img, photos, 0, 20)
	if msg == "" {
		t.Fatal(msg)
	}

	_, msg2 := GetNearestPicturesToRegion(img, photos, 20, 0)
	if msg2 == "" {
		t.Fatal(msg2)
	}
}

//This test creates output mosaic image outimage.jpg
func TestConstructGrid(t *testing.T) {
	img, _ := PrepareTargetImage(inimg)
	bounds := img.Bounds()
	photos := ReadFromDir(imgdir)
	grid, _ := GetNearestPicturesToRegion(img, photos, TileWidth, TileHeight)
	ConstructFullImageFromTiles(grid, outimg, bounds.Max.X, bounds.Max.Y, TileWidth, TileHeight)
	if _, err := os.Stat(outimg); os.IsNotExist(err) {
		t.Fatal("Output file was not created")
	}

}
