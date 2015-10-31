package snd

import (
	"fmt"
	"math"

	"golang.org/x/mobile/event/size"
	"golang.org/x/mobile/event/touch"
	"golang.org/x/mobile/exp/gl/glutil"
	"golang.org/x/mobile/gl"
)

type LoopButton struct {
	*Button
	lp *Loop
}

func NewLoopButton(btn *Button, lp *Loop) *LoopButton {
	return &LoopButton{btn, lp}
}

func (btn *LoopButton) Paint(ctx gl.Context) {
	if btn.lp.Recording() {
		btn.r, btn.g, btn.b, btn.a = 1, 0, 0, 0.5
	} else if btn.lp.Syncing() {
		btn.r, btn.g, btn.b, btn.a = 1, 1, 0, 0.5
	} else {
		btn.SetActive(false)
	}
	btn.Button.Paint(ctx)
}

type Button struct {
	program  gl.Program
	position gl.Attrib
	color    gl.Uniform
	buf      gl.Buffer

	verts []float32
	data  []byte

	active bool

	r, g, b, a float32

	x, y, w, h float32
}

func (btn *Button) GetColor() (r, g, b, a float32) {
	return btn.r, btn.g, btn.b, btn.a
}

func (btn *Button) SetActiveColor(r, g, b, a float32) {
	btn.r, btn.g, btn.b, btn.a = r, g, b, a
}

func (btn *Button) IsActive() bool { return btn.active }

func (btn *Button) SetActive(b bool) { btn.active = b }

func (btn *Button) HitTest(ev touch.Event, sz size.Event) bool {
	x := ev.X / float32(sz.WidthPx)
	y := ev.Y / float32(sz.HeightPx)

	if btn.x < x && x < (btn.x+btn.w) && btn.y < y && y < (btn.y+btn.h) {
		return true
	}
	return false
}

func NewButton(ctx gl.Context, x, y float32, w, h float32) *Button {
	btn := &Button{r: 1, g: 1, b: 1, a: 1}

	btn.x = (-(-1 - x)) / 2
	btn.y = (1 - y) / 2
	btn.w, btn.h = (w)/2, (-h)/2

	btn.verts = []float32{
		x, y, 0,
		x + w, y, 0,
		x, y + h, 0,
		x, y + h, 0,
		x + w, y, 0,
		x + w, y + h, 0,
	}

	btn.data = make([]byte, len(btn.verts)*4)

	var err error
	btn.program, err = glutil.CreateProgram(ctx, vertexShader, fragmentShader)
	if err != nil {
		panic(fmt.Errorf("error creating GL program: %v", err))
	}

	// create and alloc hw buf
	btn.buf = ctx.CreateBuffer()
	ctx.BindBuffer(gl.ARRAY_BUFFER, btn.buf)
	ctx.BufferData(gl.ARRAY_BUFFER, make([]byte, len(btn.verts)*4), gl.STATIC_DRAW)

	btn.position = ctx.GetAttribLocation(btn.program, "position")
	btn.color = ctx.GetUniformLocation(btn.program, "color")
	return btn
}

func (btn *Button) Paint(ctx gl.Context) {
	for i, x := range btn.verts {
		u := math.Float32bits(x)
		btn.data[4*i+0] = byte(u >> 0)
		btn.data[4*i+1] = byte(u >> 8)
		btn.data[4*i+2] = byte(u >> 16)
		btn.data[4*i+3] = byte(u >> 24)
	}

	ctx.UseProgram(btn.program)
	if btn.active {
		ctx.Uniform4f(btn.color, btn.r, btn.g, btn.b, btn.a)
	} else {
		ctx.Uniform4f(btn.color, 0.4, 0.4, 0.4, 0.5)
	}

	// update hw buf and draw
	ctx.BindBuffer(gl.ARRAY_BUFFER, btn.buf)
	ctx.EnableVertexAttribArray(btn.position)
	ctx.VertexAttribPointer(btn.position, 3, gl.FLOAT, false, 0, 0)
	ctx.BufferSubData(gl.ARRAY_BUFFER, 0, btn.data)
	ctx.DrawArrays(gl.TRIANGLES, 0, len(btn.verts))
	ctx.DisableVertexAttribArray(btn.position)
}

