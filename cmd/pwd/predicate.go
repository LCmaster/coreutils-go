package main

import (
	"path/filepath"

	"github.com/LCmaster/coreutils-go/pkg/fn"
	"github.com/LCmaster/coreutils-go/pkg/fsutils"
	"github.com/LCmaster/coreutils-go/pkg/ioutils"
)

func hasDotComponents(path string) bool {
	return fsutils.HasDotComponents(path)
}

type PathPredicate func(path string) bool

func allPredicates(predicates ...PathPredicate) PathPredicate {
	preds := make([]func(string) bool, len(predicates))
	for i, p := range predicates {
		preds[i] = p
	}
	return fn.All(preds...)
}

func isNonEmpty(path string) bool {
	return path != ""
}

func isAbsolute(path string) bool {
	return filepath.IsAbs(path)
}

func hasNoDotComponents(path string) bool {
	return !fsutils.HasDotComponents(path)
}

func isBrokenPipe(err error) bool {
	return ioutils.IsBrokenPipe(err)
}

var isValidLogicalPath = fn.All(isNonEmpty, isAbsolute, hasNoDotComponents, fsutils.IsCurrentDirectory)

type DirResolver func() (string, error)
