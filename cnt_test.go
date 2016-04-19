package cnt

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

var (
	testDir, gopath, cntdPath string
)

func init() {
	gopath = os.Getenv("GOPATH")

	if gopath == "" {
		panic("Please, run tests from inside GOPATH")
	}

	testDir = gopath + "/src/github.com/tiago4orion/cnt/" + "testfiles"
	cntdPath = gopath + "/src/github.com/tiago4orion/cnt/cmd/cnt/cnt"

	if _, err := os.Stat(cntdPath); err != nil {
		panic("Please, run make build before running tests")
	}
}

func TestExecute(t *testing.T) {
	testfile := testDir + "/ex1.cnt"

	sh := NewShell(false)
	sh.SetCntdPath(cntdPath)

	err := sh.Execute(testfile)

	if err != nil {
		t.Error(err)
		return
	}
}

func TestExecuteRfork(t *testing.T) {
	var out bytes.Buffer

	sh := NewShell(false)
	sh.SetCntdPath(cntdPath)
	sh.SetStdout(&out)
	sh.SetStderr(os.Stderr)

	err := sh.ExecuteString("rfork test", `
        rfork u {
            id -u
        }
        `)

	if err != nil {
		t.Error(err)
	}

	if strings.Trim(string(out.Bytes()), "\n") != "0" {
		t.Errorf("Differ '%s' != '%s'", "0", out.Bytes())
	}
}
