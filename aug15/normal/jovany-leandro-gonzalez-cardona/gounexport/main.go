// Go-Challenge 5
// solution for go challenge 5
//
//Requirements of the challenge
//
//  * Cover all exported identifiers, including functions, constants, variables, types, field names in structs, etc. If godoc shows it, it qualifies.
//  * Allow the user to decide which identifiers should be unexported. There may be good reasons to keep an identified exported even if it is not used by any packages. One simple way to do this is to have unexport generate a series of gorename commands. The user could then decide which renamings to apply. It is up to you to design the interface as you please.
//  * Use only packages found in the standard library and the golang.org/x/tools repo. Note that there are two copies of the go/types package, one in the standard library and one in golang.org/x/tools. It is ok to use either one. All else being equal, the standard library version is preferable, but the useful go/loader package in golang.org/x/tools is not (yet?) part of the standard library.
//  * Use Go 1.5.
//
//
// Usage:
//   ./gounexport [options] <pkg> <regexp>
//
// Examples of queries
//   ./gounexport github.com/bit4bit/gfsocket #list everything can be exported
//   ./gounexport github.com/bit4bit/gfsocket "@func$" #list only exported functions
//   ./gounexport github.com/bit4bit/gfsocket "@struct$" #list only structs
//   ./gounexport github.com/bit4bit/gfsocket "@type$" #list only custom types
//   ./gounexport github.com/bit4bit/gfsocket "gfsocket.Event.Type@attr$" #list specific attribute
//   ./gounexport github.com/bit4bit/gfsocket ".+Dial.+$" #list everything have Dial
//   ./gounexport github.com/bit4bit/gfsocket "gfsocket.Event@struct" #unexport event on all file
//   ./gounexport -w github.com/bit4bit/gfsocket "@func$" #unexport all funcs, overriding the file. use -w for write to files
package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/build"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode"
)


var goKeywords = map[string]bool{
	"bool":true,
	"string":true,
	"int": true, "int8":true, "int16":true, "int64": true, "int32": true, 
	"uint": true, "uint8":true, "uint16":true, "uint64":true, "uint32": true,
	"byte":true, "rune":true,
	"float32": true, "float64":true, "complex64":true, "complex128":true,
	"range":true, "var": true,  "func":true, "return":true, "map":true, 
	"if":true, "else": true, "for":true, "switch":true, "case":true, "default":true,
	"chan":true, "select":true,  "error":true,
	"type":true, "struct":true, "interface":true,
}

type queryIdent struct {
	Query string
	Ident *ast.Ident
}

type byQuery []*queryIdent

func (c byQuery) Len() int           { return len(c) }
func (c byQuery) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }
func (c byQuery) Less(i, j int) bool { return strings.Compare(c[i].Query, c[j].Query) == -1 }

var (
	fileSet = token.NewFileSet()
)

func isGoFile(file os.FileInfo) bool {
	return strings.HasSuffix(file.Name(), ".go") &&
		!strings.HasSuffix(file.Name(), "test.go")

}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s: [options] <pkg> [regexp]\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "[regexp] match for unexport. Example: \".+MiStruct+@struct\"\n")
	fmt.Fprintf(os.Stderr, "\tidentifier suffix by type: @struct, @interface, @type, @attr, @method, @func, @var, @const\n")
	flag.PrintDefaults()
}

func main() {
	log.SetPrefix("gounexport: ")
	err := do(os.Stdout, flag.CommandLine, os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}
}

