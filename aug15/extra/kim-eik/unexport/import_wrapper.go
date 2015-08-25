package main

import (
	"go/build"
	"go/importer"
	"go/types"
	"log"
	"os"
	"os/exec"
)

type importWrapper struct {
	installedPackages map[string]bool
}

func newImportWrapper() *importWrapper {
	return &importWrapper{
		installedPackages: make(map[string]bool),
	}
}

func (i *importWrapper) Import(path string) (*types.Package, error) {
	if _, exists := i.installedPackages[path]; !exists {
		//make sure package is installed, fails if it isn't
		//https://github.com/golang/go/issues/12050
		cmd := exec.Command("go", "install", path)
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		err := cmd.Run()
		if err != nil {
			log.Println(err)
		}

		_, err = build.Import(path, "", build.FindOnly)
		if err != nil {
			log.Println(err)
		}
		i.installedPackages[path] = true
	}
	return importer.Default().Import(path)
}
