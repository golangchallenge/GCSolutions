package main

import "io"

type keyboard struct {
	input io.Reader
}

func newKeyboard(r io.Reader) *keyboard {
	return &keyboard{input: r}
}

func (kb *keyboard) Read(p []byte) (n int, err error) {
	return kb.input.Read(p)
}

type display struct {
	output io.Writer
	prompt []byte
}

func newDisplay(w io.Writer, prompt string) *display {
	d := &display{output: w, prompt: []byte(prompt)}
	return d
}

func (d *display) Write(p []byte) (n int, err error) {
	n, err = d.output.Write(p)
	if err != nil {
		return n, err
	}
	return
}
