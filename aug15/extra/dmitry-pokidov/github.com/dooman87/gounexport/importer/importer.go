//Package importer provides implementation of go/type/importer.
//
//It provides an ability to collect information, such as types, definitions,
//usages about the package and  all inner packages.
//
//The main method is Collect. It will analyze package and return
//pointer to package and file set. Also, CollectInfoImporter.Info structure will
//bee filled.
//For example,
//  info := types.Info{
//  	Types: make(map[ast.Expr]types.TypeAndValue),
//  	Defs:  make(map[*ast.Ident]types.Object),
//  	Uses:  make(map[*ast.Ident]types.Object),
//  }
//  importer := new(CollectInfoImporter)
//  importer.Pkg = pkg + "/testfunc/main"
//  importer.Info = &info
//  resultPkg, fset, err := importer.Collect()
package importer

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"strings"

	"github.com/dooman87/gounexport/fs"
	"github.com/dooman87/gounexport/util"
)

// CollectInfoImporter importing packages with dependencies and
// collecting information to info field. You need to provide
// Pkg and Info prior to use it
type CollectInfoImporter struct {
	//Info struct that will be filled by Collect() method
	Info *types.Info
	//Package that should be a start point to collect info
	Pkg      string
	fset     *token.FileSet
	astFiles []*ast.File
	packages map[string]*types.Package
}

func (*CollectInfoImporter) errorHandler(err error) {
	util.Warn("error while checking source: %v", err)
}

var (
	defaultImporter = importer.Default()
)

//Collect going through package and collect info
//using conf.Check method. It's using this implementation
//of importer for check all inner packages and go/types/importer.Default()
//to check all built in packages (without sources)
func (_importer *CollectInfoImporter) Collect() (*types.Package, *token.FileSet, error) {
	var conf types.Config
	conf.Importer = _importer
	conf.Error = _importer.errorHandler

	if _importer.packages == nil {
		_importer.packages = make(map[string]*types.Package)
	}

	var pkg *types.Package
	var err error
	var files []string

	if files, err = fs.GetFiles(_importer.Pkg, false); err != nil {
		return nil, nil, err
	}
	if _importer.fset, _importer.astFiles, err = doParseFiles(files, _importer.fset); err != nil {
		return nil, nil, err
	}

	//XXX: return positive result if check() returns error.
	pkg, _ = conf.Check(_importer.Pkg, _importer.fset, _importer.astFiles, _importer.Info)
	// if pkg, err = conf.Check(_importer.Pkg, _importer.fset, _importer.astFiles, _importer.Info); err != nil {
	// 	return pkg, _importer.fset, err
	// }

	_importer.packages[_importer.Pkg] = pkg
	util.Debug("package [%s] successfully parsed\n", pkg.Name())

	return pkg, _importer.fset, nil
}

//Import parses the package or returns it from cache if it was
//already imported. Also, it collects information if path is under
//Pkg package
func (_importer *CollectInfoImporter) Import(path string) (*types.Package, error) {
	if _importer.packages[path] != nil {
		return _importer.packages[path], nil
	}

	util.Info("importing package [%s]", path)

	var pkg *types.Package
	var err error

	if strings.Contains(path, _importer.Pkg) {
		if pkg, err = _importer.doImport(path, true); err != nil {
			return pkg, err
		}
	}

	pkg, err = defaultImporter.Import(path)
	if err != nil {
		pkg, err = _importer.doImport(path, true)
	}

	if pkg != nil {
		_importer.packages[path] = pkg
	}
	util.Info("package [%s] imported: [%v] [%v]", path, pkg, err)
	return pkg, err
}

func (_importer *CollectInfoImporter) doImport(path string, collectInfo bool) (*types.Package, error) {
	var pkg *types.Package
	var err error
	var conf types.Config
	conf.Importer = _importer
	conf.Error = _importer.errorHandler

	files, err := fs.GetFiles(path, false)
	if err != nil {
		return nil, err
	}

	fset, astFiles, err := doParseFiles(files, _importer.fset)
	if err != nil {
		return nil, err
	}

	if collectInfo {
		pkg, err = conf.Check(path, fset, astFiles, _importer.Info)
	} else {
		pkg, err = conf.Check(path, fset, astFiles, nil)
	}
	return pkg, err
}

func doParseFiles(filePathes []string, fset *token.FileSet) (*token.FileSet, []*ast.File, error) {
	if fset == nil {
		fset = token.NewFileSet()
	}
	util.Info("parsing files %v", filePathes)
	astFiles := make([]*ast.File, 0, len(filePathes))
	for _, f := range filePathes {
		//XXX: Ignoring files with packages ends with _test.
		//XXX: Doing that because getting error in check()
		//XXX: cause source file is still going to current
		//XXX: packages. Need to analyze package before
		//XXX: and check both packages separately.
		tempFset := token.NewFileSet()
		astFile, err := parser.ParseFile(tempFset, f, nil, 0)
		if !strings.HasSuffix(astFile.Name.Name, "_test") {
			if err != nil {
				return nil, nil, err
			}
			astFile, _ := parser.ParseFile(fset, f, nil, 0)
			astFiles = append(astFiles, astFile)
		}
	}

	iterateFunc := func(f *token.File) bool {
		util.Debug("\t%s", f.Name())
		return true
	}
	fset.Iterate(iterateFunc)
	return fset, astFiles, nil
}
