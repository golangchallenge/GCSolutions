package main

import (
	"flag"
	"io"
	"os"
	"time"
)

type Config struct {
	Generate         bool
	Difficulty       int
	AutoSolve        bool
	ShowDifficulty   bool
	ShowColors       bool
	ShowProgress     bool
	LogLevel         string
	Finder           string
	ProgressTime     time.Duration
	UserInput        io.Reader
	InteractiveInput bool
}

func ConfigFromCli() *Config {
	c := &Config{UserInput: os.Stdin}
	defineGenerate(c)
	defineAutoSolve(c)
	defineProgressTime(c)
	defineShowProgress(c)
	defineShowPrintDifficulty(c)
	defineShowColors(c)
	defineLogLevel(c)
	defineFinder(c)

	flag.Parse()
	postParse(c)
	return c
}

func defineProgressTime(c *Config) {
	defaultValue := time.Second * 1
	defaultUsage := "Defines the amount of time elapsed between each step of showing progress. Only relevant with -s option"
	flag.DurationVar(&c.ProgressTime, "t", defaultValue, defaultUsage)
	//flag.DurationVar(&c.ProgressTime, "-time", defaultValue, defaultUsage)
}
func defineLogLevel(c *Config) {
	defaultValue := "error"
	defaultUsage := "Sets the log level. [off,error,debug,trace]"
	flag.StringVar(&c.LogLevel, "l", defaultValue, defaultUsage)
	//flag.DurationVar(&c.ProgressTime, "-time", defaultValue, defaultUsage)
}
func defineGenerate(c *Config) {
	defaultValue := -1
	defaultUsage := "Create new puzzle with diffulty between 1-5 with 5 being the most difficult. Default is 3"
	flag.IntVar(&c.Difficulty, "g", defaultValue, defaultUsage)
	//flag.IntVar(&c.Difficulty, "-generate", defaultValue, defaultUsage)
}
func defineAutoSolve(c *Config) {
	defaultValue := false
	defaultUsage := "Auto solve without prompting"
	flag.BoolVar(&c.AutoSolve, "a", defaultValue, defaultUsage)
	//flag.BoolVar(&c.AutoSolve, "-auto", defaultValue, defaultUsage)
}
func defineShowColors(c *Config) {
	defaultValue := false
	defaultUsage := "Highlight the solution"
	flag.BoolVar(&c.ShowColors, "h", defaultValue, defaultUsage)
	//flag.BoolVar(&c.ShowColors, "-color", defaultValue, defaultUsage)
}
func defineShowPrintDifficulty(c *Config) {
	defaultValue := true
	defaultUsage := "Show difficulty of puzzle"
	flag.BoolVar(&c.ShowDifficulty, "d", defaultValue, defaultUsage)
	//flag.BoolVar(&c.ShowDifficulty, "-print", defaultValue, defaultUsage)
}
func defineShowProgress(c *Config) {
	defaultValue := false
	defaultUsage := "Show progress of solving solution at each step"
	flag.BoolVar(&c.ShowProgress, "p", defaultValue, defaultUsage)
	//flag.BoolVar(&c.ShowProgress, "-show-progress", defaultValue, defaultUsage)
}
func defineFinder(c *Config) {
	defaultValue := "closest"
	defaultUsage := "Use a different coordinate finder. acceptable values: [rank|closest]"
	flag.StringVar(&c.Finder, "f", defaultValue, defaultUsage)
	//flag.BoolVar(&c.ShowProgress, "-show-progress", defaultValue, defaultUsage)
}

func postParse(c *Config) {
	if c.Difficulty > 0 {
		c.Generate = true
	}
	if c.Difficulty > 5 {
		c.Difficulty = 5
	}
	c.InteractiveInput = inputIsInteractive()
}

func inputIsInteractive() bool {
	fi, _ := os.Stdin.Stat()
	if (fi.Mode() & os.ModeCharDevice) == 0 {
		return false
	}
	return true
}
