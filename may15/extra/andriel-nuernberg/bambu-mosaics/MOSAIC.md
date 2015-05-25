# Go Challenge #3

Hi everyone,

I've been working in the mosaic algorith for this challange and I'd like to share my results with you guys.
For now, I've generated only mosaics with uniform colors in the tiles. I'm sure once you understand how it works, change to use thumbnails as tiles will be much easier.

**Disclaimer:**  I'm new in Go lang, so please don't worry about the ugly code neither the english mistakes.

PS: Any feedback is appreciated.

## Steps

1. Read the picture into an [Image](https://golang.org/pkg/image/#Image);
3. Create a new [RGB(A)](https://golang.org/pkg/image/#RGBA) image of the same size of the original image;
4. Loop through each tile image and get its average RGB(A) color;
5.  Draw a [Rectangle](https://golang.org/pkg/image/#Rectangle) using the [Uniform](https://golang.org/pkg/image/#Uniform) average color into the created RGB(A) image;
6. Save the tiled image as a new picture.


## 1. Read the picture into an Image

This is probably the easiest part. Just use the `os` lib to open the file and decode into an `Image`.

```go
reader, err := os.Open("star.jpg")
if err != nil {
  log.Fatal(err)
}
defer reader.Close()


srcImg, format, err := image.Decode(reader)
```

## 2. Create a new RGB(A) image

This will be the output image. Basically we create a new image with the same size of the original one.

```go
destImg := image.NewRGBA(image.Rect(0, 0, srcImg.Bounds().Max.X, srcImg.Bounds().Max.Y))
```

## 3. Loop through each tile ang get its average color

Before iterate over each tile, I've hardcoded a `tileSize` value. Feel free to change this value to see different results.

```go
tileSize := 10

for x := srcImg.Bounds().Min.X; x < srcImg.Bounds().Max.X; x += tileSize {
  for y := srcImg.Bounds().Min.Y; y < srcImg.Bounds().Max.Y; y += tileSize {
    r, g, b := getAvgColor(tile) # set tile by yourself :D
    c := color.RGBA{r, g, b, 255}
  }
}
```

## 4. Draw the new tile

As I said in the beginning, I'm just using uniform colors in the tiles. To use the thumbnails, you will need to process the average color of each thumbnail and get the one that has the closest average color of the original tile.

```go
# Inside the tiles loop
draw.Draw(destImg, image.Rect(x, y, x+tileSize, y+tileSize), &image.Uniform{c}, image.ZP, draw.Src)
```


## 5. Save the tiled image

Once everything is done, we just need to save it in the filesystem.

```go
outImg, err := os.Create("tiled.jpg")
if err != nil {
  log.Fatal(err)
}
defer outImg.Close()

png.Encode(outImg, destImg)
```

## Results

| ![](http://f.cl.ly/items/0o2r1l1C1f1F3G223a1H/star.jpg) | ![](http://f.cl.ly/items/151b1v203p3G1E0o0o2z/tiled.jpg) | ![](http://cl.ly/image/3K1w1f1X3d0f/tiled.jpg) |
|:-------------------------------------------------------:|:--------------------------------------------------------:|:----------------------------------------------:|
| Original picture                                        | Tiled picture (tileSize = 10)                            | Tiled picture (tileSize = 30)                  |

## Resource

- [http://www.royvanrijn.com/blog/2014/04/mosaic-algorithm/](http://www.royvanrijn.com/blog/2014/04/mosaic-algorithm/)
- [http://williamedwardscoder.tumblr.com/post/84505278488/making-image-mosaics](http://williamedwardscoder.tumblr.com/post/84505278488/making-image-mosaics)
- [http://www.drdobbs.com/understanding-photomosaics/184404848?pgno=1](http://www.drdobbs.com/understanding-photomosaics/184404848?pgno=1)
- [http://blog.golang.org/go-imagedraw-package](http://blog.golang.org/go-imagedraw-package)
- [http://blog.golang.org/go-image-package](http://blog.golang.org/go-image-package)
- [https://www.mattcutts.com/blog/photo-mosaic-effect-with-go/](https://www.mattcutts.com/blog/photo-mosaic-effect-with-go/)
