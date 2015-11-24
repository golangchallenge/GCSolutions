package solver

import "testing"

func TestBitCount(t *testing.T) {
	testcases := []struct {
		input uint16
		want  uint8
	}{
		{
			input: 0,
			want:  0,
		},
		{
			input: 32,
			want:  1,
		},
		{
			input: 7,
			want:  3,
		},
		{
			input: 127,
			want:  7,
		},
	}
	for _, tc := range testcases {
		if got := bc(tc.input); got != tc.want {
			t.Errorf("bc(%d) = %d, want %d", tc.input, got, tc.want)
		}
	}
}

func TestSingleton(t *testing.T) {
	testcases := []struct {
		input uint8
		want  uint16
	}{
		{
			input: 1,
			want:  1,
		},
		{
			input: 2,
			want:  2,
		},
		{
			input: 7,
			want:  64,
		},
	}
	for _, tc := range testcases {
		if got := singleton(tc.input); got != tc.want {
			t.Errorf("singleton(%d) = %d, want %d", tc.input, got, tc.want)
		}
	}
}
