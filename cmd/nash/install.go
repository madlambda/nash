package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
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
		return fmt.Errorf("error[%s] checking if path[%s] is dir", err, sourcepath)
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
	fail := func(err error) error {
		return fmt.Errorf(
			"error[%s] trying to copy file[%s] to [%s]", err, sourcefilepath, targetdir)
	}

	err := os.MkdirAll(targetdir, os.ModePerm)
	if err != nil {
		return fail(err)
	}

	sourcefile, err := os.Open(sourcefilepath)
	if err != nil {
		return fail(err)
	}
	defer sourcefile.Close()

	targetfilepath := filepath.Join(targetdir, filepath.Base(sourcefilepath))
	targetfile, err := os.Create(targetfilepath)
	if err != nil {
		return fail(err)
	}
	defer targetfile.Close()

	_, err = io.Copy(targetfile, sourcefile)
	return err
}
