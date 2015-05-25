package main

import (
	"crypto/rand"
	"fmt"
	"os"
)

func CheckAndCreateTempDirectory(DirName string) {
	file := DirName
	if _, err := os.Stat(file); err != nil {
		if os.IsNotExist(err) {
			if e := os.Mkdir(file, 0777); e != nil {
				panic(e)
			}
		}
	}
}

func UniueIdStr() string {
	uuid := ""
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		fmt.Println("Error: ", err)
		return uuid
	}
	uuid = fmt.Sprintf("%X-%X-%X-%X-%X", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
	return uuid
}
