package main

import (
	"os"
	"io"
	"path/filepath"
)

func NashLibDir(nashpath string) string {
	//FIXME: This is sadly duplicated from the shell implementation =(
	return filepath.Join(nashpath, "lib")
}

func InstallLib(nashpath string, installpath string) error {
	libdir := NashLibDir(nashpath)
	// TODO: Propositaly only handling a single simple case for now
	return copyfile(libdir, installpath)
}

func copyfile(targetdir string, sourcefilepath string) error {
	// TODO: error handling
	os.MkdirAll(targetdir, os.ModePerm)
	
	// TODO: error handling
	sourcefile, _ := os.Open(sourcefilepath)
	defer sourcefile.Close()
	
	targetfilepath := filepath.Join(targetdir, filepath.Base(sourcefilepath))
	// TODO: Error handling
	targetfile, _ := os.Create(targetfilepath)
	defer targetfile.Close()
	
	// TODO: Error handling moderfocker
	io.Copy(targetfile, sourcefile)
	
	return nil
}