package unexport

import (
	"fmt"
	"go/ast"
	"go/token"
	"golang.org/x/tools/go/loader"
	"golang.org/x/tools/go/types"
	"golang.org/x/tools/go/types/typeutil"
	"golang.org/x/tools/refactor/lexical"
	"golang.org/x/tools/refactor/satisfy"
	"reflect"
	"strings"
	"sync"
)

// Content of this file is copy & pasted from x/tools/refactor/rename,
// Because the check method nor the rename.reportError method are exported
// Ideally two interfaces should be exported by x/tools/refactor/rename
// * the check interface, but it should be thread-safe, as to support checking multiple object concurrently
//     it should return the collision errors instead of print to IO, preferrablly the renamer.errorf method should be atomic
// * the update interface, it should accept the affected packages for context and objects to update for bulk update
// Things I don't like:
// fmt.Fprintf(os.Stderr, ....) is very ugly, export a `ReportError` interface is OK, but use the `log` package is probably better?

// Maybe instead of expose the above mentioned functions from `x/tools/refactor/rename`, it could use the functions exposed by this package instead,
// since it only uses a subset of the functionalities. e.g. it doesn't require thread-safty

type Unexporter struct {
	path               string
	changeMethods      bool
	iprog              *loader.Program
	packages           map[*types.Package]*loader.PackageInfo // subset of iprog.AllPackages to inspect
	msets              typeutil.MethodSetCache
	satisfyConstraints map[satisfy.Constraint]bool
	warnings           chan map[types.Object]string
	Identifiers        map[types.Object]*ObjectInfo
	// memoization
	unexportableObjects []types.Object
	lexinfos            map[*loader.PackageInfo]*lexical.Info
	mutex               sync.Mutex
}

type ObjectInfo struct {
	Warning      string
	objsToUpdate map[types.Object]string
}

func (r *Unexporter) check(objsToUpdate map[types.Object]string, from types.Object, to string) {
	if _, ok := objsToUpdate[from]; ok {
		return
	}
	objsToUpdate[from] = to

	// NB: order of conditions is important.
	if from_, ok := from.(*types.PkgName); ok {
		r.checkInFileBlock(objsToUpdate, from_, to)
	} else if from_, ok := from.(*types.Label); ok {
		r.checkLabel(from_, to)
	} else if isPackageLevel(from) {
		r.checkInPackageBlock(objsToUpdate, from, to)
	} else if v, ok := from.(*types.Var); ok && v.IsField() {
		r.checkStructField(objsToUpdate, v, to)
	} else if f, ok := from.(*types.Func); ok && recv(f) != nil {
		r.checkMethod(objsToUpdate, f, to)
	} else if isLocal(from) {
		r.checkInLocalScope(objsToUpdate, from, to)
	} else {
		r.errorf(from.Pos(), "unexpected %s object %q (please report a bug)\n",
			objectKind(from), from)
	}
}

// checkInFileBlock performs safety checks for renames of objects in the file block,
// i.e. imported package names.
func (r *Unexporter) checkInFileBlock(objsToUpdate map[types.Object]string, from *types.PkgName, to string) {
	// Check import name is not "init".
	if to == "init" {
		r.errorf(from.Pos(), "%q is not a valid imported package name", to)
	}

	// Check for conflicts between file and package block.
	if prev := from.Pkg().Scope().Lookup(to); prev != nil {
		r.warn(from,
			r.errorf(from.Pos(), "renaming this %s %q to %q would conflict",
				objectKind(from), from.Name(), to),
			r.errorf(prev.Pos(), "\twith this package member %s",
				objectKind(prev)))
		return // since checkInPackageBlock would report redundant errors
	}

	// Check for conflicts in lexical scope.
	r.checkInLexicalScope(objsToUpdate, from, to, r.packages[from.Pkg()])

	// Finally, modify ImportSpec syntax to add or remove the Name as needed.
	info, path, _ := r.iprog.PathEnclosingInterval(from.Pos(), from.Pos())
	if from.Imported().Name() == to {
		// ImportSpec.Name not needed
		path[1].(*ast.ImportSpec).Name = nil
	} else {
		// ImportSpec.Name needed
		if spec := path[1].(*ast.ImportSpec); spec.Name == nil {
			spec.Name = &ast.Ident{NamePos: spec.Path.Pos(), Name: to}
			info.Defs[spec.Name] = from
		}
	}
}

