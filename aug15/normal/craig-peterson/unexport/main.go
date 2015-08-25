package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	_ "golang.org/x/tools/go/gcimporter"
	"golang.org/x/tools/go/types"
	"golang.org/x/tools/oracle"
	"golang.org/x/tools/refactor/rename"
)

var (
	flagForce    = flag.Bool("f", false, "Unexport all unused members without prompting")
	flagVerbose  = flag.Bool("v", false, "Verbose output")
	flagDownload = flag.Bool("d", false, "Look for users of package on godoc.org and download them. Can download a lot of code for widely used packages.")
)

func main() {
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "Usage: unexport (opts) package")
		flag.PrintDefaults()
		os.Exit(1)
	}
	pkgName := flag.Arg(0)
	if *flagDownload {
		DownloadPackagesUsing(pkgName)
		return
	}
	pkgDir := FindPackageDirectory(pkgName)

	verbosef("Parsing go files in %s", pkgDir)
	tPackage, fset := ParsePackage(pkgName, pkgDir)

	verbosef("Walking tree for exported symbols...")
	candidates := FindAllExports(tPackage, fset)
	if len(candidates) == 0 {
		log.Fatal("No exported symbols. Nothing to do.")
	}

	verbosef("Found %d exported symbols. Checking for external usages...", len(candidates))
	candidatesWithUsages := FindUsages(candidates, pkgDir)

	var r renamer = defaultRenamer{}
	if *flagForce {
		r = inProcRenamer{}
	}
	for _, u := range candidatesWithUsages {
		if len(u.References) == 0 {
			r.Rename(u)
		}
	}
}

//a renamer performs an action to rename a symbol.
type renamer interface {
	Rename(u CandidateWithUsage)
}

//the default renamer simply prints a gorename command to standard out.
type defaultRenamer struct{}

func (d defaultRenamer) Rename(u CandidateWithUsage) {
	fmt.Printf("gorename -offset=%s:#%d -to=%s\n", u.Pos.Filename, u.Pos.Offset, unexportedName(u.Name))
}

//inProcRenamer (-f) will automatically rename symbols in this process.
type inProcRenamer struct{}

func (i inProcRenamer) Rename(u CandidateWithUsage) {
	fmt.Println("Renaming", u.DisplayName)
	rename.Main(&build.Default, fmt.Sprintf("%s:#%d", u.Pos.Filename, u.Pos.Offset), "", unexportedName(u.Name))
	time.Sleep(3 * time.Millisecond) // hack to let output buffers flush so messages stay in order
}

func unexportedName(name string) string {
	return strings.ToLower(string(name[0])) + name[1:]
}

// Represents an exported symbol
type UnexportCandidate struct {
	Name        string
	DisplayName string
	Pos         token.Position
}

// Represents an exported symbol and usages we have found in the gopath.
type CandidateWithUsage struct {
	UnexportCandidate
	References []string
}

// Find usages for the given candidate symbols in the local system gopath.
// Relies on oracle to find referrers.
func FindUsages(candidates []UnexportCandidate, dir string) []CandidateWithUsage {
	data := []CandidateWithUsage{}
	workCh := make(chan UnexportCandidate, len(candidates))
	resultChan := make(chan CandidateWithUsage)
	quit := make(chan struct{})

	// run oracle on 10 symbols at a time in parallel.
	for i := 0; i < 10; i++ {
		go func() {
			for {
				select {
				case <-quit:
					return
				case c := <-workCh:
					q := &oracle.Query{}
					q.Build = &build.Default
					q.Mode = "referrers"
					q.Pos = fmt.Sprintf("%s:#%d", c.Pos.Filename, c.Pos.Offset)
					err := oracle.Run(q)
					if err != nil {
						resultChan <- CandidateWithUsage{c, []string{err.Error()}}
						break
					}
					refs := q.Serial().Referrers
					usage := CandidateWithUsage{c, nil}
					for _, ref := range refs.Refs {
						// filter out usages from within the package itself
						file := ref[:strings.LastIndex(ref, ":")]
						file = file[:strings.LastIndex(file, ":")]
						if filepath.Dir(file) != dir {
							usage.References = append(usage.References, ref)
						}
					}
					resultChan <- usage
				}
			}
		}()
	}
	for _, c := range candidates {
		workCh <- c
	}
	for len(data) < len(candidates) {
		r := <-resultChan
		data = append(data, r)
		verbosef("Usages of %s:", r.DisplayName)
		if len(r.References) == 0 {
			verbose("", "None Found.")
		} else {
			for _, ref := range r.References {
				verbose("", ref)
			}
		}
	}
	close(quit)
	return data
}

