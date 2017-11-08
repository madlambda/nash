// Package shell makes it easier to run nash scripts for test purposes
package sh

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"testing"

	"github.com/NeowayLabs/nash/tests/internal/assert"
)

const projectnash = "../cmd/nash/nash"

// ExecSuccess fails the test case if the script exits with failure.
// It will coalesce the stdout and stderr together and return as a string.
func ExecSuccess(
	t *testing.T,
	scriptcode string,
	scriptargs ...string,
) string {
	scriptfile, err := ioutil.TempFile("", "testshell")
	assert.NoError(t, err, "creating tmp file")

	defer func() {
		err := scriptfile.Close()
		assert.NoError(t, err, "closing tmp file")
		err = os.Remove(scriptfile.Name())
		assert.NoError(t, err, "deleting tmp file")
	}()

	_, err = io.Copy(scriptfile, bytes.NewBufferString(scriptcode))
	assert.NoError(t, err, "writing script code to tmp file")

	scriptargs = append([]string{scriptfile.Name()}, scriptargs...)
	cmd := exec.Command(projectnash, scriptargs...)

	stdout := bytes.Buffer{}
	stderr := bytes.Buffer{}

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	output := stdout.String() + stderr.String()
	if err != nil {
		t.Fatalf("error[%s] output[%s]", err, output)
	}

	return output
}
