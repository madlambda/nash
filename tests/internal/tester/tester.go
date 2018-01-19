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
	ExpectStdout          string
	ExpectStderrToContain string
	Fails                 bool
}

func Run(t *testing.T, nashcmd string, cases ...TestCase) {
	for _, tcase := range cases {
		t.Run(tcase.Name, func(t *testing.T) {
			stdout, stderr, err := sh.Exec(t, nashcmd, tcase.ScriptCode)
			if !tcase.Fails {
				if err != nil {
					t.Fatalf(
						"error[%s] stdout[%s] stderr[%s]",
						err,
						stdout,
						stderr,
					)
				}

				if stderr != "" {
					t.Fatalf(
						"unexpected stderr[%s], on success no stderr is expected",
						stderr,
					)
				}
			}

			if tcase.ExpectStdout != "" {
				assert.EqualStrings(t, tcase.ExpectStdout, stdout)
			}

			if tcase.ExpectStderrToContain != "" {
				assert.ContainsString(t, stderr, tcase.ExpectStderrToContain)
			}
		})
	}
}
