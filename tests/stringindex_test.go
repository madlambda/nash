package tests

import (
	"testing"

	"github.com/NeowayLabs/nash/tests/internal/tester"
)

func TestStringIndexing(t *testing.T) {
	tester.Run(t, nashcmd,
		tester.TestCase{
			Name: "PositionalAccess",
			ScriptCode: `
				a = "12"
				echo $a[0]
				echo $a[1]
			`,
			ExpectStdout: "1\n2\n",
		},
		tester.TestCase{
			Name: "Iteration",
			ScriptCode: `
				a = "123"
				for x in $a {
					echo $x
				}
			`,
			ExpectStdout: "1\n2\n3\n",
		},
		tester.TestCase{
			Name: "IterateEmpty",
			ScriptCode: `
				a = ""
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
				a = "123"
				echo $a[3]
			`,
			Fails: true,
			ExpectStderrToContain: "IndexError",
		},
		tester.TestCase{
			Name: "IndexEmptyFails",
			ScriptCode: `
				a = ""
				echo $a[0]
			`,
			Fails: true,
			ExpectStderrToContain: "IndexError",
		},
	)
}
