package tests

import (
	"testing"

	"github.com/NeowayLabs/nash/tests/internal/tester"
)

func TestListIndexing(t *testing.T) {
	tester.Run(t, Nashcmd,
		tester.TestCase{
			Name: "PositionalAccess",
			ScriptCode: `
				var a = ("1" "2")
				echo $a[0]
				echo $a[1]
			`,
			ExpectStdout: "1\n2\n",
		},
		tester.TestCase{
			Name: "PositionalAssigment",
			ScriptCode: `
				var a = ("1" "2")
				a[0] = "9"
				a[1] = "p"
				echo $a[0] + $a[1]
			`,
			ExpectStdout: "9p\n",
		},
		tester.TestCase{
			Name: "PositionalAccessWithVar",
			ScriptCode: `
				var a = ("1" "2")
				var i = "0"
				echo $a[$i]
				i = "1"
				echo $a[$i]
			`,
			ExpectStdout: "1\n2\n",
		},
		tester.TestCase{
			Name: "Iteration",
			ScriptCode: `
				var a = ("1" "2" "3")
				for x in $a {
					echo $x
				}
			`,
			ExpectStdout: "1\n2\n3\n",
		},
		tester.TestCase{
			Name: "IterateEmpty",
			ScriptCode: `
				var a = ()
				for x in $a {
					exit("1")
				}
				echo "ok"
			`,
			ExpectStdout: "ok\n",
		},
		tester.TestCase{
			Name: "IndexOutOfRangeFails",
			ScriptCode: `
				var a = ("1" "2" "3")
				echo $a[3]
			`,
			Fails: true,
			ExpectStderrToContain: "IndexError",
		},
		tester.TestCase{
			Name: "IndexEmptyFails",
			ScriptCode: `
				var a = ()
				echo $a[0]
			`,
			Fails: true,
			ExpectStderrToContain: "IndexError",
		},
	)
}
