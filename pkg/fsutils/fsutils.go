// Package fsutils provides filesystem, path introspection, and I/O safety utilities
// shared across core command-line tools.
package fsutils

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/LCmaster/coreutils-go/pkg/fn"
)

// SplitSegments splits a file path into its individual directory and filename components,
// stripping leading, trailing, and consecutive separator characters.
//
// Both forward slashes ('/') and platform-native separators (filepath.Separator) are
// treated as delimiters to ensure consistent behavior across Unix and Windows systems.
// If the path contains no non-separator characters, an empty or nil slice is returned.
func SplitSegments(path string) []string {
	return strings.FieldsFunc(path, func(r rune) bool {
		return r == '/' || r == filepath.Separator
	})
}

// HasDotComponents reports whether any segment in the provided path is a relative
// directory reference ("." for current directory or ".." for parent directory).
//
// This is used by path validation logic (such as in `pwd -L`) to verify that an
// environment-provided path does not contain un-canonicalized relative components.
func HasDotComponents(path string) bool {
	return fn.AnyMatch(SplitSegments(path), func(segment string) bool {
		return segment == "." || segment == ".."
	})
}

// IsCurrentDirectory reports whether targetPath resolves to the current working
// directory (".").
//
// It queries the filesystem using os.Stat and compares file identities via os.SameFile.
// This ensures that symlinks, relative paths (e.g. "../current", "./subdir/.."), and
// absolute paths that point to the current working directory are accurately identified.
// If either targetPath or "." cannot be stated (e.g. nonexistent path or permission denied),
// IsCurrentDirectory returns false.
func IsCurrentDirectory(targetPath string) bool {
	targetStat, errTarget := os.Stat(targetPath)
	currentDirStat, errCurrentDir := os.Stat(".")
	return errTarget == nil && errCurrentDir == nil && os.SameFile(targetStat, currentDirStat)
}

// IsSameAsOutput reports whether an open input file refers to the same underlying file
// as the destination output writer.
//
// This check is a critical safety safeguard in utilities like `cat`, `head`, and `tail`
// to detect when standard output redirection targets the file being read (e.g. `cat file > file`),
// which would otherwise cause an infinite loop or file truncation/corruption.
//
// It returns true if and only if both file and out refer to the same regular file according
// to os.SameFile. If out is not an *os.File or file is nil/a directory, it returns false.
func IsSameAsOutput(file *os.File, out io.Writer) bool {
	if file == nil {
		return false
	}
	stat, err := file.Stat()
	if err != nil || stat.IsDir() {
		return false
	}

	var outStat os.FileInfo
	if outFile, ok := out.(*os.File); ok {
		if s, err := outFile.Stat(); err == nil && s.Mode().IsRegular() {
			outStat = s
		}
	} else if s, err := os.Stdout.Stat(); err == nil && s.Mode().IsRegular() {
		outStat = s
	}

	if outStat != nil {
		return os.SameFile(stat, outStat)
	}
	return false
}
