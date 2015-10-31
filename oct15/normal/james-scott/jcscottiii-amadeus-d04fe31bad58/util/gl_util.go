package util

import (
	"golang.org/x/mobile/event/size"
	"golang.org/x/mobile/event/touch"
)

const (
	// MaxGLSize represents the maximum number that the OpenGL coordinate can be for X and Y.
	MaxGLSize = 2.0
)

// ConvertToOpenGLCoords converts the real world coordinates (which vary between screen sizes) into
// OpenGL coordinates which range between x = [0, MaxGLSize], y = [0, MaxGLSize].
func ConvertToOpenGLCoords(sizeEvent size.Event, touchEvent touch.Event) (float32, float32) {
	// We are drawing things in the positive x and positive y quadrant. X increases from left to right.
	// That matches the pattern for the window X coords.
	// Y increases from the bottom to top. However, a touch event in the Y direction inceases
	// from the top to bottom. As a result, we need to convert the Y.
	return (touchEvent.X / float32(sizeEvent.WidthPx)) * MaxGLSize,
		(MaxGLSize - (touchEvent.Y/float32(sizeEvent.HeightPx))*MaxGLSize)

}

// InitSizeEvent is a just a random size Event. This is used for when initializing the app when there is no
// size event yet.
var InitSizeEvent = size.Event{HeightPx: 400, WidthPx: 400}
