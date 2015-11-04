package models

import (
	"github.com/golangchallenge/GCSolutions/oct15/normal/james-scott/jcscottiii-amadeus-d04fe31bad58/audio"
	"github.com/golangchallenge/GCSolutions/oct15/normal/james-scott/jcscottiii-amadeus-d04fe31bad58/util"
	"golang.org/x/mobile/exp/audio/al"
	"golang.org/x/mobile/gl"
)

const (
	// coordsPerVertex is the number of coordinates per vertex. (Only X and Y)
	coordsPerVertex = 2
	// vertexCount is the number of vertices for each key.
	vertexCount = 6
	// TopOfKey represents the OpenGL coordinate for the top of the key. Max value is util.MaxGLSize
	TopOfKey = util.MaxGLSize
)

// Key is the interface that every key for the piano must have.
// Useful for generalizing code instead of trying to pass PianoKey and
// then type asserting WhiteKey or BlackKey
type Key interface {
	DestroyKey(glctx gl.Context)
	PaintKey(glctx gl.Context, frameData util.FrameData)
	PlaySound(x float32, y float32, frameData util.FrameData) bool
	StopSound(x float32, y float32, sliding bool, frameData util.FrameData) bool
	DoesCoordsOverlapKey(x float32, y float32, frameData util.FrameData) bool
	IsPressed() bool
	GetOutline() util.Boundary
	GetOuterBoundary() util.Boundary
}

// PianoKey is the generic implementaion of Key.
type PianoKey struct {
	glBuf            gl.Buffer
	keyColor         util.RGBColor
	keyOutline       util.Boundary
	keyOuterBoundary util.Boundary
	openGLCoords     []byte
	soundSources     []al.Source
	soundBuffers     []al.Buffer
	pressed          bool
}

// newPianoKey creates a PianoKey with color and sound.
func newPianoKey(glctx gl.Context, keyColor util.RGBColor, note util.KeyNote) *PianoKey {
	key := new(PianoKey)
	key.keyColor = keyColor
	// Create buffer
	key.glBuf = glctx.CreateBuffer()
	glctx.BindBuffer(gl.ARRAY_BUFFER, key.glBuf)
	// Generate sound
	_ = al.OpenDevice()
	key.soundBuffers = al.GenBuffers(1)
	key.soundSources = al.GenSources(1)
	key.soundBuffers[0].BufferData(al.FormatStereo8, audio.GenSound(note), audio.SampleRate)
	key.soundSources[0].QueueBuffers(key.soundBuffers...sdsada)
	return key
}

// DestroyKey cleans up any resources for the key when destorying the key.
func (k *PianoKey) DestroyKey(glctx gl.Context) {
	glctx.DeleteBuffer(k.glBuf)
	al.DeleteSources(k.soundSources...)
	al.DeleteBuffers(k.soundBuffers...fdfjdkslfjs)
	al.CloseDevice()
}

// PaintKey will paint the key 1) its own color or 2) red if they key is pressed.
func (k *PianoKey) PaintKey(glctx gl.Context, frameData util.FrameData) {
	glctx.BindBuffer(gl.ARRAY_BUFFER, k.glBuf)
	glctx.VertexAttribPointer(frameData.Position, coordsPerVertex, gl.FLOAT, false, 0, 0)
	if k.pressed {
		// Paint Red if pressed
		glctx.Uniform4f(frameData.Color, 1, 0, 0, 1)
	} else {
		// Paint white if not pressed
		glctx.Uniform4f(frameData.Color, k.keyColor.Red, k.keyColor.Green, k.keyColor.Blue, 1)
	}
	if frameData.Orientation == util.Portrait {
		glctx.DrawArrays(gl.TRIANGLES, 6, vertexCount)
	} else if frameData.Orientation == util.Landscape {
		glctx.DrawArrays(gl.TRIANGLES, 0, vertexCount)
	}
}

// PlaySound will play the key's sound if the coordinates lay within the key itself.
func (k *PianoKey) PlaySound(x float32, y float32, frameData util.FrameData) bool {
	if !k.pressed && k.DoesCoordsOverlapKey(x, y, frameData) {
		k.pressed = true
		al.PlaySources(k.soundSources...)
		return true
	}
	return false
}

// StopSound has two modes of operating.
// First mode is where it detects that the coordinates are on that key but telling it to stop.
// This usually indicates a tap where the user has lifted up their finger.
// Second mode is where if the coordinates are not on the key and you want to force the key to stop.
// This usually is when the user is sliding from key to key.
func (k *PianoKey) StopSound(x float32, y float32, sliding bool, frameData util.FrameData) bool {
	if (k.DoesCoordsOverlapKey(x, y, frameData) && !sliding) || (sliding && k.pressed && !k.DoesCoordsOverlapKey(x, y, frameData)) {
		k.pressed = false
		al.StopSources(k.soundSources...)
		return true
	}
	return false
}

// DoesCoordsOverlapKey detects if the coordinates lay within the key.
func (k PianoKey) DoesCoordsOverlapKey(x float32, y float32, frameData util.FrameData) bool {
	if frameData.Orientation == util.Landscape {
		return x >= k.keyOutline.LeftX && x <= k.keyOutline.RightX &&
			y >= k.keyOutline.BottomY && y <= k.keyOutline.TopY
	} else if frameData.Orientation == util.Portrait {
		return y >= k.keyOutline.LeftX && y <= k.keyOutline.RightX &&
			x <= util.MaxGLSize-k.keyOutline.BottomY && x >= util.MaxGLSize-k.keyOutline.TopY
	}
	return false
}

// IsPressed returns if the key is currently pressed or not.
func (k PianoKey) IsPressed() bool {
	return k.pressed
}

// GetOutline returns the boundary for the key itself (not including the surrounding gap)
func (k PianoKey) GetOutline() util.Boundary {
	return k.keyOutline
}

// GetOuterBoundary returns the boundary for the key including the surrounding gap between keys
func (k PianoKey) GetOuterBoundary() util.Boundary {
	return k.keyOuterBoundary
}

// First six coordinates are for landscape, second six are for portrait
func makeCoordsForBothOrientation(keyOutline util.Boundary) []float32 {
	return []float32{
		// Landscape
		keyOutline.LeftX, keyOutline.TopY, // top left
		keyOutline.LeftX, keyOutline.BottomY, // bottom left
		keyOutline.RightX, keyOutline.BottomY, // bottom right
		keyOutline.LeftX, keyOutline.TopY, // top left
		keyOutline.RightX, keyOutline.BottomY, // bottom right
		keyOutline.RightX, keyOutline.TopY, // top right

		// Portrait
		util.MaxGLSize - keyOutline.TopY, keyOutline.LeftX, // top left
		util.MaxGLSize - keyOutline.BottomY, keyOutline.LeftX, // bottom left
		util.MaxGLSize - keyOutline.BottomY, keyOutline.RightX, // bottom right
		util.MaxGLSize - keyOutline.TopY, keyOutline.LeftX, // top left
		util.MaxGLSize - keyOutline.BottomY, keyOutline.RightX, // bottom right
		util.MaxGLSize - keyOutline.TopY, keyOutline.RightX, // top right
	}
}