// checkInPackageBlock performs safety checks for renames of
// func/var/const/type objects in the package block.
func (r *Unexporter) checkInPackageBlock(objsToUpdate map[types.Object]string, from types.Object, to string) {
	// Check that there are no references to the name from another
	// package if the renaming would make it unexported.
	if ast.IsExported(from.Name()) && !ast.IsExported(to) {
		for pkg, info := range r.packages {
			if pkg == from.Pkg() {
				continue
			}
			if id := someUse(info, from); id != nil &&
				!r.checkExport(id, pkg, from, to) {
				break
			}
		}
	}

	info := r.packages[from.Pkg()]
	lexinfo := r.lexInfo(info)

	// Check that in the package block, "init" is a function, and never referenced.
	if to == "init" {
		kind := objectKind(from)
		if kind == "func" {
			// Reject if intra-package references to it exist.
			if refs := lexinfo.Refs[from]; len(refs) > 0 {
				r.warn(from,
					r.errorf(from.Pos(),
						"renaming this func %q to %q would make it a package initializer",
						from.Name(), to),
					r.errorf(refs[0].Id.Pos(), "\tbut references to it exist"))
			}
		} else {
			r.warn(from, r.errorf(from.Pos(), "you cannot have a %s at package level named %q",
				kind, to))
		}
	}

	// Check for conflicts between package block and all file blocks.
	for _, f := range info.Files {
		if prev, b := lexinfo.Blocks[f].Lookup(to); b == lexinfo.Blocks[f] {
			r.warn(from,
				r.errorf(from.Pos(), "renaming this %s %q to %q would conflict",
					objectKind(from), from.Name(), to),
				r.errorf(prev.Pos(), "\twith this %s",
					objectKind(prev)))
			return // since checkInPackageBlock would report redundant errors
		}
	}

	// Check for conflicts in lexical scope.
	if from.Exported() {
		for _, info := range r.packages {
			r.checkInLexicalScope(objsToUpdate, from, to, info)
		}
	} else {
		r.checkInLexicalScope(objsToUpdate, from, to, info)
	}
}

func (r *Unexporter) checkInLocalScope(objsToUpdate map[types.Object]string, from types.Object, to string) {
	info := r.packages[from.Pkg()]

	// Is this object an implicit local var for a type switch?
	// Each case has its own var, whose position is the decl of y,
	// but Ident in that decl does not appear in the Uses map.
	//
	//   switch y := x.(type) {	 // Defs[Ident(y)] is undefined
	//   case int:    print(y)       // Implicits[CaseClause(int)]    = Var(y_int)
	//   case string: print(y)       // Implicits[CaseClause(string)] = Var(y_string)
	//   }
	//
	var isCaseVar bool
	for syntax, obj := range info.Implicits {
		if _, ok := syntax.(*ast.CaseClause); ok && obj.Pos() == from.Pos() {
			isCaseVar = true
			r.check(objsToUpdate, obj, to)
		}
	}

	r.checkInLexicalScope(objsToUpdate, from, to, info)

	// Finally, if this was a type switch, change the variable y.
	if isCaseVar {
		_, path, _ := r.iprog.PathEnclosingInterval(from.Pos(), from.Pos())
		path[0].(*ast.Ident).Name = to // path is [Ident AssignStmt TypeSwitchStmt...]
	}
}

