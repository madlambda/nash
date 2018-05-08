package sh

import (
	"bytes"
	"path/filepath"
	"testing"
)


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

func TestErrorOnInvalidSearchPaths(t *testing.T) {
	type testCase struct {
		name string
		nashpath string
		nashroot string
	}
	
	// TODO: Fail on path exists but it is not dir
	// TODO: Fail if NASHROOT == NASHPATH
	
	validpath, rmdir := tmpdir(t)
	defer rmdir()
	
	cases := []testCase {
		{
			name: "EmptyNashPath",
			nashpath: "",
			nashroot: validpath,
		},
		{
			name: "NashPathDontExists",
			nashpath: filepath.Join(validpath, "dontexists"),
			nashroot: validpath,
		},
	}
	
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := NewShell(c.nashpath, c.nashroot)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
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
