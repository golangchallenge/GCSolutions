package foo

const FooConst = "FOO"

var FooVar = "Foo"

type FooStruct struct{}

func (f *FooStruct) FooPtrFunc() string {
	return FooConst
}

func (f FooStruct) FooFunc() string {
	return FooVar
}

func Foo() string {
	return "Foo"
}