// checkInLexicalScope performs safety checks that a renaming does not
// change the lexical reference structure of the specified package.
//
// For objects in lexical scope, there are three kinds of conflicts:
// same-, sub-, and super-block conflicts.  We will illustrate all three
// using this example:
//
//	var x int
//	var z int
//
//	func f(y int) {
//		print(x)
//		print(y)
//	}
//
// Renaming x to z encounters a SAME-BLOCK CONFLICT, because an object
// with the new name already exists, defined in the same lexical block
// as the old object.
//
// Renaming x to y encounters a SUB-BLOCK CONFLICT, because there exists
// a reference to x from within (what would become) a hole in its scope.
// The definition of y in an (inner) sub-block would cast a shadow in
// the scope of the renamed variable.
//
// Renaming y to x encounters a SUPER-BLOCK CONFLICT.  This is the
// converse situation: there is an existing definition of the new name
// (x) in an (enclosing) super-block, and the renaming would create a
// hole in its scope, within which there exist references to it.  The
// new name casts a shadow in scope of the existing definition of x in
// the super-block.
//
// Removing the old name (and all references to it) is always safe, and
// requires no checks.
//
func (r *Unexporter) checkInLexicalScope(objsToUpdate map[types.Object]string, from types.Object, to string, info *loader.PackageInfo) {
	lexinfo := r.lexInfo(info)

	b := lexinfo.Defs[from] // the block defining the 'from' object
	if b != nil {
		to, toBlock := b.Lookup(to)
		if toBlock == b {
			// same-block conflict
			r.warn(from,
				r.errorf(from.Pos(), "renaming this %s %q to %q",
					objectKind(from), from.Name(), to),
				r.errorf(to.Pos(), "\tconflicts with %s in same block",
					objectKind(to)))
			return
		} else if toBlock != nil {
			// Check for super-block conflict.
			// The name to is defined in a superblock.
			// Is that name referenced from within this block?
			for _, ref := range lexinfo.Refs[to] {
				if obj, _ := ref.Env.Lookup(from.Name()); obj == from {
					// super-block conflict
					r.warn(from,
						r.errorf(from.Pos(), "renaming this %s %q to %q",
							objectKind(from), from.Name(), to),
						r.errorf(ref.Id.Pos(), "\twould shadow this reference"),
						r.errorf(to.Pos(), "\tto the %s declared here",
							objectKind(to)))
					return
				}
			}
		}
	}

	// Check for sub-block conflict.
	// Is there an intervening definition of to between
	// the block defining 'from' and some reference to it?
	for _, ref := range lexinfo.Refs[from] {
		// TODO(adonovan): think about dot imports.
		// (Is b == fromBlock an invariant?)
		_, fromBlock := ref.Env.Lookup(from.Name())
		fromDepth := fromBlock.Depth()

		to, toBlock := ref.Env.Lookup(to)
		if to != nil {
			// sub-block conflict
			if toBlock.Depth() > fromDepth {
				r.warn(from,
					r.errorf(from.Pos(), "renaming this %s %q to %q",
						objectKind(from), from.Name(), to),
					r.errorf(ref.Id.Pos(), "\twould cause this reference to become shadowed"),
					r.errorf(to.Pos(), "\tby this intervening %s definition",
						objectKind(to)))
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
					r.check(objsToUpdate, field, to)
				}
			}
		}
	}
}

func (r *Unexporter) checkLabel(label *types.Label, to string) {
	// Check there are no identical labels in the function's label block.
	// (Label blocks don't nest, so this is easy.)
	if prev := label.Parent().Lookup(to); prev != nil {
		r.warn(label,
			r.errorf(label.Pos(), "renaming this label %q to %q", label.Name(), prev.Name()),
			r.errorf(prev.Pos(), "\twould conflict with this one"))
	}
}

// checkStructField checks that the field renaming will not cause
// conflicts at its declaration, or ambiguity or changes to any selection.
func (r *Unexporter) checkStructField(objsToUpdate map[types.Object]string, from *types.Var, to string) {
	// Check that the struct declaration is free of field conflicts,
	// and field/method conflicts.

	// go/types offers no easy way to get from a field (or interface
	// method) to its declaring struct (or interface), so we must
	// ascend the AST.
	info, path, _ := r.iprog.PathEnclosingInterval(from.Pos(), from.Pos())
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
	tStruct := path[i].(*ast.StructType)
	i++
	// Ascend past parens (unlikely).
	for {
		_, ok := path[i].(*ast.ParenExpr)
		if !ok {
			break
		}
		i++
	}
	if spec, ok := path[i].(*ast.TypeSpec); ok {
		// This struct is also a named type.
		// We must check for direct (non-promoted) field/field
		// and method/field conflicts.
		named := info.Defs[spec.Name].Type()
		prev, indices, _ := types.LookupFieldOrMethod(named, true, info.Pkg, to)
		if len(indices) == 1 {
			r.warn(from,
				r.errorf(from.Pos(), "renaming this field %q to %q",
					from.Name(), to),
				r.errorf(prev.Pos(), "\twould conflict with this %s",
					objectKind(prev)))
			return // skip checkSelections to avoid redundant errors
		}
	} else {
		// This struct is not a named type.
		// We need only check for direct (non-promoted) field/field conflicts.
		t := info.Types[tStruct].Type.Underlying().(*types.Struct)
		for i := 0; i < t.NumFields(); i++ {
			if prev := t.Field(i); prev.Name() == to {
				r.warn(from,
					r.errorf(from.Pos(), "renaming this field %q to %q",
						from.Name(), to),
					r.errorf(prev.Pos(), "\twould conflict with this field"))
				return // skip checkSelections to avoid redundant errors
			}
		}
	}

	// Renaming an anonymous field requires renaming the type too. e.g.
	// 	print(s.T)       // if we rename T to U,
	// 	type T int       // this and
	// 	var s struct {T} // this must change too.
	if from.Anonymous() {
		if named, ok := from.Type().(*types.Named); ok {
			r.check(objsToUpdate, named.Obj(), to)
		} else if named, ok := deref(from.Type()).(*types.Named); ok {
			r.check(objsToUpdate, named.Obj(), to)
		}
	}

	// Check integrity of existing (field and method) selections.
	r.checkSelections(objsToUpdate, from, to)
}

