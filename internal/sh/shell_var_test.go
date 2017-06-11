package sh

import "testing"

func TestVarAssign(t *testing.T) {
	for _, test := range []execTestCase{
		{
			desc:           "simple init",
			execStr:        `var a = "1"; echo -n $a`,
			expectedStdout: "1",
		},
		{
			desc:        "variable does not exists",
			execStr:     `a = "1"; echo -n $a`,
			expectedErr: `<interactive>:1:0: Variable 'a' is not initialized. Use 'var a = <value>'`,
		},
		{
			desc:        "variable already initialized",
			execStr:     `var a = "1"; var a = "2"; echo -n $a`,
			expectedErr: `<interactive>:1:17: Cannot redeclare variables (a) in current block, at least one of them must be new`,
		},
		{
			desc:           "variable set",
			execStr:        `var a = "1"; a = "2"; echo -n $a`,
			expectedStdout: "2",
		},
		{
			desc: "global variable set",
			execStr: `var global = "1"
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
