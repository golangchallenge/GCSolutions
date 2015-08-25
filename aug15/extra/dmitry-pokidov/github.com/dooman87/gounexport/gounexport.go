//Package gounexport provides functionality to unexport unused public symbols.
//
//In detail, what you can do using this package:
//
//* parse package and get package's definition
//
//* get information about all definitions in packages
//
//* get unused packages
//
//* unexport unused definitions
//
//For result Definition struct is used. It's also includes Definition.Usage array
//with all usages (internal and external) across the package.
//
package gounexport

import (
	"fmt"
	"regexp"
	"strings"

	"go/ast"
	"go/token"
	"go/types"
	"reflect"

	"github.com/dooman87/gounexport/fs"
	"github.com/dooman87/gounexport/importer"
	"github.com/dooman87/gounexport/util"
)

//Definition of symbol in package
type Definition struct {
	//Full file path for current defintion
	File string
	//Full name of the definition
	Name string
	//Simple name of the definition. We are using it
	//for renaming purposes.
	SimpleName string
	//List of interfaces that implemented current definition
	//It will be interfaces definitions for type and methods definition
	//for functions
	InterfacesDefs []*Definition
	//type of definition
	TypeOf reflect.Type
	//Number of line in the file where definition is declared
	Line int
	//Column in the file where definition is declared
	Col int
	//Offset in source file
	Offset int
	//True, if definition is exported
	IsExported bool
	//Package of the definition
	Pkg *types.Package
	//List of usages of the definition
	Usages []*Usage
}

func (def *Definition) addUsage(pos token.Position) {
	u := new(Usage)
	u.Pos = pos
	def.Usages = append(def.Usages, u)
}

//Usage is a struct that define one usage of a definition
type Usage struct {
	//Pos is a position of usage: file, line, col
	Pos token.Position
}

type objectWithIdent struct {
	obj   types.Object
	ident *ast.Ident
}

func newObjectWithIdent(obj types.Object, ident *ast.Ident) *objectWithIdent {
	result := new(objectWithIdent)
	result.obj = obj
	result.ident = ident
	return result
}

type defWithInterface struct {
	def      *Definition
	interfac *types.Interface
	named    *types.Named
}

type getDefinitionsContext struct {
	structs    map[string]string
	vars       []*objectWithIdent
	funcs      []*objectWithIdent
	interfaces []*defWithInterface
	fset       *token.FileSet
	defs       map[string]*Definition
}

func newGetDefinitionsContext(fset *token.FileSet) *getDefinitionsContext {
	ctx := new(getDefinitionsContext)
	ctx.fset = fset
	ctx.structs = make(map[string]string, 0)
	ctx.interfaces = make([]*defWithInterface, 0)
	ctx.vars = make([]*objectWithIdent, 0)
	ctx.funcs = make([]*objectWithIdent, 0)
	return ctx
}

//GetDefinitions collects information about all (exported and unexported)
//definitions and adapt them to Definition structure.
//Returns map where key is full name (package + name) of symbol.
func GetDefinitions(info *types.Info, fset *token.FileSet) map[string]*Definition {
	ctx := newGetDefinitionsContext(fset)
	ctx.defs = make(map[string]*Definition, 0)

	processTypes(info, ctx)
	processDefs(info, ctx)
	processUses(info, ctx)

	return ctx.defs
}

//processTypes is only filling interfaces from function signatures.
func processTypes(info *types.Info, ctx *getDefinitionsContext) {
	for _, t := range info.Types {
		logType(t)

		if t.Type != nil {
			switch t.Type.(type) {
			//If it's a function call then extracting
			//all params and trying to find interfaces.
			//We are doing this to find usages of interfaces
			//cause methods could be called only inside internal functions
			case *types.Signature:
				s := t.Type.(*types.Signature)
				if tuple := s.Params(); tuple != nil {
					for i := 0; i < tuple.Len(); i++ {
						v := tuple.At(i)
						if types.IsInterface(v.Type()) {
							addInterface(v, nil, ctx)
						}
					}
				}
			}
		}
	}
}

