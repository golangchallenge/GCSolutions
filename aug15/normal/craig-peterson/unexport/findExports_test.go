package main

import (
	"os"
	"os/exec"
	"testing"
)

func deleteTestDir() {
	cmd := exec.Command("rm", "-rf", "testPkg")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}
func copyTestDir() {
	deleteTestDir()
	cmd := exec.Command("cp", "-r", "testpkgTpl", "testPkg")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func TestMain(m *testing.M) {
	copyTestDir()
	status := m.Run()
	deleteTestDir() // comment out this line if you want testPkg dir to stick around
	os.Exit(status)
}

func TestGetsExportedVars(t *testing.T) {
	tPackage, fset := ParsePackage("testpkg", "testPkg")
	candidates := FindAllExports(tPackage, fset)
	foundNames := map[string]bool{}
	for _, c := range candidates {
		foundNames[c.DisplayName] = true
	}
	expectedNames := []string{
		"MFoo()",
		"MFoo2()",
		"T1",
		"T1.MFoo()",
		"T1.MT1Bar()",
		"T1.MT1Foo()",
		"T2",
		"T3",
		"T3.Foo",
		"T5",
		"T5.Foo",
		"T5.Bar",
		"T6",
		"T6.AnonInner",
		"T6.AnonInner.Foo",
		"T6.AnonInner.Bar",
		"T6.AnonInner.Qwerty",
		"T6.AnonInner.Qwerty.Super",
		"T7",
		"T8",
	}
	for _, e := range expectedNames {
		if !foundNames[e] {
			t.Errorf("Expected to find %s, but did not.", e)
		} else {
			delete(foundNames, e)
		}
	}
	for n, _ := range foundNames {
		t.Errorf("Found unexpected exported name: %s", n)
	}

}
