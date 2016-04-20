package nash

import (
	"os"
	"testing"
)

var (
	testDir, gopath, nashdPath string
)

func init() {
	gopath = os.Getenv("GOPATH")

	if gopath == "" {
		panic("Please, run tests from inside GOPATH")
	}

	testDir = gopath + "/src/github.com/tiago4orion/nash/" + "testfiles"
	nashdPath = gopath + "/src/github.com/tiago4orion/nash/cmd/nash/nash"

	if _, err := os.Stat(nashdPath); err != nil {
		panic("Please, run make build before running tests")
	}
}

func TestExecute(t *testing.T) {
	testfile := testDir + "/ex1.sh"

	sh := NewShell(false)
	sh.SetNashdPath(nashdPath)

	err := sh.Execute(testfile)

	if err != nil {
		t.Error(err)
		return
	}
}

func TestExecuteRfork(t *testing.T) {
	sh := NewShell(false)
	sh.SetNashdPath(nashdPath)

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
	sh.SetNashdPath(nashdPath)

	err := sh.ExecuteString("assignment", `
        name=i4k
        echo $name
        echo $path
        `)

	if err != nil {
		t.Error(err)
		return
	}

	err = sh.ExecuteString("list assignment", `
        name=(honda civic)
        echo $name
        `)

	if err != nil {
		t.Error(err)
		return
	}
}
