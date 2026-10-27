package main

import (
	"os"
	"path/filepath"
)

func resolvePhysicalDir() (string, error) {
	path, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(path)
}

func resolveLogicalDir() (string, error) {
	if path := os.Getenv("PWD"); isValidLogicalPath(path) {
		return path, nil
	}
	return resolvePhysicalDir()
}

func resolveWorkingDir(physical bool) (string, error) {
	if physical {
		return resolvePhysicalDir()
	}
	return resolveLogicalDir()
}
