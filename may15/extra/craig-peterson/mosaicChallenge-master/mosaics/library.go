package mosaics

import (
	"image"
)

type ThumbnailLibrary struct {
	evaluator Evaluator
	images    map[image.Image]interface{}
}

func NewLibrary(e Evaluator) *ThumbnailLibrary {
	return &ThumbnailLibrary{e, map[image.Image]interface{}{}}
}

func (l *ThumbnailLibrary) GetSampleImages(n int) []image.Image {
	list := []image.Image{}
	for img, _ := range l.images {
		list = append(list, img)
		if len(list) == n {
			break
		}
	}
	return list
}

func (l *ThumbnailLibrary) AddImage(i image.Image) {
	l.images[i] = l.evaluator.Evaluate(i, 0, 0, DefaultTileSize, DefaultTileSize)
}

func (l *ThumbnailLibrary) getBestMatch(target interface{}) image.Image {
	var bestScore float64
	var best image.Image = nil
	for img, val := range l.images {
		score := l.evaluator.Compare(target, val)
		if score < bestScore || best == nil {
			bestScore = score
			best = img
		}
	}
	return best
}
