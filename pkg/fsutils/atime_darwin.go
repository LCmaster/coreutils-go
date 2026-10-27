//go:build darwin

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
	if stat, ok := info.Sys().(*syscall.Stat_t); ok && stat != nil {
		return time.Unix(int64(stat.Atimespec.Sec), int64(stat.Atimespec.Nsec))
	}
	return info.ModTime()
}

