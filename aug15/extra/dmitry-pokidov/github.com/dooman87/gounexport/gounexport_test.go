package gounexport

import (
	"go/ast"
	"go/token"
	"go/types"
	"log"
	"regexp"
	"testing"

	"github.com/dooman87/gounexport/util"
)

const (
	pkg = "github.com/dooman87/gounexport/testdata"
)

func parsePackage(pkgStr string, t *testing.T) (*types.Package, *token.FileSet, *types.Info) {
	info := types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
		Defs:  make(map[*ast.Ident]types.Object),
		Uses:  make(map[*ast.Ident]types.Object),
	}

	packag, fset, err := ParsePackage(pkgStr, &info)
	if err != nil {
		t.Errorf("error while parsing package %v", err)
	}
	return packag, fset, &info
}

func TestParsePackageFunc(t *testing.T) {
	log.Print("-----------------------TestParsePackageFunc-----------------------")
	_, fset, _ := parsePackage(pkg+"/testfunc", t)
	fileCounter := 0
	iterator := func(f *token.File) bool {
		fileCounter++
		return true
	}
	fset.Iterate(iterator)

	if fileCounter != 2 {
		t.Errorf("expected 2 files in result file set but found %d", fileCounter)
	}
}

func TestGetDefinitionsFunc(t *testing.T) {
	log.Print("-----------------------TestGetDefinitionsFunc-----------------------")
	unimportedpkg := pkg + "/testfunc"

	_, fset, info := parsePackage(unimportedpkg, t)
	defs := GetDefinitions(info, fset)

	//Used, Unused, main
	if len(defs) != 3 {
		t.Errorf("expected 3 exported definitions, but found %d", len(defs))
	}
}

func TestGetDefenitionsUnimported(t *testing.T) {
	log.Print("-----------------------TestGetDefenitionsUnimported-----------------------")
	unimportedpkg := pkg + "/unimported"

	_, fset, info := parsePackage(unimportedpkg, t)
	defs := GetDefinitions(info, fset)

	//NeverImported
	if len(defs) != 1 {
		t.Errorf("expected 1 exported definitions, but found %d", len(defs))
	}
}

func TestGetDefenitionsToHideFunc(t *testing.T) {
	log.Print("-----------------------TestGetDefenitionsToHideFunc-----------------------")
	unimportedpkg := pkg + "/testfunc"
	unusedDefs := getDefinitionsToHide(unimportedpkg, 1, t)

	assertDef("github.com/dooman87/gounexport/testdata/testfunc.Unused", unusedDefs, t)
}

func TestGetDefenitionsToHideStruct(t *testing.T) {
	log.Print("-----------------------TestGetDefenitionsToHideStruct-----------------------")
	unimportedpkg := pkg + "/teststruct"
	unusedDefs := getDefinitionsToHide(unimportedpkg, 4, t)

	assertDef("github.com/dooman87/gounexport/testdata/teststruct.UnusedStruct", unusedDefs, t)
	assertDef("github.com/dooman87/gounexport/testdata/teststruct.UsedStruct.UnusedField", unusedDefs, t)
	assertDef("github.com/dooman87/gounexport/testdata/teststruct.UsedStruct.UnusedMethod", unusedDefs, t)
	assertDef("github.com/dooman87/gounexport/testdata/teststruct.UsedStruct.UsedInPackageMethod", unusedDefs, t)
}

func TestGetDefenitionsToHideVar(t *testing.T) {
	log.Print("-----------------------TestGetDefenitionsToHideVar-----------------------")
	unimportedpkg := pkg + "/testvar"
	unusedDefs := getDefinitionsToHide(unimportedpkg, 2, t)

	assertDef("github.com/dooman87/gounexport/testdata/testvar.UnusedVar", unusedDefs, t)
	assertDef("github.com/dooman87/gounexport/testdata/testvar.UnusedConst", unusedDefs, t)
}

func TestGetDefenitionsToHideInterface(t *testing.T) {
	log.Print("-----------------------TestGetDefenitionsToHideInterface-----------------------")
	unimportedpkg := pkg + "/testinterface"
	unusedDefs := getDefinitionsToHide(unimportedpkg, 1, t)

	assertDef("github.com/dooman87/gounexport/testdata/testinterface.UnusedInterface", unusedDefs, t)
}

func TestGetDefenitionsToHideExclusions(t *testing.T) {
	log.Print("-----------------------TestGetDefenitionsToHideInterface-----------------------")
	unimportedpkg := pkg + "/testinterface"
	regex, _ := regexp.Compile("Unused*")
	excludes := []*regexp.Regexp{regex}
	getDefinitionsToHideWithExclusions(unimportedpkg, 0, excludes, t)
}

