package strutils

import (
	"bytes"
	"testing"
)

func TestFormatNonPrinting(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected []byte
	}{
		{
			name:     "printable ASCII",
			input:    []byte("hello world 123!"),
			expected: []byte("hello world 123!"),
		},
		{
			name:     "preserve newline and tab",
			input:    []byte("line1\tvalue\n"),
			expected: []byte("line1\tvalue\n"),
		},
		{
			name:     "control characters null and bell",
			input:    []byte{0x00, 0x07},
			expected: []byte("^@^G"),
		},
		{
			name:     "delete char 127",
			input:    []byte{0x7F},
			expected: []byte("^?"),
		},
		{
			name:     "meta control character 128",
			input:    []byte{0x80},
			expected: []byte("M-^@"),
		},
		{
			name:     "meta character 160",
			input:    []byte{0xA0},
			expected: []byte("M- "),
		},
		{
			name:     "meta byte 255",
			input:    []byte{0xFF},
			expected: []byte("M-^?"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatNonPrinting(tt.input)
			if !bytes.Equal(got, tt.expected) {
				t.Errorf("FormatNonPrinting(%v) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestIsEmptyLine(t *testing.T) {
	tests := []struct {
		input    []byte
		expected bool
	}{
		{[]byte{}, true},
		{[]byte("\n"), true},
		{[]byte("\r\n"), true},
		{[]byte(" "), false},
		{[]byte("a\n"), false},
		{[]byte("\t\n"), false},
	}

	for _, tt := range tests {
		got := IsEmptyLine(tt.input)
		if got != tt.expected {
			t.Errorf("IsEmptyLine(%q) = %v, want %v", tt.input, got, tt.expected)
		}
	}
}