var fileFilter func(f os.FileInfo) bool

func FindPackageDirectory(name string) string {
	pkg, err := build.Default.Import(name, "", 0)
	if err != nil {
		log.Fatal(err)
	}
	//filter used later to only parse go files (no test or other files)
	fileFilter = func(f os.FileInfo) bool {
		for _, gof := range pkg.GoFiles {
			if gof == f.Name() {
				return true
			}
		}
		return false
	}
	return pkg.Dir
}

func ParsePackage(name, dir string) (*types.Package, *token.FileSet) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dir, fileFilter, parser.DeclarationErrors)
	astFiles := []*ast.File{}
	for _, pkg := range pkgs {
		for _, f := range pkg.Files {
			astFiles = append(astFiles, f)
		}
	}
	tPackage, err := types.Check(name, fset, astFiles)
	if err != nil {
		log.Fatal(err)
	}
	if !tPackage.Complete() {
		log.Fatal("Parsing of package incomplete.")
	}
	return tPackage, fset
}

func FindAllExports(pkg *types.Package, fset *token.FileSet) []UnexportCandidate {
	candidates := []UnexportCandidate{}
	for _, name := range pkg.Scope().Names() {
		obj := pkg.Scope().Lookup(name)
		if !obj.Exported() {
			continue
		}
		displayName := obj.Name()
		if _, ok := obj.(*types.Func); ok {
			displayName += "()"
		}
		candidate := UnexportCandidate{obj.Name(), displayName, fset.Position(obj.Pos())}
		candidates = append(candidates, candidate)
		if tn, ok := obj.(*types.TypeName); ok {
			if str, ok := tn.Type().Underlying().(*types.Struct); ok {
				candidates = append(candidates, findStructFields(str, obj.Name(), fset)...)
			}
			ptrType := types.NewPointer(tn.Type())
			methodSet := types.NewMethodSet(ptrType)
			for i := 0; i < methodSet.Len(); i++ {
				methodSel := methodSet.At(i)
				method := methodSel.Obj()
				// skip unexported functions, and functions from embedded fields.
				// The best I can figure out for embedded functions is if the selection index path is longer than 1.
				if !method.Exported() || len(methodSel.Index()) > 1 {
					continue
				}
				candidate := UnexportCandidate{method.Name(), obj.Name() + "." + method.Name() + "()", fset.Position(method.Pos())}
				candidates = append(candidates, candidate)
			}
		}
	}
	return candidates
}

//iterate over a struct's fields recursively and add its fields to the candidate list
func findStructFields(str *types.Struct, prefix string, fset *token.FileSet) []UnexportCandidate {
	candidates := []UnexportCandidate{}
	for i := 0; i < str.NumFields(); i++ {
		field := str.Field(i)
		if !field.Exported() || field.Anonymous() {
			continue
		}
		// Tags are a likely indicator that this field is used by something
		// like json or xml that uses reflection. Skip it to be safe.
		// TODO: override flag? whitelist?
		if str.Tag(i) != "" {
			continue
		}
		candidate := UnexportCandidate{field.Name(), prefix + "." + field.Name(), fset.Position(field.Pos())}
		candidates = append(candidates, candidate)
		if nested, ok := field.Type().(*types.Struct); ok {
			candidates = append(candidates, findStructFields(nested, candidate.DisplayName, fset)...)
		}
	}
	return candidates
}

func DownloadPackagesUsing(pkgName string) {
	godocUrl := fmt.Sprintf("http://api.godoc.org/importers/%s", pkgName)
	resp, err := http.Get(godocUrl)
	if err != nil {
		log.Fatal(err)
	}
	if resp.StatusCode != 200 {
		log.Fatalf("Unexpected status from godoc server: %d", resp.StatusCode)
	}
	defer resp.Body.Close()
	reader := json.NewDecoder(resp.Body)
	results := &GodocResults{}
	err = reader.Decode(results)
	if err != nil {
		log.Fatal(err)
	}
	for _, pkg := range results.Results {
		verbosef("go get -d -v %s", pkg.Path)
		cmd := exec.Command("go", "get", "-d", "-v", pkg.Path)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			log.Println(err)
		}
	}
}

type GodocResults struct {
	Results []struct {
		Path     string `json:"path"`
		Synopsis string `json:"synopsis"`
	} `json:"results"`
}

func verbose(x ...interface{}) {
	if *flagVerbose {
		log.Println(x...)
	}
}

func verbosef(f string, x ...interface{}) {
	if *flagVerbose {
		log.Printf(f, x...)
	}
}
