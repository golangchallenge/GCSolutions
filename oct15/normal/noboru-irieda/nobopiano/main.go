// TODO:
// - hittest
// - al reset on onStart

package main

import (
	"encoding/binary"
	"fmt"
	"image"
	"log"
	"sync"
	"time"

	_ "image/png"

	"golang.org/x/mobile/app"
	"golang.org/x/mobile/asset"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"
	"golang.org/x/mobile/event/touch"
	"golang.org/x/mobile/exp/audio/al"
	"golang.org/x/mobile/exp/f32"
	"golang.org/x/mobile/exp/gl/glutil"
	"golang.org/x/mobile/exp/sprite"
	"golang.org/x/mobile/exp/sprite/clock"
	"golang.org/x/mobile/exp/sprite/glsprite"
	"golang.org/x/mobile/gl"
)

const (
	NS    = 1024 // number of samples
	CZ    = 2    // bytes/1-sample for al.FormatMono16
	Fmt   = al.FormatMono16
	QUEUE = 3
)

var (
	startTime = time.Now()

	images *glutil.Images
	eng    sprite.Engine
	scene  *sprite.Node

	sz size.Event
)

var (
	whites  = []int{0, 1, 3, 5, 6, 8, 10, 12, 13}
	blacks  = []int{-1, -1, 2, 4, -1, 7, 9, 11, -1, -1}
	whiltsN = float32(len(whites))
)

func hit2key(x, y float32) int {
	w := float32(sz.WidthPx)
	h := float32(sz.HeightPx)
	offset := w / whiltsN / 2
	if y < h*0.7 {
		index := int(whiltsN * (x + offset) / w)
		if index >= len(blacks) {
			index = len(blacks) - 1
		}
		if key := blacks[index]; key >= 0 {
			return key
		}
	}
	index := int(whiltsN * x / w)
	if index >= len(whites) {
		index = len(whites) - 1
	}
	return whites[index]
}

type Context struct {
	sync.RWMutex
	source    al.Source
	queue     []al.Buffer
	oscilator Oscilator
}

func NewContext(oscilator Oscilator) *Context {
	if err := al.OpenDevice(); err != nil {
		log.Fatal(err)
	}
	s := al.GenSources(1)
	if code := al.Error(); code != 0 {
		log.Fatalln("openal error:", code)
	}
	//s[0].SetGain(s[0].MaxGain())
	//s[0].SetPosition(al.ListenerPosition())
	return &Context{
		source:    s[0],
		queue:     []al.Buffer{},
		oscilator: oscilator,
	}
}

func (c *Context) Play() {
	c.Lock()
	defer c.Unlock()
	n := c.source.BuffersProcessed()
	if n > 0 {
		rm, split := c.queue[:n], c.queue[n:]
		c.queue = split
		c.source.UnqueueBuffers(rm)
		al.DeleteBuffers(rm)
	}
	for len(c.queue) < QUEUE {
		b := al.GenBuffers(1)
		buf := make([]byte, NS*CZ)
		for n := 0; n < NS*CZ; n += CZ {
			v := int16(float32(32767) * c.oscilator())
			binary.LittleEndian.PutUint16(buf[n:n+2], uint16(v))
		}
		b[0].BufferData(Fmt, buf, SampleRate)
		c.source.QueueBuffers(b)
		c.queue = append(c.queue, b...)
	}
	if c.source.State() != al.Playing {
		al.PlaySources(c.source)
	}
}

func (c *Context) Close() {
	c.Lock()
	defer c.Unlock()
	al.StopSources(c.source)
	c.source.UnqueueBuffers(c.queue)
	al.DeleteBuffers(c.queue)
	c.queue = nil
	al.DeleteSources(c.source)
}

