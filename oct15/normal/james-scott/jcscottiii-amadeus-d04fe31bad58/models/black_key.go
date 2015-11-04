package models

import (
	"github.com/golangchallenge/GCSolutions/oct15/normal/james-scott/jcscottiii-amadeus-d04fe31bad58/util"
	"golang.org/x/mobile/event/size"
	"golang.org/x/mobile/exp/f32"
	"golang.org/x/mobile/gl"

	"encoding/binary"
)

// BlackKey is a container for the PianoKey and a pointer to the whiteKey to the left of it.
type BlackKey struct {
	*PianoKey
	leftWhiteKey *WhiteKey // Used to know where the center of the black key is.
}

// blackKeyRGBColor is a representation of black for the black keys
var blackKeyRGBColor = util.RGBColor{Red: 0.0, Green: 0.0, Blue: 0.0}

const (
	// whiteKeyToBlackKeyLengthRatio is for every white key, a black key should this ratio of it length.
	whiteKeyToBlackKeyLengthRatio = 0.67
	// whiteKeyToBlackKeyWidthRatio is for every white key, a black key should this ratio of it width.
	whiteKeyToBlackKeyWidthRatio = 0.67
)

// NewBlackKey is a constructor that creates and returns a BlackKey
func NewBlackKey(leftWhiteKey Key, glctx gl.Context, note util.KeyNote, sz size.Event) *BlackKey {
	newBlackKey := new(BlackKey)
	// Create the coloring and sound for the black key.
	newBlackKey.PianoKey = newPianoKey(glctx, blackKeyRGBColor, note)

	// Create coordinates for this specific black key
	newBlackKey.openGLCoords, newBlackKey.keyOutline, newBlackKey.keyOuterBoundary =
		makeBlackKeyVector(leftWhiteKey)
	glctx.BufferData(gl.ARRAY_BUFFER, newBlackKey.openGLCoords, gl.STATIC_DRAW)

	return newBlackKey
}

// makeBlackKeyVector creates 1) the OpenGL coordinates for the key in portrait and landscape mode
// 2) the box of the actual key and 3) the greater outer outline of the key which includes the surrounding gap
func makeBlackKeyVector(leftWhiteKey Key) ([]byte, util.Boundary, util.Boundary) {

	// First, let's get the width of the white key.
	widthOfWhiteKey := leftWhiteKey.GetOuterBoundary().RightX - leftWhiteKey.GetOuterBoundary().LeftX
	// Determine the width of the blackKey with whiteKeyToBlackKeyWidthRatio
	widthOfBlackKey := whiteKeyToBlackKeyWidthRatio * widthOfWhiteKey
	// Use the end of the white key as the center of the black key and add the offsets to the sides.
	keyOuterBoundary := util.Boundary{BottomY: TopOfKey - whiteKeyToBlackKeyLengthRatio*TopOfKey,
		TopY:   TopOfKey,
		LeftX:  leftWhiteKey.GetOuterBoundary().RightX - (widthOfBlackKey / 2),
		RightX: leftWhiteKey.GetOuterBoundary().RightX + (widthOfBlackKey / 2)}
	// Use the same gap the white key uses for the black key
	gap := leftWhiteKey.GetOutline().LeftX - leftWhiteKey.GetOuterBoundary().LeftX

	keyOutline := util.Boundary{TopY: TopOfKey, BottomY: keyOuterBoundary.BottomY + gap,
		LeftX: keyOuterBoundary.LeftX + gap, RightX: keyOuterBoundary.RightX - gap}
	return f32.Bytes(binary.LittleEndian, makeCoordsForBothOrientation(keyOutline)...),
		keyOutline, keyOuterBoundary
}
