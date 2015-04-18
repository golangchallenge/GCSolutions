package main

import (
	"flag"
	"log"
)

var curLevel = 1

func v(level int) verbose {
	if level > curLevel {
		return verbose(false)
	}

	return verbose(true)
}

type verbose bool

func (v verbose) Printf(format string, a ...interface{}) {
	if !v {
		return
	}

	log.Printf(format, a...)
}

func init() {
	flag.IntVar(&curLevel, "v", 1, "logging verbosity")
}
