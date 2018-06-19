package main

import (
	"os"
	"io"
	"fmt"
	"io/ioutil"
	"path/filepath"
)

func NashLibDir(nashpath string) string {
	//FIXME: This is sadly duplicated from the shell implementation =(
	return filepath.Join(nashpath, "lib")
}

func InstallLib(nashpath string, sourcepath string) error {
	nashlibdir := NashLibDir(nashpath)
	if filepath.HasPrefix(sourcepath, nashlibdir) {
		return fmt.Errorf(
			"lib source path[%s] can't be inside nash lib dir[%s]", sourcepath, nashlibdir) 
	}
	return installLib(nashlibdir, sourcepath)
}

func installLib(targetdir string, sourcepath string) error {
	// TODO: error handling
	f, _ := os.Stat(sourcepath)
	if !f.IsDir() {
		return copyfile(targetdir, sourcepath)
	}
	
	basedir := filepath.Base(sourcepath)
	targetdir = filepath.Join(targetdir, basedir)
	// TODO: error handling
	files, _ := ioutil.ReadDir(sourcepath)
	for _, file := range files {
		// TODO: error handling
		installLib(targetdir, filepath.Join(sourcepath, file.Name()))
	}
	return nil
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