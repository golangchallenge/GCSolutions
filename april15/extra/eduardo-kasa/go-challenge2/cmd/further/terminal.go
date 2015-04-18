package main

import "io"

// Terminal represents a computer terminal
// It is a "device" that can be used to enter data into, and
// displaying data from, a computer or a computing system.
// (wikipedia.org/wiki/Computer_terminal)
type Terminal struct {
	inputDevice     io.Reader
	displayDevice   io.Writer
	computingDevice io.ReadWriter
}

// NewTerminal Creates a new Terminal.
// input is the input to the Terminal, like io.Stdin
// display is where the Terminal will write to, like io.Stdout
// computer is what the Terminal should represent
func NewTerminal(input io.Reader, display io.Writer, computer io.ReadWriter) *Terminal {
	return &Terminal{inputDevice: input, displayDevice: display, computingDevice: computer}
}

// Run starts the terminal
func (t *Terminal) Run() error {
	end := make(chan int)
	var err error
	go func() {
		_, err = io.Copy(t.displayDevice, t.computingDevice)
		end <- 1
	}()

	go func() {
		_, err = io.Copy(t.computingDevice, t.inputDevice)
		end <- 1
	}()
	<-end
	return err
}
