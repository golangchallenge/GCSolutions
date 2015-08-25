package unexporter

import (
	"go/build"
	"go/parser"
	"os"
	"strings"

	"golang.org/x/tools/go/loader"
	"golang.org/x/tools/go/types"
	"golang.org/x/tools/go/types/typeutil"
	"golang.org/x/tools/refactor/importgraph"
	"golang.org/x/tools/refactor/satisfy"
)

// Unexporter holds all necessary information to unexport from a package
type Unexporter struct {
	pkgInfo       *loader.PackageInfo
	ctxt          *build.Context
	prog          *loader.Program
	f             map[satisfy.Constraint]map[*types.Package]bool
	msets         typeutil.MethodSetCache
	exports       []Export
	ExcludedNames []string
}

func (u *Unexporter) isInterfaceMethod(e Export) map[string][]*types.Package {
	if _, ok := e.obj.(*types.Func); !ok {
		return nil
	}
	f := e.obj.(*types.Func)
	r := f.Type().(*types.Signature).Recv()
	if r == nil {
		return nil
	}

	if u.f == nil {
		calculateConstraints(u)
	}

	interfaces := make(map[string][]*types.Package)
	for constraint := range u.f {
		if constraint.RHS == r.Type() && types.IsInterface(constraint.LHS) {
			sure, _, _ := types.LookupFieldOrMethod(constraint.LHS, true, f.Pkg(), f.Name())
			if sure != nil {
				interfaces[constraint.LHS.String()] = append(interfaces[constraint.LHS.String()], f.Pkg())
			}
		}
	}

	return interfaces
}

// ForPackage creates a new unexporter for the given build context and package
func ForPackage(ctxt *build.Context, pkgPath string) (*Unexporter, error) {
	u := &Unexporter{
		ctxt: ctxt,
	}

	err := u.loadProgram(pkgPath)
	if err != nil {
		return nil, err
	}

	u.findExports()
	return u, nil
}

func (u *Unexporter) loadProgram(pkgPath string) (err error) {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	bpkg, err := u.ctxt.Import(pkgPath, wd, build.ImportComment)
	if err != nil {
		return err
	}

	_, rev, _ := importgraph.Build(u.ctxt)
	pkgs := rev.Search(bpkg.ImportPath)

	conf := loader.Config{
		Build:       u.ctxt,
		ParserMode:  parser.ParseComments,
		AllowErrors: false,
	}

	// Optimization: don't type-check the bodies of functions in our
	// dependencies, since we only need exported package members.
	conf.TypeCheckFuncBodies = func(p string) bool {
		return pkgs[p] || pkgs[strings.TrimSuffix(p, "_test")]
	}
	for pkg := range pkgs {
		conf.ImportWithTests(pkg)
	}

	u.prog, err = conf.Load()
	if err != nil {
		return
	}
	for p, info := range u.prog.AllPackages {
		if p.Path() == bpkg.ImportPath {
			u.pkgInfo = info
			break
		}
	}

	return
}

func (u *Unexporter) IsNameExcluded(name string) bool {
	for _, n := range u.ExcludedNames {
		if name == n {
			return true
		}
	}
	return false
}

func calculateConstraints(u *Unexporter) {
	constraints := make(map[satisfy.Constraint]map[*types.Package]bool)

	for _, info := range u.prog.Imported {
		var finder satisfy.Finder
		finder.Find(&info.Info, info.Files)

		for constraint := range finder.Result {
			if _, ok := constraints[constraint]; !ok {
				constraints[constraint] = make(map[*types.Package]bool)
			}
			constraints[constraint][info.Pkg] = true
		}
	}
	u.f = constraints
}