//processDefs going through all definitions in the next order:
// - collect info about all interfaces
// - process everuthing except vars and functions to collect all structs prior vars and functions
// - process vars and functions
func processDefs(info *types.Info, ctx *getDefinitionsContext) {
	//Collect all interfaces
	for ident, obj := range info.Defs {
		if !isValidObject(obj, ident, ctx) || !types.IsInterface(obj.Type()) {
			continue
		}
		addInterface(obj, ident, ctx)
	}

	logInterfaces(ctx)

	//Collect everything except vars and functions
	for ident, obj := range info.Defs {
		if !isValidObject(obj, ident, ctx) {
			continue
		}
		var def *Definition
		if !isVar(obj) && !isFunc(obj) && !types.IsInterface(obj.Type()) {
			def = createDef(obj, ident, ctx, false)
		}
		updateGetDefinitionsContext(ctx, def, ident, obj)
	}

	for _, v := range ctx.vars {
		createDef(v.obj, v.ident, ctx, false)
	}

	for _, v := range ctx.funcs {
		createDef(v.obj, v.ident, ctx, false)
	}
}

//Fill usages inside all definitions. The special case is functions
//where all params should be processed.
func processUses(info *types.Info, ctx *getDefinitionsContext) {
	for ident, obj := range info.Uses {
		useName := getFullName(obj, ctx, false)
		if ctx.defs[useName] != nil {
			ctx.defs[useName].addUsage(ctx.fset.Position(ident.Pos()))
		} else {
			util.Warn("can't find usage for [%s] %s\n\tObject definition - %s", useName, posToStr(ctx.fset, ident.Pos()), posToStr(ctx.fset, obj.Pos()))
		}
		switch obj.Type().(type) {
		case *types.Signature:
			s := obj.Type().(*types.Signature)
			if tuple := s.Params(); tuple != nil {
				for i := 0; i < tuple.Len(); i++ {
					v := tuple.At(i)
					useName := getFullName(v, ctx, true)
					if ctx.defs[useName] != nil {
						ctx.defs[useName].addUsage(ctx.fset.Position(ident.Pos()))
					}
				}
			}
		}
	}
}

func addInterface(obj types.Object, ident *ast.Ident, ctx *getDefinitionsContext) {
	interfac := obj.Type().Underlying().(*types.Interface)

	def := createDef(obj, ident, ctx, true)
	updateGetDefinitionsContext(ctx, def, ident, obj)

	util.Debug("adding interface [%s] [%v] [%v] [%v]", def.Name, def.Pkg, obj.Type().Underlying(), obj.Type())
	//Adding all methods of interface
	for i := 0; i < interfac.NumMethods(); i++ {
		f := interfac.Method(i)
		def := createDef(f, nil, ctx, false)
		util.Debug("\tadding method [%v] [%s]", f, def.Name)
		updateGetDefinitionsContext(ctx, def, ident, f)
	}
}

func isValidObject(obj types.Object, ident *ast.Ident, ctx *getDefinitionsContext) bool {
	if obj == nil {
		return false
	}
	position := ctx.fset.Position(ident.Pos())
	typeOf := reflect.TypeOf(obj)
	fullName := getFullName(obj, ctx, false)
	if !position.IsValid() {
		util.Warn("position object is invalid for %s", ident.Name)
		return false
	}
	if len(fullName) == 0 {
		util.Warn("warning: cann't get full name for %s: %v", ident.Name, typeOf)
		return false
	}
	if typeOf == nil {
		return false
	}
	if ctx.defs[fullName] != nil {
		return false
	}
	return true
}

func updateGetDefinitionsContext(ctx *getDefinitionsContext, def *Definition, ident *ast.Ident, obj types.Object) {
	switch obj.(type) {
	case *types.Var:
		//Processing vars later to be sure that all info about structs already filled
		ctx.vars = append(ctx.vars, newObjectWithIdent(obj, ident))
	case *types.Func:
		//Processing funcs later to be sure that all info about interfaces already filled
		ctx.funcs = append(ctx.funcs, newObjectWithIdent(obj, ident))
	case *types.TypeName:
		//If the underlying type is struct, then filling
		//positions of struct's fields (key) and struct name(value)
		//to map. Then we can extract struct name for fields when
		//will be analyze them.
		t := obj.(*types.TypeName)
		underlyingType := t.Type().Underlying()
		switch underlyingType.(type) {
		case *types.Struct:
			s := underlyingType.(*types.Struct)
			for i := 0; i < s.NumFields(); i++ {
				field := s.Field(i)
				ctx.structs[posToStr(ctx.fset, field.Pos())] = obj.Name()
			}
		}
	}

	//Check for interfaces
	underlyingType := obj.Type().Underlying()
	switch underlyingType.(type) {
	case *types.Interface:
		d := new(defWithInterface)
		d.def = def
		d.interfac = underlyingType.(*types.Interface)
		ctx.interfaces = append(ctx.interfaces, d)
	}
}

