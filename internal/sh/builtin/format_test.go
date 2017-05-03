package builtin_test

import "testing"

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
		"ncallsRegressionTest": {
			script: `
				fn formatstuff() {
					r <= format("hello%s", "world")
					echo $r
				}
				formatstuff()
				formatstuff()
			`,
			output: "helloworld\nhelloworld\n",
		},
		"ncallsWithVarsRegressionTest": {
			script: `
				fn formatstuff() {
					b = "world"
					r <= format("hello%s", $b)
					s <= format("hackthe%s", $b)
					echo $r
					echo $s
				}
				formatstuff()
				formatstuff()
			`,
			output: "helloworld\nhacktheworld\nhelloworld\nhacktheworld\n",
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
			output := execSuccess(t, desc.script)
			if output != desc.output {
				t.Fatalf("got %q expected %q", output, desc.output)
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
			execFailure(t, desc.script)
		})
	}
}
