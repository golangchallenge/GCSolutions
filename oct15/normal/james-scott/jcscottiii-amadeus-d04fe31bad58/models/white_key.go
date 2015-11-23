package models

import (
	"encoding/binary"
	"github.com/golangchallenge/GCSolutions/oct15/normal/james-scott/jcscottiii-amadeus-d04fe31bad58/util"
	"golang.org/x/mobile/event/size"
	"golang.org/x/mobile/exp/f32"
	"golang.org/x/mobile/gl"
)

const (
	// KeytoGapRatio is the ratio of actual key there should be compared to gap space between keys.
	KeytoGapRatio = .95 // Increase the ratio to make gaps smaller.
	// NumberOfWhiteKeys to know how many white keys to create.
	NumberOfWhiteKeys = 9.0
)

// WhiteKey is a container for the PianoKey
type WhiteKey struct {
	*PianoKey
}

// whiteKeyRGBColor is a representation of white for the white keys
var whiteKeyRGBColor = util.RGBColor{Red: 1.0, Green: 1.0, Blue: 1.0}

// NewWhiteKey is a constructor that creates and returns a WhiteKey
func NewWhiteKey(glctx gl.Context, note util.KeyNote, sz size.Event, count int) *WhiteKey {
	newWhiteKey := new(WhiteKey)
	// Create the coloring and sound for the white key.
	newWhiteKey.PianoKey = newPianoKey(glctx, whiteKeyRGBColor, note)

	// Create coordinates for this specific white key
	newWhiteKey.openGLCoords, newWhiteKey.keyOutline, newWhiteKey.keyOuterBoundary = makeWhiteKeyVector(float32(sz.WidthPx), count)
	glctx.BufferData(gl.ARRAY_BUFFER, newWhiteKey.openGLCoords, gl.STATIC_DRAW)

	return newWhiteKey
}

// makeWhiteKeyVector creates 1) the OpenGL coordinates for the key in portrait and landscape mode
// 2) the box of the actual key and 3) the greater outer outline of the key which includes the surrounding gap
func makeWhiteKeyVector(width float32, count int) ([]byte, util.Boundary, util.Boundary) {

	keyOutline := util.Boundary{BottomY: 0.0, TopY: TopOfKey}
	// Which white key are we on. Start at that offset.
	offset := (float32(count)) / NumberOfWhiteKeys * util.MaxGLSize
	// Width of window / NumberOfWhiteKeys = width of each white key = w_k
	widthOfOneKey := width / NumberOfWhiteKeys

	keyOuterBoundary := util.Boundary{BottomY: 0.0, TopY: TopOfKey, LeftX: offset,
		RightX: offset + widthOfOneKey/width*util.MaxGLSize}

	// want KeytoGapRatio% of the width of a key to be white (allow for black separation between keys) = p
	// Top left & bottom left (x = (offset + ((1-p) / 2) *w_k) / width * MaxGLSize)
	// Top right & bottom right (x = (offset + w_k-(w_k* (1-p) / 2)) / width * MaxGLsize)
	keyOutline.RightX = offset + (widthOfOneKey-(widthOfOneKey*(1-KeytoGapRatio)/2))/width*util.MaxGLSize
	keyOutline.LeftX = offset + (widthOfOneKey*(1-KeytoGapRatio)/2)/width*util.MaxGLSize
	return f32.Bytes(binary.LittleEndian, makeCoordsForBothOrientation(keyOutline)...),
		keyOutline, keyOuterBoundary
}
