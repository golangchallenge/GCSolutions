package main

import "testing"

func validateExports(expectedTable []string, pkg *_package, t *testing.T) {
	for _, expectedKey := range expectedTable {
		if _, exists := pkg.identifiers[expectedKey]; !exists {
			t.Logf("Expected to find key '%s' in map:", expectedKey)
			for key := range pkg.identifiers {
				t.Logf(" - %s\n", key)
			}
			t.Fatalf("Expected to find key '%s' in map %#v", expectedKey, pkg.identifiers)
		}

		if len(pkg.identifiers) != len(expectedTable) {
			t.Logf("Does not equal in size! %d != %d", len(expectedTable), len(pkg.identifiers))
			for key := range pkg.identifiers {
				t.Logf(" - %s\n", key)
			}
		}
	}
}

func TestStruct(t *testing.T) {
	expectedTable := []string{
		"\"./_testdata/_struct\".Struct",
		"\"./_testdata/_struct\".Struct.A",
		"\"./_testdata/_struct\".Struct.S",
		"\"./_testdata/_struct\".Struct.Foo",
		"\"./_testdata/_struct\".Struct.Bar",
	}

	pkg, err := newPkg("./_testdata/_struct")
	if err != nil {
		t.Fatal(err)
	}
	validateExports(expectedTable, pkg, t)
}

func TestConst(t *testing.T) {
	expectedTable := []string{
		"\"./_testdata/_const\".A",
	}

	pkg, err := newPkg("./_testdata/_const")
	if err != nil {
		t.Fatal(err)
	}
	validateExports(expectedTable, pkg, t)
}

func TestVar(t *testing.T) {
	expectedTable := []string{
		"\"./_testdata/_var\".A",
	}

	pkg, err := newPkg("./_testdata/_var")
	if err != nil {
		t.Fatal(err)
	}
	validateExports(expectedTable, pkg, t)
}

func TestFunc(t *testing.T) {
	expectedTable := []string{
		"\"./_testdata/_func\".A",
	}

	pkg, err := newPkg("./_testdata/_func")
	if err != nil {
		t.Fatal(err)
	}
	validateExports(expectedTable, pkg, t)

}

func TestStructUses(t *testing.T) {

	fooPkg, err := newPkg("./_testdata/_imports/foo")
	if err != nil {
		t.Fatal(err)
	}

	testPkg, err := newPkg("./_testdata/_imports")
	if err != nil {
		t.Fatal(err)
	}

	testPkg.calculateUsesOf(fooPkg)

	for _, v := range fooPkg.identifiers {
		if len(v.usedBy) != 1 {
			t.Fatal("Expected 1 i length")
		}
	}

}
