package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
)

// clientRead sends input from Stdin to c.
func clientRead(c io.ReadWriteCloser) {
	s := bufio.NewScanner(os.Stdin)
	prompt := "> "
	fmt.Print(prompt)
	for s.Scan() {
		if _, err := c.Write([]byte(s.Text())); err != nil {
			log.Fatal(err)
		}
	}
}

// clientWrite gets data from c and prints it to Stdin.
func clientWrite(c io.ReadWriteCloser) {
	buf := make([]byte, 32000)
	for {
		n, err := c.Read(buf)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("[server]> %s\n>", buf[:n])
	}
}
