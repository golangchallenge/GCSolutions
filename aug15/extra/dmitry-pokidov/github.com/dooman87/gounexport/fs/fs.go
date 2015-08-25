//Package fs provides utility functions to work with sources files
//relate to standard golang workspace. It's using GOROOT and GOHOME
//environment variables to search.
package fs

import (
	"go/token"
	"os"
	"strings"

	"github.com/dooman87/gounexport/util"
)

//GetFiles returns all golang source files which is inside a package.
//A path to the package is constructing using standard workspace
//layout - $GOPATH/src.
//All tests files are ignored.
//If deep flag is true then will return
//files from all subpackages as well.
//It collects only files that name ends with .go extension.
//Returns list of full file names.
func GetFiles(pkg string, deep bool) ([]string, error) {
	result, err := getSourceFiles(os.Getenv("GOPATH")+"/src/"+pkg, deep)
	if err != nil {
		result, err = getSourceFiles(os.Getenv("GOROOT")+"/src/"+pkg, deep)
	}
	return result, err
}

func getSourceFiles(pkgPath string, deep bool) ([]string, error) {
	pkgDir, err := os.Open(pkgPath)
	if err != nil {
		util.Err("error while opening package at path [%s]\n%v", pkgPath, err)
		return nil, err
	}

	filesInfos, err := pkgDir.Readdir(0)
	if err != nil {
		util.Err("error reading opening package at path [%s]\n%v", pkgPath, err)
		return nil, err
	}

	var files []string
	for _, f := range filesInfos {
		if f.IsDir() && deep && isValidSourceDir(f.Name()) {
			util.Debug("append folder [%s]", f.Name())
			dirFiles, err := getSourceFiles(pkgPath+"/"+f.Name(), true)
			if err != nil {
				return dirFiles, err
			}
			files = append(files, dirFiles...)
		} else if isValidSourceFile(f.Name()) {
			util.Debug("append file [%s]", f.Name())
			if strings.HasSuffix(f.Name(), "_test.go") {
				files = append(files, pkgPath+"/"+f.Name())
			} else {
				files = append([]string{pkgPath + "/" + f.Name()}, files...)
			}
		}
	}
	util.Info("FILES [%v]", files)

	return files, nil
}

func isValidSourceFile(file string) bool {
	return strings.HasSuffix(file, ".go")
	//&& !strings.HasSuffix(file, "_test.go")
}

func isValidSourceDir(dir string) bool {
	return !strings.HasPrefix(dir, ".")
}

//GetUnusedSources returns list of source files in package that
//are not presenting in the file set
func GetUnusedSources(pkg string, fset *token.FileSet) ([]string, error) {
	unusedSource, err := GetFiles(pkg, true)

	if err != nil {
		return nil, err
	}

	iterateFiles := func(f *token.File) bool {
		idx := indexOf(unusedSource, f.Name())
		if idx >= 0 {
			unusedSource = append(unusedSource[:idx], unusedSource[idx+1:]...)
		}
		return true
	}
	fset.Iterate(iterateFiles)
	return unusedSource, nil
}

//GetPackagePath return relative path
//to standard go workspace layout - GOPATH/src
// and GOROOT/src
//and trims file name if it presents.
func GetPackagePath(dirPath string) string {
	result := GetRelativePath(dirPath)
	if strings.HasSuffix(dirPath, ".go") {
		result = result[0:strings.LastIndex(result, "/")]
	}
	return result
}

//GetRelativePath returns relative path
//to $GOPATH or $GOROOT env variable.
func GetRelativePath(path string) string {
	prefix := os.Getenv("GOPATH") + "/src/"
	result := path
	if strings.HasPrefix(path, prefix) {
		result = path[len(prefix):]
	}
	prefix = os.Getenv("GOROOT") + "/src/"
	if strings.HasPrefix(path, prefix) {
		result = path[len(prefix):]
	}

	return result
}

//ReplaceStringInFile replaces string in the file at the given offset
func ReplaceStringInFile(file string, offset int, from string, to string) error {
	sourceFile, err := os.OpenFile(file, os.O_RDWR, 0)
	if err != nil {
		return err
	}

	var info os.FileInfo
	var restFile []byte
	seekTo := offset + len(from)

	if info, err = sourceFile.Stat(); err != nil {
		goto closeAndReturn
	}

	restFile = make([]byte, int(info.Size())-seekTo)
	if _, err = sourceFile.Seek(int64(seekTo), 0); err != nil {
		goto closeAndReturn
	}
	if _, err = sourceFile.Read(restFile); err != nil {
		goto closeAndReturn
	}
	if err = sourceFile.Truncate(int64(offset)); err != nil {
		goto closeAndReturn
	}
	if _, err = sourceFile.Seek(int64(offset), 0); err != nil {
		goto closeAndReturn
	}
	if _, err = sourceFile.WriteString(to); err != nil {
		goto closeAndReturn
	}
	if _, err = sourceFile.Write(restFile); err != nil {
		goto closeAndReturn
	}
	if err = sourceFile.Close(); err != nil {
		goto closeAndReturn
	}

closeAndReturn:
	closeErr := sourceFile.Close()
	if closeErr != nil && err == nil {
		err = closeErr
	}

	return err
}

func indexOf(slice []string, find string) int {
	for i, s := range slice {
		if s == find {
			return i
		}
	}
	return -1
}
