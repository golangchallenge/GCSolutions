// The unexport command unexports exported identifiers which are not imported
// by any other Go code.
package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/build"
	"go/format"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/tools/go/buildutil"
	"golang.org/x/tools/go/loader"
	"golang.org/x/tools/go/types"
	"golang.org/x/tools/refactor/importgraph"
)

const usage = `unexport: change nonused exported identifiers to unexported identifiers

Usage:

  unexport -package <import-path>

Flags:

  -package     specifies the import path of the package 
  
  -identifier  explicit comma seperated list of identifiers. If empty (default)
               all possible exported identifiers are selected.
  
  -tags        a list of build tags to consider satisifed during the build.
  
  -dryrun      causes the tool to report conflicts but not update any files.
  
  -v           enables verbose logging.


Examples:

  $ unexport -package encoding/pem
  
  	Unexport exported identifiers in the encoding/pem package
  
  $ unexport -package github.com/fatih/color -dryrun
  
  	Process the package and display any possible changes. It doesn't unexport
  	exported identifiers because of the -dryrun flag
  
  $ unexport -package github.com/fatih/color -identifier "Color,Attribute"
  
  	Unexport only the "Color" and "Attribute" identifiers from the
  	github.com/fatih/color package. Note that if the identifiers are used by
  	other packages, it'll silently fail

Notes:

  unexport doesn't unexport identifiers which are dependent on other packages. It
  rejects identifiers in test files. If the the unexported identifier yields a
  collision, (example "Foo" -> "foo"), a warning is being displayed.

`

func main() {
	var (
		flagPackage    = flag.String("package", "", "package import path to be unexported")
		flagIdentifier = flag.String("identifier", "", "comma-separated list of identifiers names; if empty all identifiers are unexported")
		flagDryRun     = flag.Bool("dryrun", false, "show the change, but do not apply")
		flagVerbose    = flag.Bool("v", false, "enable verbose mode. Useful for debugging.")
	)

	flag.Var((*buildutil.TagsFlag)(&build.Default.BuildTags), "tags", buildutil.TagsFlagDoc)
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, usage)
	}

	flag.Parse()
	log.SetPrefix("unexport: ")

	if flag.NFlag() == 0 {
		flag.Usage()
		os.Exit(1)
	}

	identifiers := []string{}
	if *flagIdentifier != "" {
		identifiers = strings.Split(*flagIdentifier, ",")
	}

	if err := runMain(&config{
		importPath:   *flagPackage,
		identifiers:  identifiers,
		buildContext: &build.Default,
		dryRun:       *flagDryRun,
		verbose:      *flagVerbose,
	}); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

// config is used to define how unexport should be work
type config struct {
	// importPath defines the package defined with the importpath
	importPath string

	// identifiers is used to limit the changes of unexporting to certain identifiers
	identifiers []string

	// build context
	buildContext *build.Context

	// logging/development ...
	dryRun  bool
	verbose bool
}

// runMain runs the actual command. It's an helper function so we can easily
// calls defers or return errors.  most of the functionality can be seen in
// `gorename` code source. Actually unexport is something that probably
// gorename can do for us.
//
// TODO(arslan): add tests
// TODO(arslan): add vim-go integration ;)
// TODO(arslan): check how this is acting for internal/ folders
// TODO(arslan): check how this is acting for vendor/ folders
func runMain(conf *config) error {
	if conf.importPath == "" {
		return errors.New("import path of the package must be given")
	}

	path := conf.importPath

	prog, err := loadProgram(conf.buildContext, map[string]bool{path: true})
	if err != nil {
		return err
	}

	_, rev, errors := importgraph.Build(conf.buildContext)
	if len(errors) > 0 {
		// With a large GOPATH tree, errors are inevitable.
		// Report them but proceed.
		log.Printf("while scanning Go workspace:\n")
		for path, err := range errors {
			log.Printf("Package %q: %s.\n", path, err)
		}
	}

	// Enumerate the set of potentially affected packages.
	possiblePackages := make(map[string]bool)
	for _, obj := range findExportedObjects(prog, path) {
		for path := range rev.Search(obj.Pkg().Path()) {
			possiblePackages[path] = true
		}
	}

	if conf.verbose {
		log.Println("Possible affected packages:")
		for pkg := range possiblePackages {
			log.Println("\t", pkg)
		}
	}

	// reload the program with all possible packages to fetch the packageinfo's
	globalProg, err := loadProgram(conf.buildContext, possiblePackages)
	if err != nil {
		return err
	}

	objsToUpdate := make(map[types.Object]bool, 0)
	objects := findExportedObjects(globalProg, path)

	if conf.verbose {
		log.Println("Exported identififers are:")
		for _, obj := range objects {
			log.Println("\t", obj)
		}
	}

	// filter safeObjects check which exported identifiers are used by other packages
	var safeObjects map[*ast.Ident]types.Object
	for _, info := range globalProg.Imported {
		// we only check for packages other than ours
		if info.Pkg.Path() == path {
			continue
		}

		safeObjects = filterObjects(info, objects, conf.identifiers)
	}

	// filter out identifiers which can't be renamed due any collision in our package
	for _, info := range globalProg.Imported {
		// we don't care about other packages anymore
		if info.Pkg.Path() != path {
			continue
		}

		for _, obj := range safeObjects {
			// don't include collisions
			newName := toLowerCase(obj.Name())
			if info.Pkg.Path() == obj.Pkg().Path() && hasObject(info, newName) {
				log.Printf("WARNING! can't unexport %q due collision. Identifier %q already exists.\n",
					obj.Name(), newName)
				continue
			}

			objsToUpdate[obj] = true
		}
	}

	if conf.verbose {
		log.Println("Safe to unexport identifiers are:")
		for obj := range objsToUpdate {
			log.Println("\t", obj)
		}
	}

	// first create the files that needs an update and modify the fileset
	var nidents int
	var filesToUpdate = make(map[*token.File]bool)
	for _, info := range globalProg.Imported {
		for id, obj := range info.Defs {
			if objsToUpdate[obj] {
				nidents++
				id.Name = toLowerCase(obj.Name())
				filesToUpdate[globalProg.Fset.File(id.Pos())] = true
			}
		}
		for id, obj := range info.Uses {
			if objsToUpdate[obj] {
				nidents++
				id.Name = toLowerCase(obj.Name())
				filesToUpdate[globalProg.Fset.File(id.Pos())] = true
			}
		}
	}

	// now start to rewrite the files
	var nerrs, npkgs int
	for _, info := range globalProg.Imported {
		first := true
		for _, f := range info.Files {
			tokenFile := globalProg.Fset.File(f.Pos())
			if filesToUpdate[tokenFile] {
				if first {
					npkgs++
					first = false
				}

				if conf.dryRun {
					continue
				}

				if err := rewriteFile(globalProg.Fset, f, tokenFile.Name()); err != nil {
					log.Println(err)
					nerrs++
				}
			}
		}
	}

	if nidents == 0 {
		return nil
	}

	log.Printf("Unexported %d identifier%s in %d file%s in %d package%s.\n", nidents, plural(nidents),
		len(filesToUpdate), plural(len(filesToUpdate)),
		npkgs, plural(npkgs))
	log.Println("Identifiers changed:")
	for obj := range objsToUpdate {
		log.Println("\t", obj)
	}
	log.Println("Files changed:")
	for f := range filesToUpdate {
		log.Println("\t", f.Name())
	}

	if nerrs > 0 {
		return fmt.Errorf("failed to rewrite %d file%s", nerrs, plural(nerrs))
	}

	return nil
}

