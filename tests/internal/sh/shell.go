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

// Exec runs the script code and returns the result of it.
func Exec(
	t *testing.T,
	nashpath string,
	scriptcode string,
	scriptargs ...string,
) (string, string, error) {
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
	cmd := exec.Command(nashpath, scriptargs...)

	stdout := bytes.Buffer{}
	stderr := bytes.Buffer{}

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	return stdout.String(), stderr.String(), err
}