func isVar(obj types.Object) bool {
	switch obj.(type) {
	case *types.Var:
		return true
	}
	return false
}

func isFunc(obj types.Object) bool {
	switch obj.(type) {
	case *types.Func:
		return true
	}
	return false
}

func createDef(obj types.Object, ident *ast.Ident, ctx *getDefinitionsContext, isType bool) *Definition {
	fullName := getFullName(obj, ctx, isType)

	if def, ok := ctx.defs[fullName]; ok {
		return def
	}

	def := new(Definition)
	def.Name = fullName
	def.Pkg = obj.Pkg()
	def.IsExported = obj.Exported()
	def.TypeOf = reflect.TypeOf(obj)
	def.SimpleName = obj.Name()
	def.Usages = make([]*Usage, 0)
	def.InterfacesDefs = make([]*Definition, 0)

	if ident != nil {
		position := ctx.fset.Position(ident.Pos())
		def.File = position.Filename
		def.Line = position.Line
		def.Offset = position.Offset
		def.Col = position.Column
	}

	if !types.IsInterface(obj.Type()) {
		fillInterfaces(def, obj, ctx)
	}

	ctx.defs[def.Name] = def
	logDefinition(def, obj, ident, ctx)

	return def
}

func fillInterfaces(def *Definition, obj types.Object, ctx *getDefinitionsContext) {
	switch obj.(type) {
	case *types.TypeName:
		//Filling information about implemented
		//interfaces to type's definition.
		typ := obj.(*types.TypeName)
		if typ.Type() != nil {
			for _, di := range ctx.interfaces {
				if di.interfac != nil && di.def != nil && implements(typ.Type(), di.interfac, typ.Pkg()) {
					def.InterfacesDefs = append(def.InterfacesDefs, di.def)
				}
			}
		}
	case *types.Func:
		f := obj.(*types.Func)
		underlyingType := f.Type().Underlying()
		switch underlyingType.(type) {
		case *types.Signature:
			s := underlyingType.(*types.Signature)
			if s.Recv() != nil {
				//Getting all interfaces from function's receiver and
				//searching for current function in each interface.
				//If found, then adding method's definition to function's
				//interfaces
				recvTypeName := strings.Replace(s.Recv().Type().String(), "*", "", 1)
				if typeDef, ok := ctx.defs[recvTypeName]; ok {
					for _, iDef := range typeDef.InterfacesDefs {
						def.InterfacesDefs = append(def.InterfacesDefs, iDef)
						if methodDef := lookupMethod(def, iDef, ctx); methodDef != nil {
							def.InterfacesDefs = append(def.InterfacesDefs, methodDef)
						}
					}
				} else {
					util.Debug("recv type not found [%s]", s.Recv().Type().String())
				}
			}
		}
	}
}

func lookupMethod(def *Definition, iDef *Definition, ctx *getDefinitionsContext) *Definition {
	methodName := iDef.Name + "." + def.SimpleName
	externalInterfaceMethodName := "interface." + def.SimpleName
	def.InterfacesDefs = append(def.InterfacesDefs, iDef)
	if methodDef := ctx.defs[methodName]; methodDef != nil {
		return methodDef
	} else if methodDef := ctx.defs[externalInterfaceMethodName]; methodDef != nil {
		return methodDef
	} else {
		util.Debug("can't find method [%s]", methodName)
	}
	return nil
}

func implements(t types.Type, interfac *types.Interface, pkg *types.Package) bool {
	if interfac == nil || t == nil || interfac.Empty() {
		return false
	}
	if types.Implements(t, interfac) {
		return true
	}
	//For some reason, interfaces that comes
	//already built in (not from sources) are
	//not working with types.Implements method
	for i := 0; i < interfac.NumMethods(); i++ {
		m := interfac.Method(i)
		obj, _, _ := types.LookupFieldOrMethod(t, true, pkg, m.Name())
		if obj == nil {
			util.Debug("method %s not found in type %v", m.Name(), t)
			return false
		}
	}
	return true
}