// checkSelection checks that all uses and selections that resolve to
// the specified object would continue to do so after the renaming.
func (r *Unexporter) checkSelections(objsToUpdate map[types.Object]string, from types.Object, to string) {
	for pkg, info := range r.packages {
		if id := someUse(info, from); id != nil {
			if !r.checkExport(id, pkg, from, to) {
				return
			}
		}

		for syntax, sel := range info.Selections {
			// There may be extant selections of only the old
			// name or only the new name, so we must check both.
			// (If neither, the renaming is sound.)
			//
			// In both cases, we wish to compare the lengths
			// of the implicit field path (Selection.Index)
			// to see if the renaming would change it.
			//
			// If a selection that resolves to 'from', when renamed,
			// would yield a path of the same or shorter length,
			// this indicates ambiguity or a changed referent,
			// analogous to same- or sub-block lexical conflict.
			//
			// If a selection using the name 'to' would
			// yield a path of the same or shorter length,
			// this indicates ambiguity or shadowing,
			// analogous to same- or super-block lexical conflict.

			// TODO(adonovan): fix: derive from Types[syntax.X].Mode
			// TODO(adonovan): test with pointer, value, addressable value.
			isAddressable := true

			if sel.Obj() == from {
				if obj, indices, _ := types.LookupFieldOrMethod(sel.Recv(), isAddressable, from.Pkg(), to); obj != nil {
					// Renaming this existing selection of
					// 'from' may block access to an existing
					// type member named 'to'.
					delta := len(indices) - len(sel.Index())
					if delta > 0 {
						continue // no ambiguity
					}
					r.selectionConflict(objsToUpdate, from, to, delta, syntax, obj)
					return
				}

			} else if sel.Obj().Name() == to {
				if obj, indices, _ := types.LookupFieldOrMethod(sel.Recv(), isAddressable, from.Pkg(), from.Name()); obj == from {
					// Renaming 'from' may cause this existing
					// selection of the name 'to' to change
					// its meaning.
					delta := len(indices) - len(sel.Index())
					if delta > 0 {
						continue //  no ambiguity
					}
					r.selectionConflict(objsToUpdate, from, to, -delta, syntax, sel.Obj())
					return
				}
			}
		}
	}
}

func (r *Unexporter) selectionConflict(objsToUpdate map[types.Object]string, from types.Object, to string, delta int, syntax *ast.SelectorExpr, obj types.Object) {
	rename := r.errorf(from.Pos(), "renaming this %s %q to %q",
		objectKind(from), from.Name(), to)

	switch {
	case delta < 0:
		// analogous to sub-block conflict
		r.warn(from, rename,
			r.errorf(syntax.Sel.Pos(),
				"\twould change the referent of this selection"),
			r.errorf(obj.Pos(), "\tof this %s", objectKind(obj)))
	case delta == 0:
		// analogous to same-block conflict
		r.warn(from, rename,
			r.errorf(syntax.Sel.Pos(),
				"\twould make this reference ambiguous"),
			r.errorf(obj.Pos(), "\twith this %s", objectKind(obj)))
	case delta > 0:
		// analogous to super-block conflict
		r.warn(from, rename,
			r.errorf(syntax.Sel.Pos(),
				"\twould shadow this selection"),
			r.errorf(obj.Pos(), "\tof the %s declared here",
				objectKind(obj)))
	}
}

