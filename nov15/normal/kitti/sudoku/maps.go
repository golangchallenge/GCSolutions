package main

func resetMaps() {
	rows = make(map[int]map[string]struct{})
	cols = make(map[int]map[string]struct{})
	blks = make(map[int]map[string]struct{})
	allCells = make(map[int]map[string]struct{})
	for i := 0; i < 9; i++ {
		for j := 0; j < 9; j++ {
			cell := grid[i][j]
			addToRow(i, cell)
			addToCol(j, cell)
			addToBlk(i, j, cell)
		}
	}
	possibilities()
}

func addToRow(i int, v string) {
	if v == "_" {
		return
	}
	a := rows[i]
	if a == nil {
		a = make(map[string]struct{})
	}
	a[v] = struct{}{}
	rows[i] = a
}

func addToCol(j int, v string) {
	if v == "_" {
		return
	}
	a := cols[j]
	if a == nil {
		a = make(map[string]struct{})
	}
	a[v] = struct{}{}
	cols[j] = a
}

func addToBlk(i int, j int, v string) {
	if v == "_" {
		return
	}
	switch i {
	case 0, 1, 2:
		switch j {
		case 0, 1, 2:
			a := blks[0]
			if a == nil {
				a = make(map[string]struct{})
			}
			a[v] = struct{}{}
			blks[0] = a
		case 3, 4, 5:
			a := blks[1]
			if a == nil {
				a = make(map[string]struct{})
			}
			a[v] = struct{}{}
			blks[1] = a
		case 6, 7, 8:
			a := blks[2]
			if a == nil {
				a = make(map[string]struct{})
			}
			a[v] = struct{}{}
			blks[2] = a
		}
	case 3, 4, 5:
		switch j {
		case 0, 1, 2:
			a := blks[3]
			if a == nil {
				a = make(map[string]struct{})
			}
			a[v] = struct{}{}
			blks[3] = a
		case 3, 4, 5:
			a := blks[4]
			if a == nil {
				a = make(map[string]struct{})
			}
			a[v] = struct{}{}
			blks[4] = a
		case 6, 7, 8:
			a := blks[5]
			if a == nil {
				a = make(map[string]struct{})
			}
			a[v] = struct{}{}
			blks[5] = a
		}
	case 6, 7, 8:
		switch j {
		case 0, 1, 2:
			a := blks[6]
			if a == nil {
				a = make(map[string]struct{})
			}
			a[v] = struct{}{}
			blks[6] = a
		case 3, 4, 5:
			a := blks[7]
			if a == nil {
				a = make(map[string]struct{})
			}
			a[v] = struct{}{}
			blks[7] = a
		case 6, 7, 8:
			a := blks[8]
			if a == nil {
				a = make(map[string]struct{})
			}
			a[v] = struct{}{}
			blks[8] = a
		}

	}
}
