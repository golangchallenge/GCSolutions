package drum

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"
)

func TestExtractHeader(t *testing.T) {
	tests := []struct {
		input    []byte
		expected Header
	}{
		{
			[]byte{
				0x53, 0x50, 0x4c, 0x49, 0x43, 0x45, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xc5, 0x30, 0x2e,
				0x38, 0x30, 0x38, 0x2d, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0xf0, 0x42,
			},
			Header{Version: "0.808-alpha", Tempo: 120.0},
		},
	}

	for _, test := range tests {
		actual := ExtractHeader(bytes.NewBuffer(test.input[:]))
		if !reflect.DeepEqual(actual, test.expected) {
			t.Logf("actual:\n%#v\n", fmt.Sprint(actual))
			t.Logf("expected:\n%#v\n", fmt.Sprint(test.expected))
			t.Fatalf(
				"%s was not extracted as expected.\nGot:\n%s\nExpected:\n%s",
				test.expected.Version,
				actual,
				test.expected,
			)
		}
	}
}

func TestHeaderAsString(t *testing.T) {
	tests := []struct {
		subject  Header
		expected string
	}{
		{
			Header{Version: "0.808-alpha", Tempo: 120.0},
			`Saved with HW Version: 0.808-alpha
Tempo: 120`,
		},
		{
			Header{Version: "0.808-alpha", Tempo: 98.4},
			`Saved with HW Version: 0.808-alpha
Tempo: 98.4`,
		},
	}
	for _, test := range tests {
		if test.subject.String() != test.expected {
			t.Logf("actual:\n%#v\n", fmt.Sprint(test.subject))
			t.Logf("expected:\n%#v\n", fmt.Sprint(test.expected))
			t.Fatalf(
				"Got:\n%s\nExpected:\n%s",
				test.subject.String(),
				test.expected,
			)
		}
	}
}
