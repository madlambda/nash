package builtin_test

import (
	"testing"

	"github.com/NeowayLabs/nash"
)

func execSuccess(t *testing.T, scriptContents string) string {
	shell, err := nash.New()
	if err != nil {
		t.Fatalf("unexpected err: %s", err)
	}

	out, err := shell.ExecOutput("", scriptContents)

	if err != nil {
		t.Fatalf("unexpected err: %s", err)
	}
	return string(out)
}
