package main

import (
	"fmt"
	"go/build"
	"go/token"
	"go/types"
	"strings"
	"unicode"
)

type identifiers map[string]*identifier

func (o identifiers) add(ident *identifier) {
	o[ident.id()] = ident
}

func (i identifiers) addUsage(ident *identifier, pos token.Position) {
	if id, exists := i[ident.id()]; exists {
		id.usedBy = append(id.usedBy, pos)
	}
}

func (i identifiers) get(key string) *identifier {
	return i[key]
}

type identifier struct {
	buildPkg *build.Package
	parent   types.Object
	this     types.Object
	usedBy   []token.Position
}

func newObject(pkg *build.Package, obj types.Object, parent types.Object) *identifier {
	if !obj.Exported() {
		panic("Only exported objects")
	}
	if v, ok := obj.(*types.Var); ok && v.IsField() && parent == nil {
		panic("Expected a non nil parent")
	}
	return &identifier{
		buildPkg: pkg,
		parent:   parent,
		this:     obj,
		usedBy:   make([]token.Position, 0),
	}
}

func (i *identifier) id() string {
	switch i.this.(type) {
	case *types.Func:
		f := i.this.(*types.Func)
		sign := f.Type().(*types.Signature)
		if sign.Recv() != nil {
			v := sign.Recv()
			t := v.Type()
			if ptr, ok := t.(*types.Pointer); ok {
				t = ptr.Elem()
			} else if _, ok := t.(*types.Interface); ok {
				return fmt.Sprintf("\"%s\".%s.%s", i.buildPkg.ImportPath, "interface", f.Name())
			}
			named := t.(*types.Named)
			return fmt.Sprintf("\"%s\".%s.%s", i.buildPkg.ImportPath, named.Obj().Name(), f.Name())
		}
		break
	case *types.Var:
		v := i.this.(*types.Var)
		if v.IsField() {
			if i.parent == nil {
				panic("Expected a parent")
			}
			return fmt.Sprintf("\"%s\".%s.%s", i.buildPkg.ImportPath, i.parent.Name(), v.Name())
		}
		break
	}
	return fmt.Sprintf("\"%s\".%s", i.buildPkg.ImportPath, i.this.Name())
}

func (i *identifier) unexportedId() string {
	id := i.this.Name()
	if len(id) > 1 && unicode.IsUpper(rune(id[1])) {
		return strings.ToLower(id)
	}
	return fmt.Sprintf("%s%s", strings.ToLower(string(id[0])), id[1:])
}
