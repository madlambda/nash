package sh

import "testing"

func TestFunctionsClosures(t *testing.T) {
	for _, test := range []execTestCase{
		{
			desc: "simpleClosure",
			execStr: `
				fn func(a) {
					fn closure() {
						print($a)
					}
					return $closure
				}

				x <= func("1")
				y <= func("2")
				$x()
				$y()
			`,
			expectedStdout: "12",
		},
		{
			desc: "eachCallCreatesNewVar",
			execStr: `
				fn func() {
					a = ()
					fn add(elem) {
						a <= append($a, $elem)
						print("a:%s,",$a)
					}
					return $add
				}

				add <= func()
				$add("1")
				$add("3")
				$add("5")
			`,
			expectedStdout: "a:1,a:3,a:5,",
		},
	} {
		t.Run(test.desc, func(t *testing.T) {
			testExec(t, test)
		})
	}
}
