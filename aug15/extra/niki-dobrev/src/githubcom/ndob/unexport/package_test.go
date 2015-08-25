package main

import (
	"go/token"
	"path/filepath"
	"strconv"
	"testing"
)

func TestString(t *testing.T) {
	var tok token.Position

	cases := []struct {
		in   identifier
		want string
	}{
		{identifier{"pkg", "parent", "test", tok}, "\"pkg\".parent.test"},
		{identifier{"p", "p", "p", tok}, "\"p\".p.p"},
		{identifier{"Pkg", "Parent", "Test", tok}, "\"Pkg\".Parent.Test"},
		{identifier{"pkg", "", "Test", tok}, "\"pkg\".Test"},
		{identifier{"", "", "Test", tok}, "\"\".Test"},
		{identifier{"", "", "", tok}, "\"\"."},
	}

	for _, c := range cases {
		got := c.in.string()
		if got != c.want {
			t.Errorf("String(%q) == %q, want %q", c.in, got, c.want)
		}
	}
}

func TestIsExported(t *testing.T) {
	var tok token.Position

	cases := []struct {
		in   identifier
		want bool
	}{
		{identifier{"pkg", "parent", "Test", tok}, true},
		{identifier{"pkg", "parent", "TEST", tok}, true},
		{identifier{"pkg", "parent", "test", tok}, false},
		{identifier{"pkg", "parent", "tEST", tok}, false},
		{identifier{"Pkg", "Parent", "test", tok}, false},
		{identifier{"pkg", "parent", "", tok}, false},
	}

	for _, c := range cases {
		got := c.in.isExported()
		if got != c.want {
			t.Errorf("IsExported(%q) == %t, want %t", c.in, got, c.want)
		}
	}
}

func TestUnexportedName(t *testing.T) {
	var tok token.Position

	cases := []struct {
		in   identifier
		want string
	}{
		{identifier{"pkg", "parent", "Test", tok}, "test"},
		{identifier{"pkg", "parent", "TEST", tok}, "tEST"},
		{identifier{"pkg", "parent", "test", tok}, "test"},
		{identifier{"pkg", "parent", "tEST", tok}, "tEST"},
		{identifier{"Pkg", "Parent", "test", tok}, "test"},
		{identifier{"pkg", "parent", "", tok}, ""},
	}

	for _, c := range cases {
		got := c.in.unexportedName()
		if got != c.want {
			t.Errorf("UnexportedName(%q) == %q, want %q", c.in, got, c.want)
		}
	}
}

func TestNameEquals(t *testing.T) {
	tok := token.Position{Filename: "fo2.txt", Offset: 1, Line: 1, Column: 1}
	randomPos := token.Position{Filename: "foo.txt", Offset: 4444, Line: 432, Column: 1}

	cases := []struct {
		in   identifier
		in2  identifier
		want bool
	}{
		{identifier{"pkg", "parent", "test", tok}, identifier{"pkg", "parent", "test", randomPos}, true},
		{identifier{"pkg", "parent", "test", tok}, identifier{"pkg", "parent", "test", tok}, true},
		{identifier{"abc", "parent", "test", tok}, identifier{"pkg", "parent", "test", tok}, false},
		{identifier{"pkg", "abc", "test", tok}, identifier{"pkg", "parent", "test", tok}, false},
		{identifier{"pkg", "parent", "abc", tok}, identifier{"pkg", "parent", "test", tok}, false},
		{identifier{"pkg", "parent", "test", tok}, identifier{"abc", "parent", "test", tok}, false},
		{identifier{"pkg", "parent", "test", tok}, identifier{"pkg", "abc", "test", tok}, false},
		{identifier{"pkg", "parent", "test", tok}, identifier{"pkg", "parent", "abc", tok}, false},
	}

	for _, c := range cases {
		got := c.in.nameEquals(&c.in2)
		if got != c.want {
			t.Errorf("NameEquals(%q) == %t, want %t", c.in, got, c.want)
		}
	}
}

