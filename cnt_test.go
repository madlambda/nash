package cnt

import (
	"os"
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
	sh := NewShell(false)
	sh.SetCntdPath(cntdPath)

	err := sh.ExecuteString("rfork test", `
        rfork u {
            id -u
        }
        `)

	if err != nil {
		t.Error(err)
	}
}

func TestExecuteAssignment(t *testing.T) {
	sh := NewShell(false)
	sh.SetCntdPath(cntdPath)

	err := sh.ExecuteString("assignment", `
        name=i4k
        echo $name
        echo $path
        `)

	if err != nil {
		t.Error(err)
		return
	}
}
