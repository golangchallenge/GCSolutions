//Package cmd provides command line interface to gounexport tool.
//
//Command requires package name. For, example:
//  gounexport github.com/dooman87/gounexport
//
//There are next supported flags:
//
//  -exclude string
//    	File with exlude patterns for objects that shouldn't be unexported.Each pattern should be started at new line. Default pattern is Test* to exclude tests methods.
//  -out string
//    	Output file. If not set then stdout will be used
//  -rename
//    	If set, then all defenitions that will be determined as unused will be renamed in files
//  -verbose
//    	Turning on verbose mode
//
//Exclude flag is pointing to file with regular expressions to ignore
//public unexported symbols. Each expression should be starterd
//with a new line. It's a standard go/regexp package. For example,
//below we are excluding all Test methods and everything from package
//public/api/pack/:
//
//  Test*
//  public/api/packag/*
//
//Use -rename flag carefully and check output before.
//
//BUG(d): The tool is not analyzing test files if package in the test file is not
//the same as a base package. For instance, pack/pack_test.go is in package pack_test
//instead of pack
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"sort"
	"strings"

	"go/ast"
	"go/types"

	"github.com/dooman87/gounexport"
	"github.com/dooman87/gounexport/fs"
	"github.com/dooman87/gounexport/util"
)

type sortableDefinition struct {
	defs []*gounexport.Definition
}

func (sortDefs *sortableDefinition) Len() int {
	return len(sortDefs.defs)
}

func (sortDefs *sortableDefinition) Less(i int, j int) bool {
	iDef := sortDefs.defs[i]
	jDef := sortDefs.defs[j]
	if iDef.File != jDef.File {
		return iDef.File < jDef.File
	}
	if iDef.Line != jDef.Line {
		return iDef.Line < jDef.Line
	}
	return iDef.Col < jDef.Col
}

func (sortDefs *sortableDefinition) Swap(i, j int) {
	temp := sortDefs.defs[i]
	sortDefs.defs[i] = sortDefs.defs[j]
	sortDefs.defs[j] = temp
}

func main() {
	var err error

	rename := flag.Bool("rename", false,
		"If set, then all defenitions "+
			"that will be determined as unused will be renamed in files")
	verbose := flag.Bool("verbose", false, "Turning on verbose mode")
	out := flag.String("out", "", "Output file. If not set then stdout will be used")
	exclude := flag.String("exclude", "",
		"File with exlude patterns for objects that shouldn't be unexported."+
			"Each pattern should be started at new line. Default pattern is Test* to exclude tests methods.")

	flag.Parse()
	pkg := flag.Arg(0)

	//Setup logging
	if *verbose {
		util.Level = "DEBUG"
	} else {
		util.Level = "ERROR"
	}

	//Setup excludes
	defaultRegexp, _ := regexp.Compile("Test*")
	excludeRegexps := []*regexp.Regexp{defaultRegexp}
	if len(*exclude) > 0 {
		excludeRegexps, err = readExcludes(*exclude)
		if err != nil {
			util.Fatalf("error while setup logging: %v", err)
		}
	}

	//Looking up for unused definitions, print them and rename
	if len(pkg) > 0 {
		unusedDefinitions, allDefinitions, err := getUnusedDefinitions(pkg, excludeRegexps)
		if err != nil {
			util.Fatalf("error while getting definitions: %v", err)
		}
		if err := printDefinitions(*out, unusedDefinitions); err != nil {
			util.Fatalf("error while printing result: %v", err)
		}
		if *rename {
			renameDefinitions(unusedDefinitions, allDefinitions)
		}
	} else {
		fmt.Printf("Usage: gounexport [OPTIONS] package\n")
		flag.PrintDefaults()
	}
}

func readExcludes(file string) ([]*regexp.Regexp, error) {
	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	var result []*regexp.Regexp
	regexpStrings := strings.Split(string(bytes), "\n")
	for _, regexpS := range regexpStrings {
		if len(regexpS) == 0 {
			continue
		}
		var regex *regexp.Regexp
		regex, err = regexp.Compile(regexpS)
		if err != nil {
			return nil, err
		}
		result = append(result, regex)
	}
	return result, err
}

func renameDefinitions(unused []*gounexport.Definition, allDefs map[string]*gounexport.Definition) {
	for _, def := range unused {
		gounexport.Unexport(def, allDefs, fs.ReplaceStringInFile)
	}
}

func getUnusedDefinitions(pkg string, excludes []*regexp.Regexp) (
	[]*gounexport.Definition, map[string]*gounexport.Definition, error) {
	info := types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
		Defs:  make(map[*ast.Ident]types.Object),
		Uses:  make(map[*ast.Ident]types.Object),
	}
	_, fset, err := gounexport.ParsePackage(pkg, &info)
	if err != nil {
		return nil, nil, err
	}
	defs := gounexport.GetDefinitions(&info, fset)
	return gounexport.GetDefenitionsToHide(pkg, defs, excludes), defs, nil
}

func printDefinitions(filename string, defs []*gounexport.Definition) error {
	output := definitionsToString(defs)
	if len(filename) > 0 {
		if err := ioutil.WriteFile(filename, []byte(output), os.ModePerm); err != nil {
			return err
		}
	} else {
		fmt.Print(output)
	}
	return nil
}

func definitionsToString(defs []*gounexport.Definition) string {
	sDef := new(sortableDefinition)
	sDef.defs = defs
	sort.Sort(sDef)

	result := "-----------------------------------------------------\n"
	result += fmt.Sprintf("Found %d unused definitions\n", len(defs))
	for _, def := range defs {
		result += fmt.Sprintf("%s - %s:%d:%d\n", def.Name, def.File, def.Line, def.Col)
		for _, u := range def.Usages {
			result += fmt.Sprintf("\t%s:%d:%d\n", u.Pos.Filename, u.Pos.Line, u.Pos.Column)
		}
	}
	return result
}
