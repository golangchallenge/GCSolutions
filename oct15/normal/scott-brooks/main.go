// +build darwin linux

// Golang challenge 7
// Run with go run *.go
// or gomobile install and it should work on Android

// Notes: Click the notes near the bottom for now, or the sharp keys trigger.
// If you click and drag off of the piano you can rotate it around and still click on it.
// Also the light changes depending on which key you hit.
package main

import (
	"encoding/binary"
	"log"
	"os"
	"time"

	"github.com/go-gl/mathgl/mgl32"
	"golang.org/x/mobile/app"
	"golang.org/x/mobile/event/key"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"
	"golang.org/x/mobile/event/touch"
	"golang.org/x/mobile/exp/app/debug"
	"golang.org/x/mobile/exp/audio/al"
	"golang.org/x/mobile/exp/f32"
	"golang.org/x/mobile/exp/gl/glutil"
	"golang.org/x/mobile/gl"
)

var (
	images         *glutil.Images
	fps            *debug.FPS
	program        gl.Program
	position       gl.Attrib
	color          gl.Attrib
	normals        gl.Attrib
	projection     gl.Uniform
	model          gl.Uniform
	normalMatrix   gl.Uniform
	view           gl.Uniform
	tint           gl.Uniform
	lightPos       gl.Uniform
	lightIntensity gl.Uniform
	triBuf         gl.Buffer

	touchX   float32
	touchY   float32
	worldPos mgl32.Vec3

	width  int
	height int

	projectionMtx mgl32.Mat4

	arcball      ArcBall
	white        mgl32.Vec4
	red          mgl32.Vec4
	lastUpdate   time.Time
	beganOnPiano bool
)

type glLight struct {
	Pos         mgl32.Vec3
	Intensities mgl32.Vec3
}

var light = glLight{Pos: mgl32.Vec3{5, 10, 0}, Intensities: mgl32.Vec3{0.5, 0, 0}}

func main() {
	app.Main(func(a app.App) {
		var glctx gl.Context
		visible, sz := false, size.Event{}
		for e := range a.Events() {
			switch e := a.Filter(e).(type) {
			case lifecycle.Event:
				switch e.Crosses(lifecycle.StageVisible) {
				case lifecycle.CrossOn:
					visible = true
					glctx, _ = e.DrawContext.(gl.Context)
					onStart(glctx)
				case lifecycle.CrossOff:
					visible = false
					onStop(glctx)
				}
			case size.Event:
				sz = e
				width = sz.WidthPx
				height = sz.HeightPx
				touchX = float32(sz.WidthPx / 2)
				touchY = float32(sz.HeightPx / 2)
				if glctx != nil {
					glctx.Viewport(0, 0, sz.WidthPx, sz.HeightPx)
				}
			case paint.Event:
				onPaint(glctx, sz)
				a.Publish()
				if visible {
					n := time.Now()
					piano.update(n.Sub(lastUpdate))
					lastUpdate = n
					// Drive the animation by preparing to paint the next frame
					// after this one is shown.
					//
					// TODO: is paint.Event the right thing to send? Should we
					// have a dedicated publish.Event type? Should App.Publish
					// take an optional event sender and send a publish.Event?
					a.Send(paint.Event{})
				}
			case touch.Event:
				touchX = e.X
				touchY = e.Y
				worldRayStart, worldRayEnd := unproject(glctx, e.X, e.Y)

				if !piano.HandleTouch(worldRayStart, worldRayEnd, e) && !beganOnPiano {
					if e.Type == touch.TypeMove {
						arcball.move(int(e.X), int(e.Y))
					} else if e.Type == touch.TypeBegin {
						arcball.begin(int(e.X), int(e.Y))
					} else {
						arcball.end()
					}
				} else if e.Type == touch.TypeBegin {
					beganOnPiano = true
				} else if e.Type == touch.TypeEnd {
					beganOnPiano = false
				}

			case key.Event:
				if e.Code == key.CodeEscape {
					visible = false
					onStop(glctx)
					os.Exit(0)
				}
			}
		}
	})
}

