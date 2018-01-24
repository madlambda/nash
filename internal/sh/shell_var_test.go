package sh

import "testing"

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
			expectedErr: "variable do not exists",
		},
	} {
		t.Run(test.desc, func(t *testing.T) {
			testExec(t, test)
		})
	}
}
