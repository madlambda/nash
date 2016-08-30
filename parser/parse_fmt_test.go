package parser

import "testing"

type fmtTestTable struct {
	input, expected string
}

func testFmt(input string, expected string, t *testing.T) {
	p := NewParser("fmt test", input)

	tree, err := p.Parse()

	if err != nil {
		t.Error(err)
		return
	}

	fmtval := tree.String()

	if fmtval != expected {
		t.Errorf("Fmt differ: '%s' != '%s'", fmtval, expected)
		return
	}
}

func testFmtTable(testTable []fmtTestTable, t *testing.T) {
	for _, test := range testTable {
		testFmt(test.input, test.expected, t)
	}
}

func TestFmtVariables(t *testing.T) {
	testTable := []fmtTestTable{

		// correct adjust of spaces
		{`test = "a"`, `test = "a"`},
		{`test="a"`, `test = "a"`},
		{`test= "a"`, `test = "a"`},
		{`test  ="a"`, `test = "a"`},
		{`test =    "a"`, `test = "a"`},
		{`test	="a"`, `test = "a"`},
		{`test		="a"`, `test = "a"`},
		{`test =	"a"`, `test = "a"`},
		{`test =		"a"`, `test = "a"`},
		{`test = ()`, `test = ()`},
		{`test=()`, `test = ()`},
		{`test =()`, `test = ()`},
		{`test	=()`, `test = ()`},
		{`test=	()`, `test = ()`},
		{`test = (plan9)`, `test = (plan9)`},
		{`test=(plan9)`, `test = (plan9)`},
		{`test      = (plan9)`, `test = (plan9)`},
		{`test	= (plan9)`, `test = (plan9)`},
		{`test	=	(plan9)`, `test = (plan9)`},
		{`test = (	plan9)`, `test = (plan9)`},
		{`test = (     plan9)`, `test = (plan9)`},
		{`test = (plan9     )`, `test = (plan9)`},
		{`test = (plan9 from bell labs)`, `test = (plan9 from bell labs)`},
		{`test = (plan9         from bell labs)`, `test = (plan9 from bell labs)`},
		{`test = (plan9         from         bell         labs)`, `test = (plan9 from bell labs)`},
		{`test = (plan9	from	bell	labs)`, `test = (plan9 from bell labs)`},
		{`test = (
	plan9
	from
	bell
	labs
)`, `test = (plan9 from bell labs)`},
		{`test = (plan9 from bell labs windows linux freebsd netbsd openbsd)`, `test = (
	plan9
	from
	bell
	labs
	windows
	linux
	freebsd
	netbsd
	openbsd
)`},

		{`IFS = ("\n")`, `IFS = ("\n")`},

		// multiple variables
		{`test = "a"
testb = "b"`, `test = "a"
testb = "b"`},
	}

	testFmtTable(testTable, t)
}

func TestFmtGroupVariables(t *testing.T) {
	testTable := []fmtTestTable{
		{
			`test = "a"

test2 = "b"

fn cd() { echo "hello" }`,
			`test = "a"
test2 = "b"

fn cd() {
	echo "hello"
}
`,
		},
		{
			`#!/usr/bin/env nash
echo "hello"`,
			`#!/usr/bin/env nash

echo "hello"`,
		},
	}

	testFmtTable(testTable, t)
}

func TestFmtFn(t *testing.T) {
	testTable := []fmtTestTable{
		{
			`fn lala() { echo hello }
fn lele() { echo lele }`,
			`fn lala() {
	echo hello
}

fn lele() {
	echo lele
}
`,
		},
		{
			`vv = ""
fn t() {
	echo t
}`,
			`vv = ""

fn t() {
	echo t
}
`,
		},
	}

	testFmtTable(testTable, t)
}
