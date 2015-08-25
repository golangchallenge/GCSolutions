package testinterface

//UsedInterface tesing used interface
type UsedInterface interface {
	SayHello() string
}

//UnusedInterface testing unused interface
type UnusedInterface interface{}

/*
TODO:
*/

//SortImpl using as implementation for sort.Interface
type SortImpl struct {
	Arr []int
}

func (sort *SortImpl) Len() int {
	return len(sort.Arr)
}

func (sort *SortImpl) Less(i int, j int) bool {
	return sort.Arr[i] < sort.Arr[j]
}

func (sort *SortImpl) Swap(i int, j int) {
	t := sort.Arr[i]
	sort.Arr[i] = sort.Arr[j]
	sort.Arr[j] = t
}
