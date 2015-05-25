# golang challenge 3

Instructions for running.

    # Fetch images from Instagram.
    mosaicly fetch -tag cat -num 1000

    # Generate a mosaic from cat photos
    mosaicly gen -tag cat -in photo.jpg -out mosaic.jpg

    # Or generate a mosaic using images in any directory
    mosaicly gen -imgdir ~/Pictures -in photo.jpg -out mosaic.jpg

Advanced options:

    -units    - change how many mosaic tiles are used
    -unitSize - set how big the mosaic tiles are
    -shrink   - how much to reduce the the final image, as a percent

Running tests:

    # Unit tests
    go test ./...

    # Integration test the command line client
    ./tests/cli.sh

    # Integration test the JSON API
    ./tests/service.sh

---

This is an official entry.

* [Challenge](http://golang-challenge.com/go-challenge3/)
* [Mosaic](http://en.wikipedia.org/wiki/Photographic_mosaic)
* [go-colorful](https://github.com/lucasb-eyer/go-colorful)


# Algorithm

  * Build color palette of Len N from images on hand
  * Take Original image
  * Grid = Convert Original to grid
  * PatternImage = Draw Grid as pixels with FloydSteinberg and palette (dither)
  * Output = new image as pixels * unit size
  * For each pixel in PatternImage, pull an image from the palette and place it at x,y
  * http://tech-algorithm.com/articles/nearest-neighbor-image-scaling/

