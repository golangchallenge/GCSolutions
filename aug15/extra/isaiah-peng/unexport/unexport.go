package unexport

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/build"
	"go/format"
	"go/parser"
	"go/token"
	"golang.org/x/tools/go/loader"
	"golang.org/x/tools/go/types"
	"golang.org/x/tools/refactor/importgraph"
	"golang.org/x/tools/refactor/lexical"
	"io/ioutil"
	"log"
	"sort"
)

var (
	Verbose bool
)

func (u *Unexporter) unusedObjects() []types.Object {
	if len(u.unexportableObjects) != 0 {
		return u.unexportableObjects
	}
	used := u.usedObjects()
	var objs []types.Object
	for _, pkgInfo := range u.packages {
		if pkgInfo.Pkg.Path() != u.path {
			continue
		}
		for id, obj := range pkgInfo.Defs {
			if used[obj] {
				continue
			}
			if id.IsExported() {
				objs = append(objs, obj)
			}
		}
		// No need to go further if path found
		break
	}
	u.unexportableObjects = objs
	return objs
}

func (u *Unexporter) usedObjects() map[types.Object]bool {
	objs := make(map[types.Object]bool)
	for _, pkgInfo := range u.packages {
		// easy path
		for id, obj := range pkgInfo.Uses {
			// ignore builtin value
			if obj.Pkg() == nil {
				continue
			}
			// if it's a type from different package, store it
			if obj.Pkg() != pkgInfo.Pkg {
				objs[obj] = true
			}
			// embedded fields are marked as used, no much which package the original type belongs to,
			// so that they won't show up in the renaming list #16
			if field := pkgInfo.Defs[id]; field != nil {
				// embdded field identifier is the same as it's type
				objs[field] = true
			}
		}
	}
	// Check assignability
	for key := range u.satisfy() {
		var (
			lhs, rhs *types.Named
			ok       bool
		)
		if lhs, ok = key.LHS.(*types.Named); !ok {
			continue
		}
		switch r := key.RHS.(type) {
		case *types.Named:
			rhs = r
		case *types.Pointer: // the receiver could be a pointer, see #14
			rhs = r.Elem().(*types.Named)
		default:
			continue
		}

		lset := u.msets.MethodSet(key.LHS)
		rset := u.msets.MethodSet(key.RHS)
		for i := 0; i < lset.Len(); i++ {
			obj := lset.At(i).Obj()
			// LHS are the abstract methods, they are only exported if there are other packages using it
			if lhs.Obj().Pkg() != rhs.Obj().Pkg() {
				objs[obj] = true
			}
			// if satisfied by type within the same package only, it should be unexported
			// however, we should not rename from the concret method side, but from the
			// interface side, carefully exclude concret methods that don't implement an abstract method (see #14, #17)
			rsel := rset.Lookup(rhs.Obj().Pkg(), obj.Name())
			objs[rsel.Obj()] = true
		}
	}
	return objs
}

func getDeclareStructOrInterface(prog *loader.Program, v *types.Var) string {
	// From x/tools/refactor/rename/check.go(checkStructField)#L288
	// go/types offers no easy way to get from a field (or interface
	// method) to its declaring struct (or interface), so we must
	// ascend the AST.
	_, path, _ := prog.PathEnclosingInterval(v.Pos(), v.Pos())
	// path matches this pattern:
	// [Ident SelectorExpr? StarExpr? Field FieldList StructType ParenExpr* ... File]

	// Ascend to FieldList.
	var i int
	for {
		if _, ok := path[i].(*ast.FieldList); ok {
			break
		}
		i++
	}
	i++
	_ = path[i].(*ast.StructType)
	i++
	for {
		if _, ok := path[i].(*ast.ParenExpr); !ok {
			break
		}
		i++
	}
	if spec, ok := path[i].(*ast.TypeSpec); ok {
		return spec.Name.String()
	}
	return ""
}

