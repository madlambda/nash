package tests

import (
	"testing"

	"github.com/NeowayLabs/nash/tests/internal/tester"
)

func TestListIndexing(t *testing.T) {
	tester.Run(t, nashcmd,
		tester.TestCase{
			Name: "PositionalAccess",
			ScriptCode: `
				a = ("1" "2")
				echo $a[0]
				echo $a[1]
			`,
			ExpectStdout: "1\n2\n",
		},
		tester.TestCase{
			Name: "PositionalAccessWithVar",
			ScriptCode: `
				a = ("1" "2")
				i = "0"
				echo $a[$i]
				i = "1"
				echo $a[$i]
			`,
			ExpectStdout: "1\n2\n",
		},
		tester.TestCase{
			Name: "Iteration",
			ScriptCode: `
				a = ("1" "2" "3")
				for x in $a {
					echo $x
				}
			`,
			ExpectStdout: "1\n2\n3\n",
		},
		tester.TestCase{
			Name: "IterateEmpty",
			ScriptCode: `
				a = ()
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
				a = ("1" "2" "3")
				echo $a[3]
			`,
			Fails: true,
			ExpectStderrToContain: "Index out of bounds",
		},
		tester.TestCase{
			Name: "IndexEmptyFails",
			ScriptCode: `
				a = ()
				echo $a[0]
			`,
			Fails: true,
			ExpectStderrToContain: "Index out of bounds",
		},
	)
}