func TestSearchByPos(t *testing.T) {
	var arr []identifier
	var empty identifier

	randomPos := token.Position{Filename: "foo.txt", Offset: 4444, Line: 432, Column: 1}
	needlePos := token.Position{Filename: "filename.txt", Offset: 5438, Line: 433, Column: 333}
	needle := identifier{"a", "b", "c", needlePos}

	for i := 0; i < 10; i++ {
		tok := token.Position{Filename: "f.txt", Offset: i * 4, Line: i * 13, Column: i + 4}
		arr = append(arr, identifier{
			"a" + strconv.Itoa(i),
			"b" + strconv.Itoa(i),
			"c" + strconv.Itoa(i),
			tok})
	}

	arr = append(arr, needle)

	for i := 0; i < 10; i++ {
		tok := token.Position{Filename: "f.txt", Offset: i * 2, Line: i * 3, Column: i + 42}
		arr = append(arr, identifier{
			"q" + strconv.Itoa(i),
			"w" + strconv.Itoa(i),
			"e" + strconv.Itoa(i),
			tok})
	}

	cases := []struct {
		in      token.Position
		wantRes identifier
		wantOK  bool
	}{
		{needlePos, needle, true},
		{randomPos, empty, false},
	}

	for _, c := range cases {
		gotRes, gotOK := searchByPos(arr, c.in)
		if gotOK != c.wantOK || gotRes != c.wantRes {
			t.Errorf("searchByPos(%q) == %q+%t, want %q+%t", c.in, gotRes, gotOK, c.wantRes, c.wantOK)
		}
	}
}

func TestIsInternal(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"cmd/internal/test", true},
		{"cmd/i/test", false},
		{"cmd/Internal/test", false},
		{"cmd/internalization/test", false},
		{"internal/internal/test", true},
		{"internal", true},
		{"internal/internal/internal", true},
		{"internal//internal//internal", true},
		{"", false},
	}

	for _, c := range cases {
		got := isInternal(c.in)
		if got != c.want {
			t.Errorf("isInternal(%q) == %t, want %t", c.in, got, c.want)
		}
	}
}

func TestGetInternalParentPrefix(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"cmd/internal/test", "cmd"},
		{"cmd/i/test", ""},
		{"cmd/Internal/test", ""},
		{"cmd/internalization/test", ""},
		{"internal/internal/test", "internal"},
		{"internal", ""},
		{"internal/internal/internal", "internal/internal"},
		{"internal//internal//internala", "internal"},
		{"", ""},
	}

	for _, c := range cases {
		got := getInternalParentPrefix(c.in)
		if got != c.want {
			t.Errorf("getInternalParentPrefix(%q) == %q, want %q", c.in, got, c.want)
		}
	}
}

func TestPackagesToSearch(t *testing.T) {
	cases := []struct {
		in1  string
		in2  string
		want []string
	}{
		{"cmd/compile/internal/gc", "", []string{
			"cmd/compile",
			"cmd/compile/internal",
			"cmd/compile/internal/amd64",
			"cmd/compile/internal/arm",
			"cmd/compile/internal/arm64",
			"cmd/compile/internal/big",
			"cmd/compile/internal/gc/builtin",
			"cmd/compile/internal/ppc64",
			"cmd/compile/internal/x86",
		}},

		{"cmd/vet", "cmd/v", []string{
			"cmd/vet/testdata",
			"cmd/vet/testdata/tagtest",
			"cmd/vet/whitelist"}},

		{"cmd/vet", "go/ast", []string{
			"go/ast"}},

		{"cmd/compile/internal/gc", "cmd/compile/internal/amd64", []string{
			"cmd/compile/internal/amd64"}},
	}

	for _, c := range cases {
		got := packagesToSearch(c.in1, c.in2)

		for i := 0; i < len(got); i++ {
			if got[i] != c.want[i] {
				t.Errorf("packagesToSearch(%q) == %q, want %q", c.in1, got[i], c.want[i])
			}
		}
	}
}

func TestCreateProgram(t *testing.T) {
	cases := []struct {
		in   []string
		want bool
	}{
		{[]string{
			filepath.Join("github.com", "ndob", "unexport", "testdata", "prog1"),
		}, false},

		{[]string{
			filepath.Join("github.com", "ndob", "unexport", "testdata", "prog1"),
			filepath.Join("github.com", "ndob", "unexport", "testdata", "prog3"),
		}, false},

		{[]string{
			filepath.Join("github.com", "ndob", "unexport", "testdata", "prog2"),
		}, true},
	}

	for _, c := range cases {
		_, err := createProgram(c.in...)
		if (err == nil) == c.want {
			t.Errorf("createProgram(%q) == %s, want %t", c.in, err, c.want)
		}
	}
}

