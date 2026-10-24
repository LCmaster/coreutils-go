package version

import "testing"

func TestFormat(t *testing.T) {
	got := Format("cat")
	want := "cat (coreutils-go) 0.1.0"
	if got != want {
		t.Errorf("Format(%q) = %q, want %q", "cat", got, want)
	}

	gotHead := Format("head")
	wantHead := "head (coreutils-go) 0.1.0"
	if gotHead != wantHead {
		t.Errorf("Format(%q) = %q, want %q", "head", gotHead, wantHead)
	}
}
