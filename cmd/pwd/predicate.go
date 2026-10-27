package main

import (
	"path/filepath"

	"github.com/LCmaster/coreutils-go/pkg/fn"
	"github.com/LCmaster/coreutils-go/pkg/fsutils"
)

func isNonEmpty(path string) bool {
	return path != ""
}

func isAbsolute(path string) bool {
	return filepath.IsAbs(path)
}

func hasNoDotComponents(path string) bool {
	return !fsutils.HasDotComponents(path)
}

var isValidLogicalPath = fn.All(isNonEmpty, isAbsolute, hasNoDotComponents, fsutils.IsCurrentDirectory)
