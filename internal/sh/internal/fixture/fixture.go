package fixture

import (
	"testing"
	"io/ioutil"
	"os"
	"path/filepath"
	
	"github.com/NeowayLabs/nash"
)


type NashDirs struct {
	Path string
	Lib string
	Root string
	Stdlib string
	Cleanup func()
}

func SetupShell(t *testing.T) (*nash.Shell, func()) {
	dirs := SetupNashDirs(t)
	
	shell, err := nash.New(dirs.Path, dirs.Root)

	if err != nil {
		dirs.Cleanup()
		t.Fatal(err)
	}

	return shell, dirs.Cleanup
}

func SetupNashDirs(t *testing.T) NashDirs {
	testdir, rmdir := Tmpdir(t)

	nashpath := filepath.Join(testdir, "nashpath")
	nashroot := filepath.Join(testdir, "nashroot")
	
	nashlib := filepath.Join(nashpath, "lib")
	nashstdlib := filepath.Join(nashroot, "stdlib")
	
	MkdirAll(t, nashlib)
	MkdirAll(t, nashstdlib)
	
	return NashDirs{
		Path: nashpath,
		Lib: nashlib,
		Root: nashroot,
		Stdlib: nashstdlib,
		Cleanup: rmdir,
	}
}

func Tmpdir(t *testing.T) (string, func()) {
	t.Helper()
	
	dir, err := ioutil.TempDir("", "nash-tests")
	if err != nil {
		t.Fatal(err)
	}
	
	return dir, func() {
		err := os.RemoveAll(dir)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func MkdirAll(t *testing.T, nashlib string) {
	err := os.MkdirAll(nashlib, os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
}





