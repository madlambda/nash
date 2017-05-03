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

func execFailure(t *testing.T, scriptContents string) {
	shell, err := nash.New()
	if err != nil {
		t.Fatalf("unexpected err: %s", err)
	}

	out, err := shell.ExecOutput("", scriptContents)

	if err == nil {
		t.Fatalf("expected err, got success, output: %s", string(out))
	}

	if len(out) > 0 {
		t.Fatalf("expected empty output, got: %s", string(out))
	}
}
