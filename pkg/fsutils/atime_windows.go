//go:build windows

package fsutils

import (
	"os"
	"syscall"
	"time"
)

// AccessTime returns the last access time of the file described by info.
// If platform-specific metadata is unavailable, it falls back to info.ModTime().
func AccessTime(info os.FileInfo) time.Time {
	if info == nil {
		return time.Time{}
	}
	if attr, ok := info.Sys().(*syscall.Win32FileAttributeData); ok && attr != nil {
		return time.Unix(0, attr.LastAccessTime.Nanoseconds())
	}
	return info.ModTime()
}

