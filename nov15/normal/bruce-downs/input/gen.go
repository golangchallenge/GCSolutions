package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
)

func main() {
	limit := flag.Uint("limit", 10, "limit number of files (unlimited 0)")
	prefix := flag.String("prefix", "puzzle", "prefix for file name")
	dry := flag.Bool("dry", false, "dry run on writing output")
	flag.Parse()

	reader := bufio.NewReader(os.Stdin)
	var idx uint = 1

OUTER_LOOP:
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		if len(line) != 81+1 {
			log.Printf("line length not 81")
			continue
		}

		var xform bytes.Buffer
		for k, v := range line {
			if k > 0 {
				if k%9 == 0 {
					xform.WriteRune('\n')
				} else {
					xform.WriteRune(' ')
				}
			}

			switch v {
			case '_', '.', '0':
				xform.WriteRune('_')
			case '1', '2', '3', '4', '5', '6', '7', '8', '9':
				xform.WriteRune(v)
			case '\n':
				// no-op
			default:
				log.Printf("invalid field value [%s]", string(v))
				continue OUTER_LOOP
			}
		}

		filename := fmt.Sprintf("%s%d.txt", *prefix, idx)
		if *dry {
			log.Printf("did not write puzzle to %s", filename)
		} else {
			file, err := os.Create(filename)
			if err != nil {
				log.Printf("error opening %s", filename)
			}

			xform.WriteTo(file)
			file.Close()

			log.Printf("wrote puzzle to %s", filename)
		}

		// increment only upon writing file
		idx++

		if *limit > 0 && idx > *limit {
			break
		}
	}
}
