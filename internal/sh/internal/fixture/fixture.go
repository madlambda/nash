package fixture

import (
	"testing"
	"path/filepath"
	
	"github.com/NeowayLabs/nash"
	"github.com/NeowayLabs/nash/internal/testing/fixture"
)


type NashDirs struct {
	Path string
	Lib string
	Root string
	Stdlib string
	Cleanup func()
}

var MkdirAll func(*testing.T, string) = fixture.MkdirAll

var Tmpdir func(*testing.T) (string, func()) = fixture.Tmpdir

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