func main() {
	pianoPlayer := New([]float32{
		246.941650628,
		261.625565301,
		277.182630977,
		293.664767917,
		311.126983722,
		329.627556913,
		349.228231433,
		369.994422712,
		391.995435982,
		415.30469758,
		440.0,
		466.163761518,
		493.883301256,
		523.251130601,
	})
	app.Main(func(a app.App) {
		var glctx gl.Context
		var pctx *Context
		lastSeq := map[touch.Sequence]int{}
		for e := range a.Events() {
			switch e := a.Filter(e).(type) {
			case lifecycle.Event:
				switch e.Crosses(lifecycle.StageVisible) {
				case lifecycle.CrossOn:
					//fmt.Println("CrossOn")
					glctx, _ = e.DrawContext.(gl.Context)
					onStart(glctx)
					pctx = NewContext(pianoPlayer.GetOscilator())
					a.Send(paint.Event{})
				case lifecycle.CrossOff:
					//fmt.Println("CrossOff")
					pctx.Close()
					onStop()
					glctx = nil
				}
			case size.Event:
				sz = e
			case paint.Event:
				if glctx == nil || e.External {
					continue
				}
				onPaint(glctx)
				a.Publish()
				pctx.Play()
				repaint(a) // keep animating
			case touch.Event:
				key := hit2key(e.X, e.Y)
				old, ok := lastSeq[e.Sequence]
				keyChanged := !ok || (old != key)
				if ok && keyChanged {
					pianoPlayer.NoteOff(old)
					fmt.Println("notechoff:", old, e.X, e.Y, e.Type, e.Sequence)
				}
				lastSeq[e.Sequence] = key
				if e.Type == touch.TypeBegin || (keyChanged && e.Type == touch.TypeMove) {
					pianoPlayer.NoteOn(key)
					fmt.Println("noteon:", key, e.X, e.Y, e.Type, e.Sequence)
				}
				if e.Type == touch.TypeEnd {
					pianoPlayer.NoteOff(key)
					fmt.Println("noteoff:", key, e.X, e.Y, e.Type, e.Sequence)
				}
			}
		}
	})
}

var once = sync.Once{}

func onStart(glctx gl.Context) {
	images = glutil.NewImages(glctx)
	eng = glsprite.Engine(images)
	loadScene()
}

func onStop() {
	eng.Release()
	images.Release()
}

func onPaint(glctx gl.Context) {
	glctx.ClearColor(1, 1, 1, 1)
	glctx.Clear(gl.COLOR_BUFFER_BIT)
	now := clock.Time(time.Since(startTime) * 60 / time.Second)
	eng.Render(scene, now, sz)
}

func loadScene() {
	keyboard := loadKeyboard()
	scene = &sprite.Node{}
	eng.Register(scene)
	eng.SetTransform(scene, f32.Affine{
		{1, 0, 0},
		{0, 1, 0},
	})

	n := &sprite.Node{}
	eng.Register(n)
	scene.AppendChild(n)
	// TODO: Shouldn't arranger pass the size.Event?
	n.Arranger = arrangerFunc(func(eng sprite.Engine, n *sprite.Node, t clock.Time) {
		eng.SetSubTex(n, keyboard)
		eng.SetTransform(n, f32.Affine{
			{float32(sz.WidthPt), 0, 0},
			{0, float32(sz.HeightPt), 0},
		})
	})
}

func loadKeyboard() sprite.SubTex {
	a, err := asset.Open("piano-octave.png")
	if err != nil {
		log.Fatal(err)
	}
	defer a.Close()

	img, _, err := image.Decode(a)
	if err != nil {
		log.Fatal(err)
	}
	t, err := eng.LoadTexture(img)
	if err != nil {
		log.Fatal(err)
	}
	return sprite.SubTex{t, image.Rect(0, 0, 500, 249)}
}

type arrangerFunc func(e sprite.Engine, n *sprite.Node, t clock.Time)

func (a arrangerFunc) Arrange(e sprite.Engine, n *sprite.Node, t clock.Time) { a(e, n, t) }
