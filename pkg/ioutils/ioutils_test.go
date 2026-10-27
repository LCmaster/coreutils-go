package ioutils

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
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
		{"windows pipe ended", errors.New("write /dev/stdout: The pipe has been ended."), true},
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

func TestOpenInput(t *testing.T) {
	// 1. "-" returns stdin and a safe no-op closer
	stdin := strings.NewReader("stdin payload")
	r, c, err := OpenInput("-", stdin)
	if err != nil {
		t.Fatalf("OpenInput('-', stdin) unexpected error: %v", err)
	}
	data, err := io.ReadAll(r)
	if err != nil || string(data) != "stdin payload" {
		t.Errorf("read %q, err %v", data, err)
	}
	if err := c.Close(); err != nil {
		t.Errorf("Close() on stdin nopCloser returned error: %v", err)
	}

	// 2. Existing file
	path := filepath.Join(t.TempDir(), "test.txt")
	if err := os.WriteFile(path, []byte("file payload"), 0o644); err != nil {
		t.Fatal(err)
	}
	r, c, err = OpenInput(path, nil)
	if err != nil {
		t.Fatalf("OpenInput(%q) unexpected error: %v", path, err)
	}
	data, err = io.ReadAll(r)
	if err != nil || string(data) != "file payload" {
		t.Errorf("read %q, err %v", data, err)
	}
	if err := c.Close(); err != nil {
		t.Errorf("Close() returned error: %v", err)
	}

	// 3. Missing file
	_, _, err = OpenInput(filepath.Join(t.TempDir(), "missing.txt"), nil)
	if err == nil {
		t.Errorf("expected error for missing file, got nil")
	}
}
