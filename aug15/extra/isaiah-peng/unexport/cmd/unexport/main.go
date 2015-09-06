package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/isaiah/unexport"
	"go/build"
	"golang.org/x/tools/go/buildutil"
	"golang.org/x/tools/go/types"
	"log"
	"os"
	"path/filepath"
	"runtime/pprof"
	t "runtime/trace"
	"strings"
)

var (
	helpFlag = flag.Bool("help", false, "show usage message")
	runall   = flag.Bool("all", false, "run all renaming, aborts if there are unsolved conflicts")
	dryrun   = flag.Bool("dryrun", false, "show the unused identifiers, but do not apply renaming")
	profile  = flag.Bool("profile", false, "memory profile")
	trace    = flag.Bool("trace", false, "trace goroutine execution")

	errNotGoSourcePath = errors.New("path is not under GOROOT or GOPATH")
)

func init() {
	flag.BoolVar(&unexport.Verbose, "v", false, "print extra verbose information")

	flag.Usage = func() {
		usage := `unexport: a tool that finds unnecessarily exported identifiers in a package and help unexport them

By default it checks the current directory, if it's in the GOPATH, otherwise you can specify the
targeting package as argument

It prompt the unexport of each identifier, you can choose to unexport, skip or apply an alternative name.
by default the resulting name is the original name with its first letter downcased.

Usage:

  unexport <flags> [package]

Flags:
`
		fmt.Fprintf(os.Stderr, "%s", usage)
		flag.PrintDefaults()
	}
}

func main() {
	flag.Parse()
	if *helpFlag {
		flag.Usage()
		return
	}
	ctxt := &build.Default
	var path string
	if len(flag.Args()) == 0 {
		path = getwdPackages(ctxt)
	} else {
		path = flag.Args()[0]
	}
	if *trace {
		f, err := os.Create("unexport_trace.json")
		if err != nil {
			log.Fatal(err)
		}
		t.Start(f)
		defer func() {
			f.Close()
			t.Stop()
		}()
	}

	unexporter, err := unexport.New(ctxt, path)
	if err != nil {
		panic(err)
	}
	if *trace {
		os.Exit(0)
	}
	if *dryrun {
		fmt.Println(`Following identifiers are exported but not used anywhere out of the package:
(The qualifiers are valid for gorename command)
`)
		for _, obj := range unexporter.UnusedObjectsSorted() {
			info := unexporter.Identifiers[obj]

			if info.Warning == "" {
				fmt.Println(unexporter.Qualifier(obj))
			} else {
				fmt.Printf("unexport %s causes conflict:\n%s\n", unexporter.Qualifier(obj), info.Warning)
			}
		}
		os.Exit(0)
	}
	if *profile {
		f, err := os.Create("unexport.mprof")
		if err != nil {
			log.Fatal(err)
		}
		pprof.WriteHeapProfile(f)
		defer f.Close()
		os.Exit(0)
	}
	if *runall {
		var conflict bool
		for obj, info := range unexporter.Identifiers {
			if info.Warning != "" {
				fmt.Printf("unexport %s causes conflicts\n%s", unexporter.Qualifier(obj), info.Warning)
				conflict = true
			}
		}
		if conflict {
			fmt.Println("Please fix the conflicts before continue.")
			os.Exit(1)
		}
		unexporter.UpdateAll()
		os.Exit(0)
	}

	// apply the changes
	fmt.Println("Please press corresponding key to proceed, y to confirm, n to skip, r to use a different name and c to cancel:")
	for _, obj := range unexporter.UnusedObjectsSorted() {
		info := unexporter.Identifiers[obj]
		var s string
		if info.Warning == "" {
			fmt.Printf("unexport %s, y/n/r/c? ", unexporter.Qualifier(obj))
		} else {
			fmt.Printf("unexport %s causes conflicts\n%s, \nn/r/c? ", unexporter.Qualifier(obj), info.Warning)
		}
		fmt.Scanf("%s", &s)
		switch s {
		case "y", "Y":
			unexporter.Update(obj)
		case "r":
			rename(unexporter, obj, info)
		case "c":
			os.Exit(1)
		default:
			continue
		}
	}
}

func rename(unexporter *unexport.Unexporter, obj types.Object, info *unexport.ObjectInfo) {
	var to string
	fmt.Printf("please input an alternative name: ")
	fmt.Scanf("%s", &to)
	warnings := unexporter.Check(obj, to)
	if warnings == "" {
		unexporter.Update(obj)
	} else {
		fmt.Printf("rename %s to %s still causes conflicts\n%s,\nr/c/A? ",
			unexporter.Qualifier(obj), to, warnings)
		// recursive
		rename(unexporter, obj, info)
	}

}

func getImportPath(ctxt *build.Context, pathOrFilename string) (string, error) {
	dirSlash := filepath.ToSlash(pathOrFilename)

	// We assume that no source root (GOPATH[i] or GOROOT) contains any other.
	for _, srcdir := range ctxt.SrcDirs() {
		srcdirSlash := filepath.ToSlash(srcdir) + "/"
		if strings.HasPrefix(dirSlash, srcdirSlash) {
			importPath := dirSlash[len(srcdirSlash):len(dirSlash)]
			return importPath, nil
		}
	}
	return "", errNotGoSourcePath
}

func getwdPackages(ctxt *build.Context) string {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	if !buildutil.FileExists(&build.Default, wd) {
		flag.Usage()
		os.Exit(1)
	}
	importPath, err := getImportPath(ctxt, wd)
	if err != nil {
		panic(err)
	}
	return importPath
}
