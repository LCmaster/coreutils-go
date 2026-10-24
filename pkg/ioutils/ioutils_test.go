package ioutils

import (
	"errors"
	"syscall"
	"testing"
)

func TestIsBrokenPipe(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"nil error", nil, false},
		{"unrelated error", errors.New("file not found"), false},
		{"syscall EPIPE", syscall.EPIPE, true},
		{"wrapped EPIPE", errors.Join(errors.New("write failure"), syscall.EPIPE), true},
		{"text broken pipe", errors.New("write: broken pipe"), true},
		{"text uppercase BROKEN PIPE", errors.New("BROKEN PIPE"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsBrokenPipe(tt.err)
			if got != tt.expected {
				t.Errorf("IsBrokenPipe(%v) = %v, want %v", tt.err, got, tt.expected)
			}
		})
	}
}
