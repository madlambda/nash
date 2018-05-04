package sh

import (
	"bytes"
	"io/ioutil"
	"path/filepath"
	"os"
	"testing"
)

// TODO: behavior when NASHPATH and NASHROOT are invalid ?

func TestImportsLibFromNashPathLibDir(t *testing.T) {
	
	nashdirs := setupNashDirs(t)
	defer nashdirs.cleanup()
	
	writeFile(t, filepath.Join(nashdirs.lib, "lib.sh"), `
		fn test() {
			echo "hasnashpath"
		}
	`)

	newTestShell(t, nashdirs.path, nashdirs.root).ExecCheckingOutput(t, `
		import lib
		test()
	`, "hasnashpath\n")
}

func TestImportsLibFromNashPathLibDirBeforeNashRootStdlib(t *testing.T) {
	
	nashdirs := setupNashDirs(t)
	defer nashdirs.cleanup()

	writeFile(t, filepath.Join(nashdirs.lib, "lib.sh"), `
		fn test() {
			echo "libcode"
		}
	`)
	
	writeFile(t, filepath.Join(nashdirs.stdlib, "lib.sh"), `
		fn test() {
			echo "stdlibcode"
		}
	`)

	newTestShell(t, nashdirs.path, nashdirs.root).ExecCheckingOutput(t, `
		import lib
		test()
	`, "libcode\n")
}

func TestImportsLibFromNashRootStdlib(t *testing.T) {
	
	nashdirs := setupNashDirs(t)
	defer nashdirs.cleanup()
	
	writeFile(t, filepath.Join(nashdirs.stdlib, "lib.sh"), `
		fn test() {
			echo "stdlibcode"
		}
	`)

	newTestShell(t, nashdirs.path, nashdirs.root).ExecCheckingOutput(t, `
		import lib
		test()
	`, "stdlibcode\n")
}

func TestImportsLibFromWorkingDirBeforeLibAndStdlib(t *testing.T) {
	
	workingdir, rmdir := tmpdir(t)
	defer rmdir()
	
	curwd := getwd(t)
	chdir(t, workingdir)
	defer chdir(t, curwd)
	
	nashdirs := setupNashDirs(t)
	defer nashdirs.cleanup()
	
	writeFile(t, filepath.Join(workingdir, "lib.sh"), `
		fn test() {
			echo "localcode"
		}
	`)
	
	writeFile(t, filepath.Join(nashdirs.lib, "lib.sh"), `
		fn test() {
			echo "libcode"
		}
	`)
	
	writeFile(t, filepath.Join(nashdirs.stdlib, "lib.sh"), `
		fn test() {
			echo "stdlibcode"
		}
	`)
	
	newTestShell(t, nashdirs.path, nashdirs.root).ExecCheckingOutput(t, `
		import lib
		test()
	`, "localcode\n")
}


type testshell struct {
	shell  *Shell
	stdout *bytes.Buffer
}

func (s *testshell) ExecCheckingOutput(t *testing.T, code string, expectedOutupt string) {
	err := s.shell.Exec("shellenvtest", code)
	if err != nil {
		t.Fatal(err)
	}

	output := s.stdout.String()
	s.stdout.Reset()

	if output != expectedOutupt {
		t.Fatalf(
			"expected output: [%s] got: [%s]",
			expectedOutupt,
			output,
		)
	}
}

func newTestShell(t *testing.T, nashpath string, nashroot string) *testshell {

	shell, err := NewShell(nashpath, nashroot)
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	shell.SetStdout(&out)

	return &testshell{shell: shell, stdout: &out}
}

type nashDirs struct {
	path string
	lib string
	root string
	stdlib string
	cleanup func()
}

func setupNashDirs(t *testing.T) nashDirs {
	testdir, rmdir := tmpdir(t)

	nashpath := filepath.Join(testdir, "nashpath")
	nashroot := filepath.Join(testdir, "nashroot")
	
	nashlib := filepath.Join(nashpath, "lib")
	nashstdlib := filepath.Join(nashroot, "stdlib")
	
	mkdirAll(t, nashlib)
	mkdirAll(t, nashstdlib)
	
	return nashDirs{
		path: nashpath,
		lib: nashlib,
		root: nashroot,
		stdlib: nashstdlib,
		cleanup: rmdir,
	}
}

func tmpdir(t *testing.T) (string, func()) {
	t.Helper()
	
	dir, err := ioutil.TempDir("", "nash-import-tests")
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

func mkdirAll(t *testing.T, nashlib string) {
	err := os.MkdirAll(nashlib, os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
}

func writeFile(t *testing.T, filename string, data string) {
	err := ioutil.WriteFile(filename, []byte(data), os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
}

func chdir(t *testing.T, dir string) {
	t.Helper()
	
	err := os.Chdir(dir)
	if err != nil {
		t.Fatal(err)
	}
}

func getwd(t *testing.T) string {
	t.Helper()
	
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	
	return dir
}