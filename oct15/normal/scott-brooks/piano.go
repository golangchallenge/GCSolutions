package main

import (
	"bytes"
	"encoding/binary"
	"math"
	"os"
	"time"

	"github.com/go-gl/mathgl/mgl32"
	"golang.org/x/mobile/event/touch"
	"golang.org/x/mobile/exp/audio/al"
	"golang.org/x/mobile/exp/f32"
)

type PianoKey struct {
	BBox       bbox
	Color      mgl32.Vec4
	Pos        mgl32.Vec3
	LightColor mgl32.Vec3
	Angle      float32
	Frequency  float32
	Finger     int
	buffers    []al.Buffer
	source     al.Source
	white      bool
}

type bbox [2]mgl32.Vec3

func (b bbox) Contains(pos mgl32.Vec3) bool {
	return pos.X() > b[0].X() &&
		pos.Y() > b[0].Y() &&
		pos.Z() > b[0].Z() &&
		pos.X() < b[1].X() &&
		pos.Y() < b[1].Y() &&
		pos.Z() < b[1].Z()

}

type fullPiano struct {
	Keys []PianoKey
}

// NewPianoKey returns a key for our piano.
func NewPianoKey(pos mgl32.Vec3, lightColor mgl32.Vec3, white bool, freq float32) PianoKey {
	var color mgl32.Vec4
	var keySize float32
	if white {
		color = mgl32.Vec4{0.98, 0.97, 0.94}
		keySize = 2
	} else {
		color = mgl32.Vec4{0.1, 0.1, 0.1, 1.0}
		keySize = 1

	}
	pk := PianoKey{Pos: pos, Angle: 0, Color: color, Frequency: freq, Finger: -1, white: white, LightColor: lightColor}
	pk.BBox[0] = pos.Sub(mgl32.Vec3{0.5, 0.6, keySize})
	pk.BBox[1] = pos.Add(mgl32.Vec3{0.5, 0.6, keySize})
	pk.source = al.GenSources(1)[0]
	pk.source.SetGain(1.0)
	pk.source.SetPosition(al.Vector{pos.X(), pos.Y(), pos.Z()})
	pk.source.SetVelocity(al.Vector{})
	pk.buffers = al.GenBuffers(3)

	var samples [1024 * 16]int16

	sampleRate := 44100
	amplitude := float32(0.8 * 0x7FFF)

	for i := 0; i < len(samples); i++ {
		val := f32.Sin((2.0 * math.Pi * freq) / float32(sampleRate) * float32(i))
		samples[i] = int16(amplitude * val)
	}

	buf := &bytes.Buffer{}
	binary.Write(buf, binary.LittleEndian, &samples)
	pk.buffers[0].BufferData(al.FormatMono16, buf.Bytes(), 44100)

	f, _ := os.Create("audio.raw")
	binary.Write(f, binary.LittleEndian, &samples)
	f.Close()

	return pk
}

var piano fullPiano

