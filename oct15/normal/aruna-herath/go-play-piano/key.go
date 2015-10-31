package main

import (
	"encoding/binary"

	"golang.org/x/mobile/exp/f32"
	"golang.org/x/mobile/gl"
)

// Color is a color with R G B Alpha values
type Color [4]float32

// Key is a key shape in piano
// At the moment all keys are drawn as cubes
type Key struct {
	glctx                          gl.Context
	x, y, z, length, width, height float32
	data                           []byte
	color                          Color // This is the actual color drawn
	idColor                        Color // used to identify key by making it unique
	buf                            gl.Buffer
	tex                            []byte
}

// NewKey initializes a new Key
func NewKey(glctx gl.Context, x, y, z, length, width, height float32,
	color Color, idColor Color, tex []byte) *Key {

	r := &Key{
		glctx, x, y, z, length, width, height,
		[]byte{},
		color,
		idColor,
		glctx.CreateBuffer(),
		tex,
	}
	r.chart()

	glctx.BindBuffer(gl.ARRAY_BUFFER, r.buf)
	glctx.BufferData(gl.ARRAY_BUFFER, r.data, gl.STATIC_DRAW)

	return r
}

// Draw draws the key in the glctx
func (r *Key) Draw() {
	// Set the color of the key
	r.glctx.Uniform4f(color, r.color[0], r.color[1], r.color[2], r.color[3])
	r.draw()
}

// DrawI draws the key in the glctx but with idColor
// idColor is unique for each key. So we 'can tell from the pixels' what key
// it is. Used for picking
func (r *Key) DrawI() {
	r.glctx.Uniform1i(drawi, 1) // Let shader know that this draw is for picking
	r.glctx.Uniform4f(color, r.idColor[0], r.idColor[1], r.idColor[2], r.idColor[3])

	r.draw()

	r.glctx.Uniform1i(drawi, 0)
}

func (r *Key) draw() {
	r.glctx.BindBuffer(gl.ARRAY_BUFFER, r.buf)
	r.glctx.EnableVertexAttribArray(position)
	r.glctx.VertexAttribPointer(position, 3, gl.FLOAT, false, 20, 0)
	r.glctx.EnableVertexAttribArray(texCordIn)
	r.glctx.VertexAttribPointer(texCordIn, 2, gl.FLOAT, false, 20, 12)

	r.glctx.DrawArrays(gl.TRIANGLES, 0, 3)
	r.glctx.DrawArrays(gl.TRIANGLES, 1, 3)

	r.glctx.DrawArrays(gl.TRIANGLES, 4, 3)
	r.glctx.DrawArrays(gl.TRIANGLES, 5, 3)

	r.glctx.DrawArrays(gl.TRIANGLES, 8, 3)
	r.glctx.DrawArrays(gl.TRIANGLES, 9, 3)

	r.glctx.DrawArrays(gl.TRIANGLES, 12, 3)
	r.glctx.DrawArrays(gl.TRIANGLES, 13, 3)

	r.glctx.DisableVertexAttribArray(position)
	r.glctx.DisableVertexAttribArray(texCordIn)
}

func (r *Key) chart() {
	keys := []float32{
		// position    // Tex cordinates
		// top
		r.x, r.y, r.z, 0, 0, // top left
		r.x + r.length, r.y, r.z, 1, 0, // top right
		r.x, r.y + r.width, r.z, 0, 1, // bottom left
		r.x + r.length, r.y + r.width, r.z, 1, 1, // bottom right

		//front
		r.x + r.length, r.y, r.z, 1, 0,
		r.x + r.length, r.y + r.width, r.z, 1, 1,
		r.x + r.length, r.y, r.z - r.height, 1, 0,
		r.x + r.length, r.y + r.width, r.z - r.height, 1, 1,

		//left
		r.x, r.y + r.width, r.z, 0, 1,
		r.x + r.length, r.y + r.width, r.z, 1, 1,
		r.x, r.y + r.width, r.z - r.height, 1, 0,
		r.x + r.length, r.y + r.width, r.z - r.height, 0, 0,

		//right
		r.x, r.y, r.z, 0, 0,
		r.x + r.length, r.y, r.z, 1, 0,
		r.x, r.y, r.z - r.height, 0, 1,
		r.x + r.length, r.y, r.z - r.height, 1, 1,
	}

	r.data = append(r.data, f32.Bytes(binary.LittleEndian, keys...)...)
}

// Release releases buffer for the key
func (r *Key) Release() {
	r.glctx.DeleteBuffer(r.buf)
}
