package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
)

const lenMask = 0xF800000000000000
const len1 = 0x800000000000000
const len2 = 0x1000000000000000

const shiftInterval = uint(11)
const initialShift = uint(44)
const lenShift = uint(59)

const signMask = 0x400
const signedValMask = 0x7FF
const absValMask = 0x3FF

var satisfied = &[2]uint64{0xFF, 0xFF}

// SAT represents a list of clauses. When solved, the SetVars field
// represents a single solution to the SAT problem.
type SAT struct {
	SetVars               []SetVar
	Clauses               [][2]uint64
	FindMultipleSolutions bool
	MaxSolutions          int
}

// SetVar represents a variable set to either true or false used to
// satisfy the list of clauses.
type SetVar struct {
	VarNum uint64
	Value  bool
}

// NewSAT initializes a new SAT struct.
//
//  input:
//      the SAT input including header.
//
//      for example, the CNF formula:
//      (a) ? (b) ? (¬b ? c)
//
//      has the input value:
//
//      p cnf 3 3
//      1 0
//      2 0
//      -1 3 0
//
//      where the first 3 in the header is the number of variables
//      and the second 3 is the number of clauses which follow.
//      all clause lines must end with a literal 0.
//
//      more information can be found at this address:
//      http://www.satcompetition.org/2004/format-solvers2004.html
//
//  findMultipleSolutions:
//      set to true to find multiple solutions; otherwise, Solve will
//      stop after the first solution is found (if any).
//
//  maxSolutions:
//      set to limit the number of possible solutions returned. has no
//      effect if findMultipleSolutions is false.
func NewSAT(input string, findMultipleSolutions bool, maxSolutions int) (*SAT, error) {
	// load CNF input
	sr := strings.NewReader(input)
	r := bufio.NewReader(sr)

	line, err := r.ReadString('\n')
	if err != nil {
		return nil, err
	}
	for strings.HasPrefix(line, "c") {
		// skip comment
		if line, err = r.ReadString('\n'); err != nil {
			return nil, err
		}
	}
	// p cnf # # // variables, clauses
	if !strings.HasPrefix(line, "p cnf ") {
		return nil, fmt.Errorf("expected first non-comment line in format\"p cnf # #\": %q", line)
	}
	// get # of variables, # of clauses from CNF header
	line = strings.Trim(line[len("p cnf"):], " \r\n\t")
	strParts := strings.SplitN(line, " ", -1)
	parts, err := getIntArray(strParts, false, false)
	if err != nil {
		return nil, fmt.Errorf("expected first non-comment line in format \"p cnf # #\": %q %q", line, err)
	}
	if len(parts) != 2 {
		return nil, fmt.Errorf("expected first non-comment line in format \"p cnf # #\": %q", line)
	}
	variableCount := parts[0]
	clauseCount := parts[1]
	if variableCount < 0 || clauseCount < 0 {
		return nil, fmt.Errorf("variable and clause count must be non-negative: %q", line)
	}

	// TODO: validate variable, clause count
	s := &SAT{
		Clauses:               make([][2]uint64, clauseCount),
		FindMultipleSolutions: findMultipleSolutions,
		MaxSolutions:          maxSolutions,
	}

	// get clauses
	if line, err = r.ReadString('\n'); err != nil && err != io.EOF {
		// TODO: if clauses = 0, this is okay.
		return nil, err
	}

	for i := 0; line != ""; i++ {
		line = strings.Trim(line, " \r\n\t")
		if strings.HasPrefix(line, "c ") {
			// skip comment
			i--
			continue
		}
		strParts = strings.SplitN(line, " ", -1)
		parts, err := getIntArray(strParts, true, true)
		if err != nil {
			return nil, fmt.Errorf("error parsing line: %q", line)
		}
		s.Clauses[i] = intArrayToBin(parts)

		if line, err = r.ReadString('\n'); err != nil && err != io.EOF {
			return nil, err
		}
	}

	return s, nil
}

func intArrayToBin(list []int) [2]uint64 {
	var bin [2]uint64
	bin[0] = uint64(len(list)) << lenShift
	j := 0
	shift := initialShift
	for i := 0; i < len(list); i++ {
		val := abs(list[i])
		if list[i] < 0 {
			val |= signMask // add sign
		}
		bin[j] |= uint64(val << shift)

		if shift == 0 {
			shift = initialShift
			j++
		} else {
			shift -= shiftInterval
		}
	}
	return bin
}

