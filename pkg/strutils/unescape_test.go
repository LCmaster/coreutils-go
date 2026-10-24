package strutils

import (
	"testing"
)

func TestUnescape(t *testing.T) {
	tests := []struct {
		input        string
		expectedOut  string
		expectedHalt bool
	}{
		{input: `hello\nworld`, expectedOut: "hello\nworld", expectedHalt: false},
		{input: `col1\tcol2`, expectedOut: "col1\tcol2", expectedHalt: false},
		{input: `a\ab\bc\fd\re\vf`, expectedOut: "a\ab\bc\fd\re\vf", expectedHalt: false},
		{input: `\e[31m`, expectedOut: "\x1b[31m", expectedHalt: false},
		{input: `\0101`, expectedOut: "A", expectedHalt: false},
		{input: `\0`, expectedOut: "\x00", expectedHalt: false},
		{input: `\040`, expectedOut: " ", expectedHalt: false},
		{input: `\x41`, expectedOut: "A", expectedHalt: false},
		{input: `\x61`, expectedOut: "a", expectedHalt: false},
		{input: `\x`, expectedOut: `\x`, expectedHalt: false},
		{input: `stop\chere`, expectedOut: "stop", expectedHalt: true},
		{input: `literal\\backslash`, expectedOut: "literal\\backslash", expectedHalt: false},
		{input: `foo "bar" baz`, expectedOut: `foo "bar" baz`, expectedHalt: false},
		{input: `unknown\z`, expectedOut: `unknown\z`, expectedHalt: false},
		{input: `trailing\`, expectedOut: `trailing\`, expectedHalt: false},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			out, halt := Unescape(tc.input)
			if out != tc.expectedOut {
				t.Errorf("Unescape(%q) out = %q, expected %q", tc.input, out, tc.expectedOut)
			}
			if halt != tc.expectedHalt {
				t.Errorf("Unescape(%q) halt = %v, expected %v", tc.input, halt, tc.expectedHalt)
			}
		})
	}
}
