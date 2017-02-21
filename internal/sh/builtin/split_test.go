package builtin_test

import (
	"bytes"
	"testing"

	"github.com/NeowayLabs/nash"
)

func TestSplit(t *testing.T) {
	type splitDesc struct {
		script string
		word   string
		sep    string
		result string
	}

	tests := map[string]splitDesc{
		"space": {
			script: "./testdata/split.sh",
			word:   "a b c",
			sep:    " ",
			result: "a\nb\nc\n",
		},
		"pipes": {
			script: "./testdata/split.sh",
			word:   "1|2|3",
			sep:    "|",
			result: "1\n2\n3\n",
		},
		"nosplit": {
			script: "./testdata/split.sh",
			word:   "nospaces",
			sep:    " ",
			result: "nospaces\n",
		},
		"splitfunc": {
			script: "./testdata/splitfunc.sh",
			word:   "hah",
			sep:    "a",
			result: "h\nh\n",
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
			err = shell.ExecFile(desc.script, "mock cmd name", desc.word, desc.sep)

			if err != nil {
				t.Fatalf("unexpected err: %s", err)
			}

			if output.String() != desc.result {
				t.Fatalf("got %q expected %q", output.String(), desc.result)
			}
		})
	}
}
