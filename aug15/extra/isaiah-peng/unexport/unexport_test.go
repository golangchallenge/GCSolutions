package unexport

import (
	"fmt"
	"go/build"
	"golang.org/x/tools/go/buildutil"
	"golang.org/x/tools/go/loader"
	"golang.org/x/tools/go/types"
	"reflect"
	"strings"
	"testing"
)

func TestUsedIdentifiers(t *testing.T) {
	for _, test := range []struct {
		ctx *build.Context
		pkg string
	}{
		{ctx: fakeContext(
			map[string][]string{
				"foo": {`
package foo
type I interface{
F()
}
`},
				"bar": {`
package bar
import "foo"
type s int
func (s) F() {}
var _ foo.I = s(0)
`},
			},
		),
			pkg: "bar",
		},
	} {
		prog, err := loadProgram(test.ctx, []string{test.pkg})
		if err != nil {
			t.Fatal(err)
		}
		u := &Unexporter{
			iprog:       prog,
			packages:    make(map[*types.Package]*loader.PackageInfo),
			Identifiers: make(map[types.Object]*ObjectInfo),
		}
		for _, info := range prog.Imported {
			u.packages[info.Pkg] = info
		}
		used := u.usedObjects()
		if len(used) != 3 {
			t.Errorf("expected 3 used objects, got %v", used)
		}
	}
}

func TestUnusedIdentifiers(t *testing.T) {
	for _, test := range []struct {
		ctx  *build.Context
		pkg  string
		want map[string]string
	}{
		// init data
		// unused var
		{ctx: main(`package main; var Unused int = 1`),
			pkg:  "main",
			want: map[string]string{"\"main\".Unused": "unused"},
		},
		// unused const
		{ctx: main(`package main; const Unused int = 1`),
			pkg:  "main",
			want: map[string]string{"\"main\".Unused": "unused"},
		},
		// unused type
		{ctx: main(`package main; type S int`),
			pkg:  "main",
			want: map[string]string{"\"main\".S": "s"},
		},
		// unused struct type, embeded #16
		{ctx: main(`
package main
type S struct { t int }
type x struct {
S
}
`),
			pkg:  "main",
			want: map[string]string{"\"main\".S": "s"},
		},
		// unused interface type, embeded in struct type #16
		{ctx: main(`
package main
type S interface { f()int }
type x struct {
S
}
`),
			pkg:  "main",
			want: map[string]string{"\"main\".S": "s"},
		},
		// unused interface type, embeded in interface type #16
		{ctx: main(`
package main
type S interface { f()int }
type x interface {
S
}
`),
			pkg:  "main",
			want: map[string]string{"\"main\".S": "s"},
		},

		// unused type field
		{ctx: main(`package main; type s struct { T int }`),
			pkg:  "main",
			want: map[string]string{"(\"main\".s).T": "t"},
		},
		// unused type method
		{ctx: main(`package main; type s int; func (s) F(){}`),
			pkg:  "main",
			want: map[string]string{"(\"main\".s).F": "f"},
		},
		// unused interface method
		{ctx: main(`package main; type s interface { F() }`),
			pkg:  "main",
			want: map[string]string{"(\"main\".s).F": "f"},
		},
		// type used by function
		{ctx: fakeContext(map[string][]string{
			"foo": {`
package foo
type S int
type T int
`},
			"bar": {`
package bar
import "foo"

func f(t *foo.T) {}
`},
		}),
			pkg:  "foo",
			want: map[string]string{"\"foo\".S": "s"},
		},
		// type used, but field not used
		{ctx: fakeContext(map[string][]string{
			"foo": {`
package foo
type S struct {
F int
}
`},
			"bar": {`
package bar
import "foo"

var _ foo.S = foo.S{}
`},
		}),
			pkg:  "foo",
			want: map[string]string{"(\"foo\".S).F": "f"},
		},
		// type used, but field not used
		{ctx: fakeContext(map[string][]string{
			"foo": {`
package foo
type S struct {
F int
}
`},
			"bar": {`
package bar
import "foo"

var _ foo.S = foo.S{}
`},
		}),
			pkg:  "foo",
			want: map[string]string{"(\"foo\".S).F": "f"},
		},
		// type embedded, #4
		{ctx: fakeContext(map[string][]string{
			"foo": {`
package foo
type S struct {
f int
}
`},
			"bar": {`
package bar
import "foo"

type x struct {
*foo.S
}
`},
		}),
			pkg: "foo",
		},
		// unused interface type
		{ctx: fakeContext(map[string][]string{
			"foo": {`
package foo
type I interface {
}
`},
		}),
			pkg:  "foo",
			want: map[string]string{"\"foo\".I": "i"},
		},
		// interface satisfied only within package, value receiver
		{ctx: fakeContext(map[string][]string{
			"foo": {`
package foo
type i interface{
F()
}
type t int
func (t) F() {}
var _ i = t(0)
`},
		}),
			pkg: "foo",
			// should only rename the interface, the concret method
			// will be renames are part of the satisfy relationship
			want: map[string]string{"(\"foo\".i).F": "f"},
		},
		// interface satisfied only within package, pointer receiver
		{ctx: fakeContext(map[string][]string{
			"foo": {`
package foo
type i interface{
F()
}
type t struct {
x int
}
func (*t) F() {}
var _ i = &t{0}
`},
		}),
			pkg: "foo",
			// should only rename the interface, the concret method
			// will be renames are part of the satisfy relationship
			want: map[string]string{"(\"foo\".i).F": "f"},
		},
		// interface satisfied by struct type
		{ctx: fakeContext(map[string][]string{
			"foo": {`
package foo
type I interface {
F()
}
`},
			"bar": {`
package bar
import "foo"
type t int
func (t) F() {}
var _ foo.I = t(0)
`},
		}),
			pkg: "foo",
		},
		// non interface method should still be unexported #17
		{ctx: fakeContext(map[string][]string{
			"foo": {`
package foo
type I interface {
F()
}
`},
			"bar": {`
package bar
import "foo"
type t int
func (t) F() {}
func (t) G() {}
var _ foo.I = t(0)
`},
		}),
			pkg:  "bar",
			want: map[string]string{`("bar".t).G`: "g"},
		},
		// interface satisfied by interface
		{ctx: fakeContext(map[string][]string{
			"foo": {`
package foo
type I interface {
F()
}
`},
			"bar": {`
package bar
import "foo"
type j interface {
foo.I
G()
}
type t int
func (t) F() {}
var _ foo.I = t(0)
`},
		}),
			pkg:  "bar",
			want: map[string]string{"(\"bar\".j).G": "g"},
		},
		// interface used in typeswitch
		{ctx: fakeContext(map[string][]string{
			"foo": {`
package foo
type I interface {
F() int
}
`},
			"bar": {`
package bar
import "foo"
func f(z interface{}) {
		switch y := z.(type) {
				case foo.I:
						print(y.F())
				default:
						print(y)
		}
}
`},
		}),
			pkg: "foo",
		},
		// interface used by function
		{ctx: fakeContext(map[string][]string{
			"foo": {`
package foo
type I interface {
F() int
}
`},
			"bar": {`
package bar
import "foo"
func f(y foo.I) int {
return y.F()
}
`},
		}),
			pkg: "foo",
		},
	} {
		// test body
		unexporter, err := New(test.ctx, test.pkg)
		if err != nil {
			t.Fatal(err)
		}
		cmds := unexporter.Identifiers
		if len(cmds) > 1 {
			if len(test.want) != len(cmds) {
				t.Errorf("expected %d renaming, got %v", len(test.want), cmds)
			}
			var concated string
			for k := range cmds {
				concated += formatCmd(map[string]string{unexporter.Qualifier(k): k.Name()})
			}
			for k, v := range test.want {
				want := map[string]string{k: v}
				if !strings.Contains(concated, formatCmd(want)) {
					t.Errorf("command %s is not returned", formatCmd(want))
				}
			}
		} else {
			if len(test.want) > 0 {
				if len(cmds) == 0 {
					t.Errorf("expected %s, got none", formatCmd(test.want))
				} else {
					arg := make(map[string]string)
					for obj := range cmds {
						arg[unexporter.Qualifier(obj)] = obj.Name()
					}
					if formatCmd(arg) != formatCmd(test.want) {
						t.Errorf("expected %s, got %s", formatCmd(test.want), formatCmd(arg))
					}
				}
			} else {
				if len(cmds) > 0 {
					t.Errorf("expected no renaming, got\n %v", cmds)
				}
			}
		}
	}
}

