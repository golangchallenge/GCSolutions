package testrename

//UsedStruct is a struct that is exported and used
type UsedStruct struct {
	//Used field
	UsedField string
	//UnusedField not used field
	UnusedField string
}

//UsedMethod used method
func (usedStruct *UsedStruct) UsedMethod() {
	usedStruct.UsedInPackageMethod()
}

//UnusedMethod unused method
func (*UsedStruct) UnusedMethod() {
}

//UsedInPackageMethod using this method only inside package
func (*UsedStruct) UsedInPackageMethod() {}

//UnusedStructConflict is a test struct that is exported, not used
//and can't be renamed because this type is exist
type UnusedStructConflict struct{}

type unusedStructConflict struct{}

var (
	//UnusedVarConflict var that conflicts with func
	UnusedVarConflict string
)

func unusedVarConflict() string {
	return "No vars!"
}
