package unexporter

import (
	"go/ast"

	"golang.org/x/tools/go/loader"
	"golang.org/x/tools/go/types"
)

func isExported(obj types.Object) bool {
	// https://golang.org/ref/spec#Exported_identifiers
	// An identifier is exported if both:
	// the first character of the identifier's name is a Unicode upper case letter (Unicode class "Lu"); and
	// the identifier is declared in the package block or it is a field name or method name.
	// All other identifiers are not exported.

	if !obj.Exported() {
		// does not start with an upper case letter
		return false
	}

	if v, ok := obj.(*types.Var); ok && v.IsField() {
		// is a field name
		return true
	}
	if sig, ok := obj.Type().(*types.Signature); ok && sig.Recv() != nil {
		// is a method name
		return true
	}
	// is declared in the package block
	return obj.Parent() == obj.Pkg().Scope()
}

func isPackageLevel(obj types.Object) bool {
	return obj.Pkg().Scope().Lookup(obj.Name()) == obj
}

func isInterface(T types.Type) bool { return types.IsInterface(T) }

func isConcreteMethod(obj types.Object) bool {
	if f, ok := obj.(*types.Func); ok {
		r := recv(f)
		return r != nil && !isInterface(r.Type())
	}

	return false
}

// recv returns the method's receiver.
func recv(meth *types.Func) *types.Var {
	return meth.Type().(*types.Signature).Recv()
}

func deref(typ types.Type) types.Type {
	if p, _ := typ.(*types.Pointer); p != nil {
		return p.Elem()
	}
	return typ
}

// someUse returns an arbitrary use of obj within info.
func someUse(info *loader.PackageInfo, obj types.Object) *ast.Ident {
	for id, o := range info.Uses {
		if o == obj {
			return id
		}
	}
	return nil
}

func getEnclosingStruct(object types.Object) types.Type {
	pkgScope := object.Pkg().Scope()
	s := pkgScope.Innermost(object.Pos())
	if s.Parent() == pkgScope {
		s = pkgScope
	}

	var obj types.Object
	for _, name := range s.Names() {
		o := s.Lookup(name)
		if o.Pos() <= object.Pos() && (obj == nil || o.Pos() > obj.Pos()) {
			obj = o
		}
	}

	if obj == nil {
		return nil
	}
	if obj == object {
		return obj.Type()
	}

	t := obj.Type().Underlying().(*types.Struct)
	for {
		var f types.Object
		for i := 0; i < t.NumFields(); i++ {
			field := t.Field(i)
			if field == object {
				return t
			}
			if field.Pos() <= object.Pos() && (f == nil || field.Pos() > f.Pos()) {
				f = field
			}
		}
		if fs, ok := f.Type().(*types.Struct); ok {
			t = fs
		} else {
			break
		}
	}

	return t
}
