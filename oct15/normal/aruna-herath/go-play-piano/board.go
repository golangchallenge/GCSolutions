package main

import (
	"math"

	"golang.org/x/mobile/gl"
)

// Board is a piano keyboard
type Board struct {
	glctx          gl.Context
	marginFraction float32
	numKeys        float32
	keyLen         float32
	margin         float32
	bigKeys        []*Key
	smallKeys      []*Key

	idColor    float32
	idColorKey map[float32]int
	tex        []byte
}

// NewBoard returns an initialized Board
func NewBoard(glctx gl.Context, mf float32, numkeys int) *Board {
	b := new(Board)
	b.idColorKey = map[float32]int{}
	b.glctx = glctx

	b.marginFraction = float32(mf) // The space between keys as a fraction of key length
	b.numKeys = float32(numkeys)
	b.keyLen = 2 / (b.numKeys + (b.numKeys-1)*b.marginFraction)
	b.margin = b.keyLen * b.marginFraction

	b.initialize()

	return b
}

func (b *Board) initialize() {
	h := float32(-1.0)
	keys := 0

	for i := 0; i < int(b.numKeys); i++ {
		idCol := b.getIDColor()
		key := NewKey(b.glctx, -1, h, 0, 1.8, b.keyLen, 0.1, Color{1, 1, 1, 1}, idCol, b.tex)
		b.bigKeys = append(b.bigKeys, key)
		b.idColorKey[idCol[0]] = keys
		keys++

		if !(i%7 == 2 || i%7 == 6) { // if a small key is here
			idCol = b.getIDColor()
			key := NewKey(b.glctx, -1, h+(b.keyLen+b.margin/2)-(b.keyLen*0.35), 0.1, 1.2,
				b.keyLen*0.75, 0.1, Color{0.5, 0.5, 0.5, 1}, idCol, b.tex)
			b.smallKeys = append(b.smallKeys, key)
			b.idColorKey[idCol[0]] = keys
			keys++
		}

		h = h + b.keyLen + b.margin
	}
}

// Draw draws the keyboard
func (b *Board) Draw() {
	for _, key := range b.bigKeys {
		key.Draw()
	}
	for _, key := range b.smallKeys {
		key.Draw()
	}
}

// DrawI draws the keyboard but with idColors
func (b *Board) DrawI() {
	for _, key := range b.bigKeys {
		key.DrawI()
	}
	for _, key := range b.smallKeys {
		key.DrawI()
	}
}

func (b *Board) getIDColor() (c Color) {
	c = Color{b.idColor, 0.0, 0, 0}
	b.idColor = b.idColor + 0.01

	b.idColor = float32(math.Floor(float64(b.idColor*100)+0.5)) / 100
	return
}

// Release releases resources allocated for the board
func (b *Board) Release() {
	for _, key := range b.bigKeys {
		key.Release()
	}
	for _, key := range b.smallKeys {
		key.Release()
	}
}
