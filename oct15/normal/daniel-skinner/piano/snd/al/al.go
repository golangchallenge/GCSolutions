package al

import (
	"fmt"
	"log"
	"math"
	"time"

	"dasa.cc/piano/snd"

	"golang.org/x/mobile/exp/audio/al"
)

// TODO much of this functionality needs to be generalized
// but that can wait to be hashed out until other backends
// are supported, e.g. jack

const maxbufs = 80 // arbitrary soft limit

var hwa *openal

// Buffer provides adaptive buffering for real-time responses that outpace
// openal's ability to report a buffer as processed.
// Constant-time synchronization is left up to the caller.
// TODO consider naming SourceBuffers and embedding al.Source
type Buffer struct {
	src  al.Source
	bufs []al.Buffer // intrinsic to src
	size int         // size of buffers returned by Get
	idx  int         // last known index still processing or ready for reuse
}

// Get returns a slice of buffers of length b.size ready to receive data and be queued.
// If b.src reports processing at-least as many buffers as b.size, buffers up to b.size
// will be unqueued for reuse. Otherwise, new buffers will be generated.
//
// If the length of b.bufs has grown to be greater than maxbufs, a nil slice is returned.
func (b *Buffer) Get() (bufs []al.Buffer) {
	if proc := int(b.src.BuffersProcessed()); proc >= b.size {
		// advance by size, BuffersProcessed will report that many less next time.
		bufs = b.bufs[b.idx : b.idx+b.size]
		b.src.UnqueueBuffers(bufs)
		if code := al.Error(); code != 0 {
			log.Printf("snd/al: unqueue buffers failed [err=%v]\n", code)
		}
		b.idx = (b.idx + b.size) % len(b.bufs)
	} else if len(b.bufs) >= maxbufs {
		// likely programmer error, something has gone horribly wrong.
		log.Printf("snd/al: get buffers failed, maxbufs reached [len=%v]\n", len(b.bufs))
		return nil
	} else {
		// make more buffers to fill data regardless of what openal says about processed.
		bufs = al.GenBuffers(b.size)
		if code := al.Error(); code != 0 {
			log.Printf("snd/al: generate buffers failed [err=%v]\n", code)
		}
		b.bufs = append(b.bufs, bufs...)
	}
	return bufs
}

type openal struct {
	source al.Source
	buf    *Buffer

	format uint32
	in     snd.Sound
	out    []byte

	quit chan struct{}

	underruns uint64

	tdur time.Duration
	tc   uint64

	start time.Time

	inputs []*snd.Input
}

func OpenDevice(buflen int) error {
	if err := al.OpenDevice(); err != nil {
		return fmt.Errorf("snd/al: open device failed: %s", err)
	}
	if buflen == 0 || buflen&(buflen-1) != 0 {
		return fmt.Errorf("snd/al: buflen(%v) not a power of 2", buflen)
	}
	hwa = &openal{buf: &Buffer{size: buflen}}
	return nil
}

func CloseDevice() error {
	al.DeleteBuffers(hwa.buf.bufs)
	al.DeleteSources(hwa.source)
	al.CloseDevice()
	hwa = nil
	return nil
}

func AddSource(in snd.Sound) error {
	switch in.Channels() {
	case 1:
		hwa.format = al.FormatMono16
	case 2:
		hwa.format = al.FormatStereo16
	default:
		return fmt.Errorf("snd/al: can't handle input with channels(%v)", in.Channels())
	}
	hwa.in = in
	hwa.out = make([]byte, in.BufferLen()*2)

	s := al.GenSources(1)
	if code := al.Error(); code != 0 {
		return fmt.Errorf("snd/al: generate source failed [err=%v]", code)
	}
	hwa.source = s[0]
	hwa.buf.src = s[0]

	log.Println("snd/al: software latency", SoftLatency())

	hwa.inputs = snd.GetInputs(in)

	return nil
}

func Notify() {
	if hwa.in != nil {
		hwa.inputs = snd.GetInputs(hwa.in)
	}
}

func SoftLatency() time.Duration {
	nframes := float64(hwa.in.BufferLen() / hwa.in.Channels())
	return time.Duration(nframes * float64(hwa.buf.size) / hwa.in.SampleRate() * float64(time.Second))
}

func Start() {
	if hwa.quit != nil {
		panic("snd/al: hwa.quit not nil")
	}
	hwa.quit = make(chan struct{})
	go func() {
		hwa.start = time.Now()
		Tick()
		refill := time.Tick(SoftLatency())
		for {
			select {
			case <-hwa.quit:
				return
			case <-refill:
				Tick()
			}
		}
	}()
}

func Stop() { close(hwa.quit) }

var dp = new(snd.Dispatcher)

func Tick() {
	start := time.Now()

	if code := al.DeviceError(); code != 0 {
		log.Printf("snd/al: unknown device error [err=%v]\n", code)
	}
	if code := al.Error(); code != 0 {
		log.Printf("snd/al: unknown error [err=%v]\n", code)
	}

	bufs := hwa.buf.Get()

	for _, buf := range bufs {
		hwa.tc++

		// TODO the general idea here is that GetInputs is rather cheap to call, even with the
		// current first-draft implementation, so it could only return inputs that are actually
		// turned on. This would introduce software latency determined by snd.DefaultBufferLen
		// as turning an input back on would not get picked up until the next iteration.
		// if !realtime {
		// hwa.inputs = snd.GetInputs(hwa.in)
		// }

		dp.Dispatch(hwa.tc, hwa.inputs...)

		for i, x := range hwa.in.Samples() {
			// clip
			if x > 1 {
				x = 1
			} else if x < -1 {
				x = -1
			}
			n := int16(math.MaxInt16 * x)
			hwa.out[2*i] = byte(n)
			hwa.out[2*i+1] = byte(n >> 8)
		}

		buf.BufferData(hwa.format, hwa.out, int32(hwa.in.SampleRate()))
		if code := al.Error(); code != 0 {
			log.Printf("snd/al: buffer data failed [err=%v]\n", code)
		}
	}

	if len(bufs) != 0 {
		hwa.source.QueueBuffers(bufs)
	}
	if code := al.Error(); code != 0 {
		log.Printf("snd/al: queue buffer failed [err=%v]\n", code)
	}

	switch hwa.source.State() {
	case al.Initial:
		al.PlaySources(hwa.source)
	case al.Playing:
	case al.Paused:
	case al.Stopped:
		hwa.underruns++
		al.PlaySources(hwa.source)
	}

	hwa.tdur += time.Now().Sub(start)
}

func BufLen() int {
	return len(hwa.buf.bufs)
}

func Underruns() uint64 {
	return hwa.underruns
}

func TickAverge() time.Duration {
	if hwa.tc == 0 || hwa.buf.size == 0 {
		return 0
	}
	return hwa.tdur / time.Duration(int(hwa.tc)/hwa.buf.size)
}

func DriftApprox() time.Duration {
	if hwa.tc == 0 || hwa.buf.size == 0 {
		return 0
	}
	dt := int64(time.Now().Sub(hwa.start) / time.Duration(int(hwa.tc)/hwa.buf.size))
	lt := int64(SoftLatency())
	return time.Duration(lt - dt)
}