func toLowerCase(s string) string {
	if s == "" {
		return ""
	}

	r, n := utf8.DecodeRuneInString(s)
	return string(unicode.ToLower(r)) + s[n:]
}

// plural, copied from golang.org/x/tools/refactor/rename
func plural(n int) string {
	if n != 1 {
		return "s"
	}
	return ""
}

func rewriteFile(fset *token.FileSet, f *ast.File, filename string) error {
	var buf bytes.Buffer
	if err := format.Node(&buf, fset, f); err != nil {
		return fmt.Errorf("failed to pretty-print syntax tree: %v", err)
	}
	return ioutil.WriteFile(filename, buf.Bytes(), 0644)
}

// filterObjects filters the given objects and returns objects which are not in use by the given info package.
func filterObjects(info *loader.PackageInfo, exported map[*ast.Ident]types.Object, allowed []string) map[*ast.Ident]types.Object {
	isAllowed := func(id string) bool {
		for _, i := range allowed {
			if i == id {
				return true
			}
		}
		return false
	}

	if len(allowed) == 0 {
		isAllowed = func(id string) bool { return true }
	}

	filtered := make(map[*ast.Ident]types.Object, 0)
	for id, ex := range exported {
		if !hasUse(info, ex) && isAllowed(ex.Name()) {
			filtered[id] = ex
		}
	}

	return filtered
}

// hasUse returns true if the given obj is part of the use in info
func hasUse(info *loader.PackageInfo, obj types.Object) bool {
	for _, o := range info.Uses {
		if o == obj {
			return true
		}
	}
	return false
}

// hasObjects returns true if the given name exists in the definition of the package info
func hasObject(info *loader.PackageInfo, name string) bool {
	for _, obj := range info.Defs {
		if obj == nil {
			continue
		}

		if obj.Name() == name {
			return true
		}
	}

	return false
}

// exportedObjects returns objects which are exported only
func exportedObjects(info *loader.PackageInfo) map[*ast.Ident]types.Object {
	objects := make(map[*ast.Ident]types.Object, 0)
	for id, obj := range info.Defs {
		if obj == nil {
			continue
		}

		if obj.Exported() {
			objects[id] = obj
		}
	}

	return objects
}

// findExportedObjects returns a map of exported paths which are part of the
// given import path
func findExportedObjects(prog *loader.Program, path string) map[*ast.Ident]types.Object {
	var pkgObj *types.Package
	for pkg := range prog.AllPackages {
		if pkg.Path() == path {
			pkgObj = pkg
			break
		}
	}

	info := prog.AllPackages[pkgObj]
	return exportedObjects(info)
}

func loadProgram(ctxt *build.Context, pkgs map[string]bool) (*loader.Program, error) {
	conf := loader.Config{
		Build:       ctxt,
		ParserMode:  parser.ParseComments,
		AllowErrors: false,
	}

	for pkg := range pkgs {
		conf.Import(pkg)
	}
	return conf.Load()
}
