package unexporter

// This file defines the safety checks for each kind of renaming.
// Adapted for this special usecase from golang.org/x/tools/refactor/rename

import (
	"strings"

	"golang.org/x/tools/go/types"
	"golang.org/x/tools/refactor/lexical"
	"golang.org/x/tools/refactor/satisfy"
)

// check performs safety checks of the renaming of the 'from' object to 'to'.
// returns true if there is a renaming conflict
func (e *Export) check(from types.Object, to string) {
	if e.objsToUpdate[from] {
		return
	}
	e.objsToUpdate[from] = true

	e.checkUsedInAnotherPackage(from)
	if e.Conflicting {
		return
	}

	if isPackageLevel(from) {
		e.checkInPackageBlock(from, to)
	} else if v, ok := from.(*types.Var); ok && v.IsField() {
		e.checkStructField(v, to)
	} else if f, ok := from.(*types.Func); ok && recv(f) != nil {
		e.checkMethod(f, to)
	}
}

func (e *Export) checkUsedInAnotherPackage(from types.Object) {
	for _, useInfo := range e.u.prog.Imported {
		pkg := useInfo.Pkg
		if pkg != e.u.pkgInfo.Pkg {
			for _, useObj := range useInfo.Uses {
				if useObj == from {
					e.Conflicting = true
					return
				}
			}
		}
	}
}

// checkInPackageBlock performs safety checks for renames of
// func/var/const/type objects in the package block.
func (e *Export) checkInPackageBlock(from types.Object, to string) {
	info := e.u.pkgInfo
	lexinfo := lexical.Structure(e.u.prog.Fset, from.Pkg(), &info.Info, info.Files)

	// We don't rename anything in the package block to init, as that might
	// conflict or otherwise break stuff
	if to == "init" {
		e.Conflicting = true
		return
	}

	// Check for conflicts between package block and all file blocks.
	for _, f := range info.Files {
		if _, b := lexinfo.Blocks[f].Lookup(to); b == lexinfo.Blocks[f] {
			e.Conflicting = true
			return
		}
	}

	if f, ok := from.(*types.Func); ok && recv(f) == nil {
		e.checkFunction(f, to)
		if e.Conflicting {
			return
		}
	}

	// Check for conflicts in lexical scope.
	// Do not need to check all imported packages:
	// Since it's unnecessarily exported, no one else is going to be sad if I unexport it!
	e.checkInLexicalScope(from, to)
}

func (e *Export) checkInLexicalScope(from types.Object, to string) {
	info := e.u.pkgInfo
	lexinfo := lexical.Structure(e.u.prog.Fset, info.Pkg, &info.Info, info.Files)

	b := lexinfo.Defs[from] // the block defining the 'from' object
	if b != nil {
		to, toBlock := b.Lookup(to)
		if toBlock == b {
			e.Conflicting = true
			return // same-block conflict
		} else if toBlock != nil {
			for _, ref := range lexinfo.Refs[to] {
				if obj, _ := ref.Env.Lookup(from.Name()); obj == from {
					e.Conflicting = true
					return // super-block conflict
				}
			}
		}
	}

	// Check for sub-block conflict.
	// Is there an intervening definition of r.to between
	// the block defining 'from' and some reference to it?
	for _, ref := range lexinfo.Refs[from] {
		_, fromBlock := ref.Env.Lookup(from.Name())
		fromDepth := fromBlock.Depth()

		to, toBlock := ref.Env.Lookup(to)
		if to != nil {
			// sub-block conflict
			if toBlock.Depth() > fromDepth {
				e.Conflicting = true
				return
			}
		}
	}

	// Renaming a type that is used as an embedded field
	// requires renaming the field too. e.g.
	// 	type T int // if we rename this to U..
	// 	var s struct {T}
	// 	print(s.T) // ...this must change too
	if _, ok := from.(*types.TypeName); ok {
		for id, obj := range info.Uses {
			if obj == from {
				if field := info.Defs[id]; field != nil {
					e.check(field, to)
				}
			}
		}
	}
}

