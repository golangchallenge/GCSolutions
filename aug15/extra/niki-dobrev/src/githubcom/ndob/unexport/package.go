package main

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/loader"
	"golang.org/x/tools/go/types"
)

const (
	internal = "internal"
)

type identifier struct {
	pkg    string
	parent string
	name   string
	pos    token.Position
}

// String prints out the identifier in gorename-compatible format.
func (i *identifier) string() string {
	if i.parent != "" {
		return fmt.Sprintf("\"%s\".%s.%s", i.pkg, i.parent, i.name)
	}
	return fmt.Sprintf("\"%s\".%s", i.pkg, i.name)
}

func (i *identifier) isExported() bool {
	return ast.IsExported(i.name)
}

func (i *identifier) unexportedName() string {
	if len(i.name) == 0 {
		return ""
	}
	return strings.ToLower(string(i.name[0])) + i.name[1:]
}

// NameEquals returns true if name part of identifiers are equal.
// pos is ignored.
func (i *identifier) nameEquals(other *identifier) bool {
	return i.name == other.name && i.parent == other.parent && i.pkg == other.pkg
}

// searchByPos tries to find a identifier that resides in provided position.
// Uses "comma ok"-idiom.
func searchByPos(idarr []identifier, pos token.Position) (identifier, bool) {
	var ret identifier

	for _, id := range idarr {
		if id.pos == pos {
			return id, true
		}
	}
	return ret, false
}

// isInternal tells whether a package is internal (ie. if package itself
// is internal or it has internal ancestors).
func isInternal(pkgName string) bool {
	cleanedPath := filepath.Clean(pkgName)
	explodedPath := strings.Split(cleanedPath, string(os.PathSeparator))

	for _, pathPart := range explodedPath {
		if pathPart == internal {
			return true
		}
	}
	return false
}

// getInternalParentPrefix returns package name prefix until the last
// internal package.
// example: "/a/b/internal/a/c/internal/x/y/z" -> "/a/b/internal/a/c"
func getInternalParentPrefix(pkgName string) string {
	cleanedPath := filepath.Clean(pkgName)
	explodedPath := strings.Split(cleanedPath, string(os.PathSeparator))

	i := len(explodedPath) - 1
	for ; i > 0; i-- {
		if explodedPath[i] == internal {
			break
		}
	}
	return filepath.Join(explodedPath[0:i]...)
}

// packagesToSearch returns a list of package names, that should be
// searched for usages of identifiers usages from unexported package.
// pkgPrefix can narrow down the search for a certain package and it's
// subpackages.
// By default (empty pkgPrefix) all packages under build.Default.SrcDirs()
// are included in the search.
func packagesToSearch(unexportPkg string, pkgPrefix string) []string {
	var searchPackages []string
	var allPackages []string
	var internalPrefix string

	for _, dir := range build.Default.SrcDirs() {
		allPackages = append(allPackages, getSubDirectories(dir)...)
	}

	// If package is internal, the search can be narrowed down:
	// Only immediate parent (and their children) can import internal
	// packages.
	if isInternal(unexportPkg) {
		internalPrefix = getInternalParentPrefix(unexportPkg)
	}

	// Filter with prefixes.
	for _, p := range allPackages {
		if p == unexportPkg || !strings.HasPrefix(p, internalPrefix) || !strings.HasPrefix(p, pkgPrefix) {
			continue
		}
		searchPackages = append(searchPackages, p)
	}
	return searchPackages
}

// createProgram returns a program containing packages specified.
func createProgram(packages ...string) (*loader.Program, error) {
	var conf loader.Config

	for _, name := range packages {
		conf.CreateFromFilenames(name, getFileNames(name)...)
	}
	return conf.Load()
}

