package layers

import (
	"github.com/golangchallenge/GCSolutions/oct15/normal/james-scott/jcscottiii-amadeus-d04fe31bad58/util"
	"golang.org/x/mobile/event/size"
	"golang.org/x/mobile/event/touch"
	"golang.org/x/mobile/gl"
)

// Manager is the layer manager. It regulates the layers and the events passed to the layers.
// Depending on the event, it will pass the event like to the layers like a stack or a queue.
type Manager struct {
	glctx  gl.Context
	layers []layer
}

// StartLayers creates all the layers
func (m *Manager) StartLayers() {
	m.layers = make([]layer, 0, 0)
	newWhiteKeyLayer := startWhiteKeysLayer(m.glctx)
	m.layers = append(m.layers, newWhiteKeyLayer)
	m.layers = append(m.layers, startBlackKeysLayer(m.glctx, newWhiteKeyLayer.keys))
	m.layers = append(m.layers, startDebugLayer(m.glctx))
}

// PaintLayers handles the paint event for the layers. It paints the layers FIFO.
func (m *Manager) PaintLayers(sz size.Event, frameData util.FrameData) {
	for _, layer := range m.layers {
		layer.onPaint(m.glctx, sz, frameData)
	}
}

// TouchLayers handles the touch event for the layers. It handles the event like a stack.
// The top layer decides if it will consume the layer or pass it on.
func (m *Manager) TouchLayers(x float32, y float32, event touch.Event, frameData util.FrameData) {
	disableLowerLayer := false
	finished := false
	for idx := len(m.layers) - 1; idx >= 0; idx-- {
		if finished, disableLowerLayer = m.layers[idx].onTouch(x, y, event, frameData, disableLowerLayer); finished {
			break
		}
	}
}

// StopLayers handles the clean up of the layers.
func (m *Manager) StopLayers() {
	for idx := len(m.layers) - 1; idx >= 0; idx-- {
		m.layers[idx].onStop(m.glctx)
	}
}

// NewLayerManager is a constructor for the Manager
func NewLayerManager(glctx gl.Context) *Manager {
	manager := &Manager{glctx: glctx}

	return manager
}
