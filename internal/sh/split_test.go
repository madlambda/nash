package sh

import (
	"testing"

	"github.com/NeowayLabs/nash/sh"
)

type splitTests struct {
	expected []sh.Obj
	content  string
	sep      sh.Obj
}

func testSplitTable(t *testing.T, tests []splitTests) {
	for _, test := range tests {
		testSplit(t, test.content, test.sep, test.expected)
	}
}

func testSplit(t *testing.T, content string, sep sh.Obj, expected []sh.Obj) {
	shell, err := NewShell()

	if err != nil {
		t.Fatal(err)
	}

	splitfn := NewSplitFn(shell)
	splitfn.SetArgs([]sh.Obj{
		sh.NewStrObj(content),
		sep,
	})

	err = splitfn.Start()

	if err != nil {
		t.Fatal(err)
	}

	err = splitfn.Wait()

	if err != nil {
		t.Fatal(err)
	}

	result := splitfn.Results()

	if result.Type() != sh.ListType {
		t.Fatalf("Splitfn returns wrong type")
	}

	values := result.(*sh.ListObj).List()

	if len(values) != len(expected) {
		t.Fatalf("Expected %d values, but got %d",
			len(expected), len(values))
	}

	for i := 0; i < len(values); i++ {
		if values[i].Type() != sh.StringType {
			t.Fatalf("Split must return list of strings %v",
				values[i])
		}

		v := values[i].(*sh.StrObj).Str()
		e := expected[i].(*sh.StrObj).Str()

		if v != e {
			t.Fatalf("Values differ: '%s' != '%s'", e, v)
		}
	}

}

func TestSplitFnBasic(t *testing.T) {
	testSplitTable(t, []splitTests{
		{
			content: "some thing",
			expected: []sh.Obj{
				sh.NewStrObj("some"),
				sh.NewStrObj("thing"),
			},
			sep: sh.NewStrObj(" "),
		},
		{
			content: "1 2 3 4 5 6 7 8 9 10",
			expected: []sh.Obj{
				sh.NewStrObj("1"),
				sh.NewStrObj("2"),
				sh.NewStrObj("3"),
				sh.NewStrObj("4"),
				sh.NewStrObj("5"),
				sh.NewStrObj("6"),
				sh.NewStrObj("7"),
				sh.NewStrObj("8"),
				sh.NewStrObj("9"),
				sh.NewStrObj("10"),
			},
			sep: sh.NewStrObj(" "),
		},
		{
			content: "some\nthing",
			expected: []sh.Obj{
				sh.NewStrObj("some"),
				sh.NewStrObj("thing"),
			},
			sep: sh.NewStrObj("\n"),
		},
	})
}
