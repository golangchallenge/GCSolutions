package main

import (
	"log"
	"math"
	"time"

	"dasa.cc/piano/snd"
	"dasa.cc/piano/snd/al"

	"golang.org/x/mobile/app"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"
	"golang.org/x/mobile/event/touch"
	"golang.org/x/mobile/gl"
)

const buffers = 1

// TODO this file got out of hand ..

var (
	sz size.Event

	fps       int
	lastpaint = time.Now()

	ms = time.Millisecond

	sawtooth = snd.Sawtooth()
	sawsine  = snd.SawtoothSynthesis(8)
	square   = snd.Square()
	sqsine   = snd.SquareSynthesis(49)
	sine     = snd.Sine()
	triangle = snd.Triangle()

	notes = snd.EqualTempermant(12, 440, 48)
	keys  [12]Key

	bpm snd.BPM = 80

	reverb    snd.Sound
	metronome snd.Sound
	loop      *snd.Loop

	// nframes
	loopdur int = snd.Dtof(bpm.Dur(), snd.DefaultSampleRate) * 8

	btnreverb    *snd.Button
	btnlowpass   *snd.Button
	btnmetronome *snd.Button
	btnloop      *snd.LoopButton
	btnsndbank   *snd.Button

	piano   *Piano
	pianowf *snd.Waveform

	keymix  *snd.Mixer
	keygain *snd.Gain

	lowpass *snd.LowPass

	master     *snd.Mixer
	mastergain *snd.Gain
	mixwf      *snd.Waveform

	touchseq = make(map[touch.Sequence]int)

	sndbank    = []KeyFunc{NewPianoKey, NewWobbleKey, NewBeatsKey}
	sndbankpos = 0
)

type Key interface {
	snd.Sound
	Press()
	Release()
	Freeze()
}

type KeyFunc func(int) Key

type BeatsKey struct {
	*snd.Instrument
	adsr *snd.ADSR
}

func NewBeatsKey(idx int) Key {
	osc := snd.NewOscil(sawsine, notes[idx], snd.NewOscil(triangle, 4, nil))
	dmp := snd.NewDamp(bpm.Dur(), osc)
	d := snd.BPM(float64(bpm) * 1.25).Dur()
	dmp1 := snd.NewDamp(d, osc)
	drv := snd.NewDrive(d, osc)
	mix := snd.NewMixer(dmp, dmp1, drv)

	frz := snd.NewFreeze(bpm.Dur()*4, mix)

	adsr := snd.NewADSR(250*ms, 500*ms, 300*ms, 400*ms, 0.85, 1.0, frz)
	key := &BeatsKey{snd.NewInstrument(adsr), adsr}
	key.Off()
	return key
}

func (key *BeatsKey) Press() {
	key.adsr.Restart()
	key.adsr.Sustain()
	key.On()
}
func (key *BeatsKey) Release() {
	key.adsr.Release()
	key.OffIn(400 * ms)
}
func (key *BeatsKey) Freeze() {}

type WobbleKey struct {
	*snd.Instrument
	adsr *snd.ADSR
}

func NewWobbleKey(idx int) Key {
	osc := snd.NewOscil(sine, notes[idx], snd.NewOscil(triangle, 2, nil))
	adsr := snd.NewADSR(50*ms, 100*ms, 200*ms, 400*ms, 0.6, 0.9, osc)
	key := &WobbleKey{snd.NewInstrument(adsr), adsr}
	key.Off()
	return key
}

func (key *WobbleKey) Press() {
	key.adsr.Restart()
	key.adsr.Sustain()
	key.On()
}
func (key *WobbleKey) Release() {
	key.adsr.Release()
	key.OffIn(400 * ms)
}
func (key *WobbleKey) Freeze() {}

type PianoKey struct {
	*snd.Instrument

	freq float64

	osc, mod, phs    *snd.Oscil
	oscl, modl, phsl *snd.Oscil
	oscr, modr, phsr *snd.Oscil

	adsr0, adsr1 *snd.ADSR

	gain *snd.Gain

	dur    time.Duration
	reldur time.Duration

	frz *snd.Freeze
}

