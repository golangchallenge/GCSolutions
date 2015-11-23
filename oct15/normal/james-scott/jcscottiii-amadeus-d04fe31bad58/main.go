package main

import (
	"golang.org/x/mobile/app"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"
	"golang.org/x/mobile/event/touch"
	"golang.org/x/mobile/exp/gl/glutil"
	"golang.org/x/mobile/gl"

	"github.com/golangchallenge/GCSolutions/oct15/normal/james-scott/jcscottiii-amadeus-d04fe31bad58/layers"
	"github.com/golangchallenge/GCSolutions/oct15/normal/james-scott/jcscottiii-amadeus-d04fe31bad58/util"
	"log"
)

var (
	frameData util.FrameData
	manager   *layers.Manager
)

func main() {
	var glCtx gl.Context
	visible := false
	sz := util.InitSizeEvent
	app.Main(func(a app.App) {
		for e := range a.Events() {
			switch event := a.Filter(e).(type) {
			case lifecycle.Event:
				switch event.Crosses(lifecycle.StageVisible) {
				case lifecycle.CrossOn:
					visible = true
					glCtx, _ = event.DrawContext.(gl.Context)
					onStart(glCtx)
				case lifecycle.CrossOff:
					visible = false
					onStop(glCtx)
				}
			case size.Event:
				sz = event
				// Always want to make sure we draw the keys in which there is the most width.
				if (sz.WidthPx >= sz.HeightPx) && ((frameData.Orientation == util.Portrait) ||
					(frameData.Orientation == util.UnsetOrientation)) {
					// Most likely the phone is landscape and need to switch flag.
					frameData.Orientation = util.Landscape
				} else if (sz.WidthPx < sz.HeightPx) &&
					((frameData.Orientation == util.Landscape) ||
						(frameData.Orientation == util.UnsetOrientation)) {
					// Most likely the phone is portrait and need to switch flag.
					log.Printf("going portrait\n")
					frameData.Orientation = util.Portrait
				}
			case paint.Event:
				onPaint(glCtx, sz)
				a.Publish()
				if visible {
					a.Send(paint.Event{})
				}
			case touch.Event:
				log.Printf("X: %f Y: %f \n", event.X, event.Y)
				worldX, worldY := util.ConvertToOpenGLCoords(sz, event)
				manager.TouchLayers(worldX, worldY, event, frameData)
				log.Printf("X: %f Y: %f \n", worldX, worldY)
			}
		}

	})
}

func onStart(glctx gl.Context) {
	var err error
	frameData.Program, err = glutil.CreateProgram(glctx, vertexShader, fragmentShader)
	if err != nil {
		log.Printf("error creating GL program: %v", err)
		return
	}

	frameData.Position = glctx.GetAttribLocation(frameData.Program, "position")
	frameData.Color = glctx.GetUniformLocation(frameData.Program, "color")
	frameData.Offset = glctx.GetUniformLocation(frameData.Program, "offset")
	manager = layers.NewLayerManager(glctx)
	manager.StartLayers()

}

func onStop(glctx gl.Context) {
	manager.StopLayers()
	glctx.DeleteProgram(frameData.Program)
}

func onPaint(glctx gl.Context, sz size.Event) {
	glctx.ClearColor(0.5, 0.5, 0.5, 1)
	glctx.Clear(gl.COLOR_BUFFER_BIT)

	glctx.UseProgram(frameData.Program)

	glctx.Uniform4f(frameData.Color, 0.5, 0.5, 0.5, 1)

	// Put ourselves in the +x/+y quadrant (upper right quadrant)
	glctx.Uniform2f(frameData.Offset, 0.0, 1.0)

	glctx.EnableVertexAttribArray(frameData.Position)
	manager.PaintLayers(sz, frameData)
	// End drawing
	glctx.DisableVertexAttribArray(frameData.Position)

}

const vertexShader = `#version 100
uniform vec2 offset;
attribute vec4 position;
void main() {
	// offset comes in with x/y values between 0 and 1.
	// position bounds are -1 to 1.
	vec4 offset4 = vec4(2.0*offset.x-1.0, 1.0-2.0*offset.y, 0, 0);
	gl_Position = position + offset4;
}`

const fragmentShader = `#version 100
precision mediump float;
uniform vec4 color;
void main() {
	gl_FragColor = color;
}`
