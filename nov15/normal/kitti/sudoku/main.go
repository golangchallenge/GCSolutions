package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
)

var grid [9][9]string

var rows map[int]map[string]struct{}
var cols map[int]map[string]struct{}
var blks map[int]map[string]struct{}

var allCells map[int]map[string]struct{}

var out io.Writer
var fileLogger *log.Logger

var difficulty int
var back int

func init() {
	rows = make(map[int]map[string]struct{})
	cols = make(map[int]map[string]struct{})
	blks = make(map[int]map[string]struct{})
	allCells = make(map[int]map[string]struct{})
	startLogger()
	difficulty = 0
	back = 0
}

func main() {
	fileLogger.Println("sudoku solver at work")
	getGrid()
	//printGrid()
	process()
	//printGrid()
	ok := validateGrid()
	if ok {
		ok = gridHasSpace()
		if ok {
			fmt.Println("no solution found")
		}
		fmt.Println("")
		printGrid()
	} else {
		fmt.Println("no solution found")
	}
	level()
}

func startLogger() {
	logMode := "file"
	var err error
	switch logMode {
	case "file":
		out, err = os.Create("sudoku.log")
		if err != nil {
			fmt.Println(err)
		}
	case "screen":
		out = os.Stdout

	default:
		out = ioutil.Discard
	}
	fileLogger = log.New(out, "", log.Lshortfile)
}
