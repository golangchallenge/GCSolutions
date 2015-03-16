package editor

import (
	"github.com/nsf/termbox-go"
)

// A Window simply handles the termbox event loop.
type Window struct {
	listeners map[string][]func(args ...interface{})
	running   bool
	popup     bool
}

// Create a new Window. A popup Window differs in that it doesn't clear the whole screen before drawing.
func WindowNew(popup bool) *Window {
	w := &Window{popup: popup}
	w.listeners = make(map[string][]func(args ...interface{}))
	return w
}

// Register a listener for the event (e.g. "draw", "click", "exit").
func (w *Window) Listen(e string, f func(args ...interface{})) {
	w.listeners[e] = append(w.listeners[e], f)
}

// Trigger an event (i.e. call all listeners) with the provided arguments.
func (w *Window) Trigger(e string, args ...interface{}) {
	for _, f := range w.listeners[e] {
		f(args...)
	}
}

// Stop the event loop.
func (w *Window) Close() {
	w.running = false
}

// Start the event loop.
func (w *Window) Loop() {
	w.running = true
	for w.running {
		if !w.popup {
			termbox.Clear(CDefault, CDefault)
		}
		w.Trigger("draw")
		termbox.Flush()

		switch evt := termbox.PollEvent(); evt.Type {
		case termbox.EventKey:
			w.Trigger("key", evt.Key, evt.Ch)
		case termbox.EventMouse:
			w.Trigger("click", evt.MouseX, evt.MouseY)
		case termbox.EventError:
			panic(evt.Err)
		}
	}
}
