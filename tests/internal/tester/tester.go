// Package tester makes it easy to run multiple
// script test cases.
package tester

import (
	"testing"

	"github.com/NeowayLabs/nash/tests/internal/assert"
	"github.com/NeowayLabs/nash/tests/internal/sh"
)

type TestCase struct {
	Name                  string
	ScriptCode            string
	ExpectOutput          string
	ExpectOutputToContain string
	Fails                 bool
}

func Run(t *testing.T, nashcmd string, cases ...TestCase) {

	for _, tcase := range cases {
		t.Run(tcase.Name, func(t *testing.T) {
			output, err := sh.Exec(t, nashcmd, tcase.ScriptCode)
			if !tcase.Fails {
				if err != nil {
					t.Fatalf("error[%s] output[%s]", err, output)
				}
			}

			if tcase.ExpectOutput != "" {
				assert.EqualStrings(t, tcase.ExpectOutput, output)
			}

			if tcase.ExpectOutputToContain != "" {
				assert.ContainsString(t, output, tcase.ExpectOutputToContain)
			}
		})
	}
}
