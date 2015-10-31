package main

import "golang.org/x/mobile/exp/f32"

type Oscilator func() float32

func G(gain float32, f Oscilator) Oscilator {
	return func() float32 {
		return gain * f()
	}
}

func Multiplex(fs ...Oscilator) Oscilator {
	return func() float32 {
		res := float32(0)
		for _, osc := range fs {
			res += osc()
		}
		return res
	}
}

func GenOscilator(freq float32) Oscilator {
	k := 2.0 * Pi * freq
	T := 1.0 / freq
	dt := 1.0 / float32(SampleRate)
	t := float32(0.0)
	return func() float32 {
		res := f32.Sin(k * t)
		t += dt
		if t > T {
			t -= T
		}
		return res
	}
}

func GenEnvelope(press *bool, f Oscilator) Oscilator {
	top := false
	gain := float32(0.0)
	dt := 1.0 / float32(SampleRate)
	attackd := dt / 0.01
	dekeyd := dt / 0.03
	sustainlevel := float32(0.3)
	sustaind := dt / 5.0
	released := dt / 0.5
	return func() float32 {
		if *press {
			if !top {
				gain += attackd
				if gain > 1.0 {
					top = true
					gain = 1.0
				}
			} else {
				if gain > sustainlevel {
					gain -= dekeyd
				} else {
					gain -= sustaind
				}
				if gain < 0.0 {
					gain = 0.0
				}
			}
		} else {
			top = false
			gain -= released
			if gain < 0.0 {
				gain = 0.0
			}
		}
		return gain * f()
	}
}