//GetDefenitionsToHide returns list of defenitions that could be
//moved to private e.g. renamed. Criteria for renaming:
// - Definition should be exported
// - Definition should be in target package
// - Definition is not implementing external intefaces then it will be ignored
// - Definition is not used in external packages
func GetDefenitionsToHide(pkg string, defs map[string]*Definition, excludes []*regexp.Regexp) []*Definition {
	var unused []*Definition
	for _, def := range defs {
		if !def.IsExported {
			continue
		}

		if strings.HasPrefix(def.Name, pkg) && !isExcluded(def, excludes) && !isUsed(def) {
			util.Info("adding [%s] to unexport list", def.Name)
			unused = append(unused, def)
		}
	}

	return unused
}

func isExcluded(def *Definition, excludes []*regexp.Regexp) bool {
	if excludes == nil || len(excludes) == 0 {
		return false
	}

	for _, exc := range excludes {
		if exc.MatchString(def.Name) {
			util.Info("definition [%s] excluded, because matched [%s]", def.Name, exc.String())
			return true
		}
	}
	return false
}

func isUsed(def *Definition) bool {
	used := true

	if len(def.Usages) == 0 {
		used = false
	} else {
		//Checking pathes of usages to not count internal
		hasExternalUsages := false
		util.Debug("checking [%s]", def.Name)
		for _, u := range def.Usages {
			pkgPath := ""
			if def.Pkg != nil {
				pkgPath = def.Pkg.Path()
			} else if dotIdx := strings.LastIndex(def.Name, "."); dotIdx >= 0 {
				pkgPath = def.Name[0:dotIdx]
			}
			util.Debug("checking [%v]", u.Pos)
			if u.Pos.IsValid() && fs.GetPackagePath(u.Pos.Filename) != pkgPath {
				hasExternalUsages = true
				break
			}
		}
		used = hasExternalUsages
	}

	if !used {
		//Check all interfaces
		for _, i := range def.InterfacesDefs {
			if isUsed(i) {
				used = true
				break
			}
		}
	}
	return used
}

func posToStr(fset *token.FileSet, pos token.Pos) string {
	fPos := fset.Position(pos)
	return fmt.Sprintf("[%s:%d:%d]", fs.GetRelativePath(fPos.Filename), fPos.Line, fPos.Column)
}

//ParsePackage parses package and filling info structure.
//It's will fill info about all internal packages if they
//are not used in the root package.
func ParsePackage(pkgName string, info *types.Info) (*types.Package, *token.FileSet, error) {
	collectImporter := new(importer.CollectInfoImporter)
	collectImporter.Info = info

	var resultPkg *types.Package
	var resultFset *token.FileSet
	parsedPackages := make(map[string]bool)

	notParsedPackage := pkgName
	for len(notParsedPackage) > 0 {
		collectImporter.Pkg = notParsedPackage
		pkg, fset, err := collectImporter.Collect()
		if err != nil {
			return nil, nil, err
		}

		//Filling results only from first package
		//that was passed as argument to function
		if resultPkg == nil {
			resultPkg = pkg
			resultFset = fset
		}
		parsedPackages[notParsedPackage] = true

		//Searching for a new package that was not parsed before
		notParsedPackage = ""
		files, err := fs.GetUnusedSources(pkgName, fset)
		if err != nil {
			return nil, nil, err
		}
		for _, f := range files {
			newNotParsedPackage := fs.GetPackagePath(f)
			if !parsedPackages[newNotParsedPackage] {
				notParsedPackage = newNotParsedPackage
				break
			} else {
				util.Info("package %s has been already parsed, however %s file is still unused", newNotParsedPackage, f)
			}
		}
	}

	return resultPkg, resultFset, nil
}

