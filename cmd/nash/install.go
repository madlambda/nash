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
	sourcepathAbs, err := filepath.Abs(sourcepath)
	if err != nil {
		return fmt.Errorf("error[%s] getting absolute path of [%s]", err, sourcepath)
	}
	if filepath.HasPrefix(sourcepathAbs, nashlibdir) {
		return fmt.Errorf(
			"lib source path[%s] can't be inside nash lib dir[%s]", sourcepath, nashlibdir) 
	}
	return installLib(nashlibdir, sourcepathAbs)
}

func installLib(targetdir string, sourcepath string) error {

	f, err := os.Stat(sourcepath)
	if err != nil {
		return fmt.Errorf("error[%s] stating path[%s]", err, sourcepath)
	}
	
	if !f.IsDir() {
		return copyfile(targetdir, sourcepath)
	}
	
	basedir := filepath.Base(sourcepath)
	targetdir = filepath.Join(targetdir, basedir)
	
	files, err := ioutil.ReadDir(sourcepath)
	if err != nil {
		return fmt.Errorf("error[%s] reading dir[%s]", err, sourcepath)
	}
	
	for _, file := range files {
		err := installLib(targetdir, filepath.Join(sourcepath, file.Name()))
		if err != nil {
			return err
		}
	}
	return nil
}

func copyfile(targetdir string, sourcefilepath string) error {
	// TODO: error handling
	os.MkdirAll(targetdir, os.ModePerm)
	
	// TODO: error handling
	sourcefile, err := os.Open(sourcefilepath)
	if err != nil {
		return fmt.Errorf("error[%s] trying to copy file[%s] to [%s]", err, sourcefilepath, targetdir)
	}
	defer sourcefile.Close()
	
	targetfilepath := filepath.Join(targetdir, filepath.Base(sourcefilepath))
	// TODO: Error handling
	targetfile, _ := os.Create(targetfilepath)
	defer targetfile.Close()
	
	// TODO: Error handling moderfocker
	io.Copy(targetfile, sourcefile)
	
	return nil
}