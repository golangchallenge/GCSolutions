package main

import (
	"bufio"
	"flag"
	"fmt"
	"go/build"
	"io"
	"os"
	"regexp"
	"strings"
	"text/tabwriter"

	"golang.org/x/tools/go/types"

	"github.com/robhor/gochallenge5/unexporter"
)

var (
	safeFlag        bool
	listFlag        bool
	usesFlag        bool
	interactiveFlag bool
	showFilename    bool
	simpleNamesFlag bool
	excludes        string
)

func init() {
	flag.BoolVar(&safeFlag, "safe", false, "only allow unexporting objects from internal packages")
	flag.BoolVar(&listFlag, "list", false, "lists all objects that can be unexported, but doesn't actually unexport anything")
	flag.StringVar(&excludes, "e", "", "list of names that should not be unexported, separated by ','")
	flag.BoolVar(&usesFlag, "uses", false, "list packages in which exported objects are used")
	flag.BoolVar(&interactiveFlag, "i", false, "interactive mode, asks if object should be exported")
	flag.BoolVar(&simpleNamesFlag, "s", false, "Use short names instead of full object names")
	flag.BoolVar(&showFilename, "n", false, "Show filename and position in lists")
}

func main() {
	flag.Parse()

	pkg := flag.Arg(0)

	if safeFlag && !isInternal(pkg) {
		fmt.Fprintln(os.Stderr, "Safe mode only allows unexporting objects from internal packages")
		return
	}

	// Finds all packages that import this one in the workspace
	ctxt := &build.Default
	u, err := unexporter.ForPackage(ctxt, pkg)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return
	}
	u.ExcludedNames = strings.Split(excludes, `,`)

	if listFlag {
		listObjects(u)
	} else if usesFlag {
		listUses(u)
	}

	if listFlag || usesFlag {
		return
	}

	if !interactiveFlag {
		u.Unexport()
	} else {
		interactiveUnexport(u)
	}
}

func isInternal(pkg string) bool {
	matched, err := regexp.MatchString(`/internal($|/)`, pkg)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return false
	}
	return matched
}

func listObjects(u *unexporter.Unexporter) {
	for _, o := range u.UnexportableObjects() {
		printObject(u, o)
	}
}

func printObject(u *unexporter.Unexporter, o types.Object) {
	var objName string
	if simpleNamesFlag {
		objName = o.Name()
	} else {
		objName = o.String()
	}
	if showFilename {
		pos := u.PositionForObject(o)
		fmt.Printf("%s:%d:%d: %s\n", pos.Filename, pos.Line, pos.Column, objName)
	} else {
		fmt.Println(objName)
	}
}

func printExport(w io.Writer, e unexporter.Export) {
	if showFilename {
		pos := e.Position()
		fmt.Fprintf(w, "%s:%d:%d: %s\n", pos.Filename, pos.Line, pos.Column, exportName(e))
	} else {
		fmt.Fprintln(w, exportName(e))
	}
}

func exportName(e unexporter.Export) string {
	var exportName string
	if simpleNamesFlag {
		exportName = e.Name()
	} else {
		exportName = e.ObjName()
	}
	return exportName
}

func listUses(u *unexporter.Unexporter) {
	exports := u.NecessaryExports()
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 2, 3, ' ', tabwriter.DiscardEmptyColumns|tabwriter.AlignRight)
	for _, e := range exports {
		printExport(w, e)
		listInterfaces(e)
		listOccurences(w, e)
		fmt.Fprintln(w, "\t")
	}

	w.Flush()
}

func listInterfaces(e unexporter.Export) {
	for i, pkgs := range e.Interfaces {
		fmt.Printf("  Used as %s in: ", i)

		var pkgNames []string
		for _, pkg := range pkgs {
			pkgNames = append(pkgNames, pkg.String())
		}
		fmt.Println(strings.Join(pkgNames, ", "))
	}
}

func listOccurences(w io.Writer, e unexporter.Export) {
	for pkg, num := range e.Occurences {
		fmt.Fprintf(w, "%d\t %s\n", num, pkg)
	}
}

func interactiveUnexport(u *unexporter.Unexporter) {
	reader := bufio.NewReader(os.Stdin)
	exports := u.UnexportableObjects()

	for _, o := range exports {
		if u.IsNameExcluded(o.Name()) {
			continue
		}
		
		printObject(u, o)
		fmt.Printf("Unexport this [yN]: ")
		text, _ := reader.ReadString('\n')
		if strings.TrimSpace(text) != "y" {
			u.ExcludedNames = append(u.ExcludedNames, o.Name())
		}
	}

	u.Unexport()
}
