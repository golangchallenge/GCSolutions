# Drum

My solution to [Go Challenge 1](http://golang-challenge.com/go-challenge1/). And my first project in Go, very fun!

## Functionality
- Read SPLICE drum pattern files (`decoder.go`).
- Save SPLICE drum patterns to file (`encoder.go`).
- Play SPLICE drum patterns (`player.go`).
- Edit SPLICE drum patterns (`editor/`).
  - Open files.
  - Save files.
  - Edit steps.
  - Change tempo.
  - Add shuffle.

## Drum launcher
The project also includes a simple «launcher» (create with `make`), called `drumcli`:

    # print a textual representation of a SPLICE pattern
    drumcli print fixtures/pattern_1.splice

    # play a SPLICE file
    drumcli print fixtures/pattern_1.splice

    # open the editor
    drumcli edit fixtures/pattern_1.splice

## Requirements
- The Player uses [SDL2](https://github.com/veandco/go-sdl2) to play WAV files.
- The Editor uses [Termbox](https://github.com/nsf/termbox-go).

## Notes
The Player expects to find WAV files in the relative directory `sounds/` (so must be run from the root of the package).
