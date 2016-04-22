package nash

import "fmt"
import "testing"

func testTable(name, content string, expected []item, t *testing.T) {
	l := lex(name, content)

	if l == nil {
		t.Errorf("Failed to initialize lexer")
		return
	}

	if l.items == nil {
		t.Errorf("Failed to initialize lexer")
		return
	}

	result := make([]item, 0, 1024)

	for i := range l.items {
		result = append(result, i)
	}

	if len(result) != len(expected) {
		t.Errorf("Failed to parse commands, length differs %d != %d",
			len(result), len(expected))

		fmt.Printf("Parsing content: %s\n", content)

		for _, res := range result {
			fmt.Printf("parsed: %+v\n", res)
		}

		return
	}

	for i := 0; i < len(expected); i++ {
		if expected[i].typ != result[i].typ {
			t.Errorf("'%v' != '%v'", expected[i].typ, result[i].typ)
			fmt.Printf("Type: %d - %s\n", result[i].typ, result[i])
			return
		}

		if expected[i].val != result[i].val {
			t.Errorf("'%v' != '%v'", expected[i].val, result[i].val)
			return
		}
	}
}

func TestItemToString(t *testing.T) {
	it := item{
		typ: itemEOF,
	}

	if it.String() != "EOF" {
		t.Errorf("Wrong eof string: %s", it.String())
	}

	it = item{
		typ: itemError,
		val: "some error",
	}

	if it.String() != "Error: some error" {
		t.Errorf("wrong error string: %s", it.String())
	}

	it = item{
		typ: itemCommand,
		val: "echo",
	}

	if it.String() != "(itemCommand) - pos: 0, val: \"echo\"" {
		t.Errorf("wrong command name: %s", it.String())
	}

	it = item{
		typ: itemCommand,
		val: "echoooooooooooooooooooooooo",
	}

	// test if long names are truncated
	if it.String() != "(itemCommand) - pos: 0, val: \"echooooooo\"..." {
		t.Errorf("wrong command name: %s", it.String())
	}
}

func TestShebangOnly(t *testing.T) {
	expected := []item{
		item{
			typ: itemComment,
			val: "#!/bin/nash",
		},
		item{
			typ: itemEOF,
		},
	}

	testTable("testShebangonly", "#!/bin/nash\n", expected, t)
}

func TestSimpleAssignment(t *testing.T) {
	expected := []item{
		item{
			typ: itemVarName,
			val: "test",
		},
		item{
			typ: itemVarValue,
			val: "value",
		},
		item{
			typ: itemEOF,
		},
	}

	testTable("testAssignment", "test=value", expected, t)

	expected = []item{
		item{
			typ: itemVarName,
			val: "test",
		},
		item{
			typ: itemVarValue,
			val: "value",
		},
		item{
			typ: itemVarName,
			val: "other",
		},
		item{
			typ: itemVarValue,
			val: "other",
		},
		item{
			typ: itemEOF,
		},
	}

	testTable("test multiple Assignments", `
        test=value
        other=other`, expected, t)

	expected = []item{
		item{
			typ: itemVarName,
			val: "test",
		},
		item{
			typ: itemVarValue,
			val: "value",
		},
		item{
			typ: itemVarName,
			val: "other",
		},
		item{
			typ: itemVarValue,
			val: "$test",
		},
		item{
			typ: itemCommand,
			val: "echo",
		},
		item{
			typ: itemArg,
			val: "$other",
		},
		item{
			typ: itemEOF,
		},
	}

	testTable("test multiple Assignments", `
        test=value
        other=$test
        echo $other`, expected, t)
}

func TestListAssignment(t *testing.T) {
	expected := []item{
		item{
			typ: itemVarName,
			val: "test",
		},
		item{
			typ: itemListOpen,
			val: "(",
		},
		item{
			typ: itemListElem,
			val: "plan9",
		},
		item{
			typ: itemListElem,
			val: "from",
		},
		item{
			typ: itemListElem,
			val: "bell",
		},
		item{
			typ: itemListElem,
			val: "labs",
		},
		item{
			typ: itemListClose,
			val: ")",
		},
		item{
			typ: itemEOF,
		},
	}

	testTable("testListAssignment", "test=( plan9 from bell labs )", expected, t)
	testTable("testListAssignment no space", "test=(plan9 from bell labs)", expected, t)
}

func TestSimpleCommand(t *testing.T) {
	expected := []item{
		item{
			typ: itemCommand,
			val: "echo",
		},
		item{
			typ: itemString,
			val: "hello world",
		},
		item{
			typ: itemEOF,
		},
	}

	testTable("testSimpleCommand", `echo "hello world"`, expected, t)
}

