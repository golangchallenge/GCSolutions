package testpkg

func MFoo() {}
func mBar() {}

func MFoo2() {}
func mFoo2() {}

func (t *T1) MT1Foo() {}
func (t T1) MT1Bar()  {}
func (t *T1) MFoo()   {}

func (t *T1) mt1foo() {}
func (t T1) mt1bar()  {}
