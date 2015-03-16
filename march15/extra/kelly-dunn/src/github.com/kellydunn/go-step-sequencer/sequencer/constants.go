package sequencer

// Pulses Per Quarter Note.  A synchronization primative used in music technology.
var Ppqn = 24.0

// The number of seconds in a Minute as a float.
var Minute = 60.0

// The number of microseconds in a Second.
var Microsecond = 1000000000

// The Default BPM of the step sequencer
var DefaultTempo = 120.0

// The standard audio sampling rate of 44.1kHz (44100)
var SampleRate = 44100

// The number of portaudio input channels.  This value is set to 0 since we're reading from disk.
var InputChannels = 0

// The number of portaudio output channels.
var OutputChannels = 2
