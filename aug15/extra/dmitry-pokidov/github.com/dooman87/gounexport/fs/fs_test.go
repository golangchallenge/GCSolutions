package fs

import (
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strings"
	"testing"
)

const (
	pkg = "github.com/dooman87/gounexport/testdata"
)

func TestGetFilesDeep(t *testing.T) {
	log.Print("-----------------------TestGetFilesDeep-----------------------")
	basepath := os.Getenv("GOPATH") + "/src/" + pkg
	expected := []string{
		basepath + "/testfunc/main/main.go",
		basepath + "/testfunc/func.go",
	}

	files, err := GetFiles(pkg+"/testfunc", true)
	if err != nil {
		t.Errorf("%v", err)
	}
	if len(files) != len(expected) {
		t.Errorf("expected [%d] files from package [%s], but found [%d]", len(expected), pkg, len(files))
	}

	sort.Strings(files)
	sort.Strings(expected)
	for i, f := range expected {
		if files[i] != expected[i] {
			t.Errorf("file [%s] not found, error [%s]", f, files[i])
		}
	}
}

func TestGetUnusedSources(t *testing.T) {
	log.Print("-----------------------TestGetUnusedSources-----------------------")

	fset := token.NewFileSet()
	files, err := GetFiles(pkg, true)
	if err != nil {
		t.Errorf("%v", err)
	}
	for _, f := range files {
		if strings.Contains(f, "unimported.go") {
			continue
		}
		fileInfo, _ := os.Stat(f)
		fset.AddFile(f, fset.Base(), int(fileInfo.Size()))
	}

	unusedSource, err := GetUnusedSources(pkg, fset)

	if err != nil {
		t.Errorf("%v", err)
	}
	//unimported.go
	if len(unusedSource) != 1 {
		t.Errorf("expected 1 unused sources, but found %d", len(unusedSource))
	}
}

func TestGetPackagePathDir(t *testing.T) {
	log.Print("-----------------------TestGetPackagePathDir-----------------------")
	expected := "github.com/dooman87/gounexport/testdata/testfunc/main"
	packagePath := GetPackagePath(os.Getenv("GOPATH") + "/src/github.com/dooman87/gounexport/testdata/testfunc/main")
	if packagePath != expected {
		t.Errorf("expected [%s] package path but get [%s]", expected, packagePath)
	}
}

func TestGetPackagePathFile(t *testing.T) {
	log.Print("-----------------------TestGetPackagePathFile-----------------------")
	expected := "github.com/dooman87/gounexport/testdata/testfunc/main"
	packagePath := GetPackagePath(os.Getenv("GOPATH") + "/src/github.com/dooman87/gounexport/testdata/testfunc/main/file.go")
	if packagePath != expected {
		t.Errorf("expected [%s] package path but get [%s]", expected, packagePath)
	}
}

func TestReplaceStringInFile(t *testing.T) {
	log.Print("-----------------------TestReplaceStringInFile-----------------------")
	original :=
		`We want to {
   Replace the first
	 Replace word
}`
	expected :=
		`We want to {
   replace the first
	 Replace word
}`
	//packagePath := GetPackagePath(os.Getenv("GOPATH") + "/src/github.com/dooman87/gounexport/testdata/testfunc/main/file.go")
	file := os.Getenv("GOPATH") + "/src/github.com/dooman87/gounexport/testdata/testreplace.txt"
	ioutil.WriteFile(file, []byte(original), 0)
	ReplaceStringInFile(file, 16, "Replace", "replace")
	content, _ := ioutil.ReadFile(file)
	strContent := string(content)
	if strContent != expected {
		t.Errorf("expected \n[%s], but found\n[%s]", expected, strContent)
	}
}
