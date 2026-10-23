package main

import (
	"os"
	"path/filepath"
)

type ResolutionMode int

const (
	LogicalMode ResolutionMode = iota
	PhysicalMode
)

func resolvePhysicalDir() (string, error) {
	path, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(path)
}

func readEnvPwd() (string, error) {
	return os.Getenv("PWD"), nil
}

func resolveLogicalDir() (string, error) {
	path, err := readEnvPwd()
	if err == nil && isValidLogicalPath(path) {
		return path, nil
	}
	return resolvePhysicalDir()
}

func resolveWorkingDir(mode ResolutionMode) (string, error) {
	if mode == PhysicalMode {
		return resolvePhysicalDir()
	}
	return resolveLogicalDir()
}
