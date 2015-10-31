# Amadeus

Amadeus is an entry to the [Go Challenge 7](http://golang-challenge.com/go-challenge7/).

Amadeus is an on-screen piano for mobile devices. Currently, it is only tested for Android.

## Setup

```
$ go get golang.org/x/mobile/cmd/gomobile
$ gomobile init # it might take a few minutes
```

## Run
1. Start an Android Emulator or Plug in an Android phone
1. Run `gomobile install`

# Technical Details
- Support for one finger taps and sliding
- Support for multi finger simultaneous taps
- Sounds are generated with OpenAL
- Intuitive layer system (like Photoshop) for event handling

