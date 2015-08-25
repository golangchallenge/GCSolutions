/*
unexport searches unnecessarily exported identifiers from a package and
provides facilities for unexporting them with help of gorename-command.
*/
package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
)

var (
	pkg            = flag.String("pkg", "", "go package to unexport")
	searchFrom     = flag.String("search", "", "search usage of exported identifiers from this package (and its subpackages) NOTE: If this has not been specified, whole $GOPATH will be searched (and it might take a while).")
	unsafe         = flag.Bool("unsafe", false, "allow unexporting other than internal packages")
	report         = flag.Bool("report", false, "show which exported identifiers are used and by whom")
	offsetNotation = flag.Bool("offset", false, "use 'file offset'-notation instead of 'from'-notation for gorename")
	help           = flag.Bool("help", false, "print this help")
)

const usage = `
unexport searches unnecessarily exported identifiers from a package.
Identifiers are declared unnecessarily exported, if no-one outside the declaring
package is using them. This can't be 100% guaranteed for other than "internal"-packages,
so by default only "internal"-packages can be unexported.

Unexport candidates are provided as "gorename"-commands. Possible naming collisions caused by
unexporting are also reported.

Usage:
unexport -pkg <target-package> [-search <search-from-package>] [-unsafe] [-offset] [-help] [-report]

Examples:
unexport -pkg cmd/compile/internal/gc

unexport -pkg cmd/compile/internal/gc -search cmd/compile/internal/big

unexport -pkg cmd/compile/internal/gc -report

unexport -pkg cmd/vet -search cmd/fix -unsafe -offset

Flags:`

// gorenameOffsetCmd returns a string containing a
// command for running gorename in "-offset" notation.
// format is: gorename -offset file.go:#123 -to foo
func gorenameOffsetCmd(ident identifier) string {
	return fmt.Sprintf("gorename -offset %s:#%d -to %s",
		ident.pos.Filename,
		ident.pos.Offset,
		ident.unexportedName())
}

// gorenameFromCmd returns a string containing a
// command for running gorename in "-from" notation.
// format is: gorename -from '\"bytes\".Buffer.Len' -to Size
func gorenameFromCmd(ident identifier) string {
	return fmt.Sprintf("gorename -from '%s' -to %s",
		ident.string(),
		ident.unexportedName())
}

// printGorename prints out a list of gorename commands
// which can be used to unexport unused identifiers.
// Prints one command per identifier.
func printGorename(unused []identifier, offsetNotation bool) {
	var commands []string
	for _, id := range unused {
		if offsetNotation {
			commands = append(commands, gorenameOffsetCmd(id))
		} else {
			commands = append(commands, gorenameFromCmd(id))
		}
	}
	sort.Strings(commands)

	fmt.Fprintln(os.Stdout, "\nGorename commands:")
	fmt.Fprintln(os.Stdout, strings.Join(commands, "\n"))
}

// printReport prints out a report of identifier usages.
// Usages are grouped by exported identifiers. A call-site
// list is shown for each identifier.
func printReport(report map[identifier][]string) {
	fmt.Fprintln(os.Stdout, "\nUsage report:")
	for id, usages := range report {
		fmt.Fprintf(os.Stdout, "%s used %d times:\n", id.string(), len(usages))
		for _, u := range usages {
			fmt.Fprintln(os.Stdout, "\t", u)
		}
	}
}

// printCollisions prints out identifiers, which, if unexported,
// will cause a naming collision.
func printCollisions(warnings []identifier) {
	if len(warnings) == 0 {
		return
	}

	fmt.Fprintln(os.Stdout, "\n*** WARNING ***")
	fmt.Fprintln(os.Stdout, "Renaming following indentifiers will cause a naming collision:")

	var wstr []string
	for _, w := range warnings {
		wstr = append(wstr, w.string())
	}
	fmt.Fprintln(os.Stdout, strings.Join(wstr, "\n"))
}

func printHelp() {
	fmt.Fprintln(os.Stdout, usage)
	flag.PrintDefaults()
}

func main() {
	flag.Parse()

	if *help {
		printHelp()
		os.Exit(0)
	}

	if *pkg == "" {
		fmt.Fprintln(os.Stderr, "Package not specified. Specify with -pkg.")
		printHelp()
		os.Exit(2)
	}

	if !*unsafe && !isInternal(*pkg) {
		fmt.Fprintln(os.Stderr, "Package is not internal. Use -unsafe, if you know what you are doing.")
		printHelp()
		os.Exit(2)
	}

	if *report {
		report := usageReport(*pkg, *searchFrom)
		printReport(report)
		os.Exit(0)
	}

	fmt.Fprintln(os.Stdout, "unexporting package:", *pkg)

	packages := packagesToSearch(*pkg, *searchFrom)
	unused := unusedExports(*pkg, packages)
	collisions := nameCollisions(*pkg, unused)

	printGorename(unused, *offsetNotation)
	printCollisions(collisions)

	os.Exit(0)
}