//do useful for testing
func do(writer io.Writer, flags *flag.FlagSet, args []string) error {
	var pkgPath string
	var rematchUser string

	flags.Usage = usage
	writeToFilesFlag := flags.Bool("w", false, "write to files")
	simulateFlag := flags.Bool("s", false, "simulate updates")
	flags.Parse(args)
	
	pkgPath, rematchUser = parseArgs(flags)
	

	packages, err := parser.ParseDir(fileSet, pkgPath, isGoFile, 0)
	if err != nil {
		return fmt.Errorf("parserDir: %s", err.Error())
	}

	info := types.Info{
		Defs: make(map[*ast.Ident]types.Object),
		Uses: make(map[*ast.Ident]types.Object),
	}

	var files []*ast.File
	fileByPath := make(map[string]*os.File) //cache files
	for _, pkg := range packages {

		for filePath, file := range pkg.Files {
			fp, err := os.OpenFile(filePath, os.O_RDWR, 0777)
			if err != nil {
				return err
			}
			defer fp.Close()
			files = append(files, file)
			fileByPath[filePath] = fp
		}
	}

	//parse package
	var conf types.Config
	conf.Importer = importer.Default()
	conf.DisableUnusedImportCheck = true
	conf.Error = func(err error) {
		//omit errors try parse it
	}
	pkg, _ := conf.Check(pkgPath, fileSet, files, &info)

	//queries for user matching
	queries := make(chan *queryIdent)
	//queries for unexported fields used it for detect collision
	chqueriesUnexported := make(chan *queryIdent)
	go buildQueryStrings(queries, pkg, &info, true)
	go buildQueryStrings(chqueriesUnexported, pkg, &info, false)
	queriesUnexported := make(map[string]*ast.Ident)
	for queryUnexport := range chqueriesUnexported {
		queriesUnexported[queryUnexport.Query] = queryUnexport.Ident
	}

	//go routine for unexporting
	doneSave := make(chan bool)
	unexportPositions := make(chan token.Position, 1)
	unexports := make(map[string]bool)
	//start updater on files
	go unexportOnFile(doneSave, fileByPath, unexportPositions, *simulateFlag)

	//try match user query
	var where func(string) bool

	if rematchUser == "" {
		rematchUser = ".+"
	}
	where = regexp.MustCompile(rematchUser).MatchString

	//used only for showing
	var generalQueries []*queryIdent
	for query := range queries {

		if where(query.Query) {
			//detect collision
			_, isKeyword := goKeywords[nameUnexported(query.Ident.Name)]
			if queriesUnexported[query.Query] != nil || isKeyword {
				if !*writeToFilesFlag {
					generalQueries = append(generalQueries, &queryIdent{query.Query + " !!collision", query.Ident})
				} else {
					fmt.Fprintln(writer, "sorry collision detected for", query.Query)
				}
				continue
			}
			
			pos := fileSet.Position(query.Ident.Pos())
			unexports[query.Ident.Name] = true

			if !*writeToFilesFlag {
				generalQueries = append(generalQueries, query)
				continue
			}
			
			fmt.Fprintln(writer, "Unexported", query.Ident.Name, "from", query.Query)
			unexportPositions <- pos
		}

	}

	displayUses := make(map[string]map[string][]string)
	for ident := range info.Uses {
		if _, ok := unexports[ident.Name]; ok {
			pos := fileSet.Position(ident.Pos())
			if !*writeToFilesFlag {
				if _, ok := displayUses[pos.Filename]; !ok {
					displayUses[pos.Filename] = make(map[string][]string)
				}
				posLine := fmt.Sprintf("\t\t\t\t%s:%d:%d", pos.Filename, pos.Line, pos.Column)
				displayUses[pos.Filename][ident.Name] = append(displayUses[pos.Filename][ident.Name], posLine)
				continue
			}

			unexportPositions <- pos
		}
	}
	if !*writeToFilesFlag {
		sort.Sort(byQuery(generalQueries))
		prettyQueries(writer, generalQueries, displayUses)
		return nil
	}
	close(unexportPositions)
	<-doneSave

	return nil
}

//buildQueryStrings build names for matching
//if onlyExport it's true only query Exported identifier else only Unexported
func buildQueryStrings(queries chan *queryIdent, pkg *types.Package, info *types.Info, onlyExport bool) {
	defer close(queries)

	var typesDefs []types.Object
	typesByField := make(map[types.Object]types.Object)
	interfaceByMethod := make(map[types.Object]types.Object)

	//get definition of types first
	for ident, obj := range info.Defs {
		if obj == nil {
			continue
		}

		if ident.IsExported() != onlyExport {
			//querying unexport, so we need save typesDefs for finding fields func etc..
			if !onlyExport {
				typesDefs = append(typesDefs, obj)
			}
			continue
		}

		switch obj.(type) {
		case *types.TypeName:
			var typeOfObj string
			switch t := obj.Type().Underlying().(type) {
			case *types.Struct:
				for i := 0; i < t.NumFields(); i++ {
					field := t.Field(i)
					typesByField[field] = obj
				}
				typeOfObj = "@struct"
			case *types.Interface:
				for i := 0; i < t.NumMethods(); i++ {
					method := t.Method(i)
					interfaceByMethod[method] = obj
				}
				for i := 0; i < t.NumExplicitMethods(); i++ {
					method := t.ExplicitMethod(i)
					interfaceByMethod[method] = obj
				}
				typeOfObj = "@interface"
			default:
				typeOfObj = "@type"
			}
			query := pkg.Name() + "." + nameExported(obj.Name()) + typeOfObj
			queries <- &queryIdent{query, ident}
			typesDefs = append(typesDefs, obj)
		}
	}

	//get rest of definitions
	for ident, obj := range info.Defs {
		if obj == nil {
			continue
		}
		if ident.IsExported() != onlyExport {
			continue
		}

		if sobj, ok := typesByField[obj]; ok {
			queries <- &queryIdent{joinQuery(pkg, sobj, obj, "@attr"), ident}
		} else if sobj, ok := interfaceByMethod[obj]; ok {
			queries <- &queryIdent{joinQuery(pkg, sobj, obj, "@method"), ident}
		} else {

			switch t := obj.(type) {
			case *types.Const:
				queries <- &queryIdent{joinQuery(pkg, nil, obj, "@const"), ident}
			case *types.Var:
				if !t.IsField() {
					queries <- &queryIdent{joinQuery(pkg, nil, obj, "@var"), ident}
				}
			case *types.Func:
				foundLikeMethod := false
				for _, tobj := range typesDefs {
					tfield, _, _ := types.LookupFieldOrMethod(tobj.Type(), true, pkg, obj.Name())
					if tfield == nil {
						continue
					}
					if tfield.Pos() == obj.Pos() {
						queries <- &queryIdent{joinQuery(pkg, tobj, tfield, "@method"), ident}
						foundLikeMethod = true
					}
				}
				if !foundLikeMethod {
					queries <- &queryIdent{joinQuery(pkg, nil, obj, "@func"), ident}

				}
			default:
			}
		}

	}

}

