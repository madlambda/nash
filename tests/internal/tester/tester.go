// Package tester makes it easy to run multiple
// script test cases.
package tester

import (
	"testing"

	"github.com/NeowayLabs/nash/tests/internal/assert"
	"github.com/NeowayLabs/nash/tests/internal/sh"
)

type TestCase struct {
	Name           string
	ScriptCode     string
	ExpectedOutput string
}

func Run(t *testing.T, cases ...TestCase) {

	for _, tcase := range cases {
		t.Run(tcase.Name, func(t *testing.T) {
			t.Parallel()
			output := sh.ExecSuccess(t, tcase.ScriptCode)
			assert.EqualStrings(t, tcase.ExpectedOutput, output)
		})
	}
}