// checkMethod performs safety checks for renaming a method.
// There are three hazards:
// - declaration conflicts
// - selection ambiguity/changes
// - entailed renamings of assignable concrete/interface types.
//   We reject renamings initiated at concrete methods if it would
//   change the assignability relation.  For renamings of abstract
//   methods, we rename all methods transitively coupled to it via
//   assignability.
func (r *Unexporter) checkMethod(objsToUpdate map[types.Object]string, from *types.Func, to string) {
	// e.g. error.Error
	if from.Pkg() == nil {
		r.warn(from, r.errorf(from.Pos(), "you cannot rename built-in method %s", from))
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
			r.warn(from,
				r.errorf(from.Pos(), "renaming this interface method %q to %q",
					from.Name(), to),
				r.errorf(prev.Pos(), "\twould conflict with this method"))
			return
		}

		// Check all interfaces that embed this one for
		// declaration conflicts too.
		for _, info := range r.packages {
			// Start with named interface types (better errors)
			for _, obj := range info.Defs {
				if obj, ok := obj.(*types.TypeName); ok && isInterface(obj.Type()) {
					f, _, _ := types.LookupFieldOrMethod(
						obj.Type(), false, from.Pkg(), from.Name())
					if f == nil {
						continue
					}
					t, _, _ := types.LookupFieldOrMethod(
						obj.Type(), false, from.Pkg(), to)
					if t == nil {
						continue
					}
					r.warn(from,
						r.errorf(from.Pos(), "renaming this interface method %q to %q",
							from.Name(), to),
						r.errorf(t.Pos(), "\twould conflict with this method"),
						r.errorf(obj.Pos(), "\tin named interface type %q", obj.Name()))
				}
			}

			// Now look at all literal interface types (includes named ones again).
			for e, tv := range info.Types {
				if e, ok := e.(*ast.InterfaceType); ok {
					_ = e
					_ = tv.Type.(*types.Interface)
					// TODO(adonovan): implement same check as above.
				}
			}
		}

		// assignability
		//
		// Find the set of concrete or abstract methods directly
		// coupled to abstract method 'from' by some
		// satisfy.Constraint, and rename them too.
		for key := range r.satisfy() {
			// key = (lhs, rhs) where lhs is always an interface.

			lsel := r.msets.MethodSet(key.LHS).Lookup(from.Pkg(), from.Name())
			if lsel == nil {
				continue
			}
			rmethods := r.msets.MethodSet(key.RHS)
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
					rto := rtosel.Obj().(*types.Func)
					delta := len(rsel.Index()) - len(rtosel.Index())
					if delta < 0 {
						continue // no ambiguity
					}

					// TODO(adonovan): record the constraint's position.
					keyPos := token.NoPos

					rename := r.errorf(from.Pos(), "renaming this method %q to %q",
						from.Name(), to)
					if delta == 0 {
						// analogous to same-block conflict
						r.warn(from, rename,
							r.errorf(keyPos, "\twould make the %s method of %s invoked via interface %s ambiguous",
								to, key.RHS, key.LHS),
							r.errorf(rto.Pos(), "\twith (%s).%s",
								recv(rto).Type(), to))
					} else {
						// analogous to super-block conflict
						r.warn(from, rename,
							r.errorf(keyPos, "\twould change the %s method of %s invoked via interface %s",
								to, key.RHS, key.LHS),
							r.errorf(coupled.Pos(), "\tfrom (%s).%s",
								recv(coupled).Type(), to),
							r.errorf(rto.Pos(), "\tto (%s).%s",
								recv(rto).Type(), to))
					}
					return // one error is enough
				}
			}

			if !r.changeMethods {
				// This should be unreachable.
				r.warn(from,
					r.errorf(from.Pos(), "internal error: during renaming of abstract method %s", from),
					r.errorf(coupled.Pos(), "\tchangedMethods=false, coupled method=%s", coupled),
					r.errorf(from.Pos(), "\tPlease file a bug report"))
				return
			}

			// Rename the coupled method to preserve assignability.
			r.check(objsToUpdate, coupled, to)
		}
	} else {
		// Concrete method

		// declaration
		prev, indices, _ := types.LookupFieldOrMethod(R, true, from.Pkg(), to)
		if prev != nil && len(indices) == 1 {
			r.warn(from,
				r.errorf(from.Pos(), "renaming this method %q to %q",
					from.Name(), to),
				r.errorf(prev.Pos(), "\twould conflict with this %s",
					objectKind(prev)))
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
		for key := range r.satisfy() {
			// key = (lhs, rhs) where lhs is always an interface.
			if isInterface(key.RHS) {
				continue
			}
			rsel := r.msets.MethodSet(key.RHS).Lookup(from.Pkg(), from.Name())
			if rsel == nil || rsel.Obj() != from {
				continue // rhs does not have the method
			}
			lsel := r.msets.MethodSet(key.LHS).Lookup(from.Pkg(), from.Name())
			if lsel == nil {
				continue
			}
			imeth := lsel.Obj().(*types.Func)

			// imeth is the abstract method (e.g. I.f)
			// and key.RHS is the concrete coupling type (e.g. D).
			if !r.changeMethods {
				rename := r.errorf(from.Pos(), "renaming this method %q to %q",
					from.Name(), to)
				var pos token.Pos
				var iface string

				i := recv(imeth).Type()
				if named, ok := i.(*types.Named); ok {
					pos = named.Obj().Pos()
					iface = "interface " + named.Obj().Name()
				} else {
					pos = from.Pos()
					iface = i.String()
				}
				r.warn(from, rename,
					r.errorf(pos, "\twould make %s no longer assignable to %s",
						key.RHS, iface),
					r.errorf(imeth.Pos(), "\t(rename %s.%s if you intend to change both types)",
						i, from.Name()))
				return // one error is enough
			}

			// Rename the coupled interface method to preserve assignability.
			r.check(objsToUpdate, imeth, to)
		}
	}

	// Check integrity of existing (field and method) selections.
	// We skip this if there were errors above, to avoid redundant errors.
	r.checkSelections(objsToUpdate, from, to)
}

