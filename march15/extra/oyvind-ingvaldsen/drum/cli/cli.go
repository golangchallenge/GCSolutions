package main

import (
	"bufio"
	"fmt"
	"os"

	"oyvind/drum"
	"oyvind/drum/editor"
)

func main() {
	commands := []struct {
		Name        string
		Description string
		Action      func(p *drum.Pattern) int
	}{
		{"print", "print a textual representation of the pattern", func(p *drum.Pattern) int {
			fmt.Printf("%s", p)
			return 0
		}},
		{"play", "play the pattern (using libsdl2)", func(p *drum.Pattern) int {
			pl, err := drum.PlayerNew(p)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to create player: "+err.Error())
				return 1
			}
			fmt.Println("Playing pattern (press <ENTER> to exit)")
			pl.Play()
			rd := bufio.NewReader(os.Stdin)
			rd.ReadString('\n')
			return 0
		}},
		{"edit", "open the pattern in the SPLICE editor", func(p *drum.Pattern) int {
			editor.Editor(p)
			return 0
		}},
	}

	printUsage := func() {
		fmt.Println("Usage: drum COMMAND FILE")
		fmt.Println("")
		fmt.Println("COMMANDS:")
		for _, c := range commands {
			fmt.Printf("\t%s\t\t%s\n", c.Name, c.Description)
		}
		fmt.Println("")
		fmt.Println("The FILE argument is a SPLICE file.")
	}

	if len(os.Args) < 3 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	path := os.Args[2]

	p, err := drum.DecodeFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open SPLICE file: "+err.Error())
		os.Exit(1)
	}

	for _, c := range commands {
		if cmd == c.Name {
			os.Exit(c.Action(p))
		}
	}
}
