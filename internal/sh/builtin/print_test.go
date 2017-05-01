package builtin_test

import "testing"

func TestPrint(t *testing.T) {
	type printDesc struct {
		script string
		output string
	}

	tests := map[string]printDesc{
		"textonly": {
			script: `print("helloworld")`,
			output: "helloworld",
		},
		"fmtstring": {
			script: `print("%s:%s", "hello", "world")`,
			output: "hello:world",
		},
		"fmtlist": {
			script: `
				list = ("1" "2" "3")
				print("%s:%s", "list", $list)
			`,
			output: "list:1 2 3",
		},
		"funconly": {
			script: `
				fn func() {}
				print($func)
			`,
			output: "<fn func>",
		},
		"funcfmt": {
			script: `
				fn func() {}
				print("calling:%s", $func)
			`,
			output: "calling:<fn func>",
		},
		"listonly": {
			script: `
				list = ("1" "2" "3")
				print($list)
			`,
			output: "1 2 3",
		},
		"listoflists": {
			script: `
				list = (("1" "2" "3") ("4" "5" "6"))
				print("%s:%s", "listoflists", $list)
			`,
			output: "listoflists:1 2 3 4 5 6",
		},
		"listasfmt": {
			script: `
				list = ("%s" "%s")
				print($list, "1", "2")
			`,
			output: "1 2",
		},
		"invalidFmt": {
			script: `print("%d%s", "invalid")`,
			output: "%!d(string=invalid)%!s(MISSING)",
		},
	}

	for name, desc := range tests {
		t.Run(name, func(t *testing.T) {
			output := execSuccess(t, desc.script)
			if output != desc.output {
				t.Fatalf("got %q expected %q", output, desc.output)
			}
		})
	}
}

func TestPrintErrors(t *testing.T) {
	type printDesc struct {
		script string
	}

	tests := map[string]printDesc{
		"noParams": {
			script: `print()`,
		},
	}

	for name, desc := range tests {
		t.Run(name, func(t *testing.T) {
			execFailure(t, desc.script)
		})
	}
}