func TestPathCommand(t *testing.T) {
	expected := []item{
		item{
			typ: itemCommand,
			val: "/bin/echo",
		},
		item{
			typ: itemString,
			val: "hello world",
		},
		item{
			typ: itemEOF,
		},
	}

	testTable("testPathCommand", `/bin/echo "hello world"`, expected, t)
}

func TestInvalidBlock(t *testing.T) {
	expected := []item{
		item{
			typ: itemError,
			val: "Unexpected open block \"U+007B '{'\"",
		},
		item{
			typ: itemEOF,
		},
	}

	testTable("testInvalidBlock", "{", expected, t)
}

func TestQuotedStringNotFinished(t *testing.T) {
	expected := []item{
		item{
			typ: itemCommand,
			val: "echo",
		},
		item{
			typ: itemError,
			val: "Quoted string not finished: hello world",
		},
		item{
			typ: itemEOF,
			val: "hello world",
		},
	}

	testTable("testQuotedstringnotfinished", "echo \"hello world", expected, t)
}

func TestVariousCommands(t *testing.T) {
	content := `
            echo "hello world"
            mount -t proc proc /proc
        `

	expected := []item{
		item{
			typ: itemCommand,
			val: "echo",
		},
		item{
			typ: itemString,
			val: "hello world",
		},
		item{
			typ: itemCommand,
			val: "mount",
		},
		item{
			typ: itemArg,
			val: "-t",
		},
		item{
			typ: itemArg,
			val: "proc",
		},
		item{
			typ: itemArg,
			val: "proc",
		},
		item{
			typ: itemArg,
			val: "/proc",
		},
		item{
			typ: itemEOF,
		},
	}

	testTable("testVariouscommands", content, expected, t)
}

func TestRfork(t *testing.T) {
	expected := []item{
		item{
			typ: itemRfork,
			val: "rfork",
		},
		item{
			typ: itemRforkFlags,
			val: "u",
		},
		item{
			typ: itemEOF,
		},
	}

	testTable("testRfork", "rfork u\n", expected, t)

	expected = []item{
		item{
			typ: itemRfork,
			val: "rfork",
		},
		item{
			typ: itemRforkFlags,
			val: "usnm",
		},
		item{
			typ: itemLeftBlock,
			val: "{",
		},
		item{
			typ: itemCommand,
			val: "echo",
		},
		item{
			typ: itemString,
			val: "inside namespace :)",
		},
		item{
			typ: itemRightBlock,
			val: "}",
		},
		item{
			typ: itemEOF,
		},
	}

	testTable("testRforkWithBlock", `
            rfork usnm {
                echo "inside namespace :)"
            }
        `, expected, t)
}

func TestRforkInvalidArguments(t *testing.T) {
	expected := []item{
		item{
			typ: itemRfork,
			val: "rfork",
		},
		item{
			typ: itemError,
			val: "invalid rfork argument: x",
		},
		item{
			typ: itemEOF,
		},
	}

	testTable("testRfork", "rfork x\n", expected, t)
}

func TestSomethingIdontcareanymore(t *testing.T) {
	expected := []item{
		item{
			typ: itemRfork,
			val: "rfork",
		},
		item{
			typ: itemRforkFlags,
			val: "u",
		},
		item{
			typ: itemLeftBlock,
			val: "{",
		},
		item{
			typ: itemCommand,
			val: "ls",
		},
		item{
			typ: itemRightBlock,
			val: "}",
		},
		item{
			typ: itemEOF,
		},
	}

	testTable("test whatever", "rfork u { ls }", expected, t)
}

func TestBuiltinCd(t *testing.T) {
	expected := []item{
		item{
			typ: itemCd,
			val: "cd",
		},
		item{
			typ: itemString,
			val: "some place",
		},
		item{
			typ: itemEOF,
		},
	}

	testTable("testBuiltinCd", `cd "some place"`, expected, t)

	expected = []item{
		item{
			typ: itemCd,
			val: "cd",
		},
		item{
			typ: itemArg,
			val: "/proc",
		},
		item{
			typ: itemEOF,
		},
	}

	testTable("testBuiltinCdNoQuote", `cd /proc`, expected, t)
}

func TestMinusAlone(t *testing.T) {
	expected := []item{
		item{
			typ: itemError,
			val: "- requires a command",
		},
		item{
			typ: itemEOF,
		},
	}

	testTable("test minus", "-", expected, t)
}