func getDefinitionsToHide(pkg string, expectedLen int, t *testing.T) []*Definition {
	return getDefinitionsToHideWithExclusions(pkg, expectedLen, nil, t)
}

func getDefinitionsToHideWithExclusions(pkg string, expectedLen int, excludes []*regexp.Regexp, t *testing.T) []*Definition {
	_, fset, info := parsePackage(pkg, t)
	defs := GetDefinitions(info, fset)
	unusedDefs := GetDefenitionsToHide(pkg, defs, excludes)

	if expectedLen > 0 && len(unusedDefs) != expectedLen {
		t.Errorf("expected %d unused exported definitions, but found %d", expectedLen, len(unusedDefs))
	}
	return unusedDefs
}

func TestGetDefenitionsToHideThis(t *testing.T) {
	log.Print("-----------------------TestGetDefenitionsToHideThis-----------------------")
	pkg := "github.com/dooman87/gounexport"

	regex, _ := regexp.Compile("Test*")
	excludes := []*regexp.Regexp{regex}

	_, fset, info := parsePackage(pkg, t)
	defs := GetDefinitions(info, fset)
	unusedDefs := GetDefenitionsToHide(pkg, defs, excludes)

	log.Print("<<<<<<<<<<<<<<<<<<<<<<<<<<<")
	for _, d := range unusedDefs {
		util.Info("DEFINITION %s", d.Name)
		util.Info("\t%s:%d:%d", d.File, d.Line, d.Col)
	}
	log.Print("<<<<<<<<<<<<<<<<<<<<<<<<<<<")

	if len(unusedDefs) != 23 {
		t.Errorf("expected %d unused exported definitions, but found %d", 23, len(unusedDefs))
	}
}

//ExampleGetUnusedDefitions shows how to use gounexport package
//to find all definition that not used in a package. As the result,
//all unused definitions will be printed in console.
func Example() {
	//package to check
	pkg := "github.com/dooman87/gounexport"

	//Regular expression to exclude
	//tests methods from the result.
	regex, _ := regexp.Compile("Test*")
	excludes := []*regexp.Regexp{regex}

	//Internal info structure that required for
	//ParsePackage call
	info := types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
		Defs:  make(map[*ast.Ident]types.Object),
		Uses:  make(map[*ast.Ident]types.Object),
	}

	//Parsing package to fill info struct and
	//get file set.
	_, fset, err := ParsePackage(pkg, &info)
	if err != nil {
		util.Err("error while parsing package %v", err)
	}

	//Analyze info and extract all definitions with usages.
	defs := GetDefinitions(&info, fset)
	//Find all definitions that not used
	unusedDefs := GetDefenitionsToHide(pkg, defs, excludes)
	//Print all unused definition to stdout.
	for _, d := range unusedDefs {
		util.Info("DEFINITION %s", d.Name)
	}
}

func assertDef(name string, defs []*Definition, t *testing.T) {
	found := false
	for _, d := range defs {
		if name == d.Name {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("expected [%s] in defenitions", name)
	}
}

func TestUnexport(t *testing.T) {
	_, fset, info := parsePackage(pkg+"/testrename", t)
	defs := GetDefinitions(info, fset)
	unusedDefs := GetDefenitionsToHide(pkg, defs, nil)

	renamesCount := make(map[string]int)
	renameFunc := func(file string, offset int, source string, target string) error {
		log.Printf("renaming [%s] at %d from [%s] to [%s]", file, offset, source, target)
		renamesCount[source] = renamesCount[source] + 1
		return nil
	}

	for _, d := range unusedDefs {
		err := Unexport(d, defs, renameFunc)
		if d.SimpleName == "UnusedStructConflict" && err == nil {
			t.Error("expected conflict error for UnusedStructConflict")
		}
		if d.SimpleName == "UnusedVarConflict" && err == nil {
			t.Error("expected conflict error for UnusedVarConflict")
		}
	}

	assertRename(renamesCount, "UnusedField", 1, t)
	assertRename(renamesCount, "UnusedMethod", 1, t)
	assertRename(renamesCount, "UsedInPackageMethod", 2, t)
}

func assertRename(renamesCount map[string]int, name string, expected int, t *testing.T) {
	if renamesCount[name] != expected {
		t.Errorf("expected [%d] renames of [%s], but was [%d]", expected, name, renamesCount[name])
	}
}