func getIntArray(values []string, sortValues bool, trimEnd bool) ([]int, error) {
	list := make([]int, len(values))
	for i := 0; i < len(values); i++ {
		v, err := strconv.Atoi(values[i])
		if err != nil {
			return nil, fmt.Errorf("idx: %d unexpected format: %q", i, values[i])
		}
		list[i] = v
	}

	if trimEnd {
		if list[len(list)-1] != 0 {
			return nil, errors.New("error parsing line, must end in \"0\"")
		}
		list = list[:len(list)-1]
	}

	if sortValues {
		sort.Ints(list)
	}

	return list, nil
}

/*func (s *sat) getRemainingVars() []int {
	vars := make(map[int]interface{})

	for _, clause := range s.Clauses {
		length := int(clause[0]&lenMask) >> lenShift
		if length == 0 {
			continue
		}
		shift := initialShift
		cur := clause[0]
		for i := 0; i < length; i++ {
			curval := (cur >> shift) & signedValMask
			neg := (curval & signMask) == signMask
			curval &= absValMask

			var val int
			val = int(curval)
			if neg {
				val = -val
			}
			vars[val] = struct{}{}

			if shift == 0 {
				shift = initialShift
				cur = clause[1]
			} else {
				shift -= shiftInterval
			}
		}
	}

	var list []int
	for k, _ := range vars {
		list = append(list, k)
	}
	sort.Ints(list)
	return list
}*/

func (s *SAT) getAllSingleVarClauses() []SetVar {
	check := make(map[uint64]struct{})
	var list []SetVar
	var val uint64
	var on bool

	for _, clause := range s.Clauses {
		c := clause[0]
		length := c & lenMask
		if length == len1 {
			val = c >> initialShift
			on = (val&signMask == 0)
			val = val & absValMask

			if _, ok := check[c]; !ok {
				continue
			}
			check[c] = struct{}{}

			list = append(list, SetVar{VarNum: val, Value: on})
		}
	}
	return list
}

func (s *SAT) getNextSingleVar() (*uint64, *bool) {
	var val uint64

	// find a clause with a single variable
	for _, clause := range s.Clauses {
		length := clause[0] & lenMask
		if length == len1 {
			val = clause[0] >> initialShift
			on := (val&signMask == 0)
			val = val & absValMask

			return &val, &on
		}
	}

	return nil, nil
}

func (s *SAT) getNextVar() (*uint64, *bool) {
	singleVal, on := s.getNextSingleVar()
	if singleVal != nil {
		return singleVal, on
	}

	var val uint64

	for _, clause := range s.Clauses {
		length := clause[0] & lenMask
		if length == len2 {
			val = (clause[0] >> initialShift) & absValMask
			return &val, nil
		}
	}

	// let's try the first variable from the first clause
	val = (s.Clauses[0][0] >> initialShift) & absValMask
	return &val, nil
}

// Solve solves a boolean satisfiability problem.
// If the set of clauses are unsatisfiable nil is returned.
// If FindMultipleSolutions is true a slice of solutions is returned.
// If MaxSolutions is specified the number of solution is limited to this number.
func (s *SAT) Solve() []*SAT {
	val, on := s.getNextVar()
	var s2, s3 []*SAT
	if on != nil {
		s2 = set(s, *val, *on)
	} else {
		s2 = set(s, *val, false)
		if s2 == nil || s.FindMultipleSolutions {
			s3 = set(s, *val, true)
		}
	}

	var final []*SAT
	if s2 != nil {
		final = append(final, s2...)
	}
	if s3 != nil {
		final = append(final, s3...)
	}

	if len(final) == 0 {
		return nil
	}

	return final
}

func (s *SAT) solveSingleVarClauses() (*SAT, []SetVar) {
	var bigList []SetVar

	list := s.getAllSingleVarClauses()
	for len(list) != 0 {
		for _, item := range list {
			var clauses [][2]uint64
			for _, clause := range s.Clauses {
				newClause := up(&clause, item.VarNum, item.Value)
				if newClause != nil {
					if newClause != satisfied {
						clauses = append(clauses, *newClause)
					}
				} else {
					return nil, nil
				}
			}

			s.Clauses = clauses
			if len(s.Clauses) == 0 {
				break
			}
		}
		bigList = append(bigList, list...)
		if len(s.Clauses) == 0 {
			break
		}
		list = s.getAllSingleVarClauses()
	}

	return s, bigList
}

