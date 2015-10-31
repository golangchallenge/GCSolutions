// +build darwin,!arm,!arm64 linux android

package main

import (
	"golang.org/x/mobile/app"
	"golang.org/x/mobile/event/paint"
)

func repaint(a app.App) {
	a.Send(paint.Event{}) // keep animating
}
