package tests

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
)

var (
	// Nashcmd is the nash's absolute binary path in source
	Nashcmd string
	// Testdir is the test assets directory
	Testdir string

	// Stdbindir is the stdbin directory
	Stdbindir string

	// Gopath of golang
	Gopath string
)

func getGopath() (string, error) {
	gopathenv := os.Getenv("GOPATH")
	if gopathenv != "" {
		return gopathenv, nil
	}

	usr, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("failed to get user's home directory: %s", err)
	}

	gopathhome := filepath.Join(usr.HomeDir, "go")
	if _, err := os.Stat(gopathhome); err != nil {
		return "", errors.New("gopath not found")
	}

	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %s", err)
	}

	if !strings.HasPrefix(wd, gopathhome) {
		return "", errors.New("Run tests require code inside gopath")
	}
	return gopathhome, nil
}

func init() {
	gopath, err := getGopath()
	if err != nil {
		panic(err)
	}

	Gopath = gopath
	projectpath := filepath.Join(Gopath, "src", "github.com",
		"NeowayLabs", "nash")
	Testdir = filepath.Join(projectpath, "testfiles")
	Nashcmd = filepath.Join(projectpath, "cmd", "nash", "nash")
	Stdbindir = filepath.Join(projectpath, "stdbin")

	if runtime.GOOS == "windows" {
		Nashcmd += ".exe"
	}

	if _, err := os.Stat(Nashcmd); err != nil {
		panic("Please, run make build before running tests")
	}
}