func TestUnusedObjectsSorted(t *testing.T) {
	for _, test := range []struct {
		ctxt *build.Context
		pkg  string
		want []string
	}{
		{
			ctxt: main(`
package main
type S struct {
X int
}
`),
			pkg:  "main",
			want: []string{"X", "S"},
		},
		{
			ctxt: main(`
package main
type I interface {
F() int
}
`),
			pkg:  "main",
			want: []string{"F", "I"},
		},
	} {
		u, err := New(test.ctxt, test.pkg)
		if err != nil {
			t.Fatal(err)
		}
		var got []string
		for _, o := range u.UnusedObjectsSorted() {
			got = append(got, o.Name())
		}
		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("expected %v, got %v", test.want, got)
		}
	}
}

// ---------------------------------------------------------------------

// Simplifying wrapper around buildutil.FakeContext for packages whose
// filenames are sequentially numbered (%d.go).  pkgs maps a package
// import path to its list of file contents.
func fakeContext(pkgs map[string][]string) *build.Context {
	pkgs2 := make(map[string]map[string]string)
	for path, files := range pkgs {
		filemap := make(map[string]string)
		for i, contents := range files {
			filemap[fmt.Sprintf("%d.go", i)] = contents
		}
		pkgs2[path] = filemap
	}
	return buildutil.FakeContext(pkgs2)
}

// helper for single-file main packages with no imports.
func main(content string) *build.Context {
	return fakeContext(map[string][]string{"main": {content}})
}

func formatCmd(pair map[string]string) string {
	for k, _ := range pair {
		return k
	}
	return ""
}
