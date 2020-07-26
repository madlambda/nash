package tests

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

var (
	// Nashcmd is the nash's absolute binary path in source
	Nashcmd string

	// Projectpath is the path to nash source code
	Projectpath string

	// Testdir is the test assets directory
	Testdir string

	// Stdbindir is the stdbin directory
	Stdbindir string
)

func init() {

	Projectpath = findProjectRoot()
	Testdir = filepath.Join(Projectpath, "testfiles")
	Nashcmd = filepath.Join(Projectpath, "cmd", "nash", "nash")
	Stdbindir = filepath.Join(Projectpath, "stdbin")

	if runtime.GOOS == "windows" {
		Nashcmd += ".exe"
	}

	if _, err := os.Stat(Nashcmd); err != nil {
		msg := fmt.Sprintf("Unable to find nash command at %q.\n", Nashcmd)
		msg += "Please, run make build before running tests"
		panic(msg)
	}
}

func findProjectRoot() string {
	// We used to use GOPATH as a way to infer the root of the
	// project, now with Go modules this doesn't work anymore.
	// Since module definition files only appear on the root
	// of the project we use them instead, recursively going
	// backwards in the file system until we find them.
	//
	// From: https://blog.golang.org/using-go-modules
	// A module is a collection of Go packages stored in a file tree with a go.mod file at its root
	//
	// RIP GOPATH :-(

	separator := string(filepath.Separator)
	dir, err := os.Getwd()
	if err != nil {
		panic(fmt.Sprintf("failed to get current working directory:%v", err))
	}

	for !hasGoModFile(dir) {
		if dir == separator {
			// FIXME: not sure if this will work on all OS's, perhaps we need some
			// other protection against infinite loops... or just trust go test timeout.
			panic("reached root of file system without finding project root")
		}
		dir = filepath.Dir(dir)
	}

	return dir
}

func hasGoModFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	if !info.IsDir() {
		return false
	}

	gomodpath := filepath.Join(path, "go.mod")
	modinfo, err := os.Stat(gomodpath)
	if err != nil {
		return false
	}
	if modinfo.IsDir() {
		return false
	}
	return true
}
