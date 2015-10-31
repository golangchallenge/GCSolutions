package layers

import (
	"bitbucket.org/jcscottiii/amadeus/models"
	"bitbucket.org/jcscottiii/amadeus/util"
	"golang.org/x/mobile/event/size"
	"golang.org/x/mobile/event/touch"
	"golang.org/x/mobile/gl"
)

type baseKeysLayer struct {
	keys []models.Key
}

func (layer *baseKeysLayer) onPaint(glctx gl.Context, sz size.Event, frameData util.FrameData) {
	// Do the drawing
	for _, key := range layer.keys {
		key.PaintKey(glctx, frameData)
	}
}

func (layer *baseKeysLayer) onTouch(x float32, y float32, event touch.Event, frameData util.FrameData, disable bool) (bool, bool) {
	switch event.Type {
	case touch.TypeBegin:
		// User is tapping on the screen, detect if it is on a key.
		for _, key := range layer.keys {
			if key.PlaySound(x, y, frameData) {
				// Found key and played
				// No need for the other layers underneath to do anything.
				return true, false
			}
		}
	case touch.TypeEnd:
		// User's finger came off the screen, detect if it's last positon was a key.
		for _, key := range layer.keys {
			if key.StopSound(x, y, false, frameData) {
				// Found key and stopped.
				// No need for the other layers underneath to do anything.
				return true, false
			}
		}
	case touch.TypeMove:
		// User is sliding their finger on the screen.

		// On current key and not told to disable from upper layer.
		for _, key := range layer.keys {
			if key.DoesCoordsOverlapKey(x, y, frameData) && key.IsPressed() && !disable {
				// Already playing and should do nothing.
				// Also, the lower layers don't need to do anything.
				return true, true
			}
		}
		// Did we move off the key?
		for _, key := range layer.keys {
			if key.StopSound(x, y, !disable, frameData) {
				// Stopped the sound and that means another key could be playing.
				// That means, we should disable lower layers and we should tell
				// the manager to continue passing the event to lower layers
				return false, false
			}
		}
		// Did we move on the key?
		for _, key := range layer.keys {
			if key.PlaySound(x, y, frameData) {
				// Played sound. We should tell lower layers about
				// it but also tell them to disable.
				return false, true
			}
		}
	}
	return false, false
}

func (layer *baseKeysLayer) onStop(glctx gl.Context) {
	// destroy the keys in the layer
	for _, key := range layer.keys {
		key.DestroyKey(glctx)
	}
}
