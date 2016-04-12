package cnt

import (
	"os"
	"testing"
)

var testDir string

func init() {
	gopath := os.Getenv("GOPATH")
	testDir = gopath + "/src/github.com/tiago4orion/cnt/" + "testfiles"
}

func TestExecute(t *testing.T) {
	testfile := testDir + "/ex1.cnt"

	err := Execute(testfile)

	if err != nil {
		t.Error(err)
		return
	}
}
