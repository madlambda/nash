package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
)

const banner = `888b    888                888                888             888888
8888b   888                888                888             888888
88888b  888                888                888             888888
888Y88b 888 8888b. .d8888b 88888b.    .d8888b 88888b.  .d88b. 888888
888 Y88b888    "88b88K     888 "88b   88K     888 "88bd8P  Y8b888888
888  Y88888.d888888"Y8888b.888  888   "Y8888b.888  88888888888888888
888   Y8888888  888     X88888  888        X88888  888Y8b.    888888
888    Y888"Y888888 88888P'888  888    88888P'888  888 "Y8888 888888
====================================================================
|| Documentation of ~/.nash/lib projects
`

func idoc(stdout, stderr io.Writer) error {
	fmt.Fprintf(stdout, "%s\n", banner)
	return walkAll(stdout, stderr, "/home/i4k/.nash", regexp.MustCompile(".*"))
}

func walkAll(stdout, stderr io.Writer, nashpath string, pattern *regexp.Regexp) error {
	return filepath.Walk(nashpath+"/lib", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		dirpath := filepath.Dir(path)
		dirname := filepath.Base(dirpath)
		ext := filepath.Ext(path)

		if ext != "" && ext != ".sh" {
			return nil
		}

		lookFn(stdout, stderr, path, dirname, pattern)

		return nil
	})
}