func TestRedirectSimple(t *testing.T) {
	expected := []item{
		item{
			typ: itemCommand,
			val: "cmd",
		},
		item{
			typ: itemRedirRight,
			val: ">",
		},
		item{
			typ: itemRedirFile,
			val: "file.out",
		},
		item{
			typ: itemEOF,
		},
	}

	testTable("test simple redirect", "cmd > file.out", expected, t)

	expected = []item{
		item{
			typ: itemCommand,
			val: "cmd",
		},
		item{
			typ: itemRedirRight,
			val: ">",
		},
		item{
			typ: itemRedirNetAddr,
			val: "tcp://localhost:8888",
		},
		item{
			typ: itemEOF,
		},
	}

	testTable("test simple redirect", "cmd > tcp://localhost:8888", expected, t)

	expected = []item{
		item{
			typ: itemCommand,
			val: "cmd",
		},
		item{
			typ: itemRedirRight,
			val: ">",
		},
		item{
			typ: itemRedirNetAddr,
			val: "udp://localhost:8888",
		},
		item{
			typ: itemEOF,
		},
	}

	testTable("test simple redirect", "cmd > udp://localhost:8888", expected, t)

	expected = []item{
		item{
			typ: itemCommand,
			val: "cmd",
		},
		item{
			typ: itemRedirRight,
			val: ">",
		},
		item{
			typ: itemRedirNetAddr,
			val: "unix:///tmp/sock.txt",
		},
		item{
			typ: itemEOF,
		},
	}

	testTable("test simple redirect", "cmd > unix:///tmp/sock.txt", expected, t)

}

func TestRedirectMap(t *testing.T) {

	// Suppress stderr output
	expected := []item{
		item{
			typ: itemCommand,
			val: "cmd",
		},
		item{
			typ: itemRedirRight,
			val: ">",
		},
		item{
			typ: itemRedirLBracket,
			val: "[",
		},
		item{
			typ: itemRedirMapLSide,
			val: "2",
		},
		item{
			typ: itemRedirMapEqual,
			val: "=",
		},
		item{
			typ: itemRedirRBracket,
			val: "]",
		},
		item{
			typ: itemEOF,
		},
	}

	testTable("test suppress stderr", "cmd >[2=]", expected, t)

	// points stderr to stdout
	expected = []item{
		item{
			typ: itemCommand,
			val: "cmd",
		},
		item{
			typ: itemRedirRight,
			val: ">",
		},
		item{
			typ: itemRedirLBracket,
			val: "[",
		},
		item{
			typ: itemRedirMapLSide,
			val: "2",
		},
		item{
			typ: itemRedirMapEqual,
			val: "=",
		},
		item{
			typ: itemRedirMapRSide,
			val: "1",
		},
		item{
			typ: itemRedirRBracket,
			val: "]",
		},
		item{
			typ: itemEOF,
		},
	}

	testTable("test stderr=stdout", "cmd >[2=1]", expected, t)
}

func TestRedirectMapToLocation(t *testing.T) {
	// Suppress stderr output
	expected := []item{
		item{
			typ: itemCommand,
			val: "cmd",
		},
		item{
			typ: itemRedirRight,
			val: ">",
		},
		item{
			typ: itemRedirLBracket,
			val: "[",
		},
		item{
			typ: itemRedirMapLSide,
			val: "2",
		},
		item{
			typ: itemRedirMapEqual,
			val: "=",
		},
		item{
			typ: itemRedirRBracket,
			val: "]",
		},
		item{
			typ: itemRedirFile,
			val: "file.out",
		},
		item{
			typ: itemEOF,
		},
	}

	testTable("test suppress stderr", "cmd >[2=] file.out", expected, t)

	// points stderr to stdout
	expected = []item{
		item{
			typ: itemCommand,
			val: "cmd",
		},
		item{
			typ: itemRedirRight,
			val: ">",
		},
		item{
			typ: itemRedirLBracket,
			val: "[",
		},
		item{
			typ: itemRedirMapLSide,
			val: "2",
		},
		item{
			typ: itemRedirMapEqual,
			val: "=",
		},
		item{
			typ: itemRedirMapRSide,
			val: "1",
		},
		item{
			typ: itemRedirRBracket,
			val: "]",
		},
		item{
			typ: itemRedirFile,
			val: "file.out",
		},
		item{
			typ: itemEOF,
		},
	}

	testTable("test stderr=stdout", "cmd >[2=1] file.out", expected, t)
}
