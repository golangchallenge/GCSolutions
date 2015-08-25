package unexporter

import (
	"fmt"
	"go/build"
	"sort"
	"testing"

	"golang.org/x/tools/go/buildutil"
)

type byName []Export

func (e byName) Len() int           { return len(e) }
func (e byName) Swap(i, j int)      { e[i], e[j] = e[j], e[i] }
func (e byName) Less(i, j int) bool { return e[i].Name() < e[j].Name() }

func TestExports(t *testing.T) {
	ctxt := fakeContext(map[string][]string{
		"main": {`
package main

import "fmt"

var   ExportedVar   int
const ExportedConst int = 1
type  ExportedType  int
func  ExportedFunc(parameters, AreNotExported int) {
	var neitherAre int
	var VarsInFuncs int
	const OrConsts int = 42
	fmt.Println(neitherAre, VarsInFuncs, OrConsts)
}

var   unexportedVar int
const unexportedConst int = 4
type  unexportedType int
func  unexportedFunc(Par int) {}
`, `
package main

type ExportedStruct struct {
	unexportedField int
	ExportedField string
}

type unexportedStruct struct {
	ExportedFieldInUnexportedStruct string
}

func (e ExportedStruct) ExportedMethod(a, B int) {}
func (e ExportedStruct) unexportedMethod(a, B int) {}

func (e unexportedStruct) ExportedMethodAgain(a, B int) {}
func (e unexportedStruct) unexportedMethod(a, B int) {}

func unexportedFuncWithTypedef() interface{} {
	type x struct {
		AnotherExportedField string
		absolutelyNotExported int
	}

	return x{
		AnotherExportedField:  "see",
		absolutelyNotExported: 22,
	}
}
`},
	})
	wantedExports := []string{"ExportedVar", "ExportedConst", "ExportedType", "ExportedFunc", "ExportedStruct", "ExportedField", "ExportedFieldInUnexportedStruct", "ExportedMethod", "ExportedMethodAgain", "AnotherExportedField"}
	u, err := ForPackage(ctxt, "main")
	if err != nil {
		t.Fatal(err.Error())
	}
	exports := u.Exports()

	sort.Strings(wantedExports)
	sort.Sort(byName(exports))

	for i, e := range wantedExports {
		if len(exports) <= i {
			t.Fatalf(`"%s" was not marked as exported`, e)
		} else if exports[i].Name() != e {
			t.Fatalf(`"%s" is wrongly marked as exported`, exports[i].Name())
		}
	}

	if len(wantedExports) != len(exports) {
		t.Fatal("Incorrect number of exports found")
	}
}

func TestInterfaceMethod(t *testing.T) {
	ctxt := fakeContext(map[string][]string{
		"utils": {`
package utils

import "io"

type myReader int
func (m myReader) Read(p []byte) (n int, err error) {
	return 42, nil
}

func (m myReader) SomeMethod(p []string) (s string, err error) {
	return "42", nil
}

func GetReader() io.Reader {
	var r myReader
	return r
}
`},
		"main": {`
package main

import "utils"

func main() {
	var buf []byte
	utils.GetReader().Read(buf)
}
`},
	})
	u, err := ForPackage(ctxt, "utils")
	if err != nil {
		t.Fatal(err.Error())
	}
	exports := u.unnecessaryExports()

	if len(exports) != 1 {
		for _, e := range exports {
			t.Error(e.Name(), e)
		}
		t.Fatalf("Unexpected number of exports: %d instead of 1", len(exports))
	}

	if exports[0].Name() != "SomeMethod" {
		t.Fatal(`"SomeMethod" should be the only unnecessary export here`)
	}
}

func TestUsedPointer(t *testing.T) {
	ctxt := fakeContext(map[string][]string{
		"utils": {`
package utils

type MyType int
`},
		"main": {`
package main

import "fmt"
import "utils"

func main() {
	var x *utils.MyType
	fmt.Println(x)
}
`},
	})
	u, err := ForPackage(ctxt, "utils")
	if err != nil {
		t.Fatal(err.Error())
	}
	if len(u.unnecessaryExports()) != 0 {
		t.Fatal("Unexpected number of exports")
	}
}

func TestUsedType(t *testing.T) {
	ctxt := fakeContext(map[string][]string{
		"utils": {`
package utils

type UnusedType int
type DeclaredType interface {}
type TypeCheckedType int
type SwitchTypeCheckedType int
`},
		"main": {`
package main

import "fmt"
import "utils"

func main() {
	var x utils.DeclaredType
	if _, ok := x.(*utils.TypeCheckedType); ok {
		fmt.Println("how can it be this type?")
	}

	switch x.(type) {
	case utils.SwitchTypeCheckedType: fmt.Println("Interesting")
	default: fmt.Println(1)
	}
}
`},
	})
	u, err := ForPackage(ctxt, "utils")
	if err != nil {
		t.Fatal(err.Error())
	}
	if len(u.unnecessaryExports()) != 1 {
		t.Fatal("Unexpected number of unused exports")
	}
}

func TestUsedInterface(t *testing.T) {
	ctxt := fakeContext(map[string][]string{
		"utils": {`
package utils

type MyType int

func (m MyType) Read(buf []byte) (n int, err error) {
	return 0, nil
}
`},
		"main": {`
package main

import "fmt"
import "io"
import "utils"

func main() {
	var m io.Reader
	m = utils.MyType(1)
	m.Read(nil)
	fmt.Println(m)
}
`},
	})
	u, err := ForPackage(ctxt, "utils")
	if err != nil {
		t.Fatal(err.Error())
	}
	if len(u.unnecessaryExports()) != 0 {
		t.Fatal("Unexpected number of unused exports")
	}
}

// ---------------------------------------------------------------------

// Simplifying wrapper around buildutil.FakeContext for packages whose
// filenames are sequentially numbered (%d.go).  pkgs maps a package
// import path to its list of file contents.
func fakeContext(pkgs map[string][]string) *build.Context {
	pkgs["fmt"] = []string{`package fmt; func Println(args ...interface{}) {}`}
	pkgs["io"] = []string{`package io; type Reader interface { Read(p []byte) (n int, err error); }`}
	pkgs2 := make(map[string]map[string]string)
	for path, files := range pkgs {
		filemap := make(map[string]string)
		for i, contents := range files {
			filemap[fmt.Sprintf("%d.go", i)] = contents
		}
		pkgs2[path] = filemap
	}
	return buildutil.FakeContext(pkgs2)
}
