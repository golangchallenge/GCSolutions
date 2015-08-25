package teststruct

//ComplexType is just another struct to test vars in structs
type ComplexType struct{}

//UsedStruct is a struct that is exported and used
type UsedStruct struct {
	//Used field
	UsedField string
	//Used field with pointer
	UsedFieldPointer *ComplexType
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

//UnusedStruct is a test struct that is exported and not used
type UnusedStruct struct {
}
