package main

import (
	"image"
	"image/draw"
	_ "image/jpeg"
	"log"
	"math"

	"github.com/go-gl/mathgl/mgl32"
	"golang.org/x/mobile/app"
	"golang.org/x/mobile/asset"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"
	"golang.org/x/mobile/event/touch"
	"golang.org/x/mobile/exp/gl/glutil"
	"golang.org/x/mobile/gl"
)

var (
	program    gl.Program
	position   gl.Attrib
	texCordIn  gl.Attrib
	color      gl.Uniform
	drawi      gl.Uniform
	projection gl.Uniform
	camera     gl.Uniform

	touchX float32
	touchY float32

	board    *Board
	keystate map[touch.Sequence]int
)

func main() {
	app.Main(func(a app.App) {
		var glctx gl.Context

		visible := false
		sz := size.Event{}

		for e := range a.Events() {
			switch e := a.Filter(e).(type) {
			case lifecycle.Event:
				switch e.Crosses(lifecycle.StageVisible) {
				case lifecycle.CrossOn:
					visible = true
					glctx, _ = e.DrawContext.(gl.Context)
					onStart(glctx, sz)
				case lifecycle.CrossOff:
					visible = false
					onStop(glctx)
				}
			case size.Event:
				sz = e
				touchX = float32(sz.WidthPx / 2)
				touchY = float32(sz.HeightPx / 2)
			case paint.Event:
				onPaint(glctx, sz)
				a.Publish()
				if visible {
					a.Send(paint.Event{})
				}
			case touch.Event:
				onTouch(glctx, e)
			}
		}
	})
}

func onStart(glctx gl.Context, sz size.Event) {
	log.Printf("creating GL program")
	var err error
	keystate = map[touch.Sequence]int{}
	program, err = glutil.CreateProgram(glctx, vertexShader, fragmentShader)
	if err != nil {
		log.Printf("error creating GL program: %v", err)
		return
	}

	glctx.Enable(gl.DEPTH_TEST)

	position = glctx.GetAttribLocation(program, "position")
	texCordIn = glctx.GetAttribLocation(program, "texCordIn")
	color = glctx.GetUniformLocation(program, "color")
	drawi = glctx.GetUniformLocation(program, "drawi")
	projection = glctx.GetUniformLocation(program, "projection")
	camera = glctx.GetUniformLocation(program, "camera")

	loadTexture(glctx)
	glctx.UseProgram(program)

	projectionMat := mgl32.Perspective(mgl32.DegToRad(75.0), float32(1), 0.5, 40.0)
	glctx.UniformMatrix4fv(projection, projectionMat[:])

	cameraMat := mgl32.LookAtV(mgl32.Vec3{0.5, 0, 1.5}, mgl32.Vec3{0, 0, 0}, mgl32.Vec3{0, 1, 0})
	glctx.UniformMatrix4fv(camera, cameraMat[:])

	board = NewBoard(glctx, float32(0.05), 10)

	numKeys := len(board.bigKeys) + len(board.smallKeys)

	InitializeSound(numKeys)
}

func onStop(glctx gl.Context) {
	glctx.DeleteProgram(program)
	board.Release()
	ReleaseSound()
}

func onPaint(glctx gl.Context, sz size.Event) {
	glctx.ClearColor(0.2, 0.2, 0.2, 0.5)
	glctx.Clear(gl.DEPTH_BUFFER_BIT)
	glctx.Clear(gl.COLOR_BUFFER_BIT)

	glctx.UseProgram(program)

	board.Draw()
}

func onTouch(glctx gl.Context, e touch.Event) {
	touchX = e.X
	touchY = e.Y

	// When touch occurs we need to figure out which key is pressed.
	// Using color picking for this.
	// That is, scene is redrawn with each key given a unique color.
	// And then the color is read on the touched pixel to figure out which key
	// is pressed or if a key is pressed at all.

	glctx.ClearColor(0.73, 0.5, 0.75, 0.5)
	glctx.Clear(gl.DEPTH_BUFFER_BIT)
	glctx.Clear(gl.COLOR_BUFFER_BIT)
	glctx.UseProgram(program)

	board.DrawI() // Draw the board with each key in a unique color

	c := make([]byte, 12, 12)

	glctx.ReadPixels(c, int(e.X), int(e.Y), 1, 1, gl.RGB, gl.UNSIGNED_BYTE)
	// gl.RGB, gl.UNSIGNED_BYTE is the combination that is preffered by my Android
	// phone. And is said to be the one preffered by many.

	r := (float32(c[0]) / 255) * 100                 // Convert byte to float
	out := float32(math.Floor(float64(r)+0.5)) / 100 // and round up
	key, ok := board.idColorKey[out]

	if !ok {
		curKey, ok := keystate[e.Sequence] //stop already playing sound
		if ok {
			StopSound(curKey)
		}
		return
	}

	if e.Type == touch.TypeBegin {
		keystate[e.Sequence] = key
		PlaySound(key)
	} else if e.Type == touch.TypeEnd {
		delete(keystate, e.Sequence)
		StopSound(key)
	} else if e.Type == touch.TypeMove {
		if keystate[e.Sequence] != key {
			// Drag has moved out of initial key
			curKey, ok := keystate[e.Sequence] //stop already playing sound
			if ok {
				StopSound(curKey)
			}
			PlaySound(key) // play new key's sound
			keystate[e.Sequence] = key
		}
	}
}

func loadTexture(glctx gl.Context) {
	a, err := asset.Open("key.jpeg")
	if err != nil {
		log.Fatal(err)
	}
	defer a.Close()

	img, _, err := image.Decode(a)
	if err != nil {
		log.Fatal(err)
	}

	rect := img.Bounds()
	rgba := image.NewRGBA(rect)
	draw.Draw(rgba, rect, img, rect.Min, draw.Src)
	tex := glctx.CreateTexture()

	glctx.ActiveTexture(gl.TEXTURE0)
	glctx.BindTexture(gl.TEXTURE_2D, tex)

	glctx.TexImage2D(gl.TEXTURE_2D, 0, 859, 610, gl.RGBA, gl.UNSIGNED_BYTE, rgba.Pix)

	glctx.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	glctx.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	glctx.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	glctx.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)

	glctx.ActiveTexture(gl.TEXTURE0)
	glctx.BindTexture(gl.TEXTURE_2D, tex)
	glctx.TexImage2D(
		gl.TEXTURE_2D, 0,
		rect.Max.X-rect.Min.X, rect.Max.Y-rect.Min.Y,
		gl.RGBA, gl.UNSIGNED_BYTE, rgba.Pix)
}

const vertexShader = `#version 100

attribute vec4 position;
attribute vec2 texCordIn;
varying vec2 texCordOut;

uniform mat4 projection;
uniform mat4 camera;

void main() {
	// offset comes in with x/y values between 0 and 1.
	// position bounds are -1 to 1.
	gl_Position =  projection * camera * position;
  texCordOut = texCordIn;
}`

const fragmentShader = `#version 100
precision mediump float;
uniform vec4 color;
uniform int drawi;

varying lowp vec2 texCordOut;
uniform sampler2D texture;

void main() {
	if(drawi==1){
		gl_FragColor = color;
		return;
	}
	gl_FragColor = texture2D(texture, texCordOut)*color;
}`
