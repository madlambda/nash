package sh

import (
	"bytes"
	"io/ioutil"
	"path/filepath"
	"os"
	"testing"
)

func TestImportsLibFromNashPathLibDir(t *testing.T) {
	
	testdir, rmdir := tmpdir(t)
	defer rmdir()

	nashpath := filepath.Join(testdir, "nashpath")
	nashlib := filepath.Join(nashpath, "lib")
	
	mkdirAll(t, nashlib)

	writeFile(t, filepath.Join(nashlib, "lib.sh"), `
		fn test() {
			echo "hasnashpath"
		}
	`)

	newTestShell(t, nashpath, "").ExecCheckingOutput(t, `
		import lib
		test()
	`, "hasnashpath\n")
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