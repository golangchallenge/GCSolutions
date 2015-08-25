package namecollision

func Asdf() {
}

func asdf() {
}

type someHandler interface {
	ServeHTTP(int, string)
	serveHTTP(int, string)
}

type StructTest struct {
	abc bool
	Abc string
}

func (t *StructTest) Foo() {}
func (t *StructTest) foo() {}

var a, A = 2, 4.3

const Derp = "fii"
const derp = "fii"

type StructSlice []StructTest
type structSlice []StructTest
