package sh_test

import (
	"bytes"
	"path/filepath"
	"testing"
	
	"github.com/NeowayLabs/nash/internal/sh"
	"github.com/NeowayLabs/nash/internal/sh/internal/fixture"
)


func TestImportsLibFromNashPathLibDir(t *testing.T) {
	
	nashdirs := fixture.SetupNashDirs(t)
	defer nashdirs.Cleanup()
	
	writeFile(t, filepath.Join(nashdirs.Lib, "lib.sh"), `
		fn test() {
			echo "hasnashpath"
		}
	`)

	newTestShell(t, nashdirs.Path, nashdirs.Root).ExecCheckingOutput(t, `
		import lib
		test()
	`, "hasnashpath\n")
}

func TestImportsLibFromNashPathLibDirBeforeNashRootStdlib(t *testing.T) {
	
	nashdirs := fixture.SetupNashDirs(t)
	defer nashdirs.Cleanup()

	writeFile(t, filepath.Join(nashdirs.Lib, "lib.sh"), `
		fn test() {
			echo "libcode"
		}
	`)
	
	writeFile(t, filepath.Join(nashdirs.Stdlib, "lib.sh"), `
		fn test() {
			echo "stdlibcode"
		}
	`)

	newTestShell(t, nashdirs.Path, nashdirs.Root).ExecCheckingOutput(t, `
		import lib
		test()
	`, "libcode\n")
}

func TestImportsLibFromNashRootStdlib(t *testing.T) {
	
	nashdirs := fixture.SetupNashDirs(t)
	defer nashdirs.Cleanup()
	
	writeFile(t, filepath.Join(nashdirs.Stdlib, "lib.sh"), `
		fn test() {
			echo "stdlibcode"
		}
	`)

	newTestShell(t, nashdirs.Path, nashdirs.Root).ExecCheckingOutput(t, `
		import lib
		test()
	`, "stdlibcode\n")
}

func TestImportsLibFromWorkingDirBeforeLibAndStdlib(t *testing.T) {
	
	workingdir, rmdir := fixture.Tmpdir(t)
	defer rmdir()
	
	curwd := getwd(t)
	chdir(t, workingdir)
	defer chdir(t, curwd)
	
	nashdirs := fixture.SetupNashDirs(t)
	defer nashdirs.Cleanup()
	
	writeFile(t, filepath.Join(workingdir, "lib.sh"), `
		fn test() {
			echo "localcode"
		}
	`)
	
	writeFile(t, filepath.Join(nashdirs.Lib, "lib.sh"), `
		fn test() {
			echo "libcode"
		}
	`)
	
	writeFile(t, filepath.Join(nashdirs.Stdlib, "lib.sh"), `
		fn test() {
			echo "stdlibcode"
		}
	`)
	
	newTestShell(t, nashdirs.Path, nashdirs.Root).ExecCheckingOutput(t, `
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
	
	validDir, rmdir := fixture.Tmpdir(t)
	defer rmdir()

	validfile := filepath.Join(validDir, "notdir")	
	writeFile(t, validfile, "whatever")
	
	cases := []testCase {
		{
			name: "EmptyNashPath",
			nashpath: "",
			nashroot: validDir,
		},
		{
			name: "NashPathDontExists",
			nashpath: filepath.Join(validDir, "dontexists"),
			nashroot: validDir,
		},
		{
			name: "EmptyNashRoot",
			nashpath: validDir,
			nashroot: "",
		},
		{
			name: "NashRootDontExists",
			nashroot: filepath.Join(validDir, "dontexists"),
			nashpath: validDir,
		},
		{
			name: "NashPathIsFile",
			nashroot: validDir,
			nashpath: validfile,
		},
		{
			name: "NashRootIsFile",
			nashroot: validfile,
			nashpath: validDir,
		},
		{
			name: "NashPathIsRelative",
			nashroot: validDir,
			nashpath: "./",
		},
		{
			name: "NashRootIsRelative",
			nashroot: "./",
			nashpath: validDir,
		},
		{
			name: "NashRootAndNashPathAreEqual",
			nashroot: validDir,
			nashpath: validDir,
		},
	}
	
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := sh.NewShell(c.nashpath, c.nashroot)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}


type testshell struct {
	shell  *sh.Shell
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

	shell, err := sh.NewShell(nashpath, nashroot)
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	shell.SetStdout(&out)

	return &testshell{shell: shell, stdout: &out}
}
