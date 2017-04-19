package builtin_test

import (
	"bytes"
	"testing"

	"github.com/NeowayLabs/nash"
)

func TestPrintf(t *testing.T) {
	type printfDesc struct {
		script string
		output string
	}

	tests := map[string]printfDesc{
		"textonly": {
			script: `printf("helloworld")`,
			output: "helloworld",
		},
		"fmtstring": {
			script: `printf("%s:%s", "hello", "world")`,
			output: "hello:world",
		},
	}

	for name, desc := range tests {
		t.Run(name, func(t *testing.T) {
			var output bytes.Buffer
			shell, err := nash.New()

			if err != nil {
				t.Fatalf("unexpected err: %s", err)
			}

			shell.SetStdout(&output)
			err = shell.Exec("", desc.script)

			if err != nil {
				t.Fatalf("unexpected err: %s", err)
			}

			if output.String() != desc.output {
				t.Fatalf("got %q expected %q", output.String(), desc.output)
			}
		})
	}
}

func TestPrintfErrors(t *testing.T) {
	type printfDesc struct {
		script string
	}

	tests := map[string]printfDesc{
		"noParams": {
			script: `printf()`,
		},
	}

	for name, desc := range tests {
		t.Run(name, func(t *testing.T) {
			var output bytes.Buffer
			shell, err := nash.New()

			if err != nil {
				t.Fatalf("unexpected err: %s", err)
			}

			shell.SetStdout(&output)
			err = shell.Exec("", desc.script)

			if err == nil {
				t.Fatalf("expected err, got success, output: %s", output)
			}

			if output.Len() > 0 {
				t.Fatalf("expected empty output, got: %s", output)
			}
		})
	}
}
