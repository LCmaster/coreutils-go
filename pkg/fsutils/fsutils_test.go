package fsutils

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSplitSegments(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"", nil},
		{"/", nil},
		{"/a/b/c", []string{"a", "b", "c"}},
		{"a/b//c/", []string{"a", "b", "c"}},
	}

	for _, tt := range tests {
		got := SplitSegments(tt.input)
		if len(got) != len(tt.expected) {
			t.Fatalf("SplitSegments(%q) len = %d, want %d", tt.input, len(got), len(tt.expected))
		}
		for i := range got {
			if got[i] != tt.expected[i] {
				t.Errorf("SplitSegments(%q)[%d] = %q, want %q", tt.input, i, got[i], tt.expected[i])
			}
		}
	}
}

func TestHasDotComponents(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"/usr/local/bin", false},
		{"/usr/./bin", true},
		{"/usr/../bin", true},
		{".", true},
		{"..", true},
		{"file.txt", false},
		{".hidden", false},
		{"..hidden", false},
	}

	for _, tt := range tests {
		got := HasDotComponents(tt.path)
		if got != tt.expected {
			t.Errorf("HasDotComponents(%q) = %v, want %v", tt.path, got, tt.expected)
		}
	}
}

func TestIsCurrentDirectory(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd failed: %v", err)
	}

	if !IsCurrentDirectory(".") {
		t.Errorf("IsCurrentDirectory('.') want true, got false")
	}

	if !IsCurrentDirectory(wd) {
		t.Errorf("IsCurrentDirectory(%q) want true, got false", wd)
	}

	parent := filepath.Dir(wd)
	if IsCurrentDirectory(parent) {
		t.Errorf("IsCurrentDirectory(%q) want false, got true", parent)
	}

	if IsCurrentDirectory("nonexistent_path_xyz_123") {
		t.Errorf("IsCurrentDirectory for non-existent path want false, got true")
	}
}

func TestIsSameAsOutput(t *testing.T) {
	tmp := t.TempDir()
	f1Path := filepath.Join(tmp, "f1.txt")
	f2Path := filepath.Join(tmp, "f2.txt")

	if err := os.WriteFile(f1Path, []byte("hello"), 0o644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}
	if err := os.WriteFile(f2Path, []byte("world"), 0o644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	f1, err := os.Open(f1Path)
	if err != nil {
		t.Fatalf("Open f1 failed: %v", err)
	}
	defer f1.Close()

	f1Writer, err := os.OpenFile(f1Path, os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatalf("OpenFile f1Writer failed: %v", err)
	}
	defer f1Writer.Close()

	f2Writer, err := os.OpenFile(f2Path, os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatalf("OpenFile f2Writer failed: %v", err)
	}
	defer f2Writer.Close()

	if !IsSameAsOutput(f1, f1Writer) {
		t.Errorf("IsSameAsOutput(f1, f1Writer) want true, got false")
	}

	if IsSameAsOutput(f1, f2Writer) {
		t.Errorf("IsSameAsOutput(f1, f2Writer) want false, got true")
	}

	var buf bytes.Buffer
	if IsSameAsOutput(f1, &buf) {
		t.Errorf("IsSameAsOutput(f1, &buf) want false, got true")
	}

	if IsSameAsOutput(nil, f1Writer) {
		t.Errorf("IsSameAsOutput(nil, f1Writer) want false, got true")
	}

	if IsSameAsOutput(f1, nil) {
		t.Errorf("IsSameAsOutput(f1, nil) want false, got true")
	}
}

func TestAccessTime(t *testing.T) {
	if !AccessTime(nil).IsZero() {
		t.Errorf("expected zero time for nil FileInfo")
	}

	path := filepath.Join(t.TempDir(), "atime.txt")
	if err := os.WriteFile(path, []byte("data"), 0o644); err != nil {
		t.Fatal(err)
	}

	wantAtime := time.Date(2025, 3, 15, 10, 20, 30, 0, time.UTC)
	wantMtime := time.Date(2024, 1, 10, 8, 0, 0, 0, time.UTC)
	if err := os.Chtimes(path, wantAtime, wantMtime); err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}

	gotAtime := AccessTime(info)
	if !gotAtime.Equal(wantAtime) {
		t.Errorf("AccessTime = %v, want %v", gotAtime, wantAtime)
	}
	if !info.ModTime().Equal(wantMtime) {
		t.Errorf("ModTime = %v, want %v", info.ModTime(), wantMtime)
	}
}
