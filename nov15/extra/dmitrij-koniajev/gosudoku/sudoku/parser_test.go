package sudoku

import (
	"fmt"
	"strings"
	"testing"
)

var parserTable = []struct {
	input  string
	result string
	error  error
}{
	{
		`1 _ 3 _ _ 6 _ 8 _
_ 5 _ _ 8 _ 1 2 _
7 _ 9 1 _ 3 _ 5 6
_ 3 _ _ 6 7 _ 9 _
5 _ 7 8 _ _ _ 3 _
8 _ 1 _ 3 _ 5 _ 7
_ 4 _ _ 7 8 _ 1 _
6 _ 8 _ _ 2 _ 4 _
_ 1 2 _ 4 5 _ 7 8`,
		"1_3__6_8__5__8_12_7_91_3_56_3__67_9_5_78___3_8_1_3_5_7_4__78_1_6_8__2_4__12_45_78",
		nil,
	},
	{
		"",
		"",
		fmt.Errorf("line 1: unexpected end of input"),
	},
	{
		"q",
		"",
		fmt.Errorf("line 1, column 1: expected digit or underscore symbol"),
	},
	{
		"1 q",
		"",
		fmt.Errorf("line 1, column 2: expected digit or underscore symbol"),
	},
	{
		"1 2 3 4 _",
		"",
		fmt.Errorf("line 1: unexpected end of input"),
	},
	{
		"1 2 3 4 5 6 7 8 9 _",
		"",
		fmt.Errorf("line 1: too many symbols"),
	},
	{
		`1 2 3 4 5 6 7 8 9
1 2 3 4 5 6 7 8 9 q`,
		"",
		fmt.Errorf("line 2: too many symbols"),
	},
	{
		`1 2 _ 4 5 6 7 8 9
1 2 3 4 5 6 7 _ 9`,
		"",
		fmt.Errorf("line 3: unexpected end of input"),
	},
}

func parseString(s string) (string, error) {
	sr := strings.NewReader(s)
	return ParseReader(sr)
}

func TestParseReader(t *testing.T) {
	t.Parallel()

	for _, tc := range parserTable {
		r, err := parseString(tc.input)
		if r != tc.result {
			t.Errorf("Wrong result:\n%s\nfor test input:\n%s\n", r, tc.input)
		}
		if (tc.error == nil) != (err == nil) {
			t.Errorf("Unexpected error state:\n%s\nfor test input:\n%s\n", err, tc.input)
		} else if err != nil && err.Error() != tc.error.Error() {
			t.Errorf("Wrong error string:\n%s\nfor test input:\n%s\nshould be:\n%s\n", err, tc.input, tc.error)
		}
	}
}
