package main

import (
	"os"
	"path/filepath"
	"strings"
)

func anyMatch[T any](items []T, predicate func(T) bool) bool {
	for _, item := range items {
		if predicate(item) {
			return true
		}
	}
	return false
}

func splitPathSegments(path string) []string {
	return strings.FieldsFunc(path, func(r rune) bool {
		return r == '/' || r == filepath.Separator
	})
}

func hasDotComponents(path string) bool {
	return anyMatch(splitPathSegments(path), func(segment string) bool {
		return segment == "." || segment == ".."
	})
}

type PathPredicate func(path string) bool

func allPredicates(predicates ...PathPredicate) PathPredicate {
	return func(path string) bool {
		for _, pred := range predicates {
			if !pred(path) {
				return false
			}
		}
		return true
	}
}

func isNonEmpty(path string) bool {
	return path != ""
}

func isAbsolute(path string) bool {
	return filepath.IsAbs(path)
}

func hasNoDotComponents(path string) bool {
	return !hasDotComponents(path)
}

func isCurrentDirectory(path string) bool {
	pwdStat, errPwd := os.Stat(path)
	currentDirStat, errCurrentDir := os.Stat(".")
	return errPwd == nil && errCurrentDir == nil && os.SameFile(pwdStat, currentDirStat)
}

var isValidLogicalPath = allPredicates(isNonEmpty, isAbsolute, hasNoDotComponents, isCurrentDirectory)

type DirResolver func() (string, error)
