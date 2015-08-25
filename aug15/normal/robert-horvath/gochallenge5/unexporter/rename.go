package unexporter

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"io/ioutil"

	"golang.org/x/tools/go/types"
)

// In this file we do the actual renaming (unexporting) by changing the AST
// and rewriting affected files

// Unexport does the actual renaming of unnecessarily exported identifiers
// and the rewriting of affected files
func (u *Unexporter) Unexport() error {
	files := make(map[*token.File]bool)

	for _, e := range u.unnecessaryExports() {
		if u.IsNameExcluded(e.Name()) {
			continue
		}

		newName := e.UnexportedName()
		if e.Conflicting {
			fmt.Println(e.Name(), "can't be automatically unexported, skipping")
			continue
		}

		for obj := range e.objsToUpdate {
			u.rename(obj, newName, files)
		}
	}

	return u.rewriteFiles(files)
}

func (u *Unexporter) rename(obj types.Object, to string, files map[*token.File]bool) {
	r := func(o types.Object, ident *ast.Ident) {
		if o == obj {
			ident.Name = to
			files[u.prog.Fset.File(ident.NamePos)] = true
		}
	}

	for ident, o := range u.pkgInfo.Defs {
		r(o, ident)
	}
	for ident, o := range u.pkgInfo.Uses {
		r(o, ident)
	}
}

func (u *Unexporter) rewriteFiles(files map[*token.File]bool) error {
	for _, f := range u.pkgInfo.Files {
		tokenFile := u.prog.Fset.File(f.Pos())
		if files[tokenFile] {
			var buf bytes.Buffer
			err := format.Node(&buf, u.prog.Fset, f)
			if err != nil {
				return err
			}

			err = ioutil.WriteFile(tokenFile.Name(), buf.Bytes(), 0644)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
