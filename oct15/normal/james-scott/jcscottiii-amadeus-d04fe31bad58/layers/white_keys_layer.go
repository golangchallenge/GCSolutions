package layers

import (
	"github.com/golangchallenge/GCSolutions/oct15/normal/james-scott/jcscottiii-amadeus-d04fe31bad58/models"
	"github.com/golangchallenge/GCSolutions/oct15/normal/james-scott/jcscottiii-amadeus-d04fe31bad58/util"
	"golang.org/x/mobile/gl"
)

// whiteKeysLayer is a layer consisting of only white keys.
// It will sit below the layer of black keys.
type whiteKeysLayer struct {
	baseKeysLayer
}

func startWhiteKeysLayer(glctx gl.Context) *whiteKeysLayer {
	layer := whiteKeysLayer{}
	// Create white keys
	layer.keys = make([]models.Key, 0, 0)
	for idx := util.FirstKey; idx <= util.LastKey; idx++ {
		if !util.IsBlackKey(idx) {
			layer.keys = append(layer.keys,
				models.NewWhiteKey(glctx, idx, util.InitSizeEvent, len(layer.keys)))
		}
	}
	return &layer
}
