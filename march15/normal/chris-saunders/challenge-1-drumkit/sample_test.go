package drum

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"
)

func TestExtractSample(t *testing.T) {
	tests := []struct {
		input    []byte
		expected Sample
	}{
		{
			[]byte{
				// sample ID
				0x00, 0x00, 0x00, 0x00,
				// label length
				04,
				// label
				0x6b, 0x69, 0x63, 0x6b,
				// Sample
				0x01, 0x00, 0x00, 0x00,
				0x01, 0x00, 0x00, 0x00,
				0x01, 0x00, 0x00, 0x00,
				0x01, 0x00, 0x00, 0x00,
			},
			Sample{
				ID:    0,
				Label: "kick",
				Steps: [16]bool{
					true, false, false, false,
					true, false, false, false,
					true, false, false, false,
					true, false, false, false,
				},
			},
		},
	}
	for _, test := range tests {
		actual, _ := ExtractSample(bytes.NewReader(test.input))
		if !reflect.DeepEqual(actual, test.expected) {
			t.Logf("actual:\n%#v\n", fmt.Sprint(actual))
			t.Logf("expected:\n%#v\n", fmt.Sprint(test.expected))
			t.Fatalf(
				"%s was not extracted as expected.\nGot:\n%s\nExpected:\n%s",
				test.expected.Label,
				actual,
				test.expected,
			)
		}
	}
}
