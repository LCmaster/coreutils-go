// Package ioutils provides standard I/O helpers and resilient error-handling routines
// for command-line utilities.
package ioutils

import (
	"errors"
	"io"
	"os"
	"strings"
	"syscall"
)

// DefaultBufferSize is the standard 64 KB memory buffer size used across streaming
// core utilities (e.g. cat, head, wc, yes) for optimal I/O throughput.
const DefaultBufferSize = 64 * 1024

type nopCloser struct{}

func (nopCloser) Close() error { return nil }

// OpenInput returns an io.Reader and io.Closer for the given operand filename.
// If filename is "-", it returns stdin paired with a no-op Closer so callers can
// unconditionally close the returned io.Closer without closing standard input.
// Otherwise, it opens the named file on disk via os.Open.
func OpenInput(filename string, stdin io.Reader) (io.Reader, io.Closer, error) {
	if filename == "-" {
		return stdin, nopCloser{}, nil
	}
	f, err := os.Open(filename)
	if err != nil {
		return nil, nil, err
	}
	return f, f, nil
}

// IsBrokenPipe reports whether err represents a broken pipe error across platforms.
//
// In standard Unix pipelines (e.g. `cat huge.txt | head -n 5` or `yes | head`), downstream
// consumers often terminate and close their read end of the pipe while the upstream utility
// is still writing. This causes the writer to receive syscall.EPIPE (on Unix) or an operating
// system error with the phrase "broken pipe" (on Windows).
//
// POSIX core utilities are expected to treat broken pipe errors as a normal, graceful exit
// condition rather than an abnormal runtime failure or fatal diagnostic message.
//
// IsBrokenPipe returns false if err is nil. It returns true if err wraps syscall.EPIPE
// (checked via errors.Is) or if its error string contains "broken pipe" (case-insensitive).
func IsBrokenPipe(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, syscall.EPIPE) {
		return true
	}
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "broken pipe") || strings.Contains(errStr, "pipe has been ended")
}
