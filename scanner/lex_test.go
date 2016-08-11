package scanner

import (
	"fmt"
	"strconv"

	"github.com/NeowayLabs/nash/token"
)
import "testing"

func testTable(name, content string, expected []Token, t *testing.T) {
	l := Lex(name, content)

	if l == nil {
		t.Errorf("Failed to initialize lexer")
		return
	}

	if l.Tokens == nil {
		t.Errorf("Failed to initialize lexer")
		return
	}

	result := make([]Token, 0, 1024)

	for i := range l.Tokens {
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

func TestLexerTokenToString(t *testing.T) {
	it := Token{
		typ: token.EOF,
	}

	if it.String() != "EOF" {
		t.Errorf("Wrong eof string: %s", it.String())
	}

	it = Token{
		typ: token.Illegal,
		val: "some error",
	}

	if it.String() != "ERROR: some error" {
		t.Errorf("wrong error string: %s", it.String())
	}

	it = Token{
		typ: token.Command,
		val: "echo",
	}

	if it.String() != "(COMMAND) - pos: 0, val: \"echo\"" {
		t.Errorf("wrong command name: %s", it.String())
	}

	it = Token{
		typ: token.Command,
		val: "echoooooooooooooooooooooooo",
	}

	// test if long names are truncated
	if it.String() != "(COMMAND) - pos: 0, val: \"echooooooo\"..." {
		t.Errorf("wrong command name: %s", it.String())
	}
}

func TestLexerShebangOnly(t *testing.T) {
	expected := []Token{
		{
			typ: token.Comment,
			val: "#!/bin/nash",
		},
		{
			typ: token.EOF,
		},
	}

	testTable("testShebangonly", "#!/bin/nash\n", expected, t)
}

func TestLexerSimpleSetAssignment(t *testing.T) {
	expected := []Token{
		{
			typ: token.SetEnv,
			val: "setenv",
		},
		{
			typ: token.Ident,
			val: "name",
		},
		{
			typ: token.EOF,
		},
	}

	testTable("testSet", `setenv name`, expected, t)
}

func TestLexerSimpleAssignment(t *testing.T) {
	expected := []Token{
		{
			typ: token.Ident,
			val: "test",
		},
		{
			typ: token.Assign,
			val: "=",
		},
		{
			typ: token.String,
			val: "value",
		},
		{
			typ: token.EOF,
		},
	}

	testTable("testAssignment", `test="value"`, expected, t)
	testTable("testAssignment spacy", `test = "value"`, expected, t)
	testTable("testAssignment spacy", `test          ="value"`, expected, t)
	testTable("testAssignment spacy", `test=           "value"`, expected, t)
	testTable("testAssignment spacy", `test	="value"`, expected, t)
	testTable("testAssignment spacy", `test		="value"`, expected, t)
	testTable("testAssignment spacy", `test =	"value"`, expected, t)

	expected = []Token{
		{
			typ: token.Ident,
			val: "test",
		},
		{
			typ: token.Assign,
			val: "=",
		},
		{
			typ: token.String,
			val: "value",
		},
		{
			typ: token.Ident,
			val: "other",
		},
		{
			typ: token.Assign,
			val: "=",
		},
		{
			typ: token.String,
			val: "other",
		},
		{
			typ: token.EOF,
		},
	}

	testTable("test multiple Assignments", `
        test="value"
        other="other"`, expected, t)

	expected = []Token{
		{
			typ: token.Ident,
			val: "test",
		},
		{
			typ: token.Assign,
			val: "=",
		},
		{
			typ: token.String,
			val: "value",
		},
		{
			typ: token.Ident,
			val: "other",
		},
		{
			typ: token.Assign,
			val: "=",
		},
		{
			typ: token.Variable,
			val: "$test",
		},
		{
			typ: token.Command,
			val: "echo",
		},
		{
			typ: token.Variable,
			val: "$other",
		},
		{
			typ: token.EOF,
		},
	}

	testTable("test multiple Assignments", `
        test="value"
        other=$test
        echo $other`, expected, t)

	expected = []Token{
		{
			typ: token.Ident,
			val: "STALI_SRC",
		},
		{
			typ: token.Assign,
			val: "=",
		},
		{
			typ: token.Variable,
			val: "$PWD",
		},
		{
			typ: token.Concat,
			val: "+",
		},
		{
			typ: token.String,
			val: "/src",
		},
		{
			typ: token.EOF,
		},
	}

	testTable("test underscore", `STALI_SRC=$PWD + "/src"`, expected, t)

	expected = []Token{
		{
			typ: token.Ident,
			val: "PROMPT",
		},
		{
			typ: token.Assign,
			val: "=",
		},
		{
			typ: token.String,
			val: "(",
		},
		{
			typ: token.Concat,
			val: "+",
		},
		{
			typ: token.Variable,
			val: "$path",
		},
		{
			typ: token.Concat,
			val: "+",
		},
		{
			typ: token.String,
			val: ")",
		},
		{
			typ: token.Concat,
			val: "+",
		},
		{
			typ: token.Variable,
			val: "$PROMPT",
		},
		{
			typ: token.EOF,
		},
	}

	testTable("test concat with parenthesis", `PROMPT="("+$path+")"+$PROMPT`, expected, t)

}

func TestLexerListAssignment(t *testing.T) {
	expected := []Token{
		{
			typ: token.Ident,
			val: "test",
		},
		{
			typ: token.Assign,
			val: "=",
		},
		{
			typ: token.LParen,
			val: "(",
		},
		{
			typ: token.Arg,
			val: "plan9",
		},
		{
			typ: token.Arg,
			val: "from",
		},
		{
			typ: token.Arg,
			val: "bell",
		},
		{
			typ: token.Arg,
			val: "labs",
		},
		{
			typ: token.RParen,
			val: ")",
		},
		{
			typ: token.EOF,
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

	expected = []Token{
		{
			typ: token.Ident,
			val: "test",
		},
		{
			typ: token.Assign,
			val: "=",
		},
		{
			typ: token.LParen,
			val: "(",
		},
		{
			typ: token.String,
			val: "plan9",
		},
		{
			typ: token.Arg,
			val: "from",
		},
		{
			typ: token.String,
			val: "bell",
		},
		{
			typ: token.Arg,
			val: "labs",
		},
		{
			typ: token.RParen,
			val: ")",
		},
		{
			typ: token.EOF,
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

	expected = []Token{
		{
			typ: token.Ident,
			val: "test",
		},
		{
			typ: token.Assign,
			val: "=",
		},
		{
			typ: token.LParen,
			val: "(",
		},
		{
			typ: token.Variable,
			val: "$plan9",
		},
		{
			typ: token.Arg,
			val: "from",
		},
		{
			typ: token.Variable,
			val: "$bell",
		},
		{
			typ: token.Arg,
			val: "labs",
		},
		{
			typ: token.RParen,
			val: ")",
		},
		{
			typ: token.EOF,
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
	expected := []Token{
		{
			typ: token.Ident,
			val: "test",
		},
		{
			typ: token.Assign,
			val: "=",
		},
		{
			typ: token.String,
			val: "value",
		},
		{
			typ: token.Illegal,
			val: "Invalid assignment. Expected '+' or EOL, but found 'o' at pos '13'",
		},
		{
			typ: token.EOF,
		},
	}

	testTable("testInvalidAssignments", `test="value" other`, expected, t)
}

func TestLexerSimpleCommand(t *testing.T) {
	expected := []Token{
		{
			typ: token.Command,
			val: "echo",
		},
		{
			typ: token.String,
			val: "hello world",
		},
		{
			typ: token.EOF,
		},
	}

	testTable("testSimpleCommand", `echo "hello world"`, expected, t)

	expected = []Token{
		{
			typ: token.Command,
			val: "echo",
		},
		{
			typ: token.Arg,
			val: "rootfs-x86_64",
		},
		{
			typ: token.EOF,
		},
	}

	testTable("testSimpleCommand", `echo rootfs-x86_64`, expected, t)

	expected = []Token{
		{
			typ: token.Command,
			val: "git",
		},
		{
			typ: token.Arg,
			val: "clone",
		},
		{
			typ: token.Arg,
			val: "--depth=1",
		},
		{
			typ: token.Arg,
			val: "http://git.sta.li/toolchain",
		},
		{
			typ: token.EOF,
		},
	}

	testTable("testSimpleCommand", `git clone --depth=1 http://git.sta.li/toolchain`, expected, t)

	expected = []Token{
		{
			typ: token.Command,
			val: "ls",
		},
		{
			typ: token.Variable,
			val: "$GOPATH",
		},
		{
			typ: token.EOF,
		},
	}

	testTable("testSimpleCommand", `ls $GOPATH`, expected, t)

	expected = []Token{
		{
			typ: token.Command,
			val: "ls",
		},
		{
			typ: token.Variable,
			val: "$GOPATH",
		},
		{
			typ: token.Concat,
			val: "+",
		},
		{
			typ: token.String,
			val: "/src/github.com",
		},
		{
			typ: token.EOF,
		},
	}

	testTable("testSimpleCommand", `ls $GOPATH+"/src/github.com"`, expected, t)

	expected = []Token{
		{
			typ: token.Command,
			val: "ls",
		},
		{
			typ: token.String,
			val: "/src/github.com",
		},
		{
			typ: token.Concat,
			val: "+",
		},
		{
			typ: token.Variable,
			val: "$GOPATH",
		},
		{
			typ: token.EOF,
		},
	}

	testTable("testSimpleCommand", `ls "/src/github.com"+$GOPATH`, expected, t)

	expected = []Token{
		{
			typ: token.Command,
			val: "ls",
		},
		{
			typ: token.String,
			val: "/home/user",
		},
		{
			typ: token.Concat,
			val: "+",
		},
		{
			typ: token.String,
			val: "/.gvm/pkgsets/global/src",
		},
		{
			typ: token.EOF,
		},
	}

	testTable("testSimpleCommand", `ls "/home/user" + "/.gvm/pkgsets/global/src"`, expected, t)

}

func TestLexerPipe(t *testing.T) {
	expected := []Token{
		{
			typ: token.Command,
			val: "ls",
		},
		{
			typ: token.Pipe,
			val: "|",
		},
		{
			typ: token.Command,
			val: "wc",
		},
		{
			typ: token.Arg,
			val: "-l",
		},
		{
			typ: token.EOF,
		},
	}

	testTable("testPipe", `ls | wc -l`, expected, t)

	expected = []Token{
		{
			typ: token.Command,
			val: "ls",
		},
		{
			typ: token.Arg,
			val: "-l",
		},
		{
			typ: token.Pipe,
			val: "|",
		},
		{
			typ: token.Command,
			val: "wc",
		},
		{
			typ: token.Pipe,
			val: "|",
		},
		{
			typ: token.Command,
			val: "awk",
		},
		{
			typ: token.String,
			val: "{print $1}",
		},
		{
			typ: token.EOF,
		},
	}

	testTable("testPipe", `ls -l | wc | awk "{print $1}"`, expected, t)

	expected = []Token{
		{
			typ: token.Command,
			val: "go",
		},
		{
			typ: token.Arg,
			val: "tool",
		},
		{
			typ: token.Arg,
			val: "vet",
		},
		{
			typ: token.Arg,
			val: "-h",
		},
		{
			typ: token.RedirRight,
			val: ">",
		},
		{
			typ: token.LBrack,
			val: "[",
		},
		{
			typ: token.RedirMapLSide,
			val: "2",
		},
		{
			typ: token.Assign,
			val: "=",
		},
		{
			typ: token.RedirMapRSide,
			val: "1",
		},
		{
			typ: token.RBrack,
			val: "]",
		},
		{
			typ: token.Pipe,
			val: "|",
		},
		{
			typ: token.Command,
			val: "grep",
		},
		{
			typ: token.Arg,
			val: "log",
		},
		{
			typ: token.EOF,
		},
	}

	testTable("testPipe with redirection", `go tool vet -h >[2=1] | grep log`, expected, t)

	expected = []Token{
		{
			typ: token.Command,
			val: "go",
		},
		{
			typ: token.Arg,
			val: "tool",
		},
		{
			typ: token.Arg,
			val: "vet",
		},
		{
			typ: token.Arg,
			val: "-h",
		},
		{
			typ: token.RedirRight,
			val: ">",
		},
		{
			typ: token.Arg,
			val: "out.log",
		},
		{
			typ: token.Pipe,
			val: "|",
		},
		{
			typ: token.Command,
			val: "grep",
		},
		{
			typ: token.Arg,
			val: "log",
		},
		{
			typ: token.EOF,
		},
	}

	testTable("testPipe with redirection", `go tool vet -h > out.log | grep log`, expected, t)
}

func TestLexerUnquoteArg(t *testing.T) {
	expected := []Token{
		{
			typ: token.Command,
			val: "echo",
		},
		{
			typ: token.Arg,
			val: "hello",
		},
		{
			typ: token.EOF,
		},
	}

	testTable("testSimpleCommand", `echo hello`, expected, t)

	expected = []Token{
		{
			typ: token.Command,
			val: "echo",
		},
		{
			typ: token.Arg,
			val: "hello-world",
		},
		{
			typ: token.EOF,
		},
	}

	testTable("testSimpleCommand", `echo hello-world`, expected, t)

	expected = []Token{
		{
			typ: token.Command,
			val: "echo",
		},
		{
			typ: token.Comment,
			val: "#hello-world",
		},
		{
			typ: token.EOF,
		},
	}

	testTable("testSimpleCommand", `echo #hello-world`, expected, t)
}

func TestLexerDashedCommand(t *testing.T) {
	expected := []Token{
		{
			typ: token.Command,
			val: "google-chrome",
		},
		{
			typ: token.EOF,
		},
	}

	testTable("testDashedCommand", `google-chrome`, expected, t)
}

func TestLexerPathCommand(t *testing.T) {
	expected := []Token{
		{
			typ: token.Command,
			val: "/bin/echo",
		},
		{
			typ: token.String,
			val: "hello world",
		},
		{
			typ: token.EOF,
		},
	}

	testTable("testPathCommand", `/bin/echo "hello world"`, expected, t)
}

func TestLexerInvalidBlock(t *testing.T) {
	expected := []Token{
		{
			typ: token.Illegal,
			val: "Unexpected open block \"U+007B '{'\"",
		},
		{
			typ: token.EOF,
		},
	}

	testTable("testInvalidBlock", "{", expected, t)
}

func TestLexerQuotedStringNotFinished(t *testing.T) {
	expected := []Token{
		{
			typ: token.Command,
			val: "echo",
		},
		{
			typ: token.Illegal,
			val: "Quoted string not finished: hello world",
		},
		{
			typ: token.EOF,
		},
	}

	testTable("testQuotedstringnotfinished", "echo \"hello world", expected, t)
}

func TestLexerVariousCommands(t *testing.T) {
	content := `
            echo "hello world"
            mount -t proc proc /proc
        `

	expected := []Token{
		{
			typ: token.Command,
			val: "echo",
		},
		{
			typ: token.String,
			val: "hello world",
		},
		{
			typ: token.Command,
			val: "mount",
		},
		{
			typ: token.Arg,
			val: "-t",
		},
		{
			typ: token.Arg,
			val: "proc",
		},
		{
			typ: token.Arg,
			val: "proc",
		},
		{
			typ: token.Arg,
			val: "/proc",
		},
		{
			typ: token.EOF,
		},
	}

	testTable("testVariouscommands", content, expected, t)
}

func TestLexerRfork(t *testing.T) {
	expected := []Token{
		{
			typ: token.Rfork,
			val: "rfork",
		},
		{
			typ: token.String,
			val: "u",
		},
		{
			typ: token.EOF,
		},
	}

	testTable("testRfork", "rfork u\n", expected, t)

	expected = []Token{
		{
			typ: token.Rfork,
			val: "rfork",
		},
		{
			typ: token.String,
			val: "usnm",
		},
		{
			typ: token.LBrace,
			val: "{",
		},
		{
			typ: token.Command,
			val: "echo",
		},
		{
			typ: token.String,
			val: "inside namespace :)",
		},
		{
			typ: token.RBrace,
			val: "}",
		},
		{
			typ: token.EOF,
		},
	}

	testTable("testRforkWithBlock", `
            rfork usnm {
                echo "inside namespace :)"
            }
        `, expected, t)
}

func TestLexerRforkInvalidArguments(t *testing.T) {
	expected := []Token{
		{
			typ: token.Rfork,
			val: "rfork",
		},
		{
			typ: token.Illegal,
			val: "invalid rfork argument: x",
		},
		{
			typ: token.EOF,
		},
	}

	testTable("testRfork", "rfork x\n", expected, t)
}

func TestLexerSomethingIdontcareanymore(t *testing.T) {
	// maybe oneliner rfork isnt a good idea
	expected := []Token{
		{
			typ: token.Rfork,
			val: "rfork",
		},
		{
			typ: token.String,
			val: "u",
		},
		{
			typ: token.LBrace,
			val: "{",
		},
		{
			typ: token.Command,
			val: "ls",
		},
		{
			typ: token.RBrace,
			val: "}",
		},
		{
			typ: token.EOF,
		},
	}

	testTable("test whatever", "rfork u { ls }", expected, t)
}

func TestLexerBuiltinCd(t *testing.T) {
	expected := []Token{
		{
			typ: token.Cd,
			val: "cd",
		},
		{
			typ: token.String,
			val: "some place",
		},
		{
			typ: token.EOF,
		},
	}

	testTable("testBuiltinCd", `cd "some place"`, expected, t)

	expected = []Token{
		{
			typ: token.Cd,
			val: "cd",
		},
		{
			typ: token.Arg,
			val: "/proc",
		},
		{
			typ: token.EOF,
		},
	}

	testTable("testBuiltinCdNoQuote", `cd /proc`, expected, t)

	expected = []Token{
		{
			typ: token.Cd,
			val: "cd",
		},
		{
			typ: token.EOF,
		},
	}

	testTable("testBuiltincd home", `cd`, expected, t)

	expected = []Token{
		{
			typ: token.Ident,
			val: "HOME",
		},
		{
			typ: token.Assign,
			val: "=",
		},
		{
			typ: token.String,
			val: "/",
		},
		{
			typ: token.SetEnv,
			val: "setenv",
		},
		{
			typ: token.Ident,
			val: "HOME",
		},
		{
			typ: token.Cd,
			val: "cd",
		},
		{
			typ: token.Command,
			val: "pwd",
		},
		{
			typ: token.EOF,
		},
	}

	testTable("testBuiltin cd bug", `
	               HOME="/"
                       setenv HOME
	               cd
	               pwd
	           `, expected, t)

	expected = []Token{
		{
			typ: token.Cd,
			val: "cd",
		},
		{
			typ: token.Variable,
			val: "$GOPATH",
		},
		{
			typ: token.EOF,
		},
	}

	testTable("test builtin cd into variable", `cd $GOPATH`, expected, t)

	expected = []Token{
		{
			typ: token.Cd,
			val: "cd",
		},
		{
			typ: token.Variable,
			val: "$GOPATH",
		},
		{
			typ: token.Concat,
			val: "+",
		},
		{
			typ: token.String,
			val: "/src/github.com",
		},
		{
			typ: token.EOF,
		},
	}

	testTable("test cd with concat", `cd $GOPATH+"/src/github.com"`, expected, t)
}

func TestLexerMinusAlone(t *testing.T) {
	expected := []Token{
		{
			typ: token.Illegal,
			val: "- requires a command",
		},
		{
			typ: token.EOF,
		},
	}

	testTable("test minus", "-", expected, t)
}

func TestLexerRedirectSimple(t *testing.T) {
	expected := []Token{
		{
			typ: token.Command,
			val: "cmd",
		},
		{
			typ: token.RedirRight,
			val: ">",
		},
		{
			typ: token.Arg,
			val: "file.out",
		},
		{
			typ: token.EOF,
		},
	}

	testTable("test simple redirect", "cmd > file.out", expected, t)

	expected = []Token{
		{
			typ: token.Command,
			val: "cmd",
		},
		{
			typ: token.RedirRight,
			val: ">",
		},
		{
			typ: token.String,
			val: "tcp://localhost:8888",
		},
		{
			typ: token.EOF,
		},
	}

	testTable("test simple redirect", `cmd > "tcp://localhost:8888"`, expected, t)

	expected = []Token{
		{
			typ: token.Command,
			val: "cmd",
		},
		{
			typ: token.RedirRight,
			val: ">",
		},
		{
			typ: token.String,
			val: "udp://localhost:8888",
		},
		{
			typ: token.EOF,
		},
	}

	testTable("test simple redirect", `cmd > "udp://localhost:8888"`, expected, t)

	expected = []Token{
		{
			typ: token.Command,
			val: "cmd",
		},
		{
			typ: token.RedirRight,
			val: ">",
		},
		{
			typ: token.String,
			val: "unix:///tmp/sock.txt",
		},
		{
			typ: token.EOF,
		},
	}

	testTable("test simple redirect", `cmd > "unix:///tmp/sock.txt"`, expected, t)

}

func TestLexerRedirectMap(t *testing.T) {

	// Suppress stderr output
	expected := []Token{
		{
			typ: token.Command,
			val: "cmd",
		},
		{
			typ: token.RedirRight,
			val: ">",
		},
		{
			typ: token.LBrack,
			val: "[",
		},
		{
			typ: token.RedirMapLSide,
			val: "2",
		},
		{
			typ: token.Assign,
			val: "=",
		},
		{
			typ: token.RBrack,
			val: "]",
		},
		{
			typ: token.EOF,
		},
	}

	testTable("test suppress stderr", "cmd >[2=]", expected, t)

	// points stderr to stdout
	expected = []Token{
		{
			typ: token.Command,
			val: "cmd",
		},
		{
			typ: token.RedirRight,
			val: ">",
		},
		{
			typ: token.LBrack,
			val: "[",
		},
		{
			typ: token.RedirMapLSide,
			val: "2",
		},
		{
			typ: token.Assign,
			val: "=",
		},
		{
			typ: token.RedirMapRSide,
			val: "1",
		},
		{
			typ: token.RBrack,
			val: "]",
		},
		{
			typ: token.EOF,
		},
	}

	testTable("test stderr=stdout", "cmd >[2=1]", expected, t)
}

func TestLexerRedirectMapToLocation(t *testing.T) {
	// Suppress stderr output
	expected := []Token{
		{
			typ: token.Command,
			val: "cmd",
		},
		{
			typ: token.RedirRight,
			val: ">",
		},
		{
			typ: token.LBrack,
			val: "[",
		},
		{
			typ: token.RedirMapLSide,
			val: "2",
		},
		{
			typ: token.Assign,
			val: "=",
		},
		{
			typ: token.RBrack,
			val: "]",
		},
		{
			typ: token.Arg,
			val: "file.out",
		},
		{
			typ: token.EOF,
		},
	}

	testTable("test suppress stderr", "cmd >[2=] file.out", expected, t)

	// points stderr to stdout
	expected = []Token{
		{
			typ: token.Command,
			val: "cmd",
		},
		{
			typ: token.RedirRight,
			val: ">",
		},
		{
			typ: token.LBrack,
			val: "[",
		},
		{
			typ: token.RedirMapLSide,
			val: "2",
		},
		{
			typ: token.Assign,
			val: "=",
		},
		{
			typ: token.RedirMapRSide,
			val: "1",
		},
		{
			typ: token.RBrack,
			val: "]",
		},
		{
			typ: token.Arg,
			val: "file.out",
		},
		{
			typ: token.EOF,
		},
	}

	testTable("test stderr=stdout", "cmd >[2=1] file.out", expected, t)

	expected = []Token{
		{
			typ: token.Command,
			val: "cmd",
		},
		{
			typ: token.RedirRight,
			val: ">",
		},
		{
			typ: token.LBrack,
			val: "[",
		},
		{
			typ: token.RedirMapLSide,
			val: "2",
		},
		{
			typ: token.RBrack,
			val: "]",
		},
		{
			typ: token.Arg,
			val: "/var/log/service.log",
		},
		{
			typ: token.EOF,
		},
	}

	testTable("test suppress stderr", `cmd >[2] /var/log/service.log`, expected, t)
}

func TestLexerRedirectMultipleMaps(t *testing.T) {
	expected := []Token{
		{
			typ: token.Command,
			val: "cmd",
		},
		{
			typ: token.RedirRight,
			val: ">",
		},
		{
			typ: token.LBrack,
			val: "[",
		},
		{
			typ: token.RedirMapLSide,
			val: "1",
		},
		{
			typ: token.RBrack,
			val: "]",
		},
		{
			typ: token.Arg,
			val: "file.out",
		},
		{
			typ: token.RedirRight,
			val: ">",
		},
		{
			typ: token.LBrack,
			val: "[",
		},
		{
			typ: token.RedirMapLSide,
			val: "2",
		},
		{
			typ: token.RBrack,
			val: "]",
		},
		{
			typ: token.Arg,
			val: "file.err",
		},
		{
			typ: token.EOF,
		},
	}

	testTable("test suppress stderr", `cmd >[1] file.out >[2] file.err`, expected, t)

	expected = []Token{
		{
			typ: token.Command,
			val: "cmd",
		},
		{
			typ: token.RedirRight,
			val: ">",
		},
		{
			typ: token.LBrack,
			val: "[",
		},
		{
			typ: token.RedirMapLSide,
			val: "2",
		},
		{
			typ: token.Assign,
			val: "=",
		},
		{
			typ: token.RedirMapRSide,
			val: "1",
		},
		{
			typ: token.RBrack,
			val: "]",
		},
		{
			typ: token.RedirRight,
			val: ">",
		},
		{
			typ: token.LBrack,
			val: "[",
		},
		{
			typ: token.RedirMapLSide,
			val: "1",
		},
		{
			typ: token.RBrack,
			val: "]",
		},
		{
			typ: token.Arg,
			val: "/var/log/service.log",
		},
		{
			typ: token.EOF,
		},
	}

	testTable("test suppress stderr", `cmd >[2=1] >[1] /var/log/service.log`, expected, t)

	expected = []Token{
		{
			typ: token.Command,
			val: "cmd",
		},
		{
			typ: token.RedirRight,
			val: ">",
		},
		{
			typ: token.LBrack,
			val: "[",
		},
		{
			typ: token.RedirMapLSide,
			val: "1",
		},
		{
			typ: token.Assign,
			val: "=",
		},
		{
			typ: token.RedirMapRSide,
			val: "2",
		},
		{
			typ: token.RBrack,
			val: "]",
		},
		{
			typ: token.RedirRight,
			val: ">",
		},
		{
			typ: token.LBrack,
			val: "[",
		},
		{
			typ: token.RedirMapLSide,
			val: "2",
		},
		{
			typ: token.Assign,
			val: "=",
		},
		{
			typ: token.RBrack,
			val: "]",
		},
		{
			typ: token.EOF,
		},
	}

	testTable("test suppress stderr", `cmd >[1=2] >[2=]`, expected, t)
}

func TestLexerImport(t *testing.T) {
	expected := []Token{
		{
			typ: token.Import,
			val: "import",
		},
		{
			typ: token.Arg,
			val: "env.sh",
		},
		{
			typ: token.EOF,
		},
	}

	testTable("test import", `import env.sh`, expected, t)
}

func TestLexerSimpleIf(t *testing.T) {
	expected := []Token{
		{
			typ: token.If,
			val: "if",
		},
		{
			typ: token.String,
			val: "test",
		},
		{
			typ: token.Equal,
			val: "==",
		},
		{
			typ: token.String,
			val: "other",
		},
		{
			typ: token.LBrace,
			val: "{",
		},
		{
			typ: token.Command,
			val: "rm",
		},
		{
			typ: token.Arg,
			val: "-rf",
		},
		{
			typ: token.Arg,
			val: "/",
		},
		{
			typ: token.RBrace,
			val: "}",
		},
		{
			typ: token.EOF,
		},
	}

	testTable("test simple if", `if "test" == "other" { rm -rf / }`, expected, t)

	expected = []Token{
		{
			typ: token.If,
			val: "if",
		},
		{
			typ: token.String,
			val: "test",
		},
		{
			typ: token.NotEqual,
			val: "!=",
		},
		{
			typ: token.Variable,
			val: "$test",
		},
		{
			typ: token.LBrace,
			val: "{",
		},
		{
			typ: token.Command,
			val: "rm",
		},
		{
			typ: token.Arg,
			val: "-rf",
		},
		{
			typ: token.Arg,
			val: "/",
		},
		{
			typ: token.RBrace,
			val: "}",
		},
		{
			typ: token.EOF,
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
	expected := []Token{
		{
			typ: token.If,
			val: "if",
		},
		{
			typ: token.Variable,
			val: "$test",
		},
		{
			typ: token.Concat,
			val: "+",
		},
		{
			typ: token.String,
			val: "001",
		},
		{
			typ: token.NotEqual,
			val: "!=",
		},
		{
			typ: token.String,
			val: "value001",
		},
		{
			typ: token.LBrace,
			val: "{",
		},
		{
			typ: token.Command,
			val: "rm",
		},
		{
			typ: token.Arg,
			val: "-rf",
		},
		{
			typ: token.Arg,
			val: "/",
		},
		{
			typ: token.RBrace,
			val: "}",
		},
		{
			typ: token.EOF,
		},
	}

	testTable("test if concat", `if $test + "001" != "value001" {
        rm -rf /
}`, expected, t)
}

func TestLexerIfElse(t *testing.T) {
	expected := []Token{
		{
			typ: token.If,
			val: "if",
		},
		{
			typ: token.String,
			val: "test",
		},
		{
			typ: token.Equal,
			val: "==",
		},
		{
			typ: token.String,
			val: "other",
		},
		{
			typ: token.LBrace,
			val: "{",
		},
		{
			typ: token.Command,
			val: "rm",
		},
		{
			typ: token.Arg,
			val: "-rf",
		},
		{
			typ: token.Arg,
			val: "/",
		},
		{
			typ: token.RBrace,
			val: "}",
		},
		{
			typ: token.Else,
			val: "else",
		},
		{
			typ: token.LBrace,
			val: "{",
		},
		{
			typ: token.Command,
			val: "pwd",
		},
		{
			typ: token.RBrace,
			val: "}",
		},
		{
			typ: token.EOF,
		},
	}

	testTable("test simple if", `if "test" == "other" { rm -rf / } else { pwd }`, expected, t)
}

func TestLexerIfElseIf(t *testing.T) {
	expected := []Token{
		{
			typ: token.If,
			val: "if",
		},
		{
			typ: token.String,
			val: "test",
		},
		{
			typ: token.Equal,
			val: "==",
		},
		{
			typ: token.String,
			val: "other",
		},
		{
			typ: token.LBrace,
			val: "{",
		},
		{
			typ: token.Command,
			val: "rm",
		},
		{
			typ: token.Arg,
			val: "-rf",
		},
		{
			typ: token.Arg,
			val: "/",
		},
		{
			typ: token.RBrace,
			val: "}",
		},
		{
			typ: token.Else,
			val: "else",
		},
		{
			typ: token.If,
			val: "if",
		},
		{
			typ: token.String,
			val: "test",
		},
		{
			typ: token.Equal,
			val: "==",
		},
		{
			typ: token.Variable,
			val: "$var",
		},
		{
			typ: token.LBrace,
			val: "{",
		},
		{
			typ: token.Command,
			val: "pwd",
		},
		{
			typ: token.RBrace,
			val: "}",
		},
		{
			typ: token.Else,
			val: "else",
		},
		{
			typ: token.LBrace,
			val: "{",
		},
		{
			typ: token.Command,
			val: "exit",
		},
		{
			typ: token.Arg,
			val: "1",
		},
		{
			typ: token.RBrace,
			val: "}",
		},
		{
			typ: token.EOF,
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
	expected := []Token{
		{
			typ: token.FnDecl,
			val: "fn",
		},
		{
			typ: token.Ident,
			val: "build",
		},
		{
			typ: token.LParen,
			val: "(",
		},
		{
			typ: token.RParen,
			val: ")",
		},
		{
			typ: token.LBrace,
			val: "{",
		},
		{
			typ: token.RBrace,
			val: "}",
		},
		{
			typ: token.EOF,
		},
	}

	testTable("test empty fn", `fn build() {}`, expected, t)

	// lambda
	expected = []Token{
		{
			typ: token.FnDecl,
			val: "fn",
		},
		{
			typ: token.Ident,
			val: "",
		},
		{
			typ: token.LParen,
			val: "(",
		},
		{
			typ: token.RParen,
			val: ")",
		},
		{
			typ: token.LBrace,
			val: "{",
		},
		{
			typ: token.RBrace,
			val: "}",
		},
		{
			typ: token.EOF,
		},
	}

	testTable("test empty fn", `fn () {}`, expected, t)

	// IIFE
	expected = []Token{
		{
			typ: token.FnDecl,
			val: "fn",
		},
		{
			typ: token.Ident,
			val: "",
		},
		{
			typ: token.LParen,
			val: "(",
		},
		{
			typ: token.RParen,
			val: ")",
		},
		{
			typ: token.LBrace,
			val: "{",
		},
		{
			typ: token.RBrace,
			val: "}",
		},
		{
			typ: token.LParen,
			val: "(",
		},
		{
			typ: token.RParen,
			val: ")",
		},
		{
			typ: token.EOF,
		},
	}

	testTable("test empty fn", `fn () {}()`, expected, t)

	expected = []Token{
		{
			typ: token.FnDecl,
			val: "fn",
		},
		{
			typ: token.Ident,
			val: "build",
		},
		{
			typ: token.LParen,
			val: "(",
		},
		{
			typ: token.Ident,
			val: "image",
		},
		{
			typ: token.Ident,
			val: "debug",
		},
		{
			typ: token.RParen,
			val: ")",
		},
		{
			typ: token.LBrace,
			val: "{",
		},
		{
			typ: token.RBrace,
			val: "}",
		},
		{
			typ: token.EOF,
		},
	}

	testTable("test empty fn with args", `fn build(image, debug) {}`, expected, t)

	expected = []Token{
		{
			typ: token.FnDecl,
			val: "fn",
		},
		{
			typ: token.Ident,
			val: "build",
		},
		{
			typ: token.LParen,
			val: "(",
		},
		{
			typ: token.Ident,
			val: "image",
		},
		{
			typ: token.Ident,
			val: "debug",
		},
		{
			typ: token.RParen,
			val: ")",
		},
		{
			typ: token.LBrace,
			val: "{",
		},
		{
			typ: token.Command,
			val: "ls",
		},
		{
			typ: token.Command,
			val: "tar",
		},
		{
			typ: token.RBrace,
			val: "}",
		},
		{
			typ: token.EOF,
		},
	}

	testTable("test empty fn with args and body", `fn build(image, debug) {
            ls
            tar
        }`, expected, t)

	expected = []Token{
		{
			typ: token.FnDecl,
			val: "fn",
		},
		{
			typ: token.Ident,
			val: "cd",
		},
		{
			typ: token.LParen,
			val: "(",
		},
		{
			typ: token.Ident,
			val: "path",
		},
		{
			typ: token.RParen,
			val: ")",
		},
		{
			typ: token.LBrace,
			val: "{",
		},
		{
			typ: token.Cd,
			val: "cd",
		},
		{
			typ: token.Variable,
			val: "$path",
		},
		{
			typ: token.Ident,
			val: "PROMPT",
		},
		{
			typ: token.Assign,
			val: "=",
		},
		{
			typ: token.String,
			val: "(",
		},
		{
			typ: token.Concat,
			val: "+",
		},
		{
			typ: token.Variable,
			val: "$path",
		},
		{
			typ: token.Concat,
			val: "+",
		},
		{
			typ: token.String,
			val: ")",
		},
		{
			typ: token.Concat,
			val: "+",
		},
		{
			typ: token.Variable,
			val: "$PROMPT",
		},
		{
			typ: token.SetEnv,
			val: "setenv",
		},
		{
			typ: token.Ident,
			val: "PROMPT",
		},
		{
			typ: token.RBrace,
			val: "}",
		},
		{
			typ: token.EOF,
		},
	}

	testTable("test cd fn with PROMPT update", `fn cd(path) {
    cd $path
    PROMPT="(" + $path + ")"+$PROMPT
    setenv PROMPT
}`, expected, t)
}

func TestLexerFnInvocation(t *testing.T) {
	expected := []Token{
		{
			typ: token.FnInv,
			val: "build",
		},
		{
			typ: token.LParen,
			val: "(",
		},
		{
			typ: token.RParen,
			val: ")",
		},
		{
			typ: token.EOF,
		},
	}

	testTable("test fn invocation", `build()`, expected, t)

	expected = []Token{
		{
			typ: token.FnInv,
			val: "build",
		},
		{
			typ: token.LParen,
			val: "(",
		},
		{
			typ: token.String,
			val: "ubuntu",
		},

		{
			typ: token.RParen,
			val: ")",
		},
		{
			typ: token.EOF,
		},
	}

	testTable("test fn invocation", `build("ubuntu")`, expected, t)

	expected = []Token{
		{
			typ: token.FnInv,
			val: "build",
		},
		{
			typ: token.LParen,
			val: "(",
		},
		{
			typ: token.String,
			val: "ubuntu",
		},
		{
			typ: token.Variable,
			val: "$debug",
		},

		{
			typ: token.RParen,
			val: ")",
		},
		{
			typ: token.EOF,
		},
	}

	testTable("test fn invocation", `build("ubuntu", $debug)`, expected, t)

	expected = []Token{
		{
			typ: token.FnInv,
			val: "build",
		},
		{
			typ: token.LParen,
			val: "(",
		},
		{
			typ: token.Variable,
			val: "$debug",
		},

		{
			typ: token.RParen,
			val: ")",
		},
		{
			typ: token.EOF,
		},
	}

	testTable("test fn invocation", `build($debug)`, expected, t)
}

func TestLexerAssignCmdOut(t *testing.T) {
	expected := []Token{
		{
			typ: token.Ident,
			val: "ipaddr",
		},
		{
			typ: token.AssignCmd,
			val: "<=",
		},
		{
			typ: token.Command,
			val: "someprogram",
		},
		{
			typ: token.EOF,
		},
	}

	testTable("test assignCmdOut", `ipaddr <= someprogram`, expected, t)
}

func TestLexerBindFn(t *testing.T) {
	expected := []Token{
		{
			typ: token.BindFn,
			val: "bindfn",
		},
		{
			typ: token.Ident,
			val: "cd",
		},
		{
			typ: token.Ident,
			val: "cd2",
		},
		{
			typ: token.EOF,
		},
	}

	testTable("test bindfn", `bindfn cd cd2`, expected, t)

}

func TestLexerRedirectionNetwork(t *testing.T) {
	expected := []Token{
		{
			typ: token.Command,
			val: "echo",
		},
		{
			typ: token.String,
			val: "hello world",
		},
		{
			typ: token.RedirRight,
			val: ">",
		},
		{
			typ: token.LBrack,
			val: "[",
		},
		{
			typ: token.RedirMapLSide,
			val: "1",
		},
		{
			typ: token.RBrack,
			val: "]",
		},
		{
			typ: token.String,
			val: "tcp://localhost:6667",
		},
		{
			typ: token.EOF,
		},
	}

	testTable("test redirection network", `echo "hello world" >[1] "tcp://localhost:6667"`, expected, t)
}

func TestLexerDump(t *testing.T) {
	expected := []Token{
		{
			typ: token.Dump,
			val: "dump",
		},
		{
			typ: token.EOF,
		},
	}

	testTable("test dump", `dump`, expected, t)

	expected = []Token{
		{
			typ: token.Dump,
			val: "dump",
		},
		{
			typ: token.Arg,
			val: "out",
		},
		{
			typ: token.EOF,
		},
	}

	testTable("test dump", `dump out`, expected, t)

	expected = []Token{
		{
			typ: token.Dump,
			val: "dump",
		},
		{
			typ: token.Variable,
			val: "$out",
		},
		{
			typ: token.EOF,
		},
	}

	testTable("test dump", `dump $out`, expected, t)
}

func TestLexerReturn(t *testing.T) {
	expected := []Token{
		{
			typ: token.Return,
			val: "return",
		},
		{
			typ: token.EOF,
		},
	}

	testTable("test return", "return", expected, t)

	expected = []Token{
		{
			typ: token.FnDecl,
			val: "fn",
		},
		{
			typ: token.Ident,
			val: "test",
		},
		{
			typ: token.LParen,
			val: "(",
		},
		{
			typ: token.RParen,
			val: ")",
		},
		{
			typ: token.LBrace,
			val: "{",
		},
		{
			typ: token.Return,
			val: "return",
		},
		{
			typ: token.RBrace,
			val: "}",
		},
		{
			typ: token.EOF,
		},
	}

	testTable("test return", "fn test() { return }", expected, t)
	testTable("test return", `fn test() {
 return
}`, expected, t)
	testTable("test return", `fn test() {
	return
}`, expected, t)

	expected = []Token{
		{
			typ: token.FnDecl,
			val: "fn",
		},
		{
			typ: token.Ident,
			val: "test",
		},
		{
			typ: token.LParen,
			val: "(",
		},
		{
			typ: token.RParen,
			val: ")",
		},
		{
			typ: token.LBrace,
			val: "{",
		},
		{
			typ: token.Return,
			val: "return",
		},
		{
			typ: token.String,
			val: "some value",
		},
		{
			typ: token.RBrace,
			val: "}",
		},
		{
			typ: token.EOF,
		},
	}

	testTable("test return", `fn test() { return "some value"}`, expected, t)
	testTable("test return", `fn test() {
	return "some value"
}`, expected, t)

	expected = []Token{
		{
			typ: token.FnDecl,
			val: "fn",
		},
		{
			typ: token.Ident,
			val: "test",
		},
		{
			typ: token.LParen,
			val: "(",
		},
		{
			typ: token.RParen,
			val: ")",
		},
		{
			typ: token.LBrace,
			val: "{",
		},
		{
			typ: token.Ident,
			val: "value",
		},
		{
			typ: token.Assign,
			val: "=",
		},
		{
			typ: token.String,
			val: "some value",
		},
		{
			typ: token.Return,
			val: "return",
		},
		{
			typ: token.Variable,
			val: "$value",
		},
		{
			typ: token.RBrace,
			val: "}",
		},
		{
			typ: token.EOF,
		},
	}

	testTable("test return", `fn test() {
	value = "some value"
	return $value
}`, expected, t)

	expected = []Token{
		{
			typ: token.FnDecl,
			val: "fn",
		},
		{
			typ: token.Ident,
			val: "test",
		},
		{
			typ: token.LParen,
			val: "(",
		},
		{
			typ: token.RParen,
			val: ")",
		},
		{
			typ: token.LBrace,
			val: "{",
		},
		{
			typ: token.Return,
			val: "return",
		},
		{
			typ: token.LParen,
			val: "(",
		},
		{
			typ: token.String,
			val: "test",
		},
		{
			typ: token.String,
			val: "test2",
		},
		{
			typ: token.RParen,
			val: ")",
		},
		{
			typ: token.RBrace,
			val: "}",
		},
		{
			typ: token.EOF,
		},
	}

	testTable("test return", `fn test() {
	return ("test" "test2")
}`, expected, t)

	expected = []Token{
		{
			typ: token.FnDecl,
			val: "fn",
		},
		{
			typ: token.Ident,
			val: "test",
		},
		{
			typ: token.LParen,
			val: "(",
		},
		{
			typ: token.RParen,
			val: ")",
		},
		{
			typ: token.LBrace,
			val: "{",
		},
		{
			typ: token.Return,
			val: "return",
		},
		{
			typ: token.Variable,
			val: "$PWD",
		},
		{
			typ: token.RBrace,
			val: "}",
		},
		{
			typ: token.EOF,
		},
	}

	testTable("test return", `fn test() {
	return $PWD
}`, expected, t)
}

func TestLexerFor(t *testing.T) {
	expected := []Token{
		{
			typ: token.For,
			val: "for",
		},
		{
			typ: token.LBrace,
			val: "{",
		},
		{
			typ: token.RBrace,
			val: "}",
		},
		{
			typ: token.EOF,
		},
	}

	testTable("test inf loop", `for {}`, expected, t)

	expected = []Token{
		{
			typ: token.For,
			val: "for",
		},
		{
			typ: token.Ident,
			val: "f",
		},
		{
			typ: token.ForIn,
			val: "in",
		},
		{
			typ: token.Variable,
			val: "$files",
		},
		{
			typ: token.LBrace,
			val: "{",
		},
		{
			typ: token.RBrace,
			val: "}",
		},
		{
			typ: token.EOF,
		},
	}

	testTable("test inf loop", `for f in $files {}`, expected, t)

}

func TestLexerFnAsFirstClass(t *testing.T) {
	expected := []Token{
		{
			typ: token.FnDecl,
			val: "fn",
		},
		{
			typ: token.Ident,
			val: "printer",
		},
		{
			typ: token.LParen,
			val: "(",
		},
		{
			typ: token.Ident,
			val: "val",
		},
		{
			typ: token.RParen,
			val: ")",
		},
		{
			typ: token.LBrace,
			val: "{",
		},
		{
			typ: token.Command,
			val: "echo",
		},
		{
			typ: token.Arg,
			val: "-n",
		},
		{
			typ: token.Variable,
			val: "$val",
		},
		{
			typ: token.RBrace,
			val: "}",
		},
		{
			typ: token.FnDecl,
			val: "fn",
		},
		{
			typ: token.Ident,
			val: "success",
		},
		{
			typ: token.LParen,
			val: "(",
		},
		{
			typ: token.Ident,
			val: "print",
		},
		{
			typ: token.Ident,
			val: "val",
		},
		{
			typ: token.RParen,
			val: ")",
		},
		{
			typ: token.LBrace,
			val: "{",
		},
		{
			typ: token.FnInv,
			val: "$print",
		},
		{
			typ: token.LParen,
			val: "(",
		},
		{
			typ: token.String,
			val: "[SUCCESS] ",
		},
		{
			typ: token.Concat,
			val: "+",
		},
		{
			typ: token.Variable,
			val: "$val",
		},
		{
			typ: token.RParen,
			val: ")",
		},
		{
			typ: token.RBrace,
			val: "}",
		},
		{
			typ: token.FnInv,
			val: "success",
		},
		{
			typ: token.LParen,
			val: "(",
		},
		{
			typ: token.Variable,
			val: "$printer",
		},
		{
			typ: token.String,
			val: "Command executed!",
		},
		{
			typ: token.RParen,
			val: ")",
		},
		{
			typ: token.EOF,
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
	expected := []Token{
		{
			typ: token.Ident,
			val: "cmd",
		},
		{
			typ: token.Assign,
			val: "=",
		},
		{
			typ: token.Variable,
			val: "$commands",
		},
		{
			typ: token.LBrack,
			val: "[",
		},
		{
			typ: token.Number,
			val: "0",
		},
		{
			typ: token.RBrack,
			val: "]",
		},
		{
			typ: token.EOF,
		},
	}

	for i := 0; i < 1000; i++ {
		expected[4] = Token{
			typ: token.Number,
			val: strconv.Itoa(i),
		}

		testTable("test variable indexing", `cmd = $commands[`+strconv.Itoa(i)+`]`, expected, t)
	}

	expected = []Token{
		{
			typ: token.Ident,
			val: "cmd",
		},
		{
			typ: token.Assign,
			val: "=",
		},
		{
			typ: token.Variable,
			val: "$commands",
		},
		{
			typ: token.LBrack,
			val: "[",
		},
		{
			typ: token.Illegal,
			val: "Expected number or variable on variable indexing. Found 'a'",
		},
		{
			typ: token.EOF,
		},
	}

	testTable("test invalid number", `cmd = $commands[a]`, expected, t)

	expected = []Token{
		{
			typ: token.Ident,
			val: "cmd",
		},
		{
			typ: token.Assign,
			val: "=",
		},
		{
			typ: token.Variable,
			val: "$commands",
		},
		{
			typ: token.LBrack,
			val: "[",
		},
		{
			typ: token.Illegal,
			val: "Expected number or variable on variable indexing. Found ']'",
		},
		{
			typ: token.EOF,
		},
	}

	testTable("test invalid number", `cmd = $commands[]`, expected, t)

	expected = []Token{
		{
			typ: token.Command,
			val: "echo",
		},
		{
			typ: token.Arg,
			val: "test",
		},
		{
			typ: token.Variable,
			val: "$names",
		},
		{
			typ: token.LBrack,
			val: "[",
		},
		{
			typ: token.Number,
			val: "666",
		},
		{
			typ: token.RBrack,
			val: "]",
		},
		{
			typ: token.EOF,
		},
	}

	testTable("test variable index on commands", `echo test $names[666]`, expected, t)

	expected = []Token{
		{
			typ: token.If,
			val: "if",
		},
		{
			typ: token.Variable,
			val: "$crazies",
		},
		{
			typ: token.LBrack,
			val: "[",
		},
		{
			typ: token.Number,
			val: "0",
		},
		{
			typ: token.RBrack,
			val: "]",
		},
		{
			typ: token.Equal,
			val: "==",
		},
		{
			typ: token.String,
			val: "patito",
		},
		{
			typ: token.LBrace,
			val: "{",
		},
		{
			typ: token.Command,
			val: "echo",
		},
		{
			typ: token.String,
			val: ":D",
		},
		{
			typ: token.RBrace,
			val: "}",
		},
		{
			typ: token.EOF,
		},
	}

	testTable("test if with indexing", `if $crazies[0] == "patito" { echo ":D" }`, expected, t)
}
