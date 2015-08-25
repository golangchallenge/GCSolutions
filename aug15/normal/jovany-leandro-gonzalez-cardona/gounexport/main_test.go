package main


import (
	"testing"
	"bytes"
	"flag"
	"regexp"
	"strings"
)

const (
	testDir = "./testdata"
)
var tests = []struct {
	name string
	args []string
	expects []string
}{
	{
		"query collision keyword",
		[]string{
			testDir,
			"Type@const",
		},
		[]string{
			"Type@const !!collision",
		},
	},
	{
		"query exported constant",
		[]string{
			testDir,
			"^pkg.ExportedConstant@const$",
		},
		[]string{
			"pkg.ExportedConstant@const",
		},


	},

	{
		"unexport constant",
		[]string{
			"-w=true",
			"-s",
			testDir,
			"^pkg.ExportedConstant@const$",
		},
		[]string{
			"Unexported ExportedConstant from pkg.ExportedConstant@const",
		},
	},
	
	{
		"query exported all constant",
		[]string{
			testDir,
			".@const$",
		},
		[]string{
			"pkg.ExportedConstant@const",
			"pkg.ConstOne",
			"pkg.ConstTwo",
			"pkg.Type",
		},
	},

	{
		"unexport all constant",
		[]string{
			"-w=true",
			"-s",
			testDir,
			"@const$",
		},
		[]string{
			"^Unexported ExportedConstant from pkg.ExportedConstant@const$",
			"^Unexported ConstOne from pkg.ConstOne@const$",
			"^Unexported ConstTwo from pkg.ConstTwo@const$",
			"^Unexported Type from pkg.Type@const$",
			"^sorry collision detected for pkg.Type@const$",
		},
	},

	{
		"exported all variables",
		[]string{
			testDir,
			".@var",
		},
		[]string{
			"pkg.ExportedVariable",
			"pkg.ExportedVariableOne",
			"pkg.ExportedVariableTwo",
			"pkg.VarOne",
			"pkg.VarTwo",
		},
	},

	{
		"unexport all variables",
		[]string{
			"-w=true",
			"-s",
			testDir,
			"pkg\\.[a-zA-Z]+@var$",
		},
		[]string{
			"^Unexported ExportedVariable from pkg.ExportedVariable@var$",
			"^Unexported ExportedVariableOne from pkg.ExportedVariableOne@var$",
			"^Unexported ExportedVariableTwo from pkg.ExportedVariableTwo@var$",
			"^Unexported VarOne from pkg.VarOne@var$",
			"^Unexported VarTwo from pkg.VarTwo@var$",
			"^sorry collision detected for pkg.VarOne@var$", 
		},
	},

	{
		"query struct",
		[]string{
			testDir,
			".@struct",
		},
		[]string{
			"pkg.ExportedType",
		},
	},

	{
		"unexport struct",
		[]string{
			"-w=true",
			"-s",
			testDir,
			".@struct",
		},
		[]string{
			"^Unexported ExportedType from pkg.ExportedType@struct$",
			"^Unexported ExportedTypeTwo from pkg.ExportedTypeTwo@struct$",
		},
	},

	{
		"query struct method",
		[]string{
			testDir,
			"ExportedType.ExportedMethod@method$",
		},
		[]string{
			"pkg.ExportedType.ExportedMethod",
		},
	},

	{
		"unexport struct method",
		[]string{
			"-w=true",
			"-s",
			testDir,
			"ExportedType.ExportedMethod@method$",
		},
		[]string{
			"^Unexported ExportedMethod from pkg.ExportedType.ExportedMethod@method$",
		},
	},
	
	{
		"query struct collision method",
		[]string{
			testDir,
			"ExportedType.ExportedCollisionMethod@method$",
		},
		[]string{
			"pkg.ExportedType.ExportedCollisionMethod@method !!collision",
		},
	},

	{
		"unexport struct collision method",
		[]string{
			"-w=true",
			"-s",
			testDir,
			"ExportedType.ExportedCollisionMethod@method$",
		},
		[]string{
			"^sorry collision detected for pkg.ExportedType.ExportedCollisionMethod@method$",
		},
	},

	
	{
		"query struct field",
		[]string{
			testDir,
			"ExportedType.ExportedField@attr",
		},
		[]string{
			"pkg.ExportedType.ExportedField",
		},
	},

	{
		"unexport struct field",
		[]string{
			"-w=true","-s",
			testDir,
			"ExportedType.ExportedField@attr",
		},
		[]string{
			"^Unexported ExportedField from pkg.ExportedType.ExportedField@attr$",
		},
	},

	{
		"query interface",
		[]string{
			testDir,
			"ExportedInterface@interface$",
		},
		[]string{
			"pkg.ExportedInterface@interface$",
		},
	},

	{
		"unexport interface",
		[]string{
			"-w=true","-s",
			testDir,
			"ExportedInterface@interface$",
		},
		[]string{
			"^Unexported ExportedInterface from pkg.ExportedInterface@interface$",
		},
	},

	{
		"query interface method",
		[]string{
			testDir,
			"ExportedInterface..+@method$",
		},
		[]string{
			"^pkg.ExportedInterface.ExportedMethod@method$",
		},
	},

	{
		"unexport interface method",
		[]string{
			"-w=true","-s",
			testDir,
			"ExportedInterface..+@method$",
		},
		[]string{
			"^Unexported ExportedMethod from pkg.ExportedInterface.ExportedMethod@method",
		},
	},

	{
		"query function",
		[]string{
			testDir,
			"ExportedFunc@func$",
		},
		[]string{
			"^pkg.ExportedFunc@func$",
		},
	},

	{
		"unexport function",
		[]string{
			"-w=true","-s",
			testDir,
			"ExportedFunc@func$",
		},
		[]string{
			"^Unexported ExportedFunc from pkg.ExportedFunc@func",
		},
	},
	
	{
		"query function collision",
		[]string{
			testDir,
			"ExportedCollisionFunc@func",
		},
		[]string{
			"pkg.ExportedCollisionFunc@func !!collision$", 
		},
	},
	{
		"unexport function collision",
		[]string{
			"-w=true", "-s",
			testDir,
			"ExportedCollisionFunc@func",
		},
		[]string{
			"^sorry collision detected for pkg.ExportedCollisionFunc@func$", 
		},
	},
	
}

func TestUnexport(t *testing.T) {

	for _, test := range tests {
		var out bytes.Buffer
		var flags flag.FlagSet
		flag.Set("w", "false")
		flag.Set("s", "true")
		err := do(&out, &flags,test.args)

		lines := strings.Split(out.String(), "\n")
		reusesAt := regexp.MustCompile(".go.+$")
		if len(lines) == 0 || len(lines) == 1 && strings.TrimSpace(lines[0]) == ""{
			t.Errorf("%s: Invalid output", test.name)
		}
		for _, rline := range lines {
			line := strings.TrimSpace(rline)
			if line == "" || strings.HasSuffix(line, ".go")  || strings.HasSuffix(line, "Uses at")	|| reusesAt.MatchString(line) {
				continue
			}
			notMatch := true
			for _, expect := range test.expects {
				rexpect := regexp.MustCompile(expect)

				if rexpect.MatchString(line) {
					notMatch = false
				}
			}
			if notMatch {
				t.Errorf("%s: no match found for %s", test.name, line)
			}
		}

		if err != nil {
			t.Fatal(err)
		}
	}
	
}
