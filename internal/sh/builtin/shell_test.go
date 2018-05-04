package builtin_test

import (
	"testing"
	
	"github.com/NeowayLabs/nash"
)

func newShell(t *testing.T) *nash.Shell {
	shell, err := nash.New("/tmp/testnashpath", "/tmp/testnashroot")

	if err != nil {
		t.Fatal(err)
	}

	return shell
}