// XXX this is always true for the use case of this libary
func (r *Unexporter) checkExport(id *ast.Ident, pkg *types.Package, from types.Object, to string) bool {
	// Reject cross-package references if to is unexported.
	// (Such references may be qualified identifiers or field/method
	// selections.)
	if !ast.IsExported(to) && pkg != from.Pkg() {
		r.warn(from,
			r.errorf(from.Pos(),
				"renaming this %s %q to %q would make it unexported",
				objectKind(from), from.Name(), to),
			r.errorf(id.Pos(), "\tbreaking references from packages such as %q",
				pkg.Path()))
		return false
	}
	return true
}

// satisfy returns the set of interface satisfaction constraints.
func (r *Unexporter) satisfy() map[satisfy.Constraint]bool {
	if r.satisfyConstraints == nil {
		// Compute on demand: it's expensive.
		var f satisfy.Finder
		for _, info := range r.packages {
			f.Find(&info.Info, info.Files)
		}
		r.satisfyConstraints = f.Result
	}
	return r.satisfyConstraints
}

// -- helpers ----------------------------------------------------------

// recv returns the method's receiver.
func recv(meth *types.Func) *types.Var {
	return meth.Type().(*types.Signature).Recv()
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

// -- Plundered from golang.org/x/tools/go/ssa -----------------

func isInterface(t types.Type) bool { return types.IsInterface(t) }

func deref(typ types.Type) types.Type {
	if p, _ := typ.(*types.Pointer); p != nil {
		return p.Elem()
	}
	return typ
}

func isPackageLevel(obj types.Object) bool {
	return obj.Pkg().Scope().Lookup(obj.Name()) == obj
}

// isLocal reports whether obj is local to some function.
// Precondition: not a struct field or interface method.
func isLocal(obj types.Object) bool {
	// [... 5=stmt 4=func 3=file 2=pkg 1=universe]
	var depth int
	for scope := obj.Parent(); scope != nil; scope = scope.Parent() {
		depth++
	}
	return depth >= 4
}

func objectKind(obj types.Object) string {
	switch obj := obj.(type) {
	case *types.PkgName:
		return "imported package name"
	case *types.TypeName:
		return "type"
	case *types.Var:
		if obj.IsField() {
			return "field"
		}
	case *types.Func:
		if obj.Type().(*types.Signature).Recv() != nil {
			return "method"
		}
	}
	// label, func, var, const
	return strings.ToLower(strings.TrimPrefix(reflect.TypeOf(obj).String(), "*types."))
}

// errorf reports an error (e.g. conflict) and prevents file modification.
func (r *Unexporter) errorf(pos token.Pos, format string, args ...interface{}) string {
	return fmt.Sprintf("%s: %s", r.iprog.Fset.Position(pos), fmt.Sprintf(format, args...))
}
func (r *Unexporter) warn(from types.Object, warnings ...string) {
	r.warnings <- map[types.Object]string{from: strings.Join(warnings, "\n")}
}

func (r *Unexporter) lexInfo(info *loader.PackageInfo) *lexical.Info {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if lexinfo := r.lexinfos[info]; lexinfo != nil {
		return lexinfo
	} else {
		lexinfo := lexical.Structure(r.iprog.Fset, info.Pkg, &info.Info, info.Files)
		r.lexinfos[info] = lexinfo
		return lexinfo
	}
}
