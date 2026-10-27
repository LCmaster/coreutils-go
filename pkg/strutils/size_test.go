package strutils

import "testing"

func TestParseByteSize(t *testing.T) {
	tests := []struct {
		input    string
		expected int
		wantErr  bool
	}{
		{"0", 0, false},
		{"+0", 0, false},
		{"-0", 0, false},
		{"10", 10, false},
		{"+5", 5, false},
		{"-10", -10, false},
		{"1b", 512, false},
		{"1k", 1024, false},
		{"1K", 1024, false},
		{"1KiB", 1024, false},
		{"1kB", 1000, false},
		{"1KB", 1000, false},
		{"1MB", 1000 * 1000, false},
		{"1MiB", 1024 * 1024, false},
		{"1GB", 1000 * 1000 * 1000, false},
		{"1GiB", 1024 * 1024 * 1024, false},
		{"-2K", -2048, false},
		{"", 0, true},
		{"+", 0, true},
		{"-", 0, true},
		{"KiB", 0, true},
		{"abc", 0, true},
		{"9999999999EiB", 0, true}, // overflow
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseByteSize(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseByteSize(%q) err = %v, wantErr = %v", tt.input, err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.expected {
				t.Errorf("ParseByteSize(%q) = %d, want %d", tt.input, got, tt.expected)
			}
		})
	}
}

