package builtin_test

import (
	"bytes"
	"testing"

	"github.com/NeowayLabs/nash"
)

func TestFormat(t *testing.T) {
	type formatDesc struct {
		script string
		output string
	}

	tests := map[string]formatDesc{
		"textonly": {
			script: `
				r <= format("helloworld")
				echo $r
			`,
			output: "helloworld\n",
		},
		"fmtstring": {
			script: `
				r <= format("%s:%s", "hello", "world")
				echo $r
			`,
			output: "hello:world\n",
		},
		"fmtlist": {
			script: `
				list = ("1" "2" "3")
				r <= format("%s:%s", "list", $list)
				echo $r
			`,
			output: "list:1 2 3\n",
		},
		"funconly": {
			script: `
				fn func() {}
				r <= format($func)
				echo $r
			`,
			output: "<fn func>\n",
		},
		"funcfmt": {
			script: `
				fn func() {}
				r <= format("calling:%s", $func)
				echo $r
			`,
			output: "calling:<fn func>\n",
		},
		"listonly": {
			script: `
				list = ("1" "2" "3")
				r <= format($list)
				echo $r
			`,
			output: "1 2 3\n",
		},
		"listoflists": {
			script: `
				list = (("1" "2" "3") ("4" "5" "6"))
				r <= format("%s:%s", "listoflists", $list)
				echo $r
			`,
			output: "listoflists:1 2 3 4 5 6\n",
		},
		"listasfmt": {
			script: `
				list = ("%s" "%s")
				r <= format($list, "1", "2")
				echo $r
			`,
			output: "1 2\n",
		},
		"invalidFmt": {
			script: `
				r <= format("%d%s", "invalid")
				echo $r
			`,
			output: "%!d(string=invalid)%!s(MISSING)\n",
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

func TestFormatfErrors(t *testing.T) {
	type formatDesc struct {
		script string
	}

	tests := map[string]formatDesc{
		"noParams": {script: `format()`},
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
