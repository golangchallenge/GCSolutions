package unexporter

import (
	"go/token"
	"unicode"
	"unicode/utf8"

	"golang.org/x/tools/go/types"
)

// Export encapsulates an exported object and contains information about
// if it can be automatically unexported, where it occurs in other packages and its interface constraints
type Export struct {
	obj          types.Object
	objsToUpdate map[types.Object]bool
	Conflicting  bool
	Occurences   map[*types.Package]int
	Interfaces   map[string][]*types.Package
	u            *Unexporter
}

// Name returns the name of the exported object
func (e Export) Name() string {
	return e.obj.Name()
}

// ObjName returns a string representation of the exported object
func (e Export) ObjName() string {
	return e.obj.String()
}

// UnexportedName returns the name of the exported object, with the first letter lowercased
func (e Export) UnexportedName() string {
	name := e.obj.Name()
	r, size := utf8.DecodeRuneInString(name)
	return string(unicode.ToLower(r)) + name[size:]
}

// Position returns the position of the exported object as token.Position
func (e Export) Position() token.Position {
	return e.u.prog.Fset.PositionFor(e.obj.Pos(), true)
}

// PositionForObject returns the position for an object as token.Position
func (u *Unexporter) PositionForObject(o types.Object) token.Position {
	return u.prog.Fset.PositionFor(o.Pos(), true)
}

func (e Export) isUnused() bool {
	return len(e.Occurences) == 0
}

// Unnecessary is true if the export is not used by other packages and has no interface constraints
func (e Export) Unnecessary() bool {
	return e.isUnused() && len(e.Interfaces) == 0
}

func newExport(u *Unexporter, obj types.Object) Export {
	e := Export{}
	e.obj = obj
	e.Occurences = make(map[*types.Package]int)
	e.Interfaces = u.isInterfaceMethod(e)
	e.objsToUpdate = make(map[types.Object]bool)
	e.u = u
	e.check(obj, e.UnexportedName())

	for obj := range e.objsToUpdate {
		for _, useInfo := range u.prog.Imported {
			pkg := useInfo.Pkg
			if pkg != u.pkgInfo.Pkg {
				for _, useObj := range useInfo.Uses {
					if useObj == obj {
						e.Occurences[pkg]++
					}
				}
			}
		}
	}

	return e
}

func (u *Unexporter) findExports() []Export {
	for _, obj := range u.pkgInfo.Defs {
		if obj != nil && isExported(obj) {
			e := newExport(u, obj)
			u.exports = append(u.exports, e)
		}
	}

	return u.exports
}

func (u *Unexporter) unnecessaryExports() []Export {
	var unnecessaryExports []Export
	for _, e := range u.exports {
		if e.Unnecessary() {
			unnecessaryExports = append(unnecessaryExports, e)
		}
	}
	return unnecessaryExports
}

// NecessaryExports returns a list of exports that are used by other packages
func (u *Unexporter) NecessaryExports() []Export {
	var necessaryExports []Export
	for _, e := range u.exports {
		if !e.Unnecessary() {
			necessaryExports = append(necessaryExports, e)
		}
	}
	return necessaryExports
}

// Exports returns all found exported objects
func (u *Unexporter) Exports() []Export {
	return u.exports
}

// UnexportableExports returns a list of exports that cannot be automatically unexported
func (u *Unexporter) UnexportableExports() []Export {
	var exports []Export

	for _, e := range u.unnecessaryExports() {
		if !e.Conflicting && !u.IsNameExcluded(e.Name()) {
			exports = append(exports, e)
		}
	}
	return exports
}

// UnexportableObjects returns a list of exported objects that cannot be automatically unexported
func (u *Unexporter) UnexportableObjects() []types.Object {
	exports := u.UnexportableExports()
	objects := make(map[types.Object]bool)

	for _, e := range exports {
		for obj := range e.objsToUpdate {
			objects[obj] = true
		}
	}

	var objSlice []types.Object
	for obj := range objects {
		objSlice = append(objSlice, obj)
	}

	return objSlice
}
