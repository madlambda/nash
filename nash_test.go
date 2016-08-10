package nash

import (
	"bytes"
	"os"
	"testing"
)

// only testing the public API
// bypass to internal sh.Shell

var (
	gopath, testDir, nashdPath string
)

func init() {
	gopath = os.Getenv("GOPATH")

	if gopath == "" {
		panic("Please, run tests from inside GOPATH")
	}

	testDir = gopath + "/src/github.com/NeowayLabs/nash/" + "testfiles"
	nashdPath = gopath + "/src/github.com/NeowayLabs/nash/cmd/nash/nash"

	if _, err := os.Stat(nashdPath); err != nil {
		panic("Please, run make build before running tests")
	}
}

func TestExecuteFile(t *testing.T) {
	testfile := testDir + "/ex1.sh"

	var out bytes.Buffer

	sh, err := New()

	if err != nil {
		t.Error(err)
		return
	}

	sh.SetNashdPath(nashdPath)
	sh.SetStdout(&out)
	sh.SetStderr(os.Stderr)
	sh.SetStdin(os.Stdin)

	err = sh.ExecuteFile(testfile)

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
	sh, err := New()

	if err != nil {
		t.Error(err)
		return
	}

	var out bytes.Buffer

	sh.SetStdout(&out)

	err = sh.ExecuteString("-ínput-", "echo -n AAA")

	if err != nil {
		t.Error(err)
		return
	}

	if string(out.Bytes()) != "AAA" {
		t.Errorf("Unexpected '%s'", string(out.Bytes()))
		return
	}

	out.Reset()

	err = sh.ExecuteString("-input-", `
        PROMPT="humpback> "
        setenv PROMPT
        `)

	if err != nil {
		t.Error(err)
		return
	}

	prompt := sh.Prompt()

	if prompt != "humpback> " {
		t.Errorf("Invalid prompt = %s", prompt)
		return
	}

}

func TestSetDotDir(t *testing.T) {
	sh, err := New()

	if err != nil {
		t.Error(err)
		return
	}

	var out bytes.Buffer

	sh.SetStdout(&out)

	sh.SetDotDir("/tmp")

	dotDir := sh.DotDir()

	if dotDir != "/tmp" {
		t.Errorf("Invalid .nash = %s", dotDir)
		return
	}

	err = sh.ExecuteString("-ínput-", "echo -n $NASHPATH")

	if err != nil {
		t.Error(err)
		return
	}

	if string(out.Bytes()) != "/tmp" {
		t.Errorf("Unexpected '%s'", string(out.Bytes()))
		return
	}
}
