package main

import (
	"bytes"
	"fmt"
)

func getRCV(r int, c int, v int) int {
	return r*100 + c*10 + v
}

func satrcToPos(r int, c int) int {
	return (r-1)*9 + (c - 1)
}

func (b *board) hasHint(satR int, satC int, v int) bool {
	pos := satrcToPos(satR, satC)
	if b.solved[pos] != 0 {
		if b.solved[pos] == uint(v) {
			return true
		}
		return false
	}
	mask := uint(1 << (uint(v) - 1))
	if b.blits[pos]&mask == mask {
		return true
	}
	return false
}

func (b *board) getSAT() string {
	var clauses int

	singlebuf := bytes.NewBufferString("")
	buf := bytes.NewBufferString("")
	longbuf := bytes.NewBufferString("")

	// apply known values

	// each row
	for r := 1; r <= 9; r++ {
		// each column
		for c := 1; c <= 9; c++ {
			offset := (r-1)*9 + (c - 1)
			v := int(b.solved[offset])
			if v != 0 {
				cur := getRCV(r, c, v)
				singlebuf.WriteString(fmt.Sprintf("%d 0\n", cur))
				clauses++
			}
		}
	}

	// apply known negative values
	// each row
	for r := 1; r <= 9; r++ {
		// each column
		for c := 1; c <= 9; c++ {
			offset := (r-1)*9 + (c - 1)
			if b.solved[offset] != 0 {
				continue
			}
			for v := uint(1); v <= 9; v++ {
				mask := uint(1 << (v - 1))
				if b.blits[offset]&mask != mask {
					cur := getRCV(r, c, int(v))
					singlebuf.WriteString(fmt.Sprintf("-%d 0\n", cur))
					clauses++
				}
			}
		}
	}

	// not sure this RCV block will help (update: it definitely does).
	// it indicates each cell must have one of the values 1-9.

	// each row
	for r := 1; r <= 9; r++ {
		// each column
		for c := 1; c <= 9; c++ {
			// each value
			for v := 1; v <= 9; v++ {
				cur := getRCV(r, c, v)
				longbuf.WriteString(fmt.Sprintf("%d ", cur))
			}
			longbuf.WriteString("0\n")
			clauses++
		}
	}

	// each value
	for v := 1; v <= 9; v++ {
		// each row
		for r := 1; r <= 9; r++ {
			// each column
			for c := 1; c <= 9; c++ {
				cur := getRCV(r, c, v)
				longbuf.WriteString(fmt.Sprintf("%d ", cur))
			}
			longbuf.WriteString("0\n")
			clauses++

			// each combination of two values
			for c := 1; c <= 9; c++ {
				cur := getRCV(r, c, v)
				for c2 := c + 1; c2 <= 9; c2++ {
					cur2 := getRCV(r, c2, v)

					buf.WriteString(fmt.Sprintf("-%d -%d 0\n", cur, cur2))
					clauses++
				}
			}
		}

		// each column
		for c := 1; c <= 9; c++ {
			// each row
			for r := 1; r <= 9; r++ {
				cur := getRCV(r, c, v)

				longbuf.WriteString(fmt.Sprintf("%d ", cur))
			}
			longbuf.WriteString("0\n")
			clauses++
			// each combination of two values
			for r := 1; r <= 9; r++ {
				cur := getRCV(r, c, v)
				for r2 := r + 1; r2 <= 9; r2++ {
					cur2 := getRCV(r2, c, v)

					buf.WriteString(fmt.Sprintf("-%d -%d 0\n", cur, cur2))
					clauses++
				}
			}
		}

		// each box
		for b := 0; b < 9; b++ {
			rOffset := (b / 3) * 3
			cOffset := (b % 3) * 3
			// each row
			for r := rOffset + 1; r <= rOffset+3; r++ {
				// each column
				for c := cOffset + 1; c <= cOffset+3; c++ {
					cur := getRCV(r, c, v)

					longbuf.WriteString(fmt.Sprintf("%d ", cur))
				}
			}
			longbuf.WriteString("0\n")
			clauses++
			// each combination of two values
			// each row
			cheat := make(map[string]interface{})
			for r := rOffset + 1; r <= rOffset+3; r++ {
				// each column
				for c := cOffset + 1; c <= cOffset+3; c++ {
					cur := getRCV(r, c, v)
					// TODO: being lazy with the map to detect dupes
					for r2 := rOffset + 1; r2 <= rOffset+3; r2++ {
						for c2 := cOffset + 1; c2 <= cOffset+3; c2++ {
							// already checked row/col constraint
							if r == r2 || c == c2 {
								continue
							}
							cur2 := getRCV(r2, c2, v)

							clause := fmt.Sprintf("-%d -%d 0\n", cur, cur2)
							clauseReversed := fmt.Sprintf("-%d -%d 0\n", cur2, cur)
							if _, ok := cheat[clauseReversed]; ok {
								continue
							}
							cheat[clause] = struct{}{}

							buf.WriteString(clause)
							clauses++
						}
					}
				}
			}
		}
	}

	// each row
	for r := 1; r <= 9; r++ {
		// each column
		for c := 1; c <= 9; c++ {
			// each pair of values (single value per cell)
			for v1 := 1; v1 <= 9; v1++ {
				cur := getRCV(r, c, v1)
				for v2 := v1 + 1; v2 <= 9; v2++ {
					cur2 := getRCV(r, c, v2)

					buf.WriteString(fmt.Sprintf("-%d -%d 0\n", cur, cur2))
					clauses++
				}
			}
		}
	}

	header := fmt.Sprintf("p cnf %d %d", 9*9*9, clauses)
	input := fmt.Sprintf("%s\n%s%s%s", header, singlebuf, longbuf, buf)
	//ioutil.WriteFile("sat.cnf", []byte(input), 0644)
	return input
}
