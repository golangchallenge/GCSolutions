package main

import (
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"go/types"
	"log"
)

type _package struct {
	pkg         *build.Package
	identifiers identifiers
	pkgFiles    map[string]*ast.Package
}

func newPkg(pkgPath string) (*_package, error) {
	buildPackage, err := build.Import(pkgPath, ".", build.ImportComment)
	if err != nil {
		return nil, err
	}

	pkg := &_package{
		pkg:         buildPackage,
		identifiers: make(identifiers),
	}

	pkg.initPackageInfo(buildPackage)
	return pkg, nil
}

func (p *_package) initPackageInfo(pkg *build.Package) {
	fset := token.NewFileSet()
	packages, err := parser.ParseDir(fset, pkg.Dir, nil, 0)
	if err != nil {
		log.Fatal(err)
	}
	p.pkgFiles = packages

	info := &types.Info{
		Defs: make(map[*ast.Ident]types.Object),
	}

	p.parsePackages(info, fset)
	for _, obj := range info.Defs {
		if obj != nil && obj.Exported() && obj.Type() != nil {
			p.parsePackageInfo(obj)
		}
	}
}

func (p *_package) parsePackageInfo(t types.Object) {
	switch t.(type) {
	case *types.TypeName:
		tn := t.(*types.TypeName)
		n := tn.Type().(*types.Named)

		if s, ok := n.Underlying().(*types.Struct); ok {
			p.identifiers.add(newObject(p.pkg, t, nil))

			//methods
			for i := 0; i < n.NumMethods(); i++ {
				method := n.Method(i)
				if method.Exported() {
					p.identifiers.add(newObject(p.pkg, method, t))
				}
			}

			//fields
			for i := 0; i < s.NumFields(); i++ {
				field := s.Field(i)
				if field.Exported() {
					p.identifiers.add(newObject(p.pkg, field, t))
				}
			}
		}
	case *types.Const:
		p.identifiers.add(newObject(p.pkg, t, nil))
	case *types.Func:
		p.identifiers.add(newObject(p.pkg, t, nil))
	case *types.Var:
		v := t.(*types.Var)
		if !v.IsField() {
			p.identifiers.add(newObject(p.pkg, v, nil))
		}
	case *types.Label:
		break
	default:
		log.Fatalf("WARNING: Unhandled type %T\n", t)
	}

}

func (p *_package) calculateUsesOf(pkg *_package) {
	fset := token.NewFileSet()
	_, err := parser.ParseDir(fset, p.pkg.Dir, nil, 0)
	if err != nil {
		log.Fatal(err)
	}

	info := &types.Info{
		Uses: make(map[*ast.Ident]types.Object),
	}

	p.parsePackages(info, fset)
	for ident, obj := range info.Uses {
		if ident.IsExported() {
			pos := fset.Position(ident.Pos())
			pkg.identifiers.addUsage(newObject(pkg.pkg, obj, obj), pos)
		}
	}
}

func (p *_package) parsePackages(info *types.Info, fset *token.FileSet) {
	for pName, p := range p.pkgFiles {
		files := make([]*ast.File, 0, len(p.Files))
		for _, f := range p.Files {
			files = append(files, f)
		}

		conf := &types.Config{
			FakeImportC: true,
			Importer:    newImportWrapper(),
		}
		_, err := conf.Check(pName, fset, files, info)
		if err != nil {
			log.Printf("WARNING: %s", err)
		}
	}
}
