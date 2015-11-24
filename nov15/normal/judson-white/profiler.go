package main

import (
	"os"
	"runtime/pprof"
)

func startProfiler() error {
	f, err := os.Create("go-sudoku.pprof")
	if err != nil {
		return err
	}
	if err = pprof.StartCPUProfile(f); err != nil {
		return err
	}
	return nil
}

func stopProfiler() error {
	f, err := os.Create("go-sudoku.mprof")
	if err != nil {
		return err
	}
	defer f.Close()

	if err := pprof.WriteHeapProfile(f); err != nil {
		return err
	}

	pprof.StopCPUProfile()

	return nil
}
