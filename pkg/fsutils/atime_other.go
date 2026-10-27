//go:build !windows

package fsutils

import (
	"os"
	"time"
)

// AccessTime returns the last access time of the file described by info.
// On non-Windows platforms, it falls back to info.ModTime().
func AccessTime(info os.FileInfo) time.Time {
	if info == nil {
		return time.Time{}
	}
	return info.ModTime()
}