// checkStructField checks that the field renaming will not cause
// conflicts at its declaration, or ambiguity or changes to any selection.
func (e *Export) checkStructField(from *types.Var, to string) {
	// Check that the struct declaration is free of field conflicts,
	// and field/method conflicts.
	t := getEnclosingStruct(from)
	if t != t.Underlying() {
		// This struct is also a named type.
		// We must check for direct (non-promoted) field/field
		// and method/field conflicts.
		_, indices, _ := types.LookupFieldOrMethod(t, true, e.u.pkgInfo.Pkg, to)
		if len(indices) == 1 {
			e.Conflicting = true
			return
		}
	} else {
		// This struct is not a named type.
		// We need only check for direct (non-promoted) field/field conflicts.
		T := t.Underlying().(*types.Struct)
		for i := 0; i < T.NumFields(); i++ {
			if prev := T.Field(i); prev.Name() == to {
				e.Conflicting = true
				return
			}
		}
	}

	// Renaming an anonymous field requires renaming the type too. e.g.
	// 	print(s.T)       // if we rename T to U,
	// 	type T int       // this and
	// 	var s struct {T} // this must change too.
	if from.Anonymous() {
		if named, ok := from.Type().(*types.Named); ok {
			e.check(named.Obj(), to)
		} else if named, ok := deref(from.Type()).(*types.Named); ok {
			e.check(named.Obj(), to)
		}
	}

	// Check integrity of existing (field and method) selections.
	e.checkSelections(from, to)
}

// checkSelection checks that all uses and selections that resolve to
// the specified object would continue to do so after the renaming.
func (e *Export) checkSelections(from types.Object, to string) {
	info := e.u.pkgInfo
	if id := someUse(info, from); id != nil {
		e.Conflicting = true
		return
	}

	for _, sel := range info.Selections {
		if sel.Obj() == from {
			if obj, indices, _ := types.LookupFieldOrMethod(sel.Recv(), true, from.Pkg(), to); obj != nil {
				// Renaming this existing selection of
				// 'from' may block access to an existing
				// type member named 'to'.
				delta := len(indices) - len(sel.Index())
				if delta > 0 {
					continue // no ambiguity
				}
				e.Conflicting = true
				return
			}

		} else if sel.Obj().Name() == to {
			if obj, indices, _ := types.LookupFieldOrMethod(sel.Recv(), true, from.Pkg(), from.Name()); obj == from {
				// Renaming 'from' may cause this existing
				// selection of the name 'to' to change
				// its meaning.
				delta := len(indices) - len(sel.Index())
				if delta > 0 {
					continue // no ambiguity
				}
				e.Conflicting = true
				return
			}
		}
	}
}