func NewPianoKey(idx int) Key {
	const phasefac float64 = 0.5063999999999971

	k := &PianoKey{}

	k.freq = notes[idx]
	k.mod = snd.NewOscil(sqsine, k.freq/2, nil)
	k.osc = snd.NewOscil(sawtooth, k.freq, k.mod)
	k.phs = snd.NewOscil(square, k.freq*phasefac, nil)
	k.osc.SetPhase(k.phs)

	freql := k.freq * math.Pow(2, -10.0/1200)
	k.modl = snd.NewOscil(sqsine, freql/2, nil)
	k.oscl = snd.NewOscil(sawtooth, freql, k.modl)
	k.phsl = snd.NewOscil(square, freql*phasefac, nil)
	k.oscl.SetPhase(k.phsl)

	freqr := k.freq * math.Pow(2, 10.0/1200)
	k.modr = snd.NewOscil(sqsine, freqr/2, nil)
	k.oscr = snd.NewOscil(sawtooth, freqr, k.modr)
	k.phsr = snd.NewOscil(square, freqr*phasefac, nil)
	k.oscr.SetPhase(k.phsr)

	oscmix := snd.NewMixer(k.osc, k.oscl, k.oscr)

	k.reldur = 1050 * ms
	k.dur = 280*ms + k.reldur
	k.adsr0 = snd.NewADSR(30*ms, 50*ms, 200*ms, k.reldur, 0.4, 1, oscmix)
	k.adsr1 = snd.NewADSR(1*ms, 278*ms, 1*ms, k.reldur, 0.4, 1, oscmix)
	adsrmix := snd.NewMixer(k.adsr0, k.adsr1)

	k.gain = snd.NewGain(snd.Decibel(-10).Amp(), adsrmix)

	k.Instrument = snd.NewInstrument(k.gain)
	k.Off()

	return k
}

func (key *PianoKey) Freeze() {
	key.On()
	key.frz = snd.NewFreeze(key.dur, key.gain)
	key.Instrument = snd.NewInstrument(key.frz)
	key.Off()
}

func (key *PianoKey) Press() {
	key.On()
	key.OffIn(key.dur)
	if key.frz == nil {
		key.adsr0.Restart()
		key.adsr1.Restart()
	} else {
		key.frz.Restart()
	}
}

func (key *PianoKey) Release() {
	if key.frz == nil {
		if key.adsr0.Release() && key.adsr1.Release() {
			key.OffIn(key.reldur)
		}
	}
}

func makekeys() {
	keymix.Empty()
	for i := range keys {
		keys[i] = sndbank[sndbankpos](51 + i) // notes[51] is Major C
		keys[i].Freeze()
		keymix.Append(keys[i])
	}
	al.Notify()
}

func onStart(ctx gl.Context) {
	ctx.Enable(gl.BLEND)
	ctx.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	if err := al.OpenDevice(buffers); err != nil {
		log.Fatal(err)
	}

	var err error

	keymix = snd.NewMixer()
	go makekeys()
	lowpass = snd.NewLowPass(773, keymix)
	keygain = snd.NewGain(snd.Decibel(-3).Amp(), lowpass)

	dly := snd.NewDelay(29*time.Millisecond, keygain)
	tap0 := snd.NewTap(19*time.Millisecond, dly)
	tap1 := snd.NewTap(13*time.Millisecond, dly)
	tap2 := snd.NewTap(7*time.Millisecond, dly)
	cmb := snd.NewComb(snd.Decibel(-3).Amp(), 11*time.Millisecond, snd.NewMixer(dly, tap0, tap1, tap2))
	reverb = snd.NewLowPass(2000, cmb)
	dlymix := snd.NewMixer(reverb, keygain)

	loop = snd.NewLoopFrames(loopdur, dlymix)
	loop.SetBPM(bpm)
	loopmix := snd.NewMixer(dlymix, loop)

	master = snd.NewMixer(loopmix)
	mastergain = snd.NewGain(snd.Decibel(-6).Amp(), master)
	mixwf, err = snd.NewWaveform(ctx, 2, mastergain)
	if err != nil {
		log.Fatal(err)
	}
	pan := snd.NewPan(0, mixwf)

	mtrosc := snd.NewOscil(sine, 440, nil)
	mtrdmp := snd.NewDamp(bpm.Dur(), mtrosc)
	metronome = snd.NewMixer(mtrdmp)
	metronome.Off()
	master.Append(metronome)

	piano = NewPiano()
	pianowf, err = snd.NewWaveform(ctx, 1, piano)
	if err != nil {
		log.Fatal(err)
	}

	al.AddSource(pan)

	yoff := float32(-0) //.12)

	btnreverb = snd.NewButton(ctx, -0.98, 0.96+yoff, 0.2, -0.2)
	btnreverb.SetActiveColor(0, 1, 0, 0.5)
	btnreverb.SetActive(true)

	btnlowpass = snd.NewButton(ctx, -0.76, 0.96+yoff, 0.2, -0.2)
	btnlowpass.SetActiveColor(0, 1, 1, 0.5)
	btnlowpass.SetActive(true)

	btnmetronome = snd.NewButton(ctx, -0.54, 0.96+yoff, 0.2, -0.2)
	btnmetronome.SetActiveColor(0, 0, 1, 0.5)

	btnloop = snd.NewLoopButton(snd.NewButton(ctx, -0.32, 0.96+yoff, 0.2, -0.2), loop)
	btnloop.SetActiveColor(1, 0, 0, 0.5)

	btnsndbank = snd.NewButton(ctx, 0.78, 0.96+yoff, 0.2, -0.2)
	btnsndbank.SetActive(true)
	btnsndbank.SetActiveColor(sndbankcolor())
}

