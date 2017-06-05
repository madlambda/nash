package sh

import "testing"

func TestVarAssign(t *testing.T) {
	for _, test := range []execTestCase{
		{
			desc:           "simple var",
			execStr:        `var a = "1"; echo -n $a`,
			expectedStdout: "1",
		},
	} {
		t.Run(test.desc, func(t *testing.T) {
			testExec(t, test)
		})
	}
}
