package tests

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
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
	project := "github.com/madlambda/nash"
	wd, err := os.Getwd()
	if err != nil {
		panic("failed to get current directory")
	}

	pos := strings.Index(wd, project) + len(project)
	Projectpath = wd[:pos]

	Testdir = filepath.Join(Projectpath, "testfiles")
	Nashcmd = filepath.Join(Projectpath, "cmd", "nash", "nash")
	Stdbindir = filepath.Join(Projectpath, "stdbin")

	if runtime.GOOS == "windows" {
		Nashcmd += ".exe"
	}

	if _, err := os.Stat(Nashcmd); err != nil {
		panic("Please, run make build before running tests")
	}
}
