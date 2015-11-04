package layers

import (
	"github.com/golangchallenge/GCSolutions/oct15/normal/james-scott/jcscottiii-amadeus-d04fe31bad58/models"
	"github.com/golangchallenge/GCSolutions/oct15/normal/james-scott/jcscottiii-amadeus-d04fe31bad58/util"
	"golang.org/x/mobile/gl"
)

// blackKeysLayer is a layer consisting of only black keys.
// It will sit above the layer of white keys.
type blackKeysLayer struct {
	baseKeysLayer
}

func startBlackKeysLayer(glctx gl.Context, whiteKeys []models.Key) *blackKeysLayer {
	layer := blackKeysLayer{}
	// Create black keys
	layer.keys = make([]models.Key, 0, 0)
	for idx := util.FirstKey; idx <= util.LastKey; idx++ {
		if util.IsBlackKey(idx) {
			whiteKeyIdx := int((idx - util.FirstKey)) - (1 + len(layer.keys))
			layer.keys = append(layer.keys, models.NewBlackKey(whiteKeys[whiteKeyIdx],
				glctx,
				idx,
				util.InitSizeEvent))
		}
	}
	return &layer
}