func set(s1 *SAT, v uint64, isOn bool) []*SAT {
	s2 := &SAT{FindMultipleSolutions: s1.FindMultipleSolutions, MaxSolutions: s1.MaxSolutions}

	for _, clause := range s1.Clauses {
		newClause := up(&clause, v, isOn)
		if newClause != nil {
			if newClause != satisfied {
				s2.Clauses = append(s2.Clauses, *newClause)
			}
		} else {
			return nil
		}
	}

	if len(s2.Clauses) == 0 {
		s2.SetVars = append(s2.SetVars, SetVar{VarNum: v, Value: isOn})
		return []*SAT{s2}
	}

	var bigList []SetVar
	if s2.FindMultipleSolutions {
		s2, bigList = s2.solveSingleVarClauses()
		if s2 == nil {
			return nil
		}

		if len(s2.Clauses) == 0 {
			s2.SetVars = append(s2.SetVars, bigList...)
			s2.SetVars = append(s2.SetVars, SetVar{VarNum: v, Value: isOn})
			return []*SAT{s2}
		}
	}

	val, on := s2.getNextVar()

	var final []*SAT
	if on != nil {
		s3 := set(s2, *val, *on)
		if s3 != nil {
			final = append(final, s3...)
		}
	} else {
		s3 := set(s2, *val, false)
		if s3 != nil {
			final = append(final, s3...)
		}

		if s2.FindMultipleSolutions && len(final) >= s2.MaxSolutions {
			// skip
		} else {
			if s3 == nil || s2.FindMultipleSolutions {
				s4 := set(s2, *val, true)
				if s4 != nil {
					final = append(final, s4...)
				}
			}
		}
	}

	if len(final) == 0 {
		return nil
	}

	for _, item := range final {
		if s1.FindMultipleSolutions {
			item.SetVars = append(item.SetVars, bigList...)
		}
		item.SetVars = append(item.SetVars, SetVar{VarNum: v, Value: isOn})
	}

	return final
}

func up(clause *[2]uint64, v uint64, isOn bool) *[2]uint64 {
	var idx int
	for {
		if idx = indexOfValue(clause, v); idx != -1 {
			if isOn {
				return satisfied
			}
			clause = cut(clause, idx)
		} else if idx = indexOfValue(clause, v|signMask); idx != -1 {
			if !isOn {
				return satisfied
			}
			clause = cut(clause, idx)
		} else {
			return clause
		}

		if clause == nil {
			return nil
		}
	}
}

func cut(clause *[2]uint64, idx int) *[2]uint64 {
	if clause[0]&lenMask == len1 {
		return nil
	}

	var newClause [2]uint64

	length := int((clause[0] & lenMask) >> lenShift)
	newClause[0] = uint64(length-1) << lenShift

	shift := initialShift
	shift2 := initialShift
	j, k := 0, 0
	cur := clause[0]
	for i := 0; i < length; i++ {
		if i != idx {
			curval := (cur >> shift) & signedValMask
			newClause[k] |= curval << shift2

			if shift2 == 0 {
				shift2 = initialShift
				k++
			} else {
				shift2 -= shiftInterval
			}
		}

		if shift == 0 {
			shift = initialShift
			j++
			cur = clause[j]
		} else {
			shift -= shiftInterval
		}
	}
	return &newClause
}

func indexOfValue(clause *[2]uint64, val uint64) int {
	length := int((clause[0] & lenMask) >> lenShift)
	if length == 0 {
		return -1
	}

	shift := initialShift
	cur := clause[0]
	for i := 0; i < length; i++ {
		curval := (cur >> shift) & signedValMask
		if curval == val {
			return i
		}

		if shift == 0 {
			shift = initialShift
			cur = clause[1]
		} else {
			shift -= shiftInterval
		}
	}
	return -1
}

/*
// debug function
func clauseToIntArray(clause [2]uint64) []int {
	length := int((clause[0] & lenMask) >> lenShift)
	if length == 0 {
		return []int{}
	}

	var list []int

	shift := initialShift
	cur := clause[0]
	for i := 0; i < length; i++ {
		curval := (cur >> shift) & signedValMask

		on := curval&signMask == 0
		curval &= absValMask
		if !on {
			curval = -curval
		}

		list = append(list, int(curval))

		if shift == 0 {
			shift = initialShift
			cur = clause[1]
		} else {
			shift -= shiftInterval
		}
	}

	return list
}
*/