func NewPiano() fullPiano {
	return fullPiano{
		Keys: []PianoKey{
			NewPianoKey(mgl32.Vec3{-5.5, 0, 0}, mgl32.Vec3{0, 0, 1}, true, 440),
			NewPianoKey(mgl32.Vec3{-4.95, 0.5, 0}, mgl32.Vec3{0, 1, 0}, false, 466.166),
			NewPianoKey(mgl32.Vec3{-4.4, 0, 0}, mgl32.Vec3{0, 1, 1}, true, 493.883),
			NewPianoKey(mgl32.Vec3{-3.3, 0, 0}, mgl32.Vec3{1, 0, 0}, true, 523.251),
			NewPianoKey(mgl32.Vec3{-2.75, 0.5, 0}, mgl32.Vec3{1, 0, 1}, false, 554.365),
			NewPianoKey(mgl32.Vec3{-2.2, 0, 0}, mgl32.Vec3{1, 1, 0}, true, 587.330),
			NewPianoKey(mgl32.Vec3{-1.65, 0.5, 0}, mgl32.Vec3{1, 1, 1}, false, 622.254),
			NewPianoKey(mgl32.Vec3{-1.1, 0, 0}, mgl32.Vec3{1, 1, 0}, true, 659.255),
			NewPianoKey(mgl32.Vec3{0, 0, 0}, mgl32.Vec3{1, 0, 1}, true, 698.456),
			NewPianoKey(mgl32.Vec3{0.55, 0.5, 0}, mgl32.Vec3{1, 0, 0}, false, 739.989),
			NewPianoKey(mgl32.Vec3{1.1, 0, 0}, mgl32.Vec3{0, 1, 1}, true, 783.991),
			NewPianoKey(mgl32.Vec3{1.65, 0.5, 0}, mgl32.Vec3{0, 1, 0}, false, 830.609),
			NewPianoKey(mgl32.Vec3{2.2, 0, 0}, mgl32.Vec3{0, 0, 1}, true, 880),
			NewPianoKey(mgl32.Vec3{2.75, 0.5, 0}, mgl32.Vec3{0, 0, 0}, false, 932.328),
			NewPianoKey(mgl32.Vec3{3.3, 0, 0}, mgl32.Vec3{0, 1, 0}, true, 987.767),
			NewPianoKey(mgl32.Vec3{4.4, 0, 0}, mgl32.Vec3{1, 0, 1}, true, 1046.50),
			NewPianoKey(mgl32.Vec3{4.95, 0.5, 0}, mgl32.Vec3{1, 1, 0}, false, 1108.73),
			NewPianoKey(mgl32.Vec3{5.5, 0, 0}, mgl32.Vec3{0, 1, 1}, true, 1174.66),
		},
	}
}

func (p *fullPiano) update(dt time.Duration) {
	for i := range p.Keys {
		if p.Keys[i].Angle < 0.001 {
			p.Keys[i].Angle = 0
		} else {
			p.Keys[i].Angle *= 0.8 - float32(dt.Seconds()) //float32(math.Sin(float64(dt/time.Millisecond)*float64(i)) * 0.1)
		}
		if p.Keys[i].Finger != -1 {
			p.Keys[i].Hold()
		}
	}
}

func (k *PianoKey) Hit(rayStart, rayEnd mgl32.Vec3) (bool, mgl32.Vec3) {
	hit, pos := checkLineBox(k.BBox[0], k.BBox[1], rayStart, rayEnd)
	return hit, pos
}

func (k *PianoKey) Press(f int) {
	k.Angle = 3.14 / 16
	k.Finger = f
	light.Intensities = k.LightColor
	bp := k.source.BuffersProcessed()
	if bp > 0 {
		b := make([]al.Buffer, bp)
		k.source.UnqueueBuffers(b)
	}
	if k.source.BuffersQueued() == 0 {
		k.source.QueueBuffers([]al.Buffer{k.buffers[0]})
		al.PlaySources(k.source)
	}
}

func (k *PianoKey) Release() {
	k.Finger = -1
	//k.source.QueueBuffers([]al.Buffer{k.buffers[0]})
}

func (k *PianoKey) Hold() {
	k.Angle = 3.14 / 16
	bp := k.source.BuffersProcessed()
	if bp > 0 {
		b := make([]al.Buffer, bp)
		k.source.UnqueueBuffers(b)
	}
	if k.source.BuffersQueued() < 2 {
		k.source.QueueBuffers([]al.Buffer{k.buffers[0]})
	}
}

func (k PianoKey) GetMtx() mgl32.Mat4 {
	axis := mgl32.Vec3{1, 0, 0}
	keyLen := float32(2.0)
	if !k.white {
		keyLen = 1.0
	}
	modelMtx := mgl32.Ident4()
	modelMtx = modelMtx.Mul4(mgl32.Translate3D(k.Pos[0], k.Pos[1], k.Pos[2]-2))
	modelMtx = modelMtx.Mul4(mgl32.HomogRotate3D(k.Angle, axis))
	modelMtx = modelMtx.Mul4(mgl32.Translate3D(0, 0, keyLen))
	modelMtx = modelMtx.Mul4(mgl32.Scale3D(0.5, 0.5, keyLen))

	return modelMtx
}

func getIntersection(fDst1, fDst2 float32, P1, P2 mgl32.Vec3, Hit *mgl32.Vec3) bool {
	if (fDst1 * fDst2) >= 0.0 {
		return false
	}
	if fDst1 == fDst2 {
		return false
	}
	*Hit = P1.Add((P2.Sub(P1)).Mul(-fDst1 / (fDst2 - fDst1)))
	return true
}