func (e *Export) checkMethod(from *types.Func, to string) {
	// e.g. error.Error
	if from.Pkg() == nil {
		e.Conflicting = true
		return
	}

	// ASSIGNABILITY: We reject renamings of concrete methods that
	// would break a 'satisfy' constraint; but renamings of abstract
	// methods are allowed to proceed, and we rename affected
	// concrete and abstract methods as necessary.  It is the
	// initial method that determines the policy.

	// Check for conflict at point of declaration.
	// Check to ensure preservation of assignability requirements.
	R := recv(from).Type()
	if isInterface(R) {
		// Abstract method

		// declaration
		prev, _, _ := types.LookupFieldOrMethod(R, false, from.Pkg(), to)
		if prev != nil {
			e.Conflicting = true
			return
		}

		// Check all interfaces that embed this one for
		// declaration conflicts too.
		for _, info := range e.u.prog.AllPackages {
			// Start with named interface types (better errors)
			for _, obj := range info.Defs {
				if obj, ok := obj.(*types.TypeName); ok && isInterface(obj.Type()) {
					f, _, _ := types.LookupFieldOrMethod(
						obj.Type(), false, from.Pkg(), from.Name())
					if f == nil {
						continue
					}
					t, _, _ := types.LookupFieldOrMethod(obj.Type(), false, from.Pkg(), to)
					if t == nil {
						continue
					}
					e.Conflicting = true
					return
				}
			}
		}

		// assignability
		//
		// Find the set of concrete or abstract methods directly
		// coupled to abstract method 'from' by some
		// satisfy.Constraint, and rename them too.
		for key := range e.u.satisfy() {
			// key = (lhs, rhs) where lhs is always an interface.

			lsel := e.u.msets.MethodSet(key.LHS).Lookup(from.Pkg(), from.Name())
			if lsel == nil {
				continue
			}
			rmethods := e.u.msets.MethodSet(key.RHS)
			rsel := rmethods.Lookup(from.Pkg(), from.Name())
			if rsel == nil {
				continue
			}

			// If both sides have a method of this name,
			// and one of them is m, the other must be coupled.
			var coupled *types.Func
			switch from {
			case lsel.Obj():
				coupled = rsel.Obj().(*types.Func)
			case rsel.Obj():
				coupled = lsel.Obj().(*types.Func)
			default:
				continue
			}

			// We must treat concrete-to-interface
			// constraints like an implicit selection C.f of
			// each interface method I.f, and check that the
			// renaming leaves the selection unchanged and
			// unambiguous.
			//
			// Fun fact: the implicit selection of C.f
			// 	type I interface{f()}
			// 	type C struct{I}
			// 	func (C) g()
			//      var _ I = C{} // here
			// yields abstract method I.f.  This can make error
			// messages less than obvious.
			//
			if !isInterface(key.RHS) {
				// The logic below was derived from checkSelections.

				rtosel := rmethods.Lookup(from.Pkg(), to)
				if rtosel != nil {
					delta := len(rsel.Index()) - len(rtosel.Index())
					if delta < 0 {
						continue // no ambiguity
					}
					e.Conflicting = true
					return
				}
			}

			// Rename the coupled method to preserve assignability.
			e.check(coupled, to)
		}
	}

	// Concrete method

	// declaration
	prev, indices, _ := types.LookupFieldOrMethod(R, true, from.Pkg(), to)
	if prev != nil && len(indices) == 1 {
		e.Conflicting = true
		return
	}

	// assignability
	//
	// Find the set of abstract methods coupled to concrete
	// method 'from' by some satisfy.Constraint, and rename
	// them too.
	//
	// Coupling may be indirect, e.g. I.f <-> C.f via type D.
	//
	// 	type I interface {f()}
	//	type C int
	//	type (C) f()
	//	type D struct{C}
	//	var _ I = D{}
	//
	for key := range e.u.satisfy() {
		// key = (lhs, rhs) where lhs is always an interface.
		if isInterface(key.RHS) {
			continue
		}
		rsel := e.u.msets.MethodSet(key.RHS).Lookup(from.Pkg(), from.Name())
		if rsel == nil || rsel.Obj() != from {
			continue // rhs does not have the method
		}
		lsel := e.u.msets.MethodSet(key.LHS).Lookup(from.Pkg(), from.Name())
		if lsel == nil {
			continue
		}
		imeth := lsel.Obj().(*types.Func)
		e.check(imeth, to)
	}

	// Check integrity of existing (field and method) selections.
	// We skip this if there were errors above, to avoid redundant errors.
	e.checkSelections(from, to)
}

// check if function is a test function for the testing package
// we don't unexport those
func (e *Export) checkFunction(from *types.Func, to string) {
	if !strings.HasPrefix(from.Name(), "Test") {
		return
	}
	sig := from.Type().(*types.Signature)
	if sig.Params().Len() != 1 {
		return
	}
	if sig.Params().At(0).Type().String() == "*testing.T" {
		e.Conflicting = true
		return
	}
}

// satisfy returns the set of interface satisfaction constraints.
func (u *Unexporter) satisfy() map[satisfy.Constraint]bool {
	if u.f == nil {
		calculateConstraints(u)
	}
	m := make(map[satisfy.Constraint]bool)
	for constraint := range u.f {
		m[constraint] = true
	}
	return m
}
