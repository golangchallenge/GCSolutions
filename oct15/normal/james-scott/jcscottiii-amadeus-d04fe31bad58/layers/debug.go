package layers

import (
	"github.com/golangchallenge/GCSolutions/oct15/normal/james-scott/jcscottiii-amadeus-d04fe31bad58/util"
	"golang.org/x/mobile/event/size"
	"golang.org/x/mobile/event/touch"
	"golang.org/x/mobile/exp/app/debug"
	"golang.org/x/mobile/exp/gl/glutil"
	"golang.org/x/mobile/gl"
)

type debugLayer struct {
	images *glutil.Images
	fps    *debug.FPS
}

func startDebugLayer(glctx gl.Context) debugLayer {
	layer := debugLayer{}
	layer.images = glutil.NewImages(glctx)
	layer.fps = debug.NewFPS(layer.images)
	return layer
}
func (layer debugLayer) onPaint(glctx gl.Context, sz size.Event, frameData util.FrameData) {
	layer.fps.Draw(sz)
}

func (layer debugLayer) onTouch(x float32, y float32, event touch.Event, frameData util.FrameData, disable bool) (bool, bool) {
	// We are not finished with this event, continue it on to lower layers.
	return false, false
}
func (layer debugLayer) onStop(glctx gl.Context) {
	layer.fps.Release()
	layer.images.Release()
}
