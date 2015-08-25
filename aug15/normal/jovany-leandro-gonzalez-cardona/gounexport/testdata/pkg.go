package pkg

const Type = 1
const ExportedConstant = 1

const (
	// Comment before ConstOne.
	ConstOne   = 1
	ConstTwo   = 2 // Comment on line with ConstTwo.
	constThree = 3 // Comment on line with constThree.
)

// Variables

var ExportedVariable = 1
var (
	ExportedVariableOne = "a"
	ExportedVariableTwo = "b"
)
var internalVariable = 2

var (
	VarOne   = 1
	VarTwo   = 2
	varThree = 3 
	varOne = 4
)

func ExportedFunc(a int) bool

func internalFunc(a int) bool

func ExportedCollisionFunc (a int) bool
func exportedCollisionFunc (a int) bool

type ExportedType struct {
	ExportedField   int
	unexportedField int
}

func (c ExportedType) ExportedMethod(a int) bool {
	return true
}

func (c ExportedType) unexportedMethod(a int) bool {
	return true
}

func (c ExportedType) ExportedCollisionMethod(a int) bool {
	return true
}

func (c ExportedType) exportedCollisionMethod(a int) bool {
	return true
}


func ExportedTypeConstructor() *ExportedType {
	return nil
}

type ExportedTypeTwo struct {
        ExportedField   int
        unexportedField int
}

func (c ExportedTypeTwo) ExportedMethod(a int) bool {
        return true
}


type ExportedInterface interface {
	ExportedMethod(int) bool
}

func DoExportedInterface(w ExportedInterface) {
	w.ExportedMethod(0)
}

type unexportedType int

func (c unexportedType) ExportedMethod() bool {
	return true
}

func (c unexportedType) unexportedMethod() bool {
	return true
}
