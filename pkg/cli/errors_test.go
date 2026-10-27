package cli

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestErrorConstructors(t *testing.T) {
	if got := ErrInvalidOption("cat", 'z').Error(); got != "cat: invalid option -- 'z'" {
		t.Errorf("got %q", got)
	}
	if got := ErrUnrecognizedOption("cat", "bogus").Error(); got != "cat: unrecognized option '--bogus'" {
		t.Errorf("got %q", got)
	}
	if got := ErrOptionRequiresArgShort("head", 'n').Error(); got != "head: option requires an argument -- 'n'" {
		t.Errorf("got %q", got)
	}
	if got := ErrOptionRequiresArgLong("head", "lines").Error(); got != "head: option '--lines' requires an argument" {
		t.Errorf("got %q", got)
	}
	if got := ErrOptionNoArgAllowed("head", "quiet").Error(); got != "head: option '--quiet' doesn't allow an argument" {
		t.Errorf("got %q", got)
	}

	rootErr := errors.New("permission denied")
	openErr := ErrCannotOpenForReading("secret.txt", rootErr)
	if !errors.Is(openErr, rootErr) || openErr.Error() != "cannot open 'secret.txt' for reading: permission denied" {
		t.Errorf("got %v", openErr)
	}

	readErr := ErrReadingFile("data.txt", rootErr)
	if !errors.Is(readErr, rootErr) || readErr.Error() != "error reading 'data.txt': permission denied" {
		t.Errorf("got %v", readErr)
	}

	writeErr := ErrWrite(rootErr)
	if !errors.Is(writeErr, rootErr) || writeErr.Error() != "write error: permission denied" {
		t.Errorf("got %v", writeErr)
	}

	if got := ErrInvalidNumberOfLines("abc").Error(); got != "invalid number of lines: 'abc'" {
		t.Errorf("got %q", got)
	}
	if got := ErrInvalidNumberOfBytes("xyz").Error(); got != "invalid number of bytes: 'xyz'" {
		t.Errorf("got %q", got)
	}
	if got := ErrInvalidArgument("bad", "--time").Error(); got != "invalid argument 'bad' for '--time'" {
		t.Errorf("got %q", got)
	}
	if got := ErrInvalidDateFormat("not-a-date").Error(); got != "invalid date format 'not-a-date'" {
		t.Errorf("got %q", got)
	}
}

func TestCommandDiagnosticHelpers(t *testing.T) {
	cmd := NewCommand("touch")

	// Nil stderr safety
	cmd.Errorf(nil, "ignored")
	cmd.FileError(nil, "f.txt", ErrIsDirectory)
	if err := cmd.TryHelpError(nil, ErrMissingFileOperand); !errors.Is(err, ErrExitFailure) {
		t.Errorf("expected ErrExitFailure, got %v", err)
	}

	var stderr bytes.Buffer
	cmd.Errorf(&stderr, "%v", ErrZeroLengthFileName)
	if stderr.String() != "touch: invalid zero-length file name\n" {
		t.Errorf("got %q", stderr.String())
	}

	stderr.Reset()
	cmd.FileError(&stderr, "dir", ErrIsDirectory)
	if stderr.String() != "touch: dir: Is a directory\n" {
		t.Errorf("got %q", stderr.String())
	}

	stderr.Reset()
	err := cmd.TryHelpError(&stderr, ErrMissingFileOperand)
	if !errors.Is(err, ErrExitFailure) {
		t.Errorf("expected ErrExitFailure, got %v", err)
	}
	want := "touch: missing file operand\nTry 'touch --help' for more information.\n"
	if !strings.EqualFold(stderr.String(), want) {
		t.Errorf("got %q, want %q", stderr.String(), want)
	}
}

