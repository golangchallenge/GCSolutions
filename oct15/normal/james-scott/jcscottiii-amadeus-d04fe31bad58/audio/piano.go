package audio

import (
	"github.com/golangchallenge/GCSolutions/oct15/normal/james-scott/jcscottiii-amadeus-d04fe31bad58/util"
	"math"
)

const (
	// SampleRate is how many samples should be taken per second
	SampleRate = 5000.0
	// SampleDuration is how long the sample should last.
	SampleDuration = 2.0
)

// FrequencyConstant is a constant found here: https://en.wikipedia.org/wiki/Piano_key_frequencies
var FrequencyConstant = math.Pow(2.0, 1.0/12.0)

// GenSound generates a sound wave for a particular piano key.
// byte can only contain values 0-255. Since the values are only positive, we need
// to shift the sine wave up one from values [-1, 1] to [0,2]. The maximum value the wave
// can have is 255 however, since we shifted the wave up, and the maximum value of our sine
// wave is now 2, we divide 256/2 = 128 as our amplitude for the sine function to stay within
// that 255 bound region. We want a four second sample. In order to do that, we have to
// determine the sample rate (for now 5000), then we time the sample rate times the desired
// time, we get the number of samples (8800)
func GenSound(note util.KeyNote) []byte {
	hz := math.Pow(FrequencyConstant, float64(note)-49.0) * 440.0
	L := int(SampleRate * SampleDuration)
	f := (2.0 * math.Pi * hz) / SampleRate
	data := make([]byte, L, L)
	for sample := 0; sample < L; sample++ {
		data[sample] = byte(128.0 * (math.Sin(f*float64(sample)) + 1.0))
	}
	return data
}
