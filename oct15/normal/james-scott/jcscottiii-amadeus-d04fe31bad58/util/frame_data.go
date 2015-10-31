package util

import (
	"golang.org/x/mobile/gl"
)

// FrameData is a wrapper for information that different events need.
type FrameData struct {
	Color       gl.Uniform
	Program     gl.Program
	Position    gl.Attrib
	Offset      gl.Uniform
	Orientation FrameOrientation
}

// FrameOrientation is type to denote which orientation the device is in depending on width and height.
type FrameOrientation int

const (
	// UnsetOrientation is the case which the orientation has not been determined yet.
	UnsetOrientation FrameOrientation = iota
	// Portrait is when the height is greater than the width
	Portrait
	// Landscape is when the width is greater than the height
	Landscape
)