func onStart(glctx gl.Context) {
	var err error
	program, err = glutil.CreateProgram(glctx, vertexShader, fragmentShader)
	if err != nil {
		log.Printf("error creating GL program: %v", err)
		return
	}

	glctx.Enable(gl.DEPTH_TEST)

	triBuf = glctx.CreateBuffer()
	glctx.BindBuffer(gl.ARRAY_BUFFER, triBuf)
	glctx.BufferData(gl.ARRAY_BUFFER, triangleData, gl.STATIC_DRAW)

	position = glctx.GetAttribLocation(program, "vPos")
	color = glctx.GetAttribLocation(program, "vCol")
	normals = glctx.GetAttribLocation(program, "vNorm")

	projection = glctx.GetUniformLocation(program, "proj")
	view = glctx.GetUniformLocation(program, "view")
	model = glctx.GetUniformLocation(program, "model")
	tint = glctx.GetUniformLocation(program, "tint")
	normalMatrix = glctx.GetUniformLocation(program, "normalMatrix")
	lightIntensity = glctx.GetUniformLocation(program, "light.intensities")
	lightPos = glctx.GetUniformLocation(program, "light.position")

	arcball = NewArcBall(mgl32.Vec3{0, 0, 0}, mgl32.Vec3{0, 10, 10}, mgl32.Vec3{0, 1, 0})

	white = mgl32.Vec4{1.0, 1.0, 1.0, 1.0}
	red = mgl32.Vec4{1.0, 0.0, 0.0, 1.0}

	lastUpdate = time.Now()

	images = glutil.NewImages(glctx)
	fps = debug.NewFPS(images)

	err = al.OpenDevice()
	if err != nil {
		log.Printf("Err: %+v", err)
	}
	al.SetListenerPosition(al.Vector{0, 0, 0})
	al.SetListenerGain(1.0)
	piano = NewPiano()
}

func onStop(glctx gl.Context) {
	glctx.DeleteProgram(program)
	glctx.DeleteBuffer(triBuf)
	fps.Release()
	images.Release()
}

func unproject(glctx gl.Context, x, y float32) (mgl32.Vec3, mgl32.Vec3) {
	var wx, wy float32
	var viewport [4]int32
	glctx.GetIntegerv(viewport[:], gl.VIEWPORT)

	wx = x
	wy = float32(viewport[3]) - y

	posStart, err := mgl32.UnProject(mgl32.Vec3{wx, wy, 0}, arcball.getMtx(), projectionMtx, int(viewport[0]), int(viewport[1]), int(viewport[2]), int(viewport[3]))
	posEnd, err := mgl32.UnProject(mgl32.Vec3{wx, wy, 1}, arcball.getMtx(), projectionMtx, int(viewport[0]), int(viewport[1]), int(viewport[2]), int(viewport[3]))

	if err != nil {
		log.Printf("unable to unproject: %+v", err)
	}

	return posStart, posEnd
}