func loadProgram(ctx *build.Context, pkgs []string) (*loader.Program, error) {
	conf := loader.Config{
		Build:       ctx,
		ParserMode:  parser.ParseComments,
		AllowErrors: false,
	}
	for _, pkg := range pkgs {
		conf.Import(pkg)
	}
	return conf.Load()
}

// New creates a new Unexporter object that holds the states
func New(ctx *build.Context, path string) (*Unexporter, error) {
	pkgs := scanWorkspace(ctx, path)
	prog, err := loadProgram(ctx, pkgs)

	if err != nil {
		return nil, err
	}
	u := &Unexporter{
		path:          path,
		iprog:         prog,
		packages:      make(map[*types.Package]*loader.PackageInfo),
		warnings:      make(chan map[types.Object]string),
		Identifiers:   make(map[types.Object]*ObjectInfo),
		lexinfos:      make(map[*loader.PackageInfo]*lexical.Info),
		changeMethods: true, // always true for unexporter
	}

	for _, info := range prog.Imported {
		u.packages[info.Pkg] = info
	}

	for _, info := range prog.Created {
		u.packages[info.Pkg] = info
	}

	unusedObjs := u.unusedObjects()
	objs := make(chan map[types.Object]map[types.Object]string, 20)
	input := make(chan types.Object, 20)
	for _, obj := range unusedObjs {
		u.Identifiers[obj] = &ObjectInfo{}
	}
	go func() {
		for _, obj := range unusedObjs {
			input <- obj
		}
	}()
	// spawn the workers
	for i := 0; i < 10; i++ {
		go func() {
			for {
				select {
				case obj := <-input:
					toName := lowerFirst(obj.Name())
					objsToUpdate := make(map[types.Object]string)
					u.check(objsToUpdate, obj, toName)
					objs <- map[types.Object]map[types.Object]string{obj: objsToUpdate}
				}
			}
		}()
	}
	for i := 0; i < len(unusedObjs); {
		select {
		case m := <-u.warnings:
			for obj, warning := range m {
				if u.Identifiers[obj] != nil {
					u.Identifiers[obj].Warning = warning
				}
			}
		case m := <-objs:
			for obj, objsToUpdate := range m {
				u.Identifiers[obj].objsToUpdate = objsToUpdate
			}
			i++
		}
	}
DONE:
	for {
		select {
		case m := <-u.warnings:
			for obj, warning := range m {
				u.Identifiers[obj].Warning = warning
			}
		default:
			break DONE
		}
	}
	return u, nil
}

// Update unexport the specified identifier
func (u *Unexporter) Update(obj types.Object) error {
	return u.update(u.Identifiers[obj].objsToUpdate)
}

// UpdateAll apply all renaming, conflicts are ignored
func (u *Unexporter) UpdateAll() error {
	objsToUpdate := make(map[types.Object]string)
	for _, objInfo := range u.Identifiers {
		for obj, to := range objInfo.objsToUpdate {
			objsToUpdate[obj] = to
		}
	}
	return u.update(objsToUpdate)
}

// Check checks if any possible renaming conflict and return the conflict information
func (u *Unexporter) Check(from types.Object, to string) string {
	objsToUpdate := make(map[types.Object]string)
	u.check(objsToUpdate, from, to)
	close(u.warnings)
	u.Identifiers[from] = &ObjectInfo{objsToUpdate: objsToUpdate}
	for ws := range u.warnings {
		for _, warning := range ws {
			u.Identifiers[from].Warning = warning
			return warning
		}
	}
	return ""
}

// sort the objects, see #8
type typeObjects []types.Object

func (t typeObjects) Len() int      { return len(t) }
func (t typeObjects) Swap(i, j int) { t[i], t[j] = t[j], t[i] }
func (t typeObjects) Less(i, j int) bool {
	// field or method should be placed at the end
	_, field := t[i].(*types.Var)
	_, meth := t[i].(*types.Func)
	return field || meth
}

