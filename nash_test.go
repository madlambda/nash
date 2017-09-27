package nash

import (
	"fmt"
	"bytes"
	"os"
	"testing"
	"os/user"
	"path"
	"path/filepath"
	"runtime"

	"github.com/NeowayLabs/nash/sh"
)

// only testing the public API
// bypass to internal sh.Shell

var (
	gopath, testDir, nashdPath string
)

func init() {
	gopath = os.Getenv("GOPATH")

	if gopath == "" {
		usr, err := user.Current()
		if err != nil {
			panic(err)
		}
		if usr.HomeDir == "" {
			panic("Unable to discover GOPATH")	
		}
		gopath = path.Join(usr.HomeDir, "go")
	}

	testDir = filepath.FromSlash(path.Join(gopath, "/src/github.com/NeowayLabs/nash/", "testfiles"))
	nashdPath = filepath.FromSlash(path.Join(gopath, "/src/github.com/NeowayLabs/nash/cmd/nash/nash"))

	if runtime.GOOS == "windows" {
		nashdPath += ".exe"
	}

	if _, err := os.Stat(nashdPath); err != nil {
		panic(fmt.Errorf("Please, run make build before running tests: %s", err.Error()))
	}
}

func TestExecuteFile(t *testing.T) {
	testfile := testDir + "/ex1.sh"

	var out bytes.Buffer

	shell, err := New()

	if err != nil {
		t.Error(err)
		return
	}

	shell.SetNashdPath(nashdPath)
	shell.SetStdout(&out)
	shell.SetStderr(os.Stderr)
	shell.SetStdin(os.Stdin)

	err = shell.ExecuteFile(testfile)

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
	shell, err := New()

	if err != nil {
		t.Error(err)
		return
	}

	var out bytes.Buffer

	shell.SetStdout(&out)

	err = shell.ExecuteString("-ínput-", "echo -n AAA")

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

func TestSetDotDir(t *testing.T) {
	shell, err := New()
	if err != nil {
		t.Error(err)
		return
	}

	var out bytes.Buffer

	shell.SetStdout(&out)
	shell.SetDotDir("/tmp")

	dotDir := shell.DotDir()
	if dotDir != "/tmp" {
		t.Errorf("Invalid .nash = %s", dotDir)
		return
	}

	err = shell.ExecuteString("-ínput-", "echo -n $NASHPATH")
	if err != nil {
		t.Error(err)
		return
	}

	if string(out.Bytes()) != "/tmp" {
		t.Errorf("Unexpected '%s'", string(out.Bytes()))
		return
	}
}

func TestSetvar(t *testing.T) {
	shell, err := New()

	if err != nil {
		t.Error(err)
		return
	}

	shell.Setvar("__TEST__", sh.NewStrObj("something"))

	var out bytes.Buffer
	shell.SetStdout(&out)

	err = shell.Exec("TestSetvar", `echo -n $__TEST__`)

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
