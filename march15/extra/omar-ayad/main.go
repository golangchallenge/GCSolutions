package main

import (
	"./drum"
	"flag"
	"fmt"
	"os"
)

var decode bool
var help bool
var cowbell int
var play bool

func InitFlags() {
	flag.BoolVar(&decode, "decode", false, "decode a splice legacy file and prints the output")
	flag.IntVar(&cowbell, "cowbell", 0, "adds more cowbell beats (if there is a cowbell instrument). Note: use with decode")
	flag.BoolVar(&help, "help", false, "displays this small help")
	flag.Parse()
}

func main() {
	InitFlags()
	filename := os.Args[len(os.Args)-1]
	if _, err := os.Stat(filename); err != nil && help {
		fmt.Println("Usage: go run main.go [-decode, -help] filename.splice")
		flag.PrintDefaults()
		os.Exit(0)
	} else if _, err := os.Stat(filename); err != nil {
		fmt.Printf("file doesn't exist")
		os.Exit(0)
	} else {
		data, _ := drum.DecodeFile(filename)
		if cowbell > 0 {
			data.AddCowbell(cowbell)
		}
		if decode {
			decoded, _ := data.Format()
			fmt.Println(decoded)
		}
	}

}