//Unexport is unexporting def by changing first letter
//to lower case. It won't rename if there is already existing
//unexported symbol with the same name.
//renameFunc is a func that accepts four arguments: full path to file,
//offset in a file to replace, original string, string to replace. It will
//be called when renaming is possible.
func Unexport(def *Definition, allDefs map[string]*Definition,
	renameFunc func(string, int, string, string) error) error {
	util.Info("unexporting %s in %s:%d:%d", def.SimpleName, def.File, def.Line, def.Col)
	newName := strings.ToLower(def.SimpleName[0:1]) + def.SimpleName[1:]

	//Searching for conflict
	lastIdx := strings.LastIndex(def.Name, def.SimpleName)
	newFullName := def.Name[0:lastIdx] + newName + def.Name[lastIdx+len(newName):]
	if allDefs[newFullName] != nil {
		return fmt.Errorf("can't unexport %s because it conflicts with existing member", def.Name)
	}

	//rename definitions and usages
	err := renameFunc(def.File, def.Offset, def.SimpleName, newName)
	for _, u := range def.Usages {
		if err != nil {
			break
		}
		err = renameFunc(u.Pos.Filename, u.Pos.Offset, def.SimpleName, newName)
	}

	return err
}

//getFullName is returning unique name of obj.
func getFullName(obj types.Object, ctx *getDefinitionsContext, isType bool) string {
	if obj == nil {
		return ""
	}
	if isType {
		return obj.Type().String()
	}

	result := ""

	switch obj.(type) {
	case *types.Func:
		f := obj.(*types.Func)
		r := strings.NewReplacer("(", "", "*", "", ")", "")
		result = r.Replace(f.FullName())
	default:
		if obj.Pkg() != nil {
			result += obj.Pkg().Path()
			result += "."
		}

		if packageName, ok := ctx.structs[posToStr(ctx.fset, obj.Pos())]; ok {
			result += packageName
			result += "."
		}
		result += obj.Name()
	}

	return result
}

func pos(fset *token.FileSet, pos token.Pos) (int, int) {
	fPos := fset.Position(pos)
	return fPos.Line, fPos.Column
}

func logDefinition(def *Definition, obj types.Object, ident *ast.Ident, ctx *getDefinitionsContext) {
	if ident == nil {
		return
	}
	util.Info("definition [%s] [%s], exported [%v], position %s", ident.Name, def.TypeOf.String(), obj.Exported(), posToStr(ctx.fset, ident.Pos()))
	switch obj.(type) {
	case *types.TypeName:
		t := obj.(*types.TypeName)
		underlyingType := t.Type().Underlying()
		util.Info("\t [%s] [%s] [%s]", t.Type().String(), t.Type().Underlying().String(), reflect.TypeOf(t.Type().Underlying()).String())
		switch underlyingType.(type) {
		case *types.Struct:
			s := underlyingType.(*types.Struct)
			util.Info("\t\t[%d] fields", s.NumFields())
			for i := 0; i < s.NumFields(); i++ {
				field := s.Field(i)
				util.Info("\t\t\t[%s]", posToStr(ctx.fset, field.Pos()))
			}
		}
	case *types.Func:
		f := obj.(*types.Func)
		underlyingType := f.Type().Underlying()
		util.Info("\t full name: [%s] [%s] [%s]", f.FullName(), underlyingType.String(), reflect.TypeOf(underlyingType))
	}

	util.Info("\tinterfaces [%d]", len(def.InterfacesDefs))
	for _, i := range def.InterfacesDefs {
		util.Info("\tinterface [%s]", i.Name)
	}
}

func logInterfaces(ctx *getDefinitionsContext) {
	for _, i := range ctx.interfaces {
		util.Info("interface [%s]", i.def.Name)
	}
}

func logType(t types.TypeAndValue) {
	if t.Type != nil {
		util.Debug("type [%s] [%s] [%s] [%s]", reflect.TypeOf(t.Type), t.Type.String(), reflect.TypeOf(t.Type.Underlying()), t.Type.Underlying().String())
		switch t.Type.(type) {
		case *types.Signature:
			s := t.Type.(*types.Signature)
			if s.Recv() != nil {
				util.Info("\t\t[%s] [%s]", s.Recv(), s.Recv().Type().String())
			}
			if tuple := s.Params(); tuple != nil {
				for i := 0; i < tuple.Len(); i++ {
					v := tuple.At(i)
					util.Debug("\t\t%s", v.Name())
					if types.IsInterface(v.Type()) {
						util.Debug("\t\t\t<------interface")
					}
				}
			}
		}
	}
}
