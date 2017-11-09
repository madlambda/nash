package tests

import (
	"testing"

	"github.com/NeowayLabs/nash/tests/internal/tester"
)

func TestListIndexing(t *testing.T) {
	tester.Run(t, tester.TestCase{
		Name: "PositionalAccess",
		ScriptCode: `
			a = ("1" "2")
			echo $a[0]
			echo $a[1]
		`,
		ExpectedOutput: "1\n2\n",
	}, tester.TestCase{
		Name: "Iteration",
		ScriptCode: `
			a = ("1" "2" "3")
			for x in $a {
				echo $x
			}
		`,
		ExpectedOutput: "1\n2\n3\n",
	})
}