func TestFindIdentifiersString(t *testing.T) {
	thisPkg := "github.com/ndob/unexport/testdata/package1"

	var tok token.Position
	cases := []struct {
		in   string
		want []identifier
	}{
		{filepath.Join("github.com", "ndob", "unexport", "testdata", "package1"), []identifier{
			identifier{thisPkg, "", "package1", tok},
			identifier{thisPkg, "", "Asdf", tok},
			identifier{thisPkg, "", "innerStruct", tok},
			identifier{thisPkg, "innerStruct", "Fridge", tok},
			identifier{thisPkg, "", "StructTest", tok},
			identifier{thisPkg, "StructTest", "local", tok},
			identifier{thisPkg, "StructTest", "Abc", tok},
			identifier{thisPkg, "StructTest", "inner", tok},
			identifier{thisPkg, "", "StructTest2", tok},
			identifier{thisPkg, "StructTest2", "aaa", tok},
			identifier{thisPkg, "", "StructSlice", tok},
			identifier{thisPkg, "", "someHandler", tok},
			identifier{thisPkg, "someHandler", "ServeHTTP", tok},
			identifier{thisPkg, "someHandler", "serveTCP", tok},
			identifier{thisPkg, "innerStruct", "Foo", tok},
			identifier{thisPkg, "StructTest", "foo2", tok},
			identifier{thisPkg, "", "Derp", tok},
			identifier{thisPkg, "", "a", tok},
			identifier{thisPkg, "", "B", tok},
			identifier{thisPkg, "", "leaf", tok},
			identifier{thisPkg, "", "NewInnerStruct", tok},
			identifier{thisPkg, "", "p", tok},
		}},
	}

	for _, c := range cases {
		ids := findIdentifiers(c.in)
		if len(ids) != len(c.want) {
			t.Errorf("findIdentifiers(%q) == %d, want %d", c.in, len(ids), len(c.want))
		}

		for _, ident := range ids {
			found := false
			for _, ident2 := range c.want {
				if ident.string() == ident2.string() {
					found = true
				}
			}
			if !found {
				t.Errorf("findIdentifiers() not found %s", ident.string())
			}
		}
	}
}

func TestFindIdentifiersPosition(t *testing.T) {
	thisPkg := "github.com/ndob/unexport"

	cases := []struct {
		in   string
		want []identifier
	}{
		{filepath.Join("github.com", "ndob", "unexport", "testdata", "package1"), []identifier{
			identifier{thisPkg, "someHandler", "ServeHTTP", token.Position{Filename: "a.go", Offset: 299, Line: 27, Column: 2}},
			identifier{thisPkg, "", "Derp", token.Position{Filename: "a.go", Offset: 414, Line: 34, Column: 7}},
		}},
	}

	for _, c := range cases {
		ids := findIdentifiers(c.in)

		for _, ident := range c.want {

			found := false
			for _, ident2 := range ids {
				_, ident2.pos.Filename = filepath.Split(ident2.pos.Filename)
				if ident.pos == ident2.pos {
					found = true
				}
			}
			if !found {
				t.Errorf("findIdentifiers() not found %s %s %d %s %d %s %d",
					ident.string(),
					"Offset:",
					ident.pos.Offset,
					"Line:",
					ident.pos.Line,
					"Column:",
					ident.pos.Column)
			}
		}
	}
}

func TestFindExports(t *testing.T) {
	thisPkg := "github.com/ndob/unexport/testdata/package1"

	var tok token.Position
	cases := []struct {
		in   string
		want []identifier
	}{
		{filepath.Join("github.com", "ndob", "unexport", "testdata", "package1"), []identifier{
			identifier{thisPkg, "", "Asdf", tok},
			identifier{thisPkg, "innerStruct", "Fridge", tok},
			identifier{thisPkg, "", "StructTest", tok},
			identifier{thisPkg, "", "StructTest2", tok},
			identifier{thisPkg, "StructTest", "Abc", tok},
			identifier{thisPkg, "", "StructSlice", tok},
			identifier{thisPkg, "someHandler", "ServeHTTP", tok},
			identifier{thisPkg, "innerStruct", "Foo", tok},
			identifier{thisPkg, "", "Derp", tok},
			identifier{thisPkg, "", "B", tok},
			identifier{thisPkg, "", "NewInnerStruct", tok},
		}},
	}

	for _, c := range cases {
		ids := findExports(c.in)
		if len(ids) != len(c.want) {
			t.Errorf("findExports(%q) == %d, want %d", c.in, len(ids), len(c.want))
		}

		for _, ident := range ids {
			found := false
			for _, ident2 := range c.want {
				if ident.string() == ident2.string() {
					found = true
				}
			}
			if !found {
				t.Errorf("findExports() not found %s", ident.string())
			}
		}
	}
}

