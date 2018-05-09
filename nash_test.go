package nash

import (
	"bytes"
	"os"
	"testing"
	"io/ioutil"

	"github.com/NeowayLabs/nash/sh"
	"github.com/NeowayLabs/nash/tests"
)

// only testing the public API
// bypass to internal sh.Shell

func TestExecuteFile(t *testing.T) {
	testfile := tests.Testdir + "/ex1.sh"

	var out bytes.Buffer
	shell, cleanup := newShell(t)
	defer cleanup()

	shell.SetNashdPath(tests.Nashcmd)
	shell.SetStdout(&out)
	shell.SetStderr(os.Stderr)
	shell.SetStdin(os.Stdin)

	err := shell.ExecuteFile(testfile)
	if err != nil {
		t.Error(err)
		return
	}

	if string(out.Bytes()) != "hello world\n" {
		t.Errorf("Wrong command output: '%s'", string(out.Bytes()))
		return
	}
}

func TestExecuteString(t *testing.T) {
	shell, cleanup := newShell(t)
	defer cleanup()

	var out bytes.Buffer

	shell.SetStdout(&out)

	err := shell.ExecuteString("-ínput-", "echo -n AAA")
	if err != nil {
		t.Error(err)
		return
	}

	if string(out.Bytes()) != "AAA" {
		t.Errorf("Unexpected '%s'", string(out.Bytes()))
		return
	}

	out.Reset()

	err = shell.ExecuteString("-input-", `
        PROMPT="humpback> "
        setenv PROMPT
        `)
	if err != nil {
		t.Error(err)
		return
	}

	prompt := shell.Prompt()
	if prompt != "humpback> " {
		t.Errorf("Invalid prompt = %s", prompt)
		return
	}

}

func TestSetvar(t *testing.T) {
	shell,cleanup := newShell(t)
	defer cleanup()
	
	shell.Newvar("__TEST__", sh.NewStrObj("something"))

	var out bytes.Buffer
	shell.SetStdout(&out)

	err := shell.Exec("TestSetvar", `echo -n $__TEST__`)

	if err != nil {
		t.Error(err)
		return
	}

	if string(out.Bytes()) != "something" {
		t.Errorf("Value differ: '%s' != '%s'", string(out.Bytes()), "something")
		return
	}

	val, ok := shell.Getvar("__TEST__")

	if !ok || val.String() != "something" {
		t.Errorf("Getvar doesn't work: '%s' != '%s'", val, "something")
		return
	}
}

func newShell(t *testing.T) (*Shell, func()) {
	t.Helper()
	
	nashpath, pathclean := tmpdir(t)
	nashroot, rootclean := tmpdir(t)
	
	s, err := New(nashpath, nashroot)
	if err != nil {
		t.Fatal(err)
	}
	return s, func() {
		pathclean()
		rootclean()
	}
}


func tmpdir(t *testing.T) (string, func()) {
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