func sndbankcolor() (r, g, b, a float32) {
	i := float32(((sndbankpos + 1) * 29) % 256)
	a, b, c := i*8/255, i*2/255, i*4/255
	if int(i)%2 == 0 {
		a, b = b, a
	} else if int(i)%3 == 0 {
		b, c = c, b
	}
	return a, b, c, 0.5
}

func onPaint(ctx gl.Context) {
	ctx.ClearColor(0, 0, 0, 1)
	ctx.Clear(gl.COLOR_BUFFER_BIT)

	pianowf.Prepare(1)
	switch sz.Orientation {
	case size.OrientationPortrait:
		pianowf.Paint(ctx, -1, -1, 2, 0.5)
		mixwf.Paint(ctx, -1, 0.25, 2, 0.5)
	default:
		pianowf.Paint(ctx, -1, -1, 2, 1)
		mixwf.Paint(ctx, -1, 0.25, 2, 0.5)
	}

	btnreverb.Paint(ctx)
	btnlowpass.Paint(ctx)
	btnmetronome.Paint(ctx)
	btnloop.Paint(ctx)
	btnsndbank.Paint(ctx)

	now := time.Now()
	fps = int(time.Second / now.Sub(lastpaint))
	lastpaint = now
}

func onTouch(ev touch.Event) {

	idx := piano.KeyAt(ev, sz)
	if idx == -1 {
		// top half
		switch ev.Type {
		case touch.TypeBegin:
			if btnreverb.HitTest(ev, sz) {
				btnreverb.SetActive(!btnreverb.IsActive())
				if btnreverb.IsActive() {
					reverb.On()
					keygain.SetAmp(snd.Decibel(-3).Amp())
				} else {
					reverb.Off()
					keygain.SetAmp(snd.Decibel(3).Amp())
				}
			} else if btnlowpass.HitTest(ev, sz) {
				btnlowpass.SetActive(!btnlowpass.IsActive())
				lowpass.SetPassthrough(!btnlowpass.IsActive())
			} else if btnmetronome.HitTest(ev, sz) {
				btnmetronome.SetActive(!btnmetronome.IsActive())
				if btnmetronome.IsActive() {
					metronome.On()
				} else {
					metronome.Off()
				}
			} else if btnloop.HitTest(ev, sz) {
				if !btnloop.IsActive() {
					btnloop.SetActiveColor(1, 1, 0, 0.5)
					btnloop.SetActive(true)
					loop.Record()
				} else {
					loop.Stop()
				}
			} else if btnsndbank.HitTest(ev, sz) {
				sndbankpos = (sndbankpos + 1) % len(sndbank)
				btnsndbank.SetActiveColor(sndbankcolor())
				go makekeys()
			}
		case touch.TypeMove:
		}
		// TODO drag finger off piano and it still plays, shouldn't return here
		// to allow TypeMove to figure out what to turn off
		return
	}

	if keys[idx] == nil {
		return
	}

	switch ev.Type {
	case touch.TypeBegin:
		keys[idx].Press()
		touchseq[ev.Sequence] = idx
	case touch.TypeMove:
		// TODO drag finger off piano and it still plays, should stop
		if lastidx, ok := touchseq[ev.Sequence]; ok {
			if idx != lastidx {
				keys[lastidx].Release()
				keys[idx].Press()
				touchseq[ev.Sequence] = idx
			}
		}
	case touch.TypeEnd:
		keys[idx].Release()
		delete(touchseq, ev.Sequence)
	default:
	}
}

func main() {
	app.Main(func(a app.App) {
		var (
			glctx   gl.Context
			visible bool
			logdbg  = time.NewTicker(time.Second)
		)

		go func() {
			for range logdbg.C {
				log.Printf("fps=%-4v underruns=%-4v buflen=%-4v tickavg=%-12s drift=%s\n",
					fps, al.Underruns(), al.BufLen(), al.TickAverge(), al.DriftApprox())
			}
		}()

		for ev := range a.Events() {
			switch ev := a.Filter(ev).(type) {
			case lifecycle.Event:
				switch ev.Crosses(lifecycle.StageVisible) {
				case lifecycle.CrossOn:
					visible = true
					glctx = ev.DrawContext.(gl.Context)
					onStart(glctx)
					al.Start()
				case lifecycle.CrossOff:
					visible = false
					logdbg.Stop()
					al.Stop()
					al.CloseDevice()
				}
			case touch.Event:
				onTouch(ev)
			case size.Event:
				sz = ev
			case paint.Event:
				onPaint(glctx)
				a.Publish()
				if visible {
					a.Send(paint.Event{})
				}
			}
		}
	})
}
