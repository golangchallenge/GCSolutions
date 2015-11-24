package main

func intersect(a []int, b []int) []int {
	store := make(map[int]interface{})
	var list []int
	for _, i := range a {
		for _, j := range b {
			if i == j {
				if _, ok := store[i]; !ok {
					store[i] = struct{}{}
					list = append(list, i)
				}
			}
		}
	}
	return list
}

func union(a []int, b []int) []int {
	store := make(map[int]interface{})
	var list []int
	for _, val := range a {
		if _, ok := store[val]; !ok {
			store[val] = struct{}{}
			list = append(list, val)
		}
	}
	for _, val := range b {
		if _, ok := store[val]; !ok {
			store[val] = struct{}{}
			list = append(list, val)
		}
	}

	return list
}

func subtract(a []int, b []int) []int {
	store := make(map[int]interface{})
	for _, val := range b {
		store[val] = struct{}{}
	}

	var list []int
	for _, val := range a {
		if _, ok := store[val]; !ok {
			list = append(list, val)
		}
	}

	return list
}

func getPermutations(n int, pickList []int, curList []int) [][]int {
	var output [][]int

	for i := 0; i < len(pickList); i++ {
		list := make([]int, len(curList))
		copy(list, curList)              // get the source list
		list = append(list, pickList[i]) // plus the current element

		if len(list) == n {
			// if this is the length we're looking for...
			output = append(output, list)
		} else {
			// otherwise, call recursively
			perms := getPermutations(n, pickList[i+1:], list)
			if perms != nil {
				for _, v := range perms {
					output = append(output, v)
				}
			}
		}
	}

	return output
}

func abs(x int) int {
	switch {
	case x < 0:
		return -x
	case x == 0:
		return 0 // return correctly abs(-0)
	}
	return x
}
