// Package ioutils provides standard I/O helpers and resilient error-handling routines
// for command-line utilities.
package ioutils

import (
	"errors"
	"strings"
	"syscall"
)

// DefaultBufferSize is the standard 64 KB memory buffer size used across streaming
// core utilities (e.g. cat, head, yes) for optimal I/O throughput.
const DefaultBufferSize = 64 * 1024

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
	return strings.Contains(strings.ToLower(err.Error()), "broken pipe")
}
