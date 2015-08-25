package importer

import (
	"go/ast"
	"log"
	"testing"

	"go/token"
	"go/types"
)

const (
	pkg = "github.com/dooman87/gounexport/testdata"
)

func TestCollect(t *testing.T) {
	log.Print("-----------------------TestCollect-----------------------")
	info := types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
		Defs:  make(map[*ast.Ident]types.Object),
		Uses:  make(map[*ast.Ident]types.Object),
	}

	importer := new(CollectInfoImporter)
	importer.Pkg = pkg + "/testfunc/main"
	importer.Info = &info
	resultPkg, fset, err := importer.Collect()
	if err != nil {
		t.Errorf("error while collect info from %s, %v", importer.Pkg, err)
	}

	fileCounter := 0
	iterator := func(f *token.File) bool {
		fileCounter++
		return true
	}
	fset.Iterate(iterator)

	if fileCounter != 2 {
		t.Fatalf("expected 2 files in result file set but found %d", fileCounter)
	}
	if resultPkg == nil {
		t.Fatal("package should not be nil")
	}
}
