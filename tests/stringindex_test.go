package tests

import (
	"testing"

	"github.com/madlambda/nash/tests/internal/tester"
)

func TestStringIndexing(t *testing.T) {
	tester.Run(t, Nashcmd,
		tester.TestCase{
			Name: "IterateEmpty",
			ScriptCode: `
				var a = ""
				for x in $a {
					exit("1")
				}
				echo "ok"
			`,
			ExpectStdout: "ok\n",
		},
		tester.TestCase{
			Name: "IndexEmptyFails",
			ScriptCode: `
				var a = ""
				echo $a[0]
			`,
			Fails: true,
			ExpectStderrToContain: "IndexError",
		},
		tester.TestCase{
			Name: "IsImmutable",
			ScriptCode: `
				var a = "12"
				a[0] = "2"
				echo $a
			`,
			Fails: true,
			ExpectStderrToContain: "IndexError",
		},
	)
}
func TestStringIndexingASCII(t *testing.T) {
	tester.Run(t, Nashcmd,
		tester.TestCase{Name: "PositionalAccess",
			ScriptCode: `
				var a = "12"
				echo $a[0]
				echo $a[1]
			`,
			ExpectStdout: "1\n2\n",
		},
		tester.TestCase{
			Name: "PositionalAccessReturnsString",
			ScriptCode: `
				var a = "12"
				var x = $a[0] + $a[1]
				echo $x
			`,
			ExpectStdout: "12\n",
		},
		tester.TestCase{
			Name: "Len",
			ScriptCode: `
				var a = "12"
				var l <= len($a)
				echo $l
			`,
			ExpectStdout: "2\n",
		},
		tester.TestCase{
			Name: "Iterate",
			ScriptCode: `
				var a = "123"
				for x in $a {
					echo $x
				}
			`,
			ExpectStdout: "1\n2\n3\n",
		},
		tester.TestCase{
			Name: "IndexOutOfRangeFails",
			ScriptCode: `
				var a = "123"
				echo $a[3]
			`,
			Fails: true,
			ExpectStderrToContain: "IndexError",
		},
	)
}

func TestStringIndexingNonASCII(t *testing.T) {
	tester.Run(t, Nashcmd,
		tester.TestCase{Name: "PositionalAccess",
			ScriptCode: `
				var a = "⌘⌘"
				echo $a[0]
				echo $a[1]
			`,
			ExpectStdout: "⌘\n⌘\n",
		},
		tester.TestCase{
			Name: "Iterate",
			ScriptCode: `
				var a = "⌘⌘"
				for x in $a {
					echo $x
				}
			`,
			ExpectStdout: "⌘\n⌘\n",
		},
		tester.TestCase{
			Name: "PositionalAccessReturnsString",
			ScriptCode: `
				var a = "⌘⌘"
				var x = $a[0] + $a[1]
				echo $x
			`,
			ExpectStdout: "⌘⌘\n",
		},
		tester.TestCase{
			Name: "Len",
			ScriptCode: `
				var a = "⌘⌘"
				var l <= len($a)
				echo $l
			`,
			ExpectStdout: "2\n",
		},
		tester.TestCase{
			Name: "IndexOutOfRangeFails",
			ScriptCode: `
				var a = "⌘⌘"
				echo $a[2]
			`,
			Fails: true,
			ExpectStderrToContain: "IndexError",
		},
	)
}
