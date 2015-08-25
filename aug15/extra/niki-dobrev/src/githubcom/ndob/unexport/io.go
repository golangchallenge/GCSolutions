package main

import (
	"fmt"
	"go/build"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// getSubDirectories returns all subdirectories under provided
// absolute path. Directories are returned as relative paths from
// provided absolute path.
func getSubDirectories(abspath string) []string {
	var directories []string

	files, err := ioutil.ReadDir(abspath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return directories
	}

	for _, f := range files {
		// Skip hidden directories.
		if strings.HasPrefix(f.Name(), ".") {
			continue
		}

		if f.IsDir() {
			subDir := f.Name()
			directories = append(directories, subDir)
			subDirs := getSubDirectories(filepath.Join(abspath, subDir))
			for _, d := range subDirs {
				directories = append(directories, filepath.Join(subDir, d))
			}
		}

		// Follow symlinks.
		if (f.Mode() & os.ModeSymlink) == os.ModeSymlink {
			absolutePath, err := os.Readlink(filepath.Join(abspath, f.Name()))
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				continue
			}

			// Check that symlink target is a directory.
			_, err = ioutil.ReadDir(absolutePath)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				continue
			}

			directories = append(directories, f.Name())

			subDirs := getSubDirectories(absolutePath)
			for _, d := range subDirs {
				directories = append(directories, filepath.Join(f.Name(), d))
			}
		}
	}
	return directories
}

// getFileNames returns all .go-source files for provided package.
func getFileNames(pkgName string) []string {
	var filenames []string

	pkg, err := build.Default.Import(pkgName, ".", 0)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return filenames
	}

	for _, filename := range pkg.GoFiles {
		fullPath := filepath.Join(pkg.Dir, filename)
		filenames = append(filenames, fullPath)
	}
	return filenames
}