func onPaint(glctx gl.Context, sz size.Event) {

	glctx.Viewport(0, 0, sz.WidthPx, sz.HeightPx)
	glctx.ClearColor(0.5, 0.5, 0.5, 1)
	glctx.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	glctx.UseProgram(program)

	projectionMtx = mgl32.Perspective(45, float32(width)/float32(height), 0.1, 100)

	arcBallMtx := arcball.getMtx()

	glctx.UniformMatrix4fv(projection, projectionMtx[:])

	glctx.UniformMatrix4fv(view, arcBallMtx[:])

	glctx.BindBuffer(gl.ARRAY_BUFFER, triBuf)
	glctx.EnableVertexAttribArray(position)
	glctx.EnableVertexAttribArray(color)
	glctx.EnableVertexAttribArray(normals)

	vertSize := 4 * (coordsPerVertex + colorPerVertex + normalsPerVertex)

	glctx.VertexAttribPointer(position, coordsPerVertex, gl.FLOAT, false, vertSize, 0)
	glctx.VertexAttribPointer(color, colorPerVertex, gl.FLOAT, false, vertSize, 4*coordsPerVertex)
	glctx.VertexAttribPointer(normals, normalsPerVertex, gl.FLOAT, false, vertSize, 4*(coordsPerVertex+colorPerVertex))

	glctx.DepthMask(true)

	glctx.Uniform3fv(lightPos, light.Pos[:])
	glctx.Uniform3fv(lightIntensity, light.Intensities[:])

	for _, k := range piano.Keys {
		glctx.Uniform4fv(tint, k.Color[:])

		mtx := k.GetMtx()
		normMat := mtx.Mat3().Inv().Transpose()
		glctx.UniformMatrix3fv(normalMatrix, normMat[:])
		glctx.UniformMatrix4fv(model, mtx[:])
		glctx.DrawArrays(gl.TRIANGLES, 0, len(triangleData)/vertSize)
	}

	modelMtx := mgl32.Ident4()
	modelMtx = modelMtx.Mul4(mgl32.Translate3D(worldPos.X(), worldPos.Y(), worldPos.Z()))
	modelMtx = modelMtx.Mul4(mgl32.Scale3D(0.5, 0.5, 0.5))

	/*
		glctx.Uniform4fv(tint, red[:])
		// Disable depthmask so we dont get the pixel depth of the cursor cube
		glctx.DepthMask(false)
		glctx.UniformMatrix4fv(model, modelMtx[:])
		glctx.DepthMask(true)
	*/

	glctx.DisableVertexAttribArray(position)
	glctx.DisableVertexAttribArray(color)
	glctx.DisableVertexAttribArray(normals)

	fps.Draw(sz)
}

var triangleData = f32.Bytes(binary.LittleEndian,
	// front
	-1.0, -1.0, 1.0, 0.98, 0.972, 0.94, 1, 0, 0, 1,
	1.0, -1.0, 1.0, 0.98, 0.972, 0.94, 1, 0, 0, 1,
	1.0, 1.0, 1.0, 0.98, 0.972, 0.94, 1, 0, 0, 1,
	1.0, 1.0, 1.0, 0.98, 0.972, 0.94, 1, 0, 0, 1,
	-1.0, 1.0, 1.0, 0.98, 0.972, 0.94, 1, 0, 0, 1,
	-1.0, -1.0, 1.0, 0.98, 0.972, 0.94, 1, 0, 0, 1,
	// top
	-1.0, 1.0, 1.0, 0.98, 0.972, 0.94, 1, 0, 1, 0,
	1.0, 1.0, 1.0, 0.98, 0.972, 0.94, 1, 0, 1, 0,
	1.0, 1.0, -1.0, 0.98, 0.972, 0.94, 1, 0, 1, 0,
	1.0, 1.0, -1.0, 0.98, 0.972, 0.94, 1, 0, 1, 0,
	-1.0, 1.0, -1.0, 0.98, 0.972, 0.94, 1, 0, 1, 0,
	-1.0, 1.0, 1.0, 0.98, 0.972, 0.94, 1, 0, 1, 0,
	// back
	-1.0, -1.0, -1.0, 0.98, 0.972, 0.94, 1, 0, 0, -1,
	1.0, -1.0, -1.0, 0.98, 0.972, 0.94, 1, 0, 0, -1,
	1.0, 1.0, -1.0, 0.98, 0.972, 0.94, 1, 0, 0, -1,
	1.0, 1.0, -1.0, 0.98, 0.972, 0.94, 1, 0, 0, -1,
	-1.0, 1.0, -1.0, 0.98, 0.972, 0.94, 1, 0, 0, -1,
	-1.0, -1.0, -1.0, 0.98, 0.972, 0.94, 1, 0, 0, -1,
	// bottom
	-1.0, -1.0, -1.0, 0.98, 0.972, 0.94, 1, 0, -1, 0,
	1.0, -1.0, -1.0, 0.98, 0.972, 0.94, 1, 0, -1, 0,
	1.0, -1.0, 1.0, 0.98, 0.972, 0.94, 1, 0, -1, 0,
	1.0, -1.0, 1.0, 0.98, 0.972, 0.94, 1, 0, -1, 0,
	-1.0, -1.0, 1.0, 0.98, 0.972, 0.94, 1, 0, -1, 0,
	-1.0, -1.0, -1.0, 0.98, 0.972, 0.94, 1, 0, -1, 0,
	// left
	-1.0, -1.0, -1.0, 0.98, 0.972, 0.94, 1, 1, 0, 0,
	-1.0, -1.0, 1.0, 0.98, 0.972, 0.94, 1, 1, 0, 0,
	-1.0, 1.0, 1.0, 0.98, 0.972, 0.94, 1, 1, 0, 0,
	-1.0, 1.0, 1.0, 0.98, 0.972, 0.94, 1, 1, 0, 0,
	-1.0, 1.0, -1.0, 0.98, 0.972, 0.94, 1, 1, 0, 0,
	-1.0, -1.0, -1.0, 0.98, 0.972, 0.94, 1, 1, 0, 0,
	// right
	1.0, -1.0, 1.0, 0.98, 0.972, 0.94, 1, -1, 0, 0,
	1.0, -1.0, -1.0, 0.98, 0.972, 0.94, 1, -1, 0, 0,
	1.0, 1.0, -1.0, 0.98, 0.972, 0.94, 1, -1, 0, 0,
	1.0, 1.0, -1.0, 0.98, 0.972, 0.94, 1, -1, 0, 0,
	1.0, 1.0, 1.0, 0.98, 0.972, 0.94, 1, -1, 0, 0,
	1.0, -1.0, 1.0, 0.98, 0.972, 0.94, 1, -1, 0, 0,
)

