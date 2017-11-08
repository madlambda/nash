package tests

import (
	"testing"

	"github.com/NeowayLabs/nash/tests/internal/assert"
	"github.com/NeowayLabs/nash/tests/internal/sh"
)

func TestListIndex(t *testing.T) {
	output := sh.ExecSuccess(t, `
		a = ("1" "2")
		echo $a[0]
		echo $a[1]
	`)
	expectedOutput := "1\n2\n"
	assert.EqualStrings(t, expectedOutput, output)
}