// UnusedObjectsSorted place the unused field and method before everything else, so that
// they are renamed before the other, this is necessary as otherwise the we need to re-generate
// the qualifier of field and method if the type it belongs to has changed
func (u *Unexporter) UnusedObjectsSorted() []types.Object {
	objs := u.unusedObjects()
	sort.Sort(typeObjects(objs))
	return objs
}

// Qualifier the full qualifier for specified object, ready for consumption of `gorename` command
func (u *Unexporter) Qualifier(obj types.Object) string {
	return wholePath(obj, u.path, u.iprog)
}

// This is copy & pasted from x/tools/refactor/rename
// update renames the identifiers, updates the input files.
func (u *Unexporter) update(objsToUpdate map[types.Object]string) error {
	// We use token.File, not filename, since a file may appear to
	// belong to multiple packages and be parsed more than once.
	// token.File captures this distinction; filename does not.
	var nidents int
	var filesToUpdate = make(map[*token.File]bool)
	for _, info := range u.packages {
		// Mutate the ASTs and note the filenames.
		for id, obj := range info.Defs {
			if to, ok := objsToUpdate[obj]; ok {
				nidents++
				id.Name = to
				filesToUpdate[u.iprog.Fset.File(id.Pos())] = true
			}
		}
		for id, obj := range info.Uses {
			if to, ok := objsToUpdate[obj]; ok {
				nidents++
				id.Name = to
				filesToUpdate[u.iprog.Fset.File(id.Pos())] = true
			}
		}
	}

	// TODO(adonovan): don't rewrite cgo + generated files.
	var nerrs, npkgs int
	for _, info := range u.packages {
		first := true
		for _, f := range info.Files {
			tokenFile := u.iprog.Fset.File(f.Pos())
			if filesToUpdate[tokenFile] {
				if first {
					npkgs++
					first = false
					if Verbose {
						log.Printf("Updating package %s\n",
							info.Pkg.Path())
					}
				}
				if err := rewriteFile(u.iprog.Fset, f, tokenFile.Name()); err != nil {
					log.Printf("gorename: %s\n", err)
					nerrs++
				}
			}
		}
	}
	log.Printf("Renamed %d occurrence%s in %d file%s in %d package%s.\n",
		nidents, plural(nidents),
		len(filesToUpdate), plural(len(filesToUpdate)),
		npkgs, plural(npkgs))
	if nerrs > 0 {
		return fmt.Errorf("failed to rewrite %d file%s", nerrs, plural(nerrs))
	}
	return nil
}

func plural(n int) string {
	if n != 1 {
		return "s"
	}
	return ""
}

var rewriteFile = func(fset *token.FileSet, f *ast.File, filename string) (err error) {
	// TODO(adonovan): print packages and filenames in a form useful
	// to editors (so they can reload files).
	if Verbose {
		log.Printf("\t%s\n", filename)
	}
	var buf bytes.Buffer
	if err := format.Node(&buf, fset, f); err != nil {
		return fmt.Errorf("failed to pretty-print syntax tree: %v", err)
	}
	return ioutil.WriteFile(filename, buf.Bytes(), 0644)
}

func scanWorkspace(ctxt *build.Context, path string) []string {
	// Scan the workspace and build the import graph.
	_, rev, errors := importgraph.Build(ctxt)
	if len(errors) > 0 {
		// With a large GOPATH tree, errors are inevitable.
		// Report them but proceed.
		log.Printf("While scanning Go workspace:\n")
		for path, err := range errors {
			log.Printf("Package %q: %s.\n", path, err)
		}
	}

	// Enumerate the set of potentially affected packages.
	var affectedPackages []string
	// External test packages are never imported,
	// so they will never appear in the graph.
	for pkg := range rev.Search(path) {
		affectedPackages = append(affectedPackages, pkg)
	}
	return affectedPackages
}