// findIdentifiers returns a list of all identifiers (both exported and local)
// found in provided package.
func findIdentifiers(pkgName string) []identifier {
	var ret []identifier

	addID := func(newID identifier) {
		for _, id := range ret {
			// Compare names only, because if type embedding is used
			// the same symbols is added multiple times.
			if id.nameEquals(&newID) {
				return
			}
		}
		ret = append(ret, newID)
	}

	prog, err := createProgram(pkgName)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return ret
	}

	// Traverse definitions and convert them to identifiers.
	for dident, dobj := range prog.Package(pkgName).Defs {

		if dident.Obj == nil {
			// This branch handles receivers.
			var parent string
			if t, ok := dobj.(*types.Func); ok {
				if t2, ok := t.Type().(*types.Signature); ok && t2.Recv() != nil {
					switch t3 := t2.Recv().Type().(type) {
					case *types.Pointer:
						if t4, ok := t3.Elem().(*types.Named); ok {
							parent = t4.Obj().Name()
						}
					case *types.Named:
						parent = t3.Obj().Name()
					}

				}
			}
			addID(identifier{pkgName, parent, dident.Name, prog.Fset.Position(dident.Pos())})

		} else if typeSpec, ok := dident.Obj.Decl.(*ast.TypeSpec); ok {
			/// This branch handles structs and interfaces and their fields.
			addID(identifier{pkgName, "", dident.Name, prog.Fset.Position(dident.Pos())})

			var fl *ast.FieldList
			switch t := typeSpec.Type.(type) {
			case *ast.StructType:
				fl = t.Fields
			case *ast.InterfaceType:
				fl = t.Methods
			}

			// Handle field names.
			if fl != nil {
				for _, v := range fl.List {
					for _, n := range v.Names {
						addID(identifier{pkgName, dident.Name, n.Name, prog.Fset.Position(n.Pos())})
					}
				}
			}

		} else if _, ok := dident.Obj.Decl.(*ast.Field); ok {
			// This branch handles Fields.
			// Fields get added when handling structs, interfaces and FieldLists (previous branch),
			// so no need to export them here.
			continue

		} else {
			// This branch handles all other top level declarations (vars, consts)
			addID(identifier{pkgName, "", dident.Name, prog.Fset.Position(dident.Pos())})

		}
	}
	return ret
}

// findExports returns a list of exported identifiers found in provided package.
func findExports(pkgName string) []identifier {
	var ret []identifier

	ids := findIdentifiers(pkgName)
	for _, id := range ids {
		if id.isExported() {
			ret = append(ret, id)
		}
	}
	return ret
}

// usageCounts returns a list of usages of "pkgName"-package in "searchUsageFrom"-package
// per exported identifier. Each usage-element contains call-site's exact position in a file.
// targetExports is a list of exported identifiers in "pkgName"-package.
func usageCounts(pkgName string, targetExports []identifier, searchUsageFrom string) map[identifier][]string {
	ret := make(map[identifier][]string)

	prog, err := createProgram(pkgName, searchUsageFrom)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return ret
	}

	targetInfo := prog.Package(pkgName)
	fromInfo := prog.Package(searchUsageFrom)

	for uident, uobj := range fromInfo.Uses {
		if uobj.Pkg() != nil && uobj.Pkg().Path() == pkgName {
			pos := prog.Fset.Position(uident.Pos())
			lineCol := fmt.Sprintf("%s:%d:%d", pos.Filename, pos.Line, pos.Column)

			for dident, dobj := range targetInfo.Defs {
				if uobj == dobj {
					if id, ok := searchByPos(targetExports, prog.Fset.Position(dident.Pos())); ok {
						ret[id] = append(ret[id], lineCol)
					}
				}
			}
		}
	}
	return ret
}

// unusedExports returns a list of unused package-exports for pkgName.
// Export usage is searched from provided "searchPackages"-packages.
func unusedExports(pkgName string, searchPackages []string) []identifier {
	var ret []identifier
	used := make(map[identifier]bool)

	exports := findExports(pkgName)
	for _, s := range searchPackages {
		fmt.Fprintln(os.Stdout, "searching usages from:", s)

		u := usageCounts(pkgName, exports, s)
		for id := range u {
			used[id] = true
		}
	}

	for _, id := range exports {
		if !used[id] {
			ret = append(ret, id)
		}
	}
	return ret
}

// nameCollisions returns a list of identifiers, which will cause a naming collision
// if exported.
func nameCollisions(pkgName string, candidates []identifier) []identifier {
	var collisions []identifier
	unexportCollision := func(a *identifier, b *identifier) bool {
		return a.pkg == b.pkg && a.parent == b.parent && a.unexportedName() == b.name
	}

	allIds := findIdentifiers(pkgName)
	for _, cid := range candidates {
		for _, id := range allIds {
			if unexportCollision(&cid, &id) {
				collisions = append(collisions, cid)
				break
			}
		}
	}
	return collisions
}

// usageReport returns a list of users (file+location)
// for each exported identifier in provided package.
func usageReport(pkgName string, searchFrom string) map[identifier][]string {
	ret := make(map[identifier][]string)

	exports := findExports(pkgName)
	searchPackages := packagesToSearch(pkgName, searchFrom)
	for _, s := range searchPackages {
		newuc := usageCounts(pkgName, exports, s)
		for id, str := range newuc {
			ret[id] = append(ret[id], str...)
		}
	}
	return ret
}
