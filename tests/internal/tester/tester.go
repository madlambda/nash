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
	Fails          bool
}

func Run(t *testing.T, cases ...TestCase) {

	for _, tcase := range cases {
		t.Run(tcase.Name, func(t *testing.T) {
			output, err := sh.Exec(t, tcase.ScriptCode)
			if !tcase.Fails {
				if err != nil {
					t.Fatalf("error[%s] output[%s]", err, output)
				}
			}

			if tcase.ExpectedOutput != "" {
				assert.EqualStrings(t, tcase.ExpectedOutput, output)
			}
		})
	}
}
