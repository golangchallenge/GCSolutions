package main

func (b *board) SolveSAT() error {
	satInput := b.getSAT()
	satSolver, err := NewSAT(satInput, b.countSolutions, b.maxSolutions)
	if err != nil {
		return err
	}

	slns := satSolver.Solve()
	if slns == nil || len(slns) == 0 {
		return NewErrUnsolvable("could not solve with SAT %v")
	}

	if b.countSolutions {
		b.solutionCount = len(slns)
	}

	sln1 := slns[0]
	for _, setvar := range sln1.SetVars {
		k := int(setvar.VarNum)
		v := setvar.Value
		if v {
			r := k/100 - 1
			c := (k%100)/10 - 1
			pos := r*9 + c
			if b.solved[pos] == 0 {
				val := k % 10
				b.SolvePositionNoValidate(pos, uint(val))
			}
		}
	}

	if err = b.Validate(); err != nil {
		return err
	}

	b.AddLog("SAT", nil, "Solved with SAT")

	return nil
}
