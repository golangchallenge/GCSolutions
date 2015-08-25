package testpkg

import (
	"sync"
)

type T1 int
type (
	T2 byte
	T3 struct {
		Foo string //exported field on exported type
		bar int    //unexported field on exported type
	}
	t4 struct {
		Foo string //exported field on unexported type
		bar T1
	}
	t5 byte
	//T5 can't be renamed due to conflict
	T5 struct {
		Foo int // make sure this won't conflict in renaming
		Bar *T3
		baz *T6
	}
	T6 struct {
		AnonInner struct {
			Foo    T5
			Bar    *T6
			baz    T1
			Qwerty struct {
				Super int
				deep  string
			}
		}
		unexportAnonInner struct {
			Foo int
		}
	}
	T7 struct {
		// Embedded fields should not be considered, nor should their methods
		sync.Mutex
		*T1
		T2
	}
	T8 struct {
		// don't unexport tagged struct fields or json and friends will stop working
		X string `json:"Foo"`
	}
)
