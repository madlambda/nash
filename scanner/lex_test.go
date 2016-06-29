package nash

import (
	"fmt"
	"strconv"
)
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
			t.Errorf("Parsing '%s':\n\terror: '%v' != '%v'", content, expected[i].val, result[i].val)
			return
		}
	}
}

func TestLexerItemToString(t *testing.T) {
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

func TestLexerShebangOnly(t *testing.T) {
	expected := []item{
		{
			typ: itemComment,
			val: "#!/bin/nash",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("testShebangonly", "#!/bin/nash\n", expected, t)
}

func TestLexerShowEnv(t *testing.T) {
	expected := []item{
		{
			typ: itemShowEnv,
			val: "showenv",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("testShowEnv", `showenv`, expected, t)

	expected = []item{
		{
			typ: itemShowEnv,
			val: "showenv",
		},
		{
			typ: itemError,
			val: "Unexpected character 'a' at pos 9. Showenv doesn't have arguments.",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("testShowEnv", `showenv a`, expected, t)
}

func TestLexerSimpleSetAssignment(t *testing.T) {
	expected := []item{
		{
			typ: itemSetEnv,
			val: "setenv",
		},
		{
			typ: itemIdentifier,
			val: "name",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("testSet", `setenv name`, expected, t)
}

func TestLexerSimpleAssignment(t *testing.T) {
	expected := []item{
		{
			typ: itemIdentifier,
			val: "test",
		},
		{
			typ: itemAssign,
			val: "=",
		},
		{
			typ: itemString,
			val: "value",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("testAssignment", `test="value"`, expected, t)
	testTable("testAssignment spacy", `test = "value"`, expected, t)
	testTable("testAssignment spacy", `test          ="value"`, expected, t)
	testTable("testAssignment spacy", `test=           "value"`, expected, t)
	testTable("testAssignment spacy", `test	="value"`, expected, t)
	testTable("testAssignment spacy", `test		="value"`, expected, t)
	testTable("testAssignment spacy", `test =	"value"`, expected, t)

	expected = []item{
		{
			typ: itemIdentifier,
			val: "test",
		},
		{
			typ: itemAssign,
			val: "=",
		},
		{
			typ: itemString,
			val: "value",
		},
		{
			typ: itemIdentifier,
			val: "other",
		},
		{
			typ: itemAssign,
			val: "=",
		},
		{
			typ: itemString,
			val: "other",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("test multiple Assignments", `
        test="value"
        other="other"`, expected, t)

	expected = []item{
		{
			typ: itemIdentifier,
			val: "test",
		},
		{
			typ: itemAssign,
			val: "=",
		},
		{
			typ: itemString,
			val: "value",
		},
		{
			typ: itemIdentifier,
			val: "other",
		},
		{
			typ: itemAssign,
			val: "=",
		},
		{
			typ: itemVariable,
			val: "$test",
		},
		{
			typ: itemCommand,
			val: "echo",
		},
		{
			typ: itemVariable,
			val: "$other",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("test multiple Assignments", `
        test="value"
        other=$test
        echo $other`, expected, t)

	expected = []item{
		{
			typ: itemIdentifier,
			val: "STALI_SRC",
		},
		{
			typ: itemAssign,
			val: "=",
		},
		{
			typ: itemVariable,
			val: "$PWD",
		},
		{
			typ: itemConcat,
			val: "+",
		},
		{
			typ: itemString,
			val: "/src",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("test underscore", `STALI_SRC=$PWD + "/src"`, expected, t)

	expected = []item{
		{
			typ: itemIdentifier,
			val: "PROMPT",
		},
		{
			typ: itemAssign,
			val: "=",
		},
		{
			typ: itemString,
			val: "(",
		},
		{
			typ: itemConcat,
			val: "+",
		},
		{
			typ: itemVariable,
			val: "$path",
		},
		{
			typ: itemConcat,
			val: "+",
		},
		{
			typ: itemString,
			val: ")",
		},
		{
			typ: itemConcat,
			val: "+",
		},
		{
			typ: itemVariable,
			val: "$PROMPT",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("test concat with parenthesis", `PROMPT="("+$path+")"+$PROMPT`, expected, t)

}

func TestLexerListAssignment(t *testing.T) {
	expected := []item{
		{
			typ: itemIdentifier,
			val: "test",
		},
		{
			typ: itemAssign,
			val: "=",
		},
		{
			typ: itemListOpen,
			val: "(",
		},
		{
			typ: itemArg,
			val: "plan9",
		},
		{
			typ: itemArg,
			val: "from",
		},
		{
			typ: itemArg,
			val: "bell",
		},
		{
			typ: itemArg,
			val: "labs",
		},
		{
			typ: itemListClose,
			val: ")",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("testListAssignment", "test=( plan9 from bell labs )", expected, t)
	testTable("testListAssignment no space", "test=(plan9 from bell labs)", expected, t)
	testTable("testListAssignment multiline", `test = (
	plan9
	from
	bell
	labs
)`, expected, t)

	expected = []item{
		{
			typ: itemIdentifier,
			val: "test",
		},
		{
			typ: itemAssign,
			val: "=",
		},
		{
			typ: itemListOpen,
			val: "(",
		},
		{
			typ: itemString,
			val: "plan9",
		},
		{
			typ: itemArg,
			val: "from",
		},
		{
			typ: itemString,
			val: "bell",
		},
		{
			typ: itemArg,
			val: "labs",
		},
		{
			typ: itemListClose,
			val: ")",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("testListAssignment mixed args", `test=( "plan9" from "bell" labs )`, expected, t)
	testTable("testListAssignment mixed args", `test=("plan9" from "bell" labs)`, expected, t)
	testTable("testListAssignment mixed args", `test = (
        "plan9"
        from
        "bell"
        labs
)`, expected, t)

	expected = []item{
		{
			typ: itemIdentifier,
			val: "test",
		},
		{
			typ: itemAssign,
			val: "=",
		},
		{
			typ: itemListOpen,
			val: "(",
		},
		{
			typ: itemVariable,
			val: "$plan9",
		},
		{
			typ: itemArg,
			val: "from",
		},
		{
			typ: itemVariable,
			val: "$bell",
		},
		{
			typ: itemArg,
			val: "labs",
		},
		{
			typ: itemListClose,
			val: ")",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("testListAssignment mixed args", `test=( $plan9 from $bell labs )`, expected, t)
	testTable("testListAssignment mixed args", `test=($plan9 from $bell labs)`, expected, t)
	testTable("testListAssignment mixed args", `test = (
        $plan9
        from
        $bell
        labs
)`, expected, t)

}

func TestLexerInvalidAssignments(t *testing.T) {
	expected := []item{
		{
			typ: itemIdentifier,
			val: "test",
		},
		{
			typ: itemAssign,
			val: "=",
		},
		{
			typ: itemString,
			val: "value",
		},
		{
			typ: itemError,
			val: "Invalid assignment. Expected '+' or EOL, but found 'o' at pos '13'",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("testInvalidAssignments", `test="value" other`, expected, t)
}

func TestLexerSimpleCommand(t *testing.T) {
	expected := []item{
		{
			typ: itemCommand,
			val: "echo",
		},
		{
			typ: itemString,
			val: "hello world",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("testSimpleCommand", `echo "hello world"`, expected, t)

	expected = []item{
		{
			typ: itemCommand,
			val: "echo",
		},
		{
			typ: itemArg,
			val: "rootfs-x86_64",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("testSimpleCommand", `echo rootfs-x86_64`, expected, t)

	expected = []item{
		{
			typ: itemCommand,
			val: "git",
		},
		{
			typ: itemArg,
			val: "clone",
		},
		{
			typ: itemArg,
			val: "--depth=1",
		},
		{
			typ: itemArg,
			val: "http://git.sta.li/toolchain",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("testSimpleCommand", `git clone --depth=1 http://git.sta.li/toolchain`, expected, t)

	expected = []item{
		{
			typ: itemCommand,
			val: "ls",
		},
		{
			typ: itemVariable,
			val: "$GOPATH",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("testSimpleCommand", `ls $GOPATH`, expected, t)

	expected = []item{
		{
			typ: itemCommand,
			val: "ls",
		},
		{
			typ: itemVariable,
			val: "$GOPATH",
		},
		{
			typ: itemConcat,
			val: "+",
		},
		{
			typ: itemString,
			val: "/src/github.com",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("testSimpleCommand", `ls $GOPATH+"/src/github.com"`, expected, t)

	expected = []item{
		{
			typ: itemCommand,
			val: "ls",
		},
		{
			typ: itemString,
			val: "/src/github.com",
		},
		{
			typ: itemConcat,
			val: "+",
		},
		{
			typ: itemVariable,
			val: "$GOPATH",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("testSimpleCommand", `ls "/src/github.com"+$GOPATH`, expected, t)

	expected = []item{
		{
			typ: itemCommand,
			val: "ls",
		},
		{
			typ: itemString,
			val: "/home/user",
		},
		{
			typ: itemConcat,
			val: "+",
		},
		{
			typ: itemString,
			val: "/.gvm/pkgsets/global/src",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("testSimpleCommand", `ls "/home/user" + "/.gvm/pkgsets/global/src"`, expected, t)

}

func TestLexerPipe(t *testing.T) {
	expected := []item{
		{
			typ: itemCommand,
			val: "ls",
		},
		{
			typ: itemPipe,
			val: "|",
		},
		{
			typ: itemCommand,
			val: "wc",
		},
		{
			typ: itemArg,
			val: "-l",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("testPipe", `ls | wc -l`, expected, t)

	expected = []item{
		{
			typ: itemCommand,
			val: "ls",
		},
		{
			typ: itemArg,
			val: "-l",
		},
		{
			typ: itemPipe,
			val: "|",
		},
		{
			typ: itemCommand,
			val: "wc",
		},
		{
			typ: itemPipe,
			val: "|",
		},
		{
			typ: itemCommand,
			val: "awk",
		},
		{
			typ: itemString,
			val: "{print $1}",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("testPipe", `ls -l | wc | awk "{print $1}"`, expected, t)

	expected = []item{
		{
			typ: itemCommand,
			val: "go",
		},
		{
			typ: itemArg,
			val: "tool",
		},
		{
			typ: itemArg,
			val: "vet",
		},
		{
			typ: itemArg,
			val: "-h",
		},
		{
			typ: itemRedirRight,
			val: ">",
		},
		{
			typ: itemRedirLBracket,
			val: "[",
		},
		{
			typ: itemRedirMapLSide,
			val: "2",
		},
		{
			typ: itemRedirMapEqual,
			val: "=",
		},
		{
			typ: itemRedirMapRSide,
			val: "1",
		},
		{
			typ: itemRedirRBracket,
			val: "]",
		},
		{
			typ: itemPipe,
			val: "|",
		},
		{
			typ: itemCommand,
			val: "grep",
		},
		{
			typ: itemArg,
			val: "log",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("testPipe with redirection", `go tool vet -h >[2=1] | grep log`, expected, t)

	expected = []item{
		{
			typ: itemCommand,
			val: "go",
		},
		{
			typ: itemArg,
			val: "tool",
		},
		{
			typ: itemArg,
			val: "vet",
		},
		{
			typ: itemArg,
			val: "-h",
		},
		{
			typ: itemRedirRight,
			val: ">",
		},
		{
			typ: itemArg,
			val: "out.log",
		},
		{
			typ: itemPipe,
			val: "|",
		},
		{
			typ: itemCommand,
			val: "grep",
		},
		{
			typ: itemArg,
			val: "log",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("testPipe with redirection", `go tool vet -h > out.log | grep log`, expected, t)
}

func TestLexerUnquoteArg(t *testing.T) {
	expected := []item{
		{
			typ: itemCommand,
			val: "echo",
		},
		{
			typ: itemArg,
			val: "hello",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("testSimpleCommand", `echo hello`, expected, t)

	expected = []item{
		{
			typ: itemCommand,
			val: "echo",
		},
		{
			typ: itemArg,
			val: "hello-world",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("testSimpleCommand", `echo hello-world`, expected, t)

	expected = []item{
		{
			typ: itemCommand,
			val: "echo",
		},
		{
			typ: itemComment,
			val: "#hello-world",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("testSimpleCommand", `echo #hello-world`, expected, t)
}

func TestLexerPathCommand(t *testing.T) {
	expected := []item{
		{
			typ: itemCommand,
			val: "/bin/echo",
		},
		{
			typ: itemString,
			val: "hello world",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("testPathCommand", `/bin/echo "hello world"`, expected, t)
}

func TestLexerInvalidBlock(t *testing.T) {
	expected := []item{
		{
			typ: itemError,
			val: "Unexpected open block \"U+007B '{'\"",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("testInvalidBlock", "{", expected, t)
}

func TestLexerQuotedStringNotFinished(t *testing.T) {
	expected := []item{
		{
			typ: itemCommand,
			val: "echo",
		},
		{
			typ: itemError,
			val: "Quoted string not finished: hello world",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("testQuotedstringnotfinished", "echo \"hello world", expected, t)
}

func TestLexerVariousCommands(t *testing.T) {
	content := `
            echo "hello world"
            mount -t proc proc /proc
        `

	expected := []item{
		{
			typ: itemCommand,
			val: "echo",
		},
		{
			typ: itemString,
			val: "hello world",
		},
		{
			typ: itemCommand,
			val: "mount",
		},
		{
			typ: itemArg,
			val: "-t",
		},
		{
			typ: itemArg,
			val: "proc",
		},
		{
			typ: itemArg,
			val: "proc",
		},
		{
			typ: itemArg,
			val: "/proc",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("testVariouscommands", content, expected, t)
}

func TestLexerRfork(t *testing.T) {
	expected := []item{
		{
			typ: itemRfork,
			val: "rfork",
		},
		{
			typ: itemRforkFlags,
			val: "u",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("testRfork", "rfork u\n", expected, t)

	expected = []item{
		{
			typ: itemRfork,
			val: "rfork",
		},
		{
			typ: itemRforkFlags,
			val: "usnm",
		},
		{
			typ: itemBracesOpen,
			val: "{",
		},
		{
			typ: itemCommand,
			val: "echo",
		},
		{
			typ: itemString,
			val: "inside namespace :)",
		},
		{
			typ: itemBracesClose,
			val: "}",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("testRforkWithBlock", `
            rfork usnm {
                echo "inside namespace :)"
            }
        `, expected, t)
}

func TestLexerRforkInvalidArguments(t *testing.T) {
	expected := []item{
		{
			typ: itemRfork,
			val: "rfork",
		},
		{
			typ: itemError,
			val: "invalid rfork argument: x",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("testRfork", "rfork x\n", expected, t)
}

func TestLexerSomethingIdontcareanymore(t *testing.T) {
	// maybe oneliner rfork isnt a good idea
	expected := []item{
		{
			typ: itemRfork,
			val: "rfork",
		},
		{
			typ: itemRforkFlags,
			val: "u",
		},
		{
			typ: itemBracesOpen,
			val: "{",
		},
		{
			typ: itemCommand,
			val: "ls",
		},
		{
			typ: itemBracesClose,
			val: "}",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("test whatever", "rfork u { ls }", expected, t)
}

func TestLexerBuiltinCd(t *testing.T) {
	expected := []item{
		{
			typ: itemCd,
			val: "cd",
		},
		{
			typ: itemString,
			val: "some place",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("testBuiltinCd", `cd "some place"`, expected, t)

	expected = []item{
		{
			typ: itemCd,
			val: "cd",
		},
		{
			typ: itemArg,
			val: "/proc",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("testBuiltinCdNoQuote", `cd /proc`, expected, t)

	expected = []item{
		{
			typ: itemCd,
			val: "cd",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("testBuiltincd home", `cd`, expected, t)

	expected = []item{
		{
			typ: itemIdentifier,
			val: "HOME",
		},
		{
			typ: itemAssign,
			val: "=",
		},
		{
			typ: itemString,
			val: "/",
		},
		{
			typ: itemSetEnv,
			val: "setenv",
		},
		{
			typ: itemIdentifier,
			val: "HOME",
		},
		{
			typ: itemCd,
			val: "cd",
		},
		{
			typ: itemCommand,
			val: "pwd",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("testBuiltin cd bug", `
	               HOME="/"
                       setenv HOME
	               cd
	               pwd
	           `, expected, t)

	expected = []item{
		{
			typ: itemCd,
			val: "cd",
		},
		{
			typ: itemVariable,
			val: "$GOPATH",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("test builtin cd into variable", `cd $GOPATH`, expected, t)

	expected = []item{
		{
			typ: itemCd,
			val: "cd",
		},
		{
			typ: itemVariable,
			val: "$GOPATH",
		},
		{
			typ: itemConcat,
			val: "+",
		},
		{
			typ: itemString,
			val: "/src/github.com",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("test cd with concat", `cd $GOPATH+"/src/github.com"`, expected, t)
}

func TestLexerMinusAlone(t *testing.T) {
	expected := []item{
		{
			typ: itemError,
			val: "- requires a command",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("test minus", "-", expected, t)
}

func TestLexerRedirectSimple(t *testing.T) {
	expected := []item{
		{
			typ: itemCommand,
			val: "cmd",
		},
		{
			typ: itemRedirRight,
			val: ">",
		},
		{
			typ: itemArg,
			val: "file.out",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("test simple redirect", "cmd > file.out", expected, t)

	expected = []item{
		{
			typ: itemCommand,
			val: "cmd",
		},
		{
			typ: itemRedirRight,
			val: ">",
		},
		{
			typ: itemString,
			val: "tcp://localhost:8888",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("test simple redirect", `cmd > "tcp://localhost:8888"`, expected, t)

	expected = []item{
		{
			typ: itemCommand,
			val: "cmd",
		},
		{
			typ: itemRedirRight,
			val: ">",
		},
		{
			typ: itemString,
			val: "udp://localhost:8888",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("test simple redirect", `cmd > "udp://localhost:8888"`, expected, t)

	expected = []item{
		{
			typ: itemCommand,
			val: "cmd",
		},
		{
			typ: itemRedirRight,
			val: ">",
		},
		{
			typ: itemString,
			val: "unix:///tmp/sock.txt",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("test simple redirect", `cmd > "unix:///tmp/sock.txt"`, expected, t)

}

func TestLexerRedirectMap(t *testing.T) {

	// Suppress stderr output
	expected := []item{
		{
			typ: itemCommand,
			val: "cmd",
		},
		{
			typ: itemRedirRight,
			val: ">",
		},
		{
			typ: itemRedirLBracket,
			val: "[",
		},
		{
			typ: itemRedirMapLSide,
			val: "2",
		},
		{
			typ: itemRedirMapEqual,
			val: "=",
		},
		{
			typ: itemRedirRBracket,
			val: "]",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("test suppress stderr", "cmd >[2=]", expected, t)

	// points stderr to stdout
	expected = []item{
		{
			typ: itemCommand,
			val: "cmd",
		},
		{
			typ: itemRedirRight,
			val: ">",
		},
		{
			typ: itemRedirLBracket,
			val: "[",
		},
		{
			typ: itemRedirMapLSide,
			val: "2",
		},
		{
			typ: itemRedirMapEqual,
			val: "=",
		},
		{
			typ: itemRedirMapRSide,
			val: "1",
		},
		{
			typ: itemRedirRBracket,
			val: "]",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("test stderr=stdout", "cmd >[2=1]", expected, t)
}

func TestLexerRedirectMapToLocation(t *testing.T) {
	// Suppress stderr output
	expected := []item{
		{
			typ: itemCommand,
			val: "cmd",
		},
		{
			typ: itemRedirRight,
			val: ">",
		},
		{
			typ: itemRedirLBracket,
			val: "[",
		},
		{
			typ: itemRedirMapLSide,
			val: "2",
		},
		{
			typ: itemRedirMapEqual,
			val: "=",
		},
		{
			typ: itemRedirRBracket,
			val: "]",
		},
		{
			typ: itemArg,
			val: "file.out",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("test suppress stderr", "cmd >[2=] file.out", expected, t)

	// points stderr to stdout
	expected = []item{
		{
			typ: itemCommand,
			val: "cmd",
		},
		{
			typ: itemRedirRight,
			val: ">",
		},
		{
			typ: itemRedirLBracket,
			val: "[",
		},
		{
			typ: itemRedirMapLSide,
			val: "2",
		},
		{
			typ: itemRedirMapEqual,
			val: "=",
		},
		{
			typ: itemRedirMapRSide,
			val: "1",
		},
		{
			typ: itemRedirRBracket,
			val: "]",
		},
		{
			typ: itemArg,
			val: "file.out",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("test stderr=stdout", "cmd >[2=1] file.out", expected, t)

	expected = []item{
		{
			typ: itemCommand,
			val: "cmd",
		},
		{
			typ: itemRedirRight,
			val: ">",
		},
		{
			typ: itemRedirLBracket,
			val: "[",
		},
		{
			typ: itemRedirMapLSide,
			val: "2",
		},
		{
			typ: itemRedirRBracket,
			val: "]",
		},
		{
			typ: itemArg,
			val: "/var/log/service.log",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("test suppress stderr", `cmd >[2] /var/log/service.log`, expected, t)
}

func TestLexerRedirectMultipleMaps(t *testing.T) {
	expected := []item{
		{
			typ: itemCommand,
			val: "cmd",
		},
		{
			typ: itemRedirRight,
			val: ">",
		},
		{
			typ: itemRedirLBracket,
			val: "[",
		},
		{
			typ: itemRedirMapLSide,
			val: "1",
		},
		{
			typ: itemRedirRBracket,
			val: "]",
		},
		{
			typ: itemArg,
			val: "file.out",
		},
		{
			typ: itemRedirRight,
			val: ">",
		},
		{
			typ: itemRedirLBracket,
			val: "[",
		},
		{
			typ: itemRedirMapLSide,
			val: "2",
		},
		{
			typ: itemRedirRBracket,
			val: "]",
		},
		{
			typ: itemArg,
			val: "file.err",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("test suppress stderr", `cmd >[1] file.out >[2] file.err`, expected, t)

	expected = []item{
		{
			typ: itemCommand,
			val: "cmd",
		},
		{
			typ: itemRedirRight,
			val: ">",
		},
		{
			typ: itemRedirLBracket,
			val: "[",
		},
		{
			typ: itemRedirMapLSide,
			val: "2",
		},
		{
			typ: itemRedirMapEqual,
			val: "=",
		},
		{
			typ: itemRedirMapRSide,
			val: "1",
		},
		{
			typ: itemRedirRBracket,
			val: "]",
		},
		{
			typ: itemRedirRight,
			val: ">",
		},
		{
			typ: itemRedirLBracket,
			val: "[",
		},
		{
			typ: itemRedirMapLSide,
			val: "1",
		},
		{
			typ: itemRedirRBracket,
			val: "]",
		},
		{
			typ: itemArg,
			val: "/var/log/service.log",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("test suppress stderr", `cmd >[2=1] >[1] /var/log/service.log`, expected, t)

	expected = []item{
		{
			typ: itemCommand,
			val: "cmd",
		},
		{
			typ: itemRedirRight,
			val: ">",
		},
		{
			typ: itemRedirLBracket,
			val: "[",
		},
		{
			typ: itemRedirMapLSide,
			val: "1",
		},
		{
			typ: itemRedirMapEqual,
			val: "=",
		},
		{
			typ: itemRedirMapRSide,
			val: "2",
		},
		{
			typ: itemRedirRBracket,
			val: "]",
		},
		{
			typ: itemRedirRight,
			val: ">",
		},
		{
			typ: itemRedirLBracket,
			val: "[",
		},
		{
			typ: itemRedirMapLSide,
			val: "2",
		},
		{
			typ: itemRedirMapEqual,
			val: "=",
		},
		{
			typ: itemRedirRBracket,
			val: "]",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("test suppress stderr", `cmd >[1=2] >[2=]`, expected, t)
}

func TestLexerImport(t *testing.T) {
	expected := []item{
		{
			typ: itemImport,
			val: "import",
		},
		{
			typ: itemArg,
			val: "env.sh",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("test import", `import env.sh`, expected, t)
}

func TestLexerSimpleIf(t *testing.T) {
	expected := []item{
		{
			typ: itemIf,
			val: "if",
		},
		{
			typ: itemString,
			val: "test",
		},
		{
			typ: itemComparison,
			val: "==",
		},
		{
			typ: itemString,
			val: "other",
		},
		{
			typ: itemBracesOpen,
			val: "{",
		},
		{
			typ: itemCommand,
			val: "rm",
		},
		{
			typ: itemArg,
			val: "-rf",
		},
		{
			typ: itemArg,
			val: "/",
		},
		{
			typ: itemBracesClose,
			val: "}",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("test simple if", `if "test" == "other" { rm -rf / }`, expected, t)

	expected = []item{
		{
			typ: itemIf,
			val: "if",
		},
		{
			typ: itemString,
			val: "test",
		},
		{
			typ: itemComparison,
			val: "!=",
		},
		{
			typ: itemVariable,
			val: "$test",
		},
		{
			typ: itemBracesOpen,
			val: "{",
		},
		{
			typ: itemCommand,
			val: "rm",
		},
		{
			typ: itemArg,
			val: "-rf",
		},
		{
			typ: itemArg,
			val: "/",
		},
		{
			typ: itemBracesClose,
			val: "}",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("test simple if", `if "test" != $test {
            rm -rf /
        }`, expected, t)

	testTable("test simple if", `

        if "test" != $test {
            rm -rf /
        }`, expected, t)
}

func TestLexerIfWithConcat(t *testing.T) {
	expected := []item{
		{
			typ: itemIf,
			val: "if",
		},
		{
			typ: itemVariable,
			val: "$test",
		},
		{
			typ: itemConcat,
			val: "+",
		},
		{
			typ: itemString,
			val: "001",
		},
		{
			typ: itemComparison,
			val: "!=",
		},
		{
			typ: itemString,
			val: "value001",
		},
		{
			typ: itemBracesOpen,
			val: "{",
		},
		{
			typ: itemCommand,
			val: "rm",
		},
		{
			typ: itemArg,
			val: "-rf",
		},
		{
			typ: itemArg,
			val: "/",
		},
		{
			typ: itemBracesClose,
			val: "}",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("test if concat", `if $test + "001" != "value001" {
        rm -rf /
}`, expected, t)
}

func TestLexerIfElse(t *testing.T) {
	expected := []item{
		{
			typ: itemIf,
			val: "if",
		},
		{
			typ: itemString,
			val: "test",
		},
		{
			typ: itemComparison,
			val: "==",
		},
		{
			typ: itemString,
			val: "other",
		},
		{
			typ: itemBracesOpen,
			val: "{",
		},
		{
			typ: itemCommand,
			val: "rm",
		},
		{
			typ: itemArg,
			val: "-rf",
		},
		{
			typ: itemArg,
			val: "/",
		},
		{
			typ: itemBracesClose,
			val: "}",
		},
		{
			typ: itemElse,
			val: "else",
		},
		{
			typ: itemBracesOpen,
			val: "{",
		},
		{
			typ: itemCommand,
			val: "pwd",
		},
		{
			typ: itemBracesClose,
			val: "}",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("test simple if", `if "test" == "other" { rm -rf / } else { pwd }`, expected, t)
}

func TestLexerIfElseIf(t *testing.T) {
	expected := []item{
		{
			typ: itemIf,
			val: "if",
		},
		{
			typ: itemString,
			val: "test",
		},
		{
			typ: itemComparison,
			val: "==",
		},
		{
			typ: itemString,
			val: "other",
		},
		{
			typ: itemBracesOpen,
			val: "{",
		},
		{
			typ: itemCommand,
			val: "rm",
		},
		{
			typ: itemArg,
			val: "-rf",
		},
		{
			typ: itemArg,
			val: "/",
		},
		{
			typ: itemBracesClose,
			val: "}",
		},
		{
			typ: itemElse,
			val: "else",
		},
		{
			typ: itemIf,
			val: "if",
		},
		{
			typ: itemString,
			val: "test",
		},
		{
			typ: itemComparison,
			val: "==",
		},
		{
			typ: itemVariable,
			val: "$var",
		},
		{
			typ: itemBracesOpen,
			val: "{",
		},
		{
			typ: itemCommand,
			val: "pwd",
		},
		{
			typ: itemBracesClose,
			val: "}",
		},
		{
			typ: itemElse,
			val: "else",
		},
		{
			typ: itemBracesOpen,
			val: "{",
		},
		{
			typ: itemCommand,
			val: "exit",
		},
		{
			typ: itemArg,
			val: "1",
		},
		{
			typ: itemBracesClose,
			val: "}",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("test simple if", `
        if "test" == "other" {
                rm -rf /
        } else if "test" == $var {
                pwd
        } else {
                exit 1
        }`, expected, t)
}

func TestLexerFnBasic(t *testing.T) {
	expected := []item{
		{
			typ: itemFnDecl,
			val: "fn",
		},
		{
			typ: itemIdentifier,
			val: "build",
		},
		{
			typ: itemParenOpen,
			val: "(",
		},
		{
			typ: itemParenClose,
			val: ")",
		},
		{
			typ: itemBracesOpen,
			val: "{",
		},
		{
			typ: itemBracesClose,
			val: "}",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("test empty fn", `fn build() {}`, expected, t)

	// lambda
	expected = []item{
		{
			typ: itemFnDecl,
			val: "fn",
		},
		{
			typ: itemIdentifier,
			val: "",
		},
		{
			typ: itemParenOpen,
			val: "(",
		},
		{
			typ: itemParenClose,
			val: ")",
		},
		{
			typ: itemBracesOpen,
			val: "{",
		},
		{
			typ: itemBracesClose,
			val: "}",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("test empty fn", `fn () {}`, expected, t)

	// IIFE
	expected = []item{
		{
			typ: itemFnDecl,
			val: "fn",
		},
		{
			typ: itemIdentifier,
			val: "",
		},
		{
			typ: itemParenOpen,
			val: "(",
		},
		{
			typ: itemParenClose,
			val: ")",
		},
		{
			typ: itemBracesOpen,
			val: "{",
		},
		{
			typ: itemBracesClose,
			val: "}",
		},
		{
			typ: itemParenOpen,
			val: "(",
		},
		{
			typ: itemParenClose,
			val: ")",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("test empty fn", `fn () {}()`, expected, t)

	expected = []item{
		{
			typ: itemFnDecl,
			val: "fn",
		},
		{
			typ: itemIdentifier,
			val: "build",
		},
		{
			typ: itemParenOpen,
			val: "(",
		},
		{
			typ: itemIdentifier,
			val: "image",
		},
		{
			typ: itemIdentifier,
			val: "debug",
		},
		{
			typ: itemParenClose,
			val: ")",
		},
		{
			typ: itemBracesOpen,
			val: "{",
		},
		{
			typ: itemBracesClose,
			val: "}",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("test empty fn with args", `fn build(image, debug) {}`, expected, t)

	expected = []item{
		{
			typ: itemFnDecl,
			val: "fn",
		},
		{
			typ: itemIdentifier,
			val: "build",
		},
		{
			typ: itemParenOpen,
			val: "(",
		},
		{
			typ: itemIdentifier,
			val: "image",
		},
		{
			typ: itemIdentifier,
			val: "debug",
		},
		{
			typ: itemParenClose,
			val: ")",
		},
		{
			typ: itemBracesOpen,
			val: "{",
		},
		{
			typ: itemCommand,
			val: "ls",
		},
		{
			typ: itemCommand,
			val: "tar",
		},
		{
			typ: itemBracesClose,
			val: "}",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("test empty fn with args and body", `fn build(image, debug) {
            ls
            tar
        }`, expected, t)

	expected = []item{
		{
			typ: itemFnDecl,
			val: "fn",
		},
		{
			typ: itemIdentifier,
			val: "cd",
		},
		{
			typ: itemParenOpen,
			val: "(",
		},
		{
			typ: itemIdentifier,
			val: "path",
		},
		{
			typ: itemParenClose,
			val: ")",
		},
		{
			typ: itemBracesOpen,
			val: "{",
		},
		{
			typ: itemCd,
			val: "cd",
		},
		{
			typ: itemVariable,
			val: "$path",
		},
		{
			typ: itemIdentifier,
			val: "PROMPT",
		},
		{
			typ: itemAssign,
			val: "=",
		},
		{
			typ: itemString,
			val: "(",
		},
		{
			typ: itemConcat,
			val: "+",
		},
		{
			typ: itemVariable,
			val: "$path",
		},
		{
			typ: itemConcat,
			val: "+",
		},
		{
			typ: itemString,
			val: ")",
		},
		{
			typ: itemConcat,
			val: "+",
		},
		{
			typ: itemVariable,
			val: "$PROMPT",
		},
		{
			typ: itemSetEnv,
			val: "setenv",
		},
		{
			typ: itemIdentifier,
			val: "PROMPT",
		},
		{
			typ: itemBracesClose,
			val: "}",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("test cd fn with PROMPT update", `fn cd(path) {
    cd $path
    PROMPT="(" + $path + ")"+$PROMPT
    setenv PROMPT
}`, expected, t)
}

func TestLexerFnInvocation(t *testing.T) {
	expected := []item{
		{
			typ: itemFnInv,
			val: "build",
		},
		{
			typ: itemParenOpen,
			val: "(",
		},
		{
			typ: itemParenClose,
			val: ")",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("test fn invocation", `build()`, expected, t)

	expected = []item{
		{
			typ: itemFnInv,
			val: "build",
		},
		{
			typ: itemParenOpen,
			val: "(",
		},
		{
			typ: itemString,
			val: "ubuntu",
		},

		{
			typ: itemParenClose,
			val: ")",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("test fn invocation", `build("ubuntu")`, expected, t)

	expected = []item{
		{
			typ: itemFnInv,
			val: "build",
		},
		{
			typ: itemParenOpen,
			val: "(",
		},
		{
			typ: itemString,
			val: "ubuntu",
		},
		{
			typ: itemVariable,
			val: "$debug",
		},

		{
			typ: itemParenClose,
			val: ")",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("test fn invocation", `build("ubuntu", $debug)`, expected, t)

	expected = []item{
		{
			typ: itemFnInv,
			val: "build",
		},
		{
			typ: itemParenOpen,
			val: "(",
		},
		{
			typ: itemVariable,
			val: "$debug",
		},

		{
			typ: itemParenClose,
			val: ")",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("test fn invocation", `build($debug)`, expected, t)
}

func TestLexerAssignCmdOut(t *testing.T) {
	expected := []item{
		{
			typ: itemIdentifier,
			val: "ipaddr",
		},
		{
			typ: itemAssignCmd,
			val: "<=",
		},
		{
			typ: itemCommand,
			val: "someprogram",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("test assignCmdOut", `ipaddr <= someprogram`, expected, t)
}

func TestLexerBindFn(t *testing.T) {
	expected := []item{
		{
			typ: itemBindFn,
			val: "bindfn",
		},
		{
			typ: itemIdentifier,
			val: "cd",
		},
		{
			typ: itemIdentifier,
			val: "cd2",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("test bindfn", `bindfn cd cd2`, expected, t)

}

func TestLexerRedirectionNetwork(t *testing.T) {
	expected := []item{
		{
			typ: itemCommand,
			val: "echo",
		},
		{
			typ: itemString,
			val: "hello world",
		},
		{
			typ: itemRedirRight,
			val: ">",
		},
		{
			typ: itemRedirLBracket,
			val: "[",
		},
		{
			typ: itemRedirMapLSide,
			val: "1",
		},
		{
			typ: itemRedirRBracket,
			val: "]",
		},
		{
			typ: itemString,
			val: "tcp://localhost:6667",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("test redirection network", `echo "hello world" >[1] "tcp://localhost:6667"`, expected, t)
}

func TestLexerDump(t *testing.T) {
	expected := []item{
		{
			typ: itemDump,
			val: "dump",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("test dump", `dump`, expected, t)

	expected = []item{
		{
			typ: itemDump,
			val: "dump",
		},
		{
			typ: itemArg,
			val: "out",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("test dump", `dump out`, expected, t)

	expected = []item{
		{
			typ: itemDump,
			val: "dump",
		},
		{
			typ: itemVariable,
			val: "$out",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("test dump", `dump $out`, expected, t)
}

func TestLexerReturn(t *testing.T) {
	expected := []item{
		{
			typ: itemReturn,
			val: "return",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("test return", "return", expected, t)

	expected = []item{
		{
			typ: itemFnDecl,
			val: "fn",
		},
		{
			typ: itemIdentifier,
			val: "test",
		},
		{
			typ: itemParenOpen,
			val: "(",
		},
		{
			typ: itemParenClose,
			val: ")",
		},
		{
			typ: itemBracesOpen,
			val: "{",
		},
		{
			typ: itemReturn,
			val: "return",
		},
		{
			typ: itemBracesClose,
			val: "}",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("test return", "fn test() { return }", expected, t)
	testTable("test return", `fn test() {
 return
}`, expected, t)
	testTable("test return", `fn test() {
	return
}`, expected, t)

	expected = []item{
		{
			typ: itemFnDecl,
			val: "fn",
		},
		{
			typ: itemIdentifier,
			val: "test",
		},
		{
			typ: itemParenOpen,
			val: "(",
		},
		{
			typ: itemParenClose,
			val: ")",
		},
		{
			typ: itemBracesOpen,
			val: "{",
		},
		{
			typ: itemReturn,
			val: "return",
		},
		{
			typ: itemString,
			val: "some value",
		},
		{
			typ: itemBracesClose,
			val: "}",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("test return", `fn test() { return "some value"}`, expected, t)
	testTable("test return", `fn test() {
	return "some value"
}`, expected, t)

	expected = []item{
		{
			typ: itemFnDecl,
			val: "fn",
		},
		{
			typ: itemIdentifier,
			val: "test",
		},
		{
			typ: itemParenOpen,
			val: "(",
		},
		{
			typ: itemParenClose,
			val: ")",
		},
		{
			typ: itemBracesOpen,
			val: "{",
		},
		{
			typ: itemIdentifier,
			val: "value",
		},
		{
			typ: itemAssign,
			val: "=",
		},
		{
			typ: itemString,
			val: "some value",
		},
		{
			typ: itemReturn,
			val: "return",
		},
		{
			typ: itemVariable,
			val: "$value",
		},
		{
			typ: itemBracesClose,
			val: "}",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("test return", `fn test() {
	value = "some value"
	return $value
}`, expected, t)

	expected = []item{
		{
			typ: itemFnDecl,
			val: "fn",
		},
		{
			typ: itemIdentifier,
			val: "test",
		},
		{
			typ: itemParenOpen,
			val: "(",
		},
		{
			typ: itemParenClose,
			val: ")",
		},
		{
			typ: itemBracesOpen,
			val: "{",
		},
		{
			typ: itemReturn,
			val: "return",
		},
		{
			typ: itemListOpen,
			val: "(",
		},
		{
			typ: itemString,
			val: "test",
		},
		{
			typ: itemString,
			val: "test2",
		},
		{
			typ: itemListClose,
			val: ")",
		},
		{
			typ: itemBracesClose,
			val: "}",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("test return", `fn test() {
	return ("test" "test2")
}`, expected, t)

	expected = []item{
		{
			typ: itemFnDecl,
			val: "fn",
		},
		{
			typ: itemIdentifier,
			val: "test",
		},
		{
			typ: itemParenOpen,
			val: "(",
		},
		{
			typ: itemParenClose,
			val: ")",
		},
		{
			typ: itemBracesOpen,
			val: "{",
		},
		{
			typ: itemReturn,
			val: "return",
		},
		{
			typ: itemVariable,
			val: "$PWD",
		},
		{
			typ: itemBracesClose,
			val: "}",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("test return", `fn test() {
	return $PWD
}`, expected, t)
}

func TestLexerFor(t *testing.T) {
	expected := []item{
		{
			typ: itemFor,
			val: "for",
		},
		{
			typ: itemBracesOpen,
			val: "{",
		},
		{
			typ: itemBracesClose,
			val: "}",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("test inf loop", `for {}`, expected, t)

	expected = []item{
		{
			typ: itemFor,
			val: "for",
		},
		{
			typ: itemIdentifier,
			val: "f",
		},
		{
			typ: itemForIn,
			val: "in",
		},
		{
			typ: itemVariable,
			val: "$files",
		},
		{
			typ: itemBracesOpen,
			val: "{",
		},
		{
			typ: itemBracesClose,
			val: "}",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("test inf loop", `for f in $files {}`, expected, t)

}

func TestLexerFnAsFirstClass(t *testing.T) {
	expected := []item{
		{
			typ: itemFnDecl,
			val: "fn",
		},
		{
			typ: itemIdentifier,
			val: "printer",
		},
		{
			typ: itemParenOpen,
			val: "(",
		},
		{
			typ: itemIdentifier,
			val: "val",
		},
		{
			typ: itemParenClose,
			val: ")",
		},
		{
			typ: itemBracesOpen,
			val: "{",
		},
		{
			typ: itemCommand,
			val: "echo",
		},
		{
			typ: itemArg,
			val: "-n",
		},
		{
			typ: itemVariable,
			val: "$val",
		},
		{
			typ: itemBracesClose,
			val: "}",
		},
		{
			typ: itemFnDecl,
			val: "fn",
		},
		{
			typ: itemIdentifier,
			val: "success",
		},
		{
			typ: itemParenOpen,
			val: "(",
		},
		{
			typ: itemIdentifier,
			val: "print",
		},
		{
			typ: itemIdentifier,
			val: "val",
		},
		{
			typ: itemParenClose,
			val: ")",
		},
		{
			typ: itemBracesOpen,
			val: "{",
		},
		{
			typ: itemFnInv,
			val: "$print",
		},
		{
			typ: itemParenOpen,
			val: "(",
		},
		{
			typ: itemString,
			val: "[SUCCESS] ",
		},
		{
			typ: itemConcat,
			val: "+",
		},
		{
			typ: itemVariable,
			val: "$val",
		},
		{
			typ: itemParenClose,
			val: ")",
		},
		{
			typ: itemBracesClose,
			val: "}",
		},
		{
			typ: itemFnInv,
			val: "success",
		},
		{
			typ: itemParenOpen,
			val: "(",
		},
		{
			typ: itemVariable,
			val: "$printer",
		},
		{
			typ: itemString,
			val: "Command executed!",
		},
		{
			typ: itemParenClose,
			val: ")",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("test fn as first class", `
        fn printer(val) {
                echo -n $val
        }

        fn success(print, val) {
                $print("[SUCCESS] " + $val)
        }

        success($printer, "Command executed!")
        `, expected, t)
}

func TestLexerListIndexing(t *testing.T) {
	expected := []item{
		{
			typ: itemIdentifier,
			val: "cmd",
		},
		{
			typ: itemAssign,
			val: "=",
		},
		{
			typ: itemVariable,
			val: "$commands",
		},
		{
			typ: itemBracketOpen,
			val: "[",
		},
		{
			typ: itemNumber,
			val: "0",
		},
		{
			typ: itemBracketClose,
			val: "]",
		},
		{
			typ: itemEOF,
		},
	}

	for i := 0; i < 1000; i++ {
		expected[4] = item{
			typ: itemNumber,
			val: strconv.Itoa(i),
		}

		testTable("test variable indexing", `cmd = $commands[`+strconv.Itoa(i)+`]`, expected, t)
	}

	expected = []item{
		{
			typ: itemIdentifier,
			val: "cmd",
		},
		{
			typ: itemAssign,
			val: "=",
		},
		{
			typ: itemVariable,
			val: "$commands",
		},
		{
			typ: itemBracketOpen,
			val: "[",
		},
		{
			typ: itemError,
			val: "Expected number or variable on variable indexing. Found 'a'",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("test invalid number", `cmd = $commands[a]`, expected, t)

	expected = []item{
		{
			typ: itemIdentifier,
			val: "cmd",
		},
		{
			typ: itemAssign,
			val: "=",
		},
		{
			typ: itemVariable,
			val: "$commands",
		},
		{
			typ: itemBracketOpen,
			val: "[",
		},
		{
			typ: itemError,
			val: "Expected number or variable on variable indexing. Found ']'",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("test invalid number", `cmd = $commands[]`, expected, t)

	expected = []item{
		{
			typ: itemCommand,
			val: "echo",
		},
		{
			typ: itemArg,
			val: "test",
		},
		{
			typ: itemVariable,
			val: "$names",
		},
		{
			typ: itemBracketOpen,
			val: "[",
		},
		{
			typ: itemNumber,
			val: "666",
		},
		{
			typ: itemBracketClose,
			val: "]",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("test variable index on commands", `echo test $names[666]`, expected, t)

	expected = []item{
		{
			typ: itemIf,
			val: "if",
		},
		{
			typ: itemVariable,
			val: "$crazies",
		},
		{
			typ: itemBracketOpen,
			val: "[",
		},
		{
			typ: itemNumber,
			val: "0",
		},
		{
			typ: itemBracketClose,
			val: "]",
		},
		{
			typ: itemComparison,
			val: "==",
		},
		{
			typ: itemString,
			val: "patito",
		},
		{
			typ: itemBracesOpen,
			val: "{",
		},
		{
			typ: itemCommand,
			val: "echo",
		},
		{
			typ: itemString,
			val: ":D",
		},
		{
			typ: itemBracesClose,
			val: "}",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("test if with indexing", `if $crazies[0] == "patito" { echo ":D" }`, expected, t)
}
