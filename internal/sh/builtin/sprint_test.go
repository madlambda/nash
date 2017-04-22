package builtin_test

import (
	"bytes"
	"testing"

	"github.com/NeowayLabs/nash"
)

func TestSprint(t *testing.T) {
	type sprintDesc struct {
		script string
		output string
	}

	tests := map[string]sprintDesc{
		"textonly": {
			script: `
				r <= sprint("helloworld")
				echo $r
			`,
			output: "helloworld\n",
		},
		"fmtstring": {
			script: `
				r <= sprint("%s:%s", "hello", "world")
				echo $r
			`,
			output: "hello:world\n",
		},
		"fmtlist": {
			script: `
				list = ("1" "2" "3")
				r <= sprint("%s:%s", "list", $list)
				echo $r
			`,
			output: "list:1 2 3\n",
		},
		"funconly": {
			script: `
				fn func() {}
				r <= sprint($func)
				echo $r
			`,
			output: "<fn func>\n",
		},
		"funcfmt": {
			script: `
				fn func() {}
				r <= sprint("calling:%s", $func)
				echo $r
			`,
			output: "calling:<fn func>\n",
		},
		"listonly": {
			script: `
				list = ("1" "2" "3")
				r <= sprint($list)
				echo $r
			`,
			output: "1 2 3\n",
		},
		"listoflists": {
			script: `
				list = (("1" "2" "3") ("4" "5" "6"))
				r <= sprint("%s:%s", "listoflists", $list)
				echo $r
			`,
			output: "listoflists:1 2 3 4 5 6\n",
		},
		"listasfmt": {
			script: `
				list = ("%s" "%s")
				r <= sprint($list, "1", "2")
				echo $r
			`,
			output: "1 2\n",
		},
		"invalidFmt": {
			script: `
				r <= sprint("%d%s", "invalid")
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

func TestSprintfErrors(t *testing.T) {
	type sprintDesc struct {
		script string
	}

	tests := map[string]sprintDesc{
		"noParams": {script: `sprint()`},
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
