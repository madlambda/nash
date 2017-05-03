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
			desc: "closuresSharingState",
			execStr: `
				fn func() {
					a = ()
					fn add(elem) {
						a <= append($a, $elem)
					}
					fn view() {
						print($a)
					}
					return $add, $view
				}

				add, view <= func()
				$add("1")
				$add("3")
				$view()
			`,
			expectedStdout: "1 3",
		},
	} {
		t.Run(test.desc, func(t *testing.T) {
			testExec(t, test)
		})
	}
}
