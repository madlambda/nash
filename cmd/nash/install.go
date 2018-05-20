package main

import (
	"path/filepath"
)

func NashLibDir(nashpath string) string {
	//FIXME: This is sadly duplicated from the shell implementation =(
	return filepath.Join(nashpath, "lib")
}

func InstallLib(nashpath string, installdir string) error {
	return nil
}