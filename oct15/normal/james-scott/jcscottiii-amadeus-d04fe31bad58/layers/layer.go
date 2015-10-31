package layers

import (
	"bitbucket.org/jcscottiii/amadeus/util"
	"golang.org/x/mobile/event/size"
	"golang.org/x/mobile/event/touch"
	"golang.org/x/mobile/gl"
)

// Layer is an object that contains methods of
// how to resolve a particular event. Layers are intended
// of being in a stack like structure or queue. A layer can choose
// to handle an event or pass it to another layer.
type layer interface {
	onPaint(gl.Context, size.Event, util.FrameData)  // queue
	// onTouch returns two outputs, 1) whether the event has been consumed or
	// 2) should tell all other layers to disable
	onTouch(float32, float32, touch.Event, util.FrameData, bool) (bool, bool) // stack
	onStop(gl.Context)
}
