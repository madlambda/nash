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
		"fmtlist": {
			script: `
				list = ("1" "2" "3")
				printf("%s:%s", "list", $list)
			`,
			output: "list:1 2 3",
		},
		"listonly": {
			script: `
				list = ("1" "2" "3")
				printf($list)
			`,
			output: "1 2 3",
		},
		"listoflists": {
			script: `
				list = (("1" "2" "3") ("4" "5" "6"))
				printf("%s:%s", "listoflists", $list)
			`,
			output: "listoflists:1 2 3 4 5 6",
		},
		"listasfmt": {
			script: `
				list = ("%s" "%s")
				printf($list, "1", "2")
			`,
			output: "1 2",
		},
		"invalidFmt": {
			script: `printf("%d%s", "invalid")`,
			output: "%!d(string=invalid)%!s(MISSING)",
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
