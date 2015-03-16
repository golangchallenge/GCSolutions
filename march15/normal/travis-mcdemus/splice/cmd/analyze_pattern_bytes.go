package main

// TODO(aoeu): This program was coded in a sprint mostly while commuting on the L train, clean. it. up.

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"text/tabwriter"
)

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

var spliceSuffixRe = regexp.MustCompile("^.*\\.splice$")

func getSpliceFileInfos(path string) (spliceFileInfos []os.FileInfo) {
	fileInfos, err := ioutil.ReadDir(path)
	check(err)
	for _, fileInfo := range fileInfos {
		if spliceSuffixRe.Match([]byte(fileInfo.Name())) {
			spliceFileInfos = append(spliceFileInfos, fileInfo)
		}
	}
	return
}

func readFiles(path string, fileInfos []os.FileInfo) map[string][]byte {
	allFiles := make(map[string][]byte, 0)
	for _, fileInfo := range fileInfos {
		fileContents, err := ioutil.ReadFile(path + fileInfo.Name())
		check(err)
		allFiles[fileInfo.Name()] = fileContents
	}
	return allFiles
}

func getLongestFileLengthInBytes(files map[string][]byte) (longest int, allEqual bool) {
	allEqual = true
	longest = -1
	for _, contents := range files {
		length := len(contents)
		if length > longest {
			longest = length
			allEqual = false
		}
	}
	return
}

func getMapKeys(aMap map[string][]byte) (keys []string) {
	for key := range aMap {
		keys = append(keys, key)
	}

	return
}

type valueFreqs map[byte]int

func (v valueFreqs) String() string {
	s := ""
	for key, freq := range v {
		s += fmt.Sprintf("%s/%d/%X/%v:%v\t", string(key), key, key, key, freq)
	}
	for i := len(v); i <= maxLen; i++ {
		s += " \t"
	}
	return s
}

type byteDelta struct { // TODO: A less horrible no good very bad name.
	uniform bool
	valueFreqs
}

func (b byteDelta) String() string {
	return fmt.Sprintf("%v\t%v", b.uniform, b.valueFreqs)
}

var maxLen int // TODO: Is there no alternative since this gets used in String()?
var path string

// TODO: This is gross to read and too nested, fix it.
func main() {
	flag.StringVar(&path, "dir", "../encoding/drum/patterns/", "Path to a patterns (.splice) directory")
	flag.Parse()
	fileInfos := getSpliceFileInfos(path)
	allFiles := readFiles(path, fileInfos)
	longest, _ := getLongestFileLengthInBytes(allFiles)
	byteDeltas := make([]byteDelta, longest)
	fileNames := getMapKeys(allFiles)
	for i := 0; i < longest; i++ {
		byteDelta := byteDelta{uniform: false, valueFreqs: make(map[byte]int)}
		checkedAllFiles := true
		for _, fileName := range fileNames {
			if len(allFiles[fileName]) > i {
				byteAtOffset := allFiles[fileName][i]
				byteDelta.valueFreqs[byteAtOffset]++
			} else {
				checkedAllFiles = false
			}

		}
		if len(byteDelta.valueFreqs) == 1 && checkedAllFiles {
			byteDelta.uniform = true
		}
		byteDeltas[i] = byteDelta
	}
	writer := tabwriter.NewWriter(os.Stdout, 0, 16, 0, '\t', tabwriter.Debug|tabwriter.AlignRight)
	for _, byteDelta := range byteDeltas {
		if maxLen < len(byteDelta.valueFreqs) {
			maxLen = len(byteDelta.valueFreqs)
		}
	}
	for i, byteDelta := range byteDeltas {
		out := fmt.Sprintf("%v\t%v\t", i, byteDelta)
		fmt.Fprintln(writer, out)
	}
}
