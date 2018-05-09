package builtin_test

import (
	"testing"
	
	"github.com/NeowayLabs/nash/internal/sh/internal/fixture"
)

func execSuccess(t *testing.T, scriptContents string) string {
	shell, cleanup := fixture.SetupShell(t)
	defer cleanup()

	out, err := shell.ExecOutput("", scriptContents)

	if err != nil {
		t.Fatalf("unexpected err: %s", err)
	}
	return string(out)
}

func execFailure(t *testing.T, scriptContents string) {
	shell, cleanup := fixture.SetupShell(t)
	defer cleanup()
	
	out, err := shell.ExecOutput("", scriptContents)

	if err == nil {
		t.Fatalf("expected err, got success, output: %s", string(out))
	}

	if len(out) > 0 {
		t.Fatalf("expected empty output, got: %s", string(out))
	}
}