func TestUsageCounts(t *testing.T) {
	var tok token.Position
	thisPkg := "github.com/ndob/unexport"
	pkg1 := thisPkg + "/testdata/package1"
	pkg2 := thisPkg + "/testdata/package3"

	exp := findExports(pkg1)
	usages := usageCounts(pkg1, exp, pkg2)

	want := []struct {
		id     identifier
		usages int
	}{
		{identifier{thisPkg, "", "Derp", tok}, 2},
		{identifier{thisPkg, "", "B", tok}, 1},
		{identifier{thisPkg, "", "Asdf", tok}, 1},
	}

	for id, idusages := range usages {
		for _, w := range want {
			if w.id.string() == id.string() {
				if len(idusages) != w.usages {
					t.Errorf("usageCounts() wrong [%s] expected: %d got: %d", id.string(), w.usages, len(idusages))
				}
			}
		}

	}
}

func TestUnusedExports(t *testing.T) {
	var tok token.Position
	thisPkg := "github.com/ndob/unexport"
	pkg1 := thisPkg + "/testdata/package1"
	pkg2 := thisPkg + "/testdata/package3"

	packages := packagesToSearch(pkg1, pkg2)
	unused := unusedExports(pkg1, packages)

	want := []identifier{
		identifier{pkg1, "innerStruct", "Fridge", tok},
		identifier{pkg1, "", "StructTest", tok},
		identifier{pkg1, "", "StructTest2", tok},
		identifier{pkg1, "StructTest", "Abc", tok},
		identifier{pkg1, "", "StructSlice", tok},
		identifier{pkg1, "someHandler", "ServeHTTP", tok},
		identifier{pkg1, "innerStruct", "Foo", tok},
		identifier{pkg1, "", "NewInnerStruct", tok},
	}

	if len(unused) != len(want) {
		t.Errorf("unusedExports() lengths differ %d, want %d", len(unused), len(want))
	}

	for _, id := range want {
		found := false
		for _, w := range unused {
			if w.string() == id.string() {
				found = true
			}
		}
		if !found {
			t.Errorf("unusedExports() not found [%s]", id.string())
		}
	}
}

func TestUnusedExports2(t *testing.T) {
	var tok token.Position
	thisPkg := "github.com/ndob/unexport"
	pkg1 := thisPkg + "/testdata/package1"
	pkg2 := thisPkg + "/testdata/package2"

	packages := packagesToSearch(pkg1, pkg2)
	unused := unusedExports(pkg1, packages)

	want := []identifier{
		identifier{pkg1, "", "StructTest", tok},
		identifier{pkg1, "", "StructSlice", tok},
		identifier{pkg1, "someHandler", "ServeHTTP", tok},
	}

	if len(unused) != len(want) {
		t.Errorf("unusedExports() lengths differ %d, want %d", len(unused), len(want))
	}

	for _, id := range want {
		found := false
		for _, w := range unused {
			if w.string() == id.string() {
				found = true
			}
		}
		if !found {
			t.Errorf("unusedExports() not found [%s]", id.string())
		}
	}
}

func TestNameCollisions(t *testing.T) {
	var tok token.Position
	thisPkg := "github.com/ndob/unexport"
	pkg1 := thisPkg + "/testdata/namecollision"
	pkg2 := thisPkg + "/testdata/package2"

	packages := packagesToSearch(pkg1, pkg2)
	unused := unusedExports(pkg1, packages)

	collisions := nameCollisions(pkg1, unused)

	want := []identifier{
		identifier{pkg1, "", "Asdf", tok},
		identifier{pkg1, "someHandler", "ServeHTTP", tok},
		identifier{pkg1, "StructTest", "Abc", tok},
		identifier{pkg1, "StructTest", "Foo", tok},
		identifier{pkg1, "", "A", tok},
		identifier{pkg1, "", "Derp", tok},
		identifier{pkg1, "", "StructSlice", tok},
	}

	if len(collisions) != len(want) {
		t.Errorf("nameCollisions() lengths differ %d, want %d", len(collisions), len(want))
	}

	for _, id := range want {
		found := false
		for _, w := range collisions {
			if w.string() == id.string() {
				found = true
			}
		}
		if !found {
			t.Errorf("nameCollisions() not found [%s]", id.string())
		}
	}
}
