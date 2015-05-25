Instructions to run:

`go run main.go --target input_image.jpg --dir /path/to/picturedir --outpath output_image.jpg --tilewidth 20 --tileheight 20`

Where:

* --target - input image which will be used as photomosaic
* --dir - directory which contains tile images
* --outpath - output image name. If this is not defined, output image name will be output.jpg
* --tilewidth - Width of the region on the input image
* --tileheight - Height of the region on the input image

**Note**: Tile pictures must be the same size. I.E, tilewidth x tileheight