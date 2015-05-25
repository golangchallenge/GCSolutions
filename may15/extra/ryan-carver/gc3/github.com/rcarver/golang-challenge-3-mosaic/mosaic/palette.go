package mosaic

import (
	"image"
	"image/color"
)

// ImagePalette is a color.Paletted that is backed by images. Each entry in the
// color palette may have multiple images associated, in which case the image
// used is rotated through available options.
type ImagePalette struct {
	color.Palette
	solidFallback bool
	images        map[int][]image.Image
	indices       map[int]int
}

// NewImagePalette initializes an ImagePalette of a number of colors, and
// images of a certain size. This palette must be populated with images to be
// useful.
func NewImagePalette(colors int) *ImagePalette {
	return &ImagePalette{
		Palette:       make(color.Palette, 0, colors),
		solidFallback: false,
		images:        make(map[int][]image.Image),
		indices:       make(map[int]int),
	}
}

// NewSolidPalette initializes an ImagePalette with a given color palette and
// images of a certain size. This palette does not need to be populated with
// images - instead, solid images will be returned for any color index.
func NewSolidPalette(palette color.Palette) *ImagePalette {
	return &ImagePalette{
		Palette:       palette,
		solidFallback: true,
	}
}

// Add builds the color palette by assigning a new color from the given image.
// If the palette is full, or the palette already contains the color of the
// image then the image is added as an option to the nearest color.
func (p *ImagePalette) Add(m image.Image) {
	c := average(m, m.Bounds(), 1)
	// If we don't have a full color palette, use every image as a new
	// entry (unless it's a dup).
	if len(p.Palette) < cap(p.Palette) {
		var found bool
		for _, x := range p.Palette {
			if x == c {
				found = true
				break
			}
		}
		if found {
			//fmt.Printf("Add(%v) has color\n", c)
		} else {
			p.Palette = append(p.Palette, c)
			//fmt.Printf("Add(%v) new color\n", c)
		}
	}
	// Index images by their nearest color in the palette.
	i := p.Index(c)
	// TODO: crop image to ImgX, ImgY
	p.images[i] = append(p.images[i], m)
	//fmt.Printf("Add(%v) %d\n", c, len(p.images[i]))
}

// AtColor returns an image whose average color is closest to c in the palette.
func (p *ImagePalette) AtColor(c color.Color) image.Image {
	i := p.Index(c)
	images, ok := p.images[i]
	if ok {
		idx := p.indices[i]
		p.indices[i]++
		if p.indices[i] > len(images) {
			p.indices[i] = 0
			idx = 0
		}
		//fmt.Printf("%v %d is %v\n", c, idx, images[idx].At(0, 0))
		return images[idx]
	}
	if p.solidFallback {
		x := p.Convert(c)
		return image.NewUniform(x)
	}
	return nil
}

// NumColors returns the number of colors in the palette.
func (p *ImagePalette) NumColors() int {
	return len(p.Palette)
}

// NumImages returns the number of images in the palette.
func (p *ImagePalette) NumImages() int {
	var c = 0
	for _, images := range p.images {
		c += len(images)
	}
	return c
}