func inBox(Hit, B1, B2 mgl32.Vec3, Axis int) bool {
	if Axis == 1 && Hit.Z() > B1.Z() && Hit.Z() < B2.Z() && Hit.Y() > B1.Y() && Hit.Y() < B2.Y() {
		return true
	}
	if Axis == 2 && Hit.Z() > B1.Z() && Hit.Z() < B2.Z() && Hit.X() > B1.X() && Hit.X() < B2.X() {
		return true
	}
	if Axis == 3 && Hit.X() > B1.X() && Hit.X() < B2.X() && Hit.Y() > B1.Y() && Hit.Y() < B2.Y() {
		return true
	}
	return false
}

// returns true if line (L1, L2) intersects with the box (B1, B2)
// returns intersection point in Hit
func checkLineBox(B1, B2, L1, L2 mgl32.Vec3) (bool, mgl32.Vec3) {
	var Hit mgl32.Vec3

	if L2.X() < B1.X() && L1.X() < B1.X() {
		return false, mgl32.Vec3{}
	}
	if L2.X() > B2.X() && L1.X() > B2.X() {
		return false, mgl32.Vec3{}
	}
	if L2.Y() < B1.Y() && L1.Y() < B1.Y() {
		return false, mgl32.Vec3{}
	}
	if L2.Y() > B2.Y() && L1.Y() > B2.Y() {
		return false, mgl32.Vec3{}
	}
	if L2.Z() < B1.Z() && L1.Z() < B1.Z() {
		return false, mgl32.Vec3{}
	}
	if L2.Z() > B2.Z() && L1.Z() > B2.Z() {
		return false, mgl32.Vec3{}
	}
	if L1.X() > B1.X() && L1.X() < B2.X() &&
		L1.Y() > B1.Y() && L1.Y() < B2.Y() &&
		L1.Z() > B1.Z() && L1.Z() < B2.Z() {
		Hit = L1
		return true, Hit
	}
	if (getIntersection(L1.X()-B1.X(), L2.X()-B1.X(), L1, L2, &Hit) && inBox(Hit, B1, B2, 1)) ||
		(getIntersection(L1.Y()-B1.Y(), L2.Y()-B1.Y(), L1, L2, &Hit) && inBox(Hit, B1, B2, 2)) ||
		(getIntersection(L1.Z()-B1.Z(), L2.Z()-B1.Z(), L1, L2, &Hit) && inBox(Hit, B1, B2, 3)) ||
		(getIntersection(L1.X()-B2.X(), L2.X()-B2.X(), L1, L2, &Hit) && inBox(Hit, B1, B2, 1)) ||
		(getIntersection(L1.Y()-B2.Y(), L2.Y()-B2.Y(), L1, L2, &Hit) && inBox(Hit, B1, B2, 2)) ||
		(getIntersection(L1.Z()-B2.Z(), L2.Z()-B2.Z(), L1, L2, &Hit) && inBox(Hit, B1, B2, 3)) {
		return true, Hit
	}

	return false, mgl32.Vec3{}
}

func (p *fullPiano) HandleTouch(rayStart mgl32.Vec3, rayEnd mgl32.Vec3, e touch.Event) bool {
	k := p.GetKeyFromRay(rayStart, rayEnd)
	p.clearKeysForFinger(int(e.Sequence))
	if k != nil && (e.Type == touch.TypeBegin || e.Type == touch.TypeMove) {
		k.Press(int(e.Sequence))
		return true
	} else {
		return false
	}
}

func (p *fullPiano) clearKeysForFinger(f int) {
	for i, k := range p.Keys {
		if k.Finger == f {
			p.Keys[i].Release()
		}
	}
}

func (p *fullPiano) GetKeyFromRay(rayStart, rayEnd mgl32.Vec3) *PianoKey {
	minDist := float32(10000)
	minDistIdx := -1

	for i, k := range p.Keys {
		hit, pos := k.Hit(rayStart, rayEnd)
		if hit {
			dist := rayStart.Sub(pos).Len()
			if dist < minDist {
				minDistIdx = i
				minDist = dist
			}
		}
	}

	if minDistIdx != -1 {
		return &p.Keys[minDistIdx]
	}

	return nil
}
