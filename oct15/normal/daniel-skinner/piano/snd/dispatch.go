package snd

import (
	"sort"
	"sync"
)

// TODO most of this probably doesn't need to be exposed
// but those details can be worked out once additional audio
// drivers are supported.

type Dispatcher struct{ sync.WaitGroup }

// Dispatch blocks until all inputs are prepared.
func (dp *Dispatcher) Dispatch(tc uint64, inps ...*Input) {
	wt := inps[0].wt
	for _, inp := range inps {
		if inp.wt != wt {
			dp.Wait()
			wt = inp.wt
		}
		dp.Add(1)
		go func(sd Sound, tc uint64) {
			sd.Prepare(tc)
			dp.Done()
		}(inp.sd, tc)
	}
	dp.Wait()
}

type Input struct {
	sd Sound
	wt int
}

type ByWT []*Input

func (a ByWT) Len() int           { return len(a) }
func (a ByWT) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByWT) Less(i, j int) bool { return a[i].wt > a[j].wt }

func (a ByWT) Slice() (sl [][]*Input) {
	if len(a) == 0 {
		return nil
	}
	wt := a[0].wt
	i := 0
	for j, p := range a {
		if p.wt != wt {
			sl = append(sl, a[i:j])
			i = j
			wt = p.wt
		}
	}
	return append(sl, a[i:])
}

func GetInputs(sd Sound) []*Input {
	inps := []*Input{{sd, 0}}
	getinputs(sd, 1, &inps)
	sort.Sort(ByWT(inps))
	return inps
}

// TODO janky func
func getinputs(sd Sound, wt int, out *[]*Input) {
	for _, in := range sd.Inputs() {
		if in == nil { // TODO for !realtime || in.IsOff() {
			continue
		}
		at := -1
		for i, p := range *out {
			if p.sd == in {
				if p.wt >= wt {
					return // object has or will be traversed on different path
				}
				at = i
				break
			}
		}
		if at != -1 {
			(*out)[at].sd = in
			(*out)[at].wt = wt
		} else {
			*out = append(*out, &Input{in, wt})
		}
		getinputs(in, wt+1, out)
	}
}
