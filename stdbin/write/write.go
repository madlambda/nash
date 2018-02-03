package main

import (
	"io"
	"os"
	"path/filepath"
)

func toabs(path string) (string, error) {
	if !filepath.IsAbs(path) {
		wd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		path = filepath.Join(wd, path)
	}
	return path, nil
}

func outfd(fname string) (io.WriteCloser, error) {
	fname, err := toabs(fname)
	if err != nil {
		return nil, err
	}

	var out io.WriteCloser

	out, ok := specialFile(fname)
	if !ok {
		f, err := os.OpenFile(fname, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			return nil, err
		}
		out = f
	}
	return out, nil
}

func write(fname string, in io.Reader) (err error) {
	out, err := outfd(fname)
	if err != nil {
		return err
	}

	defer func() {
		err2 := out.Close()
		if err == nil {
			err = err2
		}
	}()

	_, err = io.Copy(out, in)
	return err
}
