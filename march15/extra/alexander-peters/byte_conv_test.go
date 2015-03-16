package drum

import (
	"testing"
)

func TestByteToBool(t *testing.T) {
	testCases := []struct {
		in             byte
		expectedResult bool
		expectedError  bool
	}{
		{0x00, false, false},
		{0x01, true, false},
		{0xFF, false, true},
	}
	for _, testCase := range testCases {
		got, err := byteToBool(testCase.in)
		if got != testCase.expectedResult {
			t.Errorf("Expected result '%v' but got %v'", testCase.expectedResult, got)
		}

		if testCase.expectedError != (err != nil) {
			t.Errorf("Expected error %v but got '%v'", testCase.expectedError, err)
		}
	}
}
func TestCropToString(t *testing.T) {
	testCases := []struct {
		in  []byte
		exp string
	}{
		{[]byte("foo"), "foo"},
		{[]byte{102, 111, 111, 0, 0, 0}, "foo"},
	}
	for _, testCase := range testCases {
		got := cropToString(testCase.in)
		if got != testCase.exp {
			t.Errorf("Expected '%v' but got '%v' for input '%v'", testCase.exp, got, testCase.in)
		}
		if len(got) != len(testCase.exp) {
			t.Errorf("Expected len '%v' but got '%v' for input '%v'", len(testCase.exp), len(got), testCase.in)
		}
	}

}
