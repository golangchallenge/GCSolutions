package main

import (
	"math"

	"golang.org/x/mobile/exp/f32"

	"github.com/go-gl/mathgl/mgl32"
)

// ArcBall lets you rotate around a point on screen(origin) in a nice way
type ArcBall struct {
	lastMx, lastMy int
	mtx            mgl32.Mat4
	viewMtx        mgl32.Mat4
	on             bool
	origin         mgl32.Vec3
	camera         mgl32.Vec3
	up             mgl32.Vec3
}

// NewArcBall creates a new ArcBall at camera looking at origin
func NewArcBall(origin, camera, up mgl32.Vec3) ArcBall {
	a := ArcBall{mtx: mgl32.Ident4(), origin: origin, camera: camera, up: up}
	a.updateViewMtx()
	return a
}

func getArcballVector(x, y int) mgl32.Vec3 {
	p := mgl32.Vec3{float32(1.0*float32(x)/float32(width)*2 - 1.0),
		float32(1.0*float32(y)/float32(height)*2 - 1.0),
		0}

	p[1] = -p[1]

	opSquared := p[0]*p[0] + p[1]*p[1]
	if opSquared <= 1 {
		p[2] = f32.Sqrt(1 - opSquared)
	} else {
		p.Normalize()
	}

	return p
}

func (a *ArcBall) updateViewMtx() {
	a.viewMtx = mgl32.LookAtV(a.camera, a.origin, a.up)
}

func (a *ArcBall) begin(mx, my int) {
	a.on = true
	a.lastMx = mx
	a.lastMy = my
}
func (a *ArcBall) end() {
	a.on = false
}

func (a *ArcBall) move(mx, my int) {
	if mx != a.lastMx || my != a.lastMy {
		va := getArcballVector(a.lastMx, a.lastMy)
		vb := getArcballVector(mx, my)

		angle := float32(math.Acos(math.Min(1.0, float64(va.Dot(vb)))))

		axisInCameraCoords := va.Cross(vb).Normalize()
		camera2object := a.mtx.Mul4(a.viewMtx).Mat3().Inv()
		axisInObjectCoords := camera2object.Mul3x1(axisInCameraCoords).Normalize()
		a.mtx = a.mtx.Mul4(mgl32.HomogRotate3D(angle, axisInObjectCoords))

		a.lastMx = mx
		a.lastMy = my
	}
}

func (a *ArcBall) getMtx() mgl32.Mat4 {

	return a.viewMtx.Mul4(a.mtx)
}
