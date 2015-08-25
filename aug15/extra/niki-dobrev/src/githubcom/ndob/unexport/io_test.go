package main

import (
	"go/build"
	"path/filepath"
	"testing"
)

func TestGetSubDirectories(t *testing.T) {

	// Figure out the correct path for this package.
	var absPath string
	for _, dir := range build.Default.SrcDirs() {
		abs := filepath.Join(dir, "github.com", "ndob", "unexport", "testdata")
		if len(getSubDirectories(abs)) > 0 {
			absPath = abs
		}
	}

	cases := []struct {
		in   string
		want []string
	}{
		// dir2 contains a hidden subdir, which should be ignored.
		{filepath.Join(absPath, "subdir", "dir2"), []string{
			"dir3",
		}},

		{filepath.Join(absPath, "subdir"), []string{
			"dir2",
			"dir2/dir3",
		}},

		{"github.com/ndob/unexport/testdata/notfound", []string{}},
	}

	for _, c := range cases {
		got := getSubDirectories(c.in)

		if len(got) != len(c.want) {
			t.Errorf("getSubDirectories(%q) == %q, want %q", c.in, got, c.want)
		}

		if len(c.want) == 0 {
			continue
		}

		found := false
		for _, f := range got {
			for _, file := range c.want {
				if f == file {
					found = true
				}
			}
		}

		if !found {
			t.Errorf("getFileNames(%q) == %q, want %q", c.in, got, c.want)
		}

	}
}

func TestGetFileNames(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{"github.com/ndob/unexport/testdata/prog1", []string{
			"github.com/ndob/unexport/testdata/prog1/prog1.go",
			"github.com/ndob/unexport/testdata/prog1/lib1.go",
		}},

		// prog2 has one test-file, that should not appear on the list.
		{"github.com/ndob/unexport/testdata/prog2", []string{
			"github.com/ndob/unexport/testdata/prog2/prog2.go",
			"github.com/ndob/unexport/testdata/prog2/lib2.go",
		}},

		{"github.com/ndob/unexport/testdata/notfound", []string{}},
	}

	for _, c := range cases {
		got := getFileNames(c.in)

		if len(got) != len(c.want) {
			t.Errorf("getFileNames(%q) == %q, want %q", c.in, got, c.want)
		}

		if len(c.want) == 0 {
			continue
		}

		found := false
		for _, f := range got {
			for _, file := range c.want {
				for _, dir := range build.Default.SrcDirs() {
					if filepath.Join(dir, file) == f {
						found = true
					}
				}
			}
		}

		if !found {
			t.Errorf("getFileNames(%q) == %q, want %q", c.in, got, c.want)
		}

	}
}
