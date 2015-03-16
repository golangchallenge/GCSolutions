package main

import (
	"bufio"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"splice/encoding/drum"
)

var patternPath string
var defaultPath = "../encoding/drum/patterns/decoded/pattern_3.txt"

func main() {
	flag.StringVar(&patternPath, "file",
		defaultPath,
		"Path to a text file representing a pattern.")
	flag.Parse()
	data, err := ioutil.ReadFile(patternPath)
	if err != nil {
		log.Fatal(err)
	}
	p, err := drum.NewPatternFromBackup(string(data))
	if err != nil {
		log.Fatal(err)
	}
	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()
	e := drum.NewEncoder(w)
	err = e.Encode(*p)
	if err != nil {
		log.Fatal(err)
	}
}