//unexportOnFile rewrite the first character to lower case
func unexportOnFile(done chan bool, fileByPath map[string]*os.File, positions chan token.Position, simulate bool) {
	for pos := range positions {
		if simulate {
			continue
		}

		if fp, ok := fileByPath[pos.Filename]; ok {

			char := make([]byte, 1)

			_, err := fp.ReadAt(char, int64(pos.Offset))
			if err != nil {
				log.Println(err)
			}
			char = bytes.ToLower(char)
			_, err = fp.WriteAt(char, int64(pos.Offset))
			if err != nil {
				log.Println(err)
			}

		}
	}
	done <- true
}

//nameExported uppercase first character
func nameExported(name string) string {
	c := unicode.ToUpper(rune(name[0]))
	return string(c) + string(name[1:])
}

func nameUnexported(name string) string {
	c := unicode.ToLower(rune(name[0]))
	return string(c) + string(name[1:])
}

func joinQuery(pkg *types.Package, parent types.Object, obj types.Object, suffix string) string {
	var args []string
	args = append(args, pkg.Name())
	if parent != nil {
		args = append(args, parent.Name())
	}

	args = append(args, nameExported(obj.Name()))
	return strings.Join(args, ".") + suffix
}

//prettyQueries show queries group by file
//uses it's like map[filename]map[ast.Ident.Name][]query
func prettyQueries(w io.Writer, queries []*queryIdent, uses map[string]map[string][]string) {

	omit := make(map[int]bool)
	orderByFilename := make(map[string][]string)

	for _, suffix := range []string{"@struct", "@interface"} {
		for idx, query := range queries {
			if strings.HasSuffix(query.Query, suffix) {
				pos := fileSet.Position(query.Ident.Pos())
				orderByFilename[pos.Filename] = append(orderByFilename[pos.Filename], "\t"+query.Query)

				if us, ok := uses[pos.Filename][query.Ident.Name]; ok {
					orderByFilename[pos.Filename] = append(orderByFilename[pos.Filename], "\t\t\tUses at\n"+strings.Join(us, "\n"))
				}
				omit[idx] = true
				for idx2, query2 := range queries {
					if strings.HasPrefix(query2.Query, strings.Replace(query.Query, suffix, ".", -1)) {
						omit[idx2] = true
						orderByFilename[pos.Filename] = append(orderByFilename[pos.Filename], "\t\t"+query2.Query)
						if us, ok := uses[pos.Filename][query2.Ident.Name]; ok {
							orderByFilename[pos.Filename] = append(orderByFilename[pos.Filename], "\t\t\tUses at\n"+strings.Join(us, "\n"))
						}
					}
				}
			}
		}
	}

	for idx, query := range queries {
		if _, ok := omit[idx]; ok {
			continue
		}
		pos := fileSet.Position(query.Ident.Pos())
		if us, ok := uses[pos.Filename][query.Ident.Name]; ok {
			orderByFilename[pos.Filename] = append(orderByFilename[pos.Filename], "\t\t\tUses at\n"+strings.Join(us, "\n"))
		}
		orderByFilename[pos.Filename] = append(orderByFilename[pos.Filename], "\t"+query.Query)
	}

	var filesSort []string
	for filePath := range orderByFilename {
		filesSort = append(filesSort, filePath)

	}
	sort.Strings(filesSort)
	for _, filePath := range filesSort {
		fmt.Fprintln(w, filePath)
		fmt.Fprintln(w, strings.Join(orderByFilename[filePath], "\n"))
	}
}


func parseArgs(flags *flag.FlagSet) (string, string) {
	var homePkg string
	if strings.HasPrefix(flags.Arg(0), "/") {
		homePkg = ""
	} else	if flags.Arg(0) != "" && !strings.HasPrefix(flags.Arg(0), "./") {
		homePkg = filepath.Join(build.Default.GOPATH, "src")
	} else {
		pwd, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		homePkg = pwd
	}
	pkgPath := filepath.Join(homePkg, flags.Arg(0))
	return pkgPath, flags.Arg(1)
}