const (
	coordsPerVertex  = 3
	colorPerVertex   = 4
	normalsPerVertex = 3
)

const vertexShader = `#version 100
precision mediump float;

uniform mat4 proj;
uniform mat4 view;
uniform mat4 model;
uniform vec4 tint;

attribute vec4 vPos;
attribute vec4 vCol;
attribute vec4 vNorm;

varying vec4 fCol;
varying vec3 fVert;
varying vec3 fNormal;
void main() {
	fCol = vCol;
	fCol *= tint;

	fVert = vPos.xyz;
	fNormal = vNorm.xyz;

	gl_Position = proj * view * model * vPos;
}`

const fragmentShader = `#version 100
precision mediump float;

uniform mat4 model;
uniform mat3 normalMatrix;

uniform struct Light {
	vec3 position;
	vec3 intensities;
} light;

varying lowp vec4 fCol;
varying mediump vec3 fVert;
varying mediump vec3 fNormal;


void main() {
    //calculate normal in world coordinates
    vec3 normal = normalize(normalMatrix * fNormal);
    
    //calculate the location of this fragment (pixel) in world coordinates
    vec3 fragPosition = vec3(model * vec4(fVert, 1));
    
    //calculate the vector from this pixels surface to the light source
    vec3 surfaceToLight = light.position - fragPosition;

    //calculate the cosine of the angle of incidence
    float brightness = dot(normal, surfaceToLight) / (length(surfaceToLight) * length(normal));
    brightness = clamp(brightness, 0.0, 1.0);

    //calculate final color of the pixel, based on:
    // 1. The angle of incidence: brightness
    // 2. The color/intensities of the light: light.intensities
    // 3. The texture and texture coord: texture(tex, fragTexCoord)
    gl_FragColor = vec4(brightness * light.intensities * fCol.rgb, fCol.a);
}

`

/*
const fragmentShader = `#version 100
precision mediump float;
varying lowp vec4 fCol;
void main() {
	gl_FragColor = fCol;
}`
*/