// TODO this is intended to graphically represent sound using opengl
// but the package is "snd". It doesn't make much sense to require
// go mobile gl to build snd (increasing complexity of portability)
// so move this to a subpkg requiring explicit importing.

type Waveform struct {
	Sound

	program  gl.Program
	position gl.Attrib
	color    gl.Uniform
	buf      gl.Buffer

	outs    [][]float64
	samples []float64

	// align    bool
	// alignamp float64
	// aligned  []float64

	verts []float32

	data []byte
}

// TODO just how many samples do we want/need to display something useful?
func NewWaveform(ctx gl.Context, n int, in Sound) (*Waveform, error) {
	wf := &Waveform{Sound: in}

	wf.outs = make([][]float64, n)
	for i := range wf.outs {
		wf.outs[i] = make([]float64, in.BufferLen()*in.Channels())
	}
	wf.samples = make([]float64, in.BufferLen()*in.Channels()*n)
	// wf.aligned = make([]float64, in.BufferLen()*in.Channels()*n)

	wf.verts = make([]float32, len(wf.samples)*3)
	wf.data = make([]byte, len(wf.verts)*4)

	if ctx == nil {
		return wf, nil
	}

	var err error
	wf.program, err = glutil.CreateProgram(ctx, vertexShader, fragmentShader)
	if err != nil {
		return nil, fmt.Errorf("error creating GL program: %v", err)
	}

	// create and alloc hw buf
	wf.buf = ctx.CreateBuffer()
	ctx.BindBuffer(gl.ARRAY_BUFFER, wf.buf)
	ctx.BufferData(gl.ARRAY_BUFFER, make([]byte, len(wf.samples)*12), gl.STREAM_DRAW)

	wf.position = ctx.GetAttribLocation(wf.program, "position")
	wf.color = ctx.GetUniformLocation(wf.program, "color")
	return wf, nil
}

// func (wf *Waveform) Align(amp float64) {
// wf.align = true
// wf.alignamp = amp
// }

func (wf *Waveform) Prepare(tc uint64) {
	wf.Sound.Prepare(tc)

	// cycle outputs
	out := wf.outs[0]
	for i := 0; i+1 < len(wf.outs); i++ {
		wf.outs[i] = wf.outs[i+1]
	}
	for i, x := range wf.Sound.Samples() {
		out[i] = x
	}
	wf.outs[len(wf.outs)-1] = out

	//
	for i, out := range wf.outs {
		idx := i * len(out)
		sl := wf.samples[idx : idx+len(out)]
		for j, x := range out {
			sl[j] = x
		}
	}
}

func (wf *Waveform) Paint(ctx gl.Context, xps, yps, width, height float32) {
	// TODO this is racey and samples can be in the middle of changing
	// move the slice copy to Prepare and sync with playback, or feed over chan
	// TODO assumes mono

	var (
		xstep float32 = width / float32(len(wf.samples))
		xpos  float32 = xps
	)

	for i, x := range wf.samples {
		// clip
		if x > 1 {
			x = 1
		} else if x < -1 {
			x = -1
		}

		wf.verts[i*3] = float32(xpos)
		wf.verts[i*3+1] = yps + (height * float32((x+1)/2))
		wf.verts[i*3+2] = 0
		xpos += xstep
	}

	for i, x := range wf.verts {
		u := math.Float32bits(x)
		wf.data[4*i+0] = byte(u >> 0)
		wf.data[4*i+1] = byte(u >> 8)
		wf.data[4*i+2] = byte(u >> 16)
		wf.data[4*i+3] = byte(u >> 24)
	}

	ctx.UseProgram(wf.program)
	ctx.Uniform4f(wf.color, 1, 1, 1, 1)

	// update hw buf and draw
	ctx.BindBuffer(gl.ARRAY_BUFFER, wf.buf)
	ctx.EnableVertexAttribArray(wf.position)
	ctx.VertexAttribPointer(wf.position, 3, gl.FLOAT, false, 0, 0)
	ctx.BufferSubData(gl.ARRAY_BUFFER, 0, wf.data)
	ctx.DrawArrays(gl.LINE_STRIP, 0, len(wf.samples))
	ctx.DisableVertexAttribArray(wf.position)
}

const vertexShader = `#version 100
attribute vec4 position;
void main() {
  gl_Position = position;
}`

const fragmentShader = `#version 100
precision mediump float;
uniform vec4 color;
void main() {
  gl_FragColor = color;
}`
