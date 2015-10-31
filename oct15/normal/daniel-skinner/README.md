To build from source, the source expects to exist at $GOPATH/src/dasa.cc/piano

This is due to the piano/snd package within.

to support intended use on android (landscape, fullscreen), build with go mobile from at least the following commit: 38a56c4998acb3d9e92277296a18e10302225654

For convenience, there is a piano.apk within the zip.

Please note, this generates sounds and some of them can take a while to load. There's software latency of 5.8ms and buffer settings are fairly aggressive on mobile.

There will still be hardware latency making these kinds of things difficult on Android in general. See [https://source.android.com/devices/audio/latency_measurements.html](https://source.android.com/devices/audio/latency_measurements.html) and if your device is listed there to get a general idea of what to expect. The nexus 9 on 6.0 seems the only acceptable latency for real time but the app is still fun on my Moto G 2nd gen.

There are three sounds in the app (though you could go in and add some more to main.go). Two of these three sounds freeze portions of the audio to fit within the 5.8ms software latency i was aiming for on my test device. The difference being heard is not noticeable with the exception of piano sound which has no immediate release.

A list of what the buttons do from left to right:
[reverb] [lowpass] [metronome] [record]    [switch sound]

reverb, lowpass, metronome are either on or off.
metronome is set to 80bpm
record will start in-sync with metronome and sample what you play for 8 measures. It does not crossfade ..
switch-sound will cycle between 3 sounds.

The first sound is like a piano and is intended for use with reverb and lowpass on.

The second sound is at a low freq. If you can't hear it, you can connect better speakers to your phone, or try turning lowpass and reverb off.

The third sound is intended for use by holding a key (or keys) down for extended periods.

Anyway, have fun.