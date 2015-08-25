package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

var goPathSrc = filepath.Join(os.Getenv("GOPATH"), "src")
var goRootSrc = filepath.Join(os.Getenv("GOROOT"), "src")

var unexportAll bool
var pkgPath string
var safe bool

func init() {
	log.SetFlags(log.Flags() | log.Llongfile)
	flag.StringVar(&pkgPath, "pkg", "", "Which package to scan for exported identifiers")
	flag.BoolVar(&unexportAll, "unexportAll", false, "Don't interactively ask for each unused exported identifier, just export all.")
	flag.BoolVar(&safe, "safe", false, "Make unexport only work on internal packages")
}
func main() {
	flag.Parse()

	if pkgPath == "" {
		flag.PrintDefaults()
		return
	}

	fmt.Println("Parsing packages, please wait...")
	pkg, err := newPkg(pkgPath)
	if err != nil {
		log.Fatal(err)
	}

	var internalPkgPath string
	if safe {
		internalPkg := strings.SplitN(pkgPath, "internal", 2)
		if len(internalPkg) == 1 {
			log.Fatal("Expected an internal package")
		}
		internalPkgPath = internalPkg[0] + "internal"
		internalPkgPath = filepath.Join(strings.Split(pkg.pkg.Dir, internalPkgPath)[0], internalPkgPath)
	}

	var pkgLocations []string
	if pkg.pkg.Goroot {
		pkgLocations = []string{goRootSrc}
	} else {
		pkgLocations = []string{goPathSrc, goRootSrc}
	}

	packageList := make([]string, 0)

	for _, basePath := range pkgLocations {
		if err := filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
			if safe && !filepath.HasPrefix(path, internalPkgPath) {
				return nil
			}

			if info.IsDir() && !strings.Contains(path, "/testdata") && !strings.Contains(path, "/_") && !strings.Contains(path, "/.") {
				relPath, err := filepath.Rel(basePath, path)
				if err != nil {
					return err
				}

				matches, err := filepath.Glob(filepath.Join(path, "*.go"))
				if err != nil {
					log.Fatal(err)
				}
				if relPath != "." && len(matches) > 0 {
					packageList = append(packageList, relPath)
				}
			}
			return nil
		}); err != nil {
			log.Fatal(err)
		}
	}

	for _, relPath := range packageList {
		fmt.Println(relPath)
		otherPkg, err := newPkg(relPath)
		if err != nil {
			fmt.Println(err)
			break
		}

		otherPkg.calculateUsesOf(pkg)
	}
	fmt.Print("\n\n")

	var cmds []*exec.Cmd
	var yn string
	var firstUnused bool
	for _, ident := range sortIdentifiersByUsedByCount(pkg.identifiers) {
		if len(ident.usedBy) == 0 {
			if !firstUnused {
				fmt.Printf("\n\nThe following exported identifiers is not used by any other packages in %s\n", pkgLocations)
				firstUnused = true
			}
			fmt.Printf(" - %s\n", ident.id())
			fmt.Println("Would you like to invoke gorename to unexport this identifier? [Yn]: ")
			if !unexportAll {
				fmt.Scanf("%s", &yn)
			}
			if yn == "" || yn[0] == 'y' || yn[0] == 'Y' {
				cmd := exec.Command("gorename", "-from", ident.id(), "-to", ident.unexportedId())
				cmds = append(cmds, cmd)
			}
		} else {
			for i, pos := range ident.usedBy {
				if i == 0 {
					fmt.Printf("%s is referenced by:\n", ident.id())
				}
				fmt.Printf("  %s\n", pos)
			}
		}
	}

	var success []*exec.Cmd
	var failures []*exec.Cmd
	fmt.Println("Running gorename commands:")
	for _, cmd := range cmds {
		cmd.Stdout = &bytes.Buffer{}
		cmd.Stderr = &bytes.Buffer{}

		fmt.Printf(" # %s\n", strings.Join(cmd.Args, " "))
		err := cmd.Run()
		if err != nil {
			failures = append(failures, cmd)
		} else {
			success = append(success, cmd)
		}
	}

	for i, cmd := range success {
		if i == 0 {
			fmt.Println("The following commands completed successfully:")
		}
		fmt.Printf(" # %s\n", strings.Join(cmd.Args, " "))
		fmt.Println(cmd.Stdout)
	}

	for i, cmd := range failures {
		if i == 0 {
			fmt.Println("The following commands failed horribly:")
		}
		fmt.Printf(" # %s\n", strings.Join(cmd.Args, " "))
		fmt.Println(cmd.Stdout)
		fmt.Println(cmd.Stderr)

		fmt.Println("Would you like to rerun the command with the force flag? [yN]: ")
		fmt.Scanf("%s", &yn)
		if len(yn) > 0 && (yn[0] == 'y' || yn[0] == 'Y') {
			newArgs := []string{"-force"}
			newArgs = append(newArgs, cmd.Args[1:]...)
			cmd = exec.Command(cmd.Args[0], newArgs...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			fmt.Println("Running command with force flag!")
			fmt.Printf(" # %s\n", strings.Join(cmd.Args, " "))
			err := cmd.Run()
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

func sortIdentifiersByUsedByCount(o identifiers) []*identifier {
	s := &byUsedBy{
		o: make([]*identifier, 0, len(o)),
	}
	for _, v := range o {
		s.o = append(s.o, v)
	}
	sort.Sort(s)
	return s.o
}

type byUsedBy struct {
	o []*identifier
}

func (s *byUsedBy) Len() int { return len(s.o) }

func (s *byUsedBy) Swap(i, j int) { s.o[i], s.o[j] = s.o[j], s.o[i] }

func (s *byUsedBy) Less(i, j int) bool { return len(s.o[i].usedBy) > len(s.o[j].usedBy) }
