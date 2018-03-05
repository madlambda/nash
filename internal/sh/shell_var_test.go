package sh

import (
	"fmt"
	"testing"

	"github.com/NeowayLabs/nash/tests"
)

func TestVarAssign(t *testing.T) {
	for _, test := range []execTestCase{
		{
			desc:           "simple init",
			code:           `var a = "1"; echo -n $a`,
			expectedStdout: "1",
		},
		{
			desc:        "variable does not exists",
			code:        `a = "1"; echo -n $a`,
			expectedErr: `<interactive>:1:0: Variable 'a' is not initialized. Use 'var a = <value>'`,
		},
		{
			desc:           "variable already initialized",
			code:           `var a = "1"; var a = "2"; echo -n $a`,
			expectedStdout: "2",
		},
		{
			desc:           "variable set",
			code:           `var a = "1"; a = "2"; echo -n $a`,
			expectedStdout: "2",
		},
		{
			desc: "global variable set",
			code: `var global = "1"
				fn somefunc() { global = "2" }
				somefunc()
				echo -n $global`,
			expectedStdout: "2",
		},
	} {
		t.Run(test.desc, func(t *testing.T) {
			testExec(t, test)
		})
	}
}

func TestVarExecAssign(t *testing.T) {

	for _, test := range []execTestCase{
		{
			desc: "simple exec var",
			code: `var heart <= echo -n "feed both wolves"
				echo -n $heart`,
			expectedStdout: "feed both wolves",
		},
		{
			desc:        "var do not exists",
			code:        `__a <= echo -n "fury"`,
			expectedErr: "<interactive>:1:0: Variable '__a' is not initialized. Use 'var __a = <value>'",
		},
		{
			desc: "multiple var same name",
			code: `var a = "1"
					var a = "2"
					var a = "3"
					echo -n $a`,
			expectedStdout: "3",
		},
		{
			desc: "multiple var same name with exec",
			code: `var a <= echo -n "1"
				var a <= echo -n "hello"
				echo -n $a`,
			expectedStdout: "hello",
		},
		{
			desc: "first variable is stdout",
			code: `var out <= echo -n "hello"
				echo -n $out`,
			expectedStdout: "hello",
		},
		{
			desc: "two variable, first stdout and second is status",
			code: `var stdout, status <= echo -n "bleh"
			echo -n $stdout $status`,
			expectedStdout: "bleh 0",
		},
		{
			desc: "three variables, stdout empty, stderr with data, status",
			code: fmt.Sprintf(`var out, err, st <= %s/write/write /dev/stderr "hello"
					echo $out
					echo $err
					echo -n $st`, tests.Stdbindir),
			expectedStdout: "\nhello\n0",
		},
		{
			desc: "three variables, stdout with data, stderr empty, status",
			code: fmt.Sprintf(`var out, err, st <= %s/write/write /dev/stdout "hello"
					echo $out
					echo $err
					echo -n $st`, tests.Stdbindir),
			expectedStdout: "hello\n\n0",
		},
	} {
		t.Run(test.desc, func(t *testing.T) {
			testExec(t, test)
		})
	}
}
