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

		fmt.Printf("\n")

		for _, exp := range expected {
			fmt.Printf("expect: %+v\n", exp)
		}

		return
	}

	for i := 0; i < len(expected); i++ {
		if expected[i].typ != result[i].typ {
			t.Errorf("'%s (%s)' != '%s (%s)'", expected[i].typ, expected[i].val, result[i].typ, result[i].val)
			return
		}

		if expected[i].val != result[i].val {
			t.Errorf("Parsing '%s':\n\terror: '%s' != '%s'", content, expected[i].val, result[i].val)
			return
		}
	}
}

func TestLexerCommandStringArgs(t *testing.T) {
	expected := []Token{
		{typ: token.Ident, val: "echo"},
		{typ: token.Ident, val: "hello"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.Ident, val: "echo"},
		{typ: token.String, val: "hello"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.EOF},
	}

	testTable("test args", `echo hello
echo "hello"`, expected, t)
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
		typ: token.Ident,
		val: "echo",
	}

	if it.String() != "IDENT" {
		t.Errorf("wrong command name: %s", it.String())
	}

	it = Token{
		typ: token.Ident,
		val: "echoooooooooooooooooooooooo",
	}

	// test if long names are truncated
	if it.String() != "IDENT" {
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
		{typ: token.SetEnv, val: "setenv"},
		{typ: token.Ident, val: "name"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.EOF},
	}

	testTable("testSet", `setenv name`, expected, t)
}

func TestLexerSimpleAssignment(t *testing.T) {
	expected := []Token{
		{typ: token.Ident, val: "test"},
		{typ: token.Assign, val: "="},
		{typ: token.String, val: "value"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.EOF},
	}

	testTable("testAssignment", `test="value"`, expected, t)
	testTable("testAssignment spacy", `test = "value"`, expected, t)
	testTable("testAssignment spacy", `test          ="value"`, expected, t)
	testTable("testAssignment spacy", `test=           "value"`, expected, t)
	testTable("testAssignment spacy", `test	="value"`, expected, t)
	testTable("testAssignment spacy", `test		="value"`, expected, t)
	testTable("testAssignment spacy", `test =	"value"`, expected, t)

	expected = []Token{
		{typ: token.Ident, val: "test"},
		{typ: token.Assign, val: "="},
		{typ: token.String, val: "value"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.Ident, val: "other"},
		{typ: token.Assign, val: "="},
		{typ: token.String, val: "other"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.EOF},
	}

	testTable("test multiple Assignments", `
        test="value"
        other="other"`, expected, t)

	expected = []Token{
		{typ: token.Ident, val: "test"},
		{typ: token.Assign, val: "="},
		{typ: token.String, val: "value"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.Ident, val: "other"},
		{typ: token.Assign, val: "="},
		{typ: token.Variable, val: "$test"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.Ident, val: "echo"},
		{typ: token.Variable, val: "$other"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.EOF},
	}

	testTable("test multiple Assignments", `
        test="value"
        other=$test
        echo $other`, expected, t)

	expected = []Token{
		{typ: token.Ident, val: "STALI_SRC"},
		{typ: token.Assign, val: "="},
		{typ: token.Variable, val: "$PWD"},
		{typ: token.Plus, val: "+"},
		{typ: token.String, val: "/src"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.EOF},
	}

	testTable("test underscore", `STALI_SRC = $PWD + "/src"`, expected, t)

	expected = []Token{
		{typ: token.Ident, val: "PROMPT"},
		{typ: token.Assign, val: "="},
		{typ: token.String, val: "("},
		{typ: token.Plus, val: "+"},
		{typ: token.Variable, val: "$path"},
		{typ: token.Plus, val: "+"},
		{typ: token.String, val: ")"},
		{typ: token.Plus, val: "+"},
		{typ: token.Variable, val: "$PROMPT"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.EOF},
	}

	testTable("test concat with parenthesis", `PROMPT = "("+$path+")"+$PROMPT`, expected, t)

	expected = []Token{
		{typ: token.Ident, val: "a"},
		{typ: token.LBrack, val: "["},
		{typ: token.Number, val: "0"},
		{typ: token.RBrack, val: "]"},
		{typ: token.Assign, val: "="},
		{typ: token.String, val: "test"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.EOF},
	}

	testTable("test index assignment", `a[0] = "test"`, expected, t)

}

func TestLexerListAssignment(t *testing.T) {
	expected := []Token{
		{typ: token.Ident, val: "test"},
		{typ: token.Assign, val: "="},
		{typ: token.LParen, val: "("},
		{typ: token.Ident, val: "plan9"},
		{typ: token.Ident, val: "from"},
		{typ: token.Ident, val: "bell"},
		{typ: token.Ident, val: "labs"},
		{typ: token.RParen, val: ")"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.EOF},
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
		{typ: token.Ident, val: "test"},
		{typ: token.Assign, val: "="},
		{typ: token.LParen, val: "("},
		{typ: token.String, val: "plan9"},
		{typ: token.Ident, val: "from"},
		{typ: token.String, val: "bell"},
		{typ: token.Ident, val: "labs"},
		{typ: token.RParen, val: ")"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.EOF},
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
		{typ: token.Ident, val: "test"},
		{typ: token.Assign, val: "="},
		{typ: token.LParen, val: "("},
		{typ: token.Variable, val: "$plan9"},
		{typ: token.Ident, val: "from"},
		{typ: token.Variable, val: "$bell"},
		{typ: token.Ident, val: "labs"},
		{typ: token.RParen, val: ")"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.EOF},
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

func TestLexerListOfLists(t *testing.T) {
	expected := []Token{
		{typ: token.Ident, val: "l"},
		{typ: token.Assign, val: "="},
		{typ: token.LParen, val: "("},
		{typ: token.LParen, val: "("},
		{typ: token.RParen, val: ")"},
		{typ: token.RParen, val: ")"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.EOF},
	}

	testTable("testlistoflists", `l = (())`, expected, t)
	testTable("testlistoflists", `l = (
		()
	)`, expected, t)

	expected = []Token{
		{typ: token.Ident, val: "l"},
		{typ: token.Assign, val: "="},
		{typ: token.LParen, val: "("},

		{typ: token.LParen, val: "("},
		{typ: token.Ident, val: "plan9"},
		{typ: token.Ident, val: "from"},
		{typ: token.Ident, val: "bell"},
		{typ: token.Ident, val: "labs"},
		{typ: token.RParen, val: ")"},

		{typ: token.LParen, val: "("},
		{typ: token.Ident, val: "linux"},
		{typ: token.RParen, val: ")"},

		{typ: token.RParen, val: ")"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.EOF},
	}

	testTable("testlistoflists", `l = ((plan9 from bell labs) (linux))`, expected, t)
	testTable("testlistoflists", `l = (
		(plan9 from bell labs)
		(linux)
	)`, expected, t)

}

func TestLexerInvalidAssignments(t *testing.T) {
	expected := []Token{
		{typ: token.Ident, val: "test"},
		{typ: token.Assign, val: "="},
		{typ: token.String, val: "value"},
		{typ: token.Ident, val: "other"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.EOF},
	}

	testTable("testInvalidAssignments", `test="value" other`, expected, t)
}

func TestLexerSimpleCommand(t *testing.T) {
	expected := []Token{
		{typ: token.Ident, val: "echo"},
		{typ: token.String, val: "hello world"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.EOF},
	}

	testTable("testSimpleCommand", `echo "hello world"`, expected, t)

	expected = []Token{
		{typ: token.Ident, val: "echo"},
		{typ: token.Arg, val: "rootfs-x86_64"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.EOF},
	}

	testTable("testSimpleCommand", `echo rootfs-x86_64`, expected, t)

	expected = []Token{
		{typ: token.Ident, val: "git"},
		{typ: token.Ident, val: "clone"},
		{typ: token.Arg, val: "--depth=1"},
		{typ: token.Arg, val: "http://git.sta.li/toolchain"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.EOF},
	}

	testTable("testSimpleCommand", `git clone --depth=1 http://git.sta.li/toolchain`, expected, t)

	expected = []Token{
		{typ: token.Ident, val: "ls"},
		{typ: token.Variable, val: "$GOPATH"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.EOF},
	}

	testTable("testSimpleCommand", `ls $GOPATH`, expected, t)

	expected = []Token{
		{typ: token.Ident, val: "ls"},
		{typ: token.Variable, val: "$GOPATH"},
		{typ: token.Plus, val: "+"},
		{typ: token.String, val: "/src/github.com"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.EOF},
	}

	testTable("testSimpleCommand", `ls $GOPATH+"/src/github.com"`, expected, t)

	expected = []Token{
		{typ: token.Ident, val: "ls"},
		{typ: token.String, val: "/src/github.com"},
		{typ: token.Plus, val: "+"},
		{typ: token.Variable, val: "$GOPATH"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.EOF},
	}

	testTable("testSimpleCommand", `ls "/src/github.com"+$GOPATH`, expected, t)

	expected = []Token{
		{typ: token.Ident, val: "ls"},
		{typ: token.String, val: "/home/user"},
		{typ: token.Plus, val: "+"},
		{typ: token.String, val: "/.gvm/pkgsets/global/src"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.EOF},
	}

	testTable("testSimpleCommand", `ls "/home/user" + "/.gvm/pkgsets/global/src"`, expected, t)

}

func TestLexerPipe(t *testing.T) {
	expected := []Token{
		{typ: token.Ident, val: "ls"},
		{typ: token.Pipe, val: "|"},
		{typ: token.Ident, val: "wc"},
		{typ: token.Arg, val: "-l"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.EOF},
	}

	testTable("testPipe", `ls | wc -l`, expected, t)

	expected = []Token{
		{typ: token.Ident, val: "ls"},
		{typ: token.Arg, val: "-l"},
		{typ: token.Pipe, val: "|"},
		{typ: token.Ident, val: "wc"},
		{typ: token.Pipe, val: "|"},
		{typ: token.Ident, val: "awk"},
		{typ: token.String, val: "{print $1}"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.EOF},
	}

	testTable("testPipe", `ls -l | wc | awk "{print $1}"`, expected, t)

	expected = []Token{
		{typ: token.Ident, val: "go"},
		{typ: token.Ident, val: "tool"},
		{typ: token.Ident, val: "vet"},
		{typ: token.Arg, val: "-h"},
		{typ: token.Gt, val: ">"},
		{typ: token.LBrack, val: "["},
		{typ: token.Number, val: "2"},
		{typ: token.Assign, val: "="},
		{typ: token.Number, val: "1"},
		{typ: token.RBrack, val: "]"},
		{typ: token.Pipe, val: "|"},
		{typ: token.Ident, val: "grep"},
		{typ: token.Ident, val: "log"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.EOF},
	}

	testTable("testPipe with redirection", `go tool vet -h >[2=1] | grep log`, expected, t)

	expected = []Token{
		{typ: token.Ident, val: "go"},
		{typ: token.Ident, val: "tool"},
		{typ: token.Ident, val: "vet"},
		{typ: token.Arg, val: "-h"},
		{typ: token.Gt, val: ">"},
		{typ: token.Arg, val: "out.log"},
		{typ: token.Pipe, val: "|"},
		{typ: token.Ident, val: "grep"},
		{typ: token.Ident, val: "log"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.EOF},
	}

	testTable("testPipe with redirection", `go tool vet -h > out.log | grep log`, expected, t)
}

func TestPipeFunctions(t *testing.T) {
	expected := []Token{
		{typ: token.Ident, val: "echo"},
		{typ: token.String, val: "some thing"},
		{typ: token.Pipe, val: "|"},
		{typ: token.Ident, val: "replace"},
		{typ: token.LParen, val: "("},
		{typ: token.String, val: " "},
		{typ: token.Comma, val: ","},
		{typ: token.String, val: "|"},
		{typ: token.RParen, val: ")"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.EOF},
	}

	testTable("test pipe with function",
		`echo "some thing" | replace(" ", "|")`,
		expected, t)
}

func TestLexerUnquoteArg(t *testing.T) {
	expected := []Token{
		{typ: token.Ident, val: "echo"},
		{typ: token.Ident, val: "hello"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.EOF},
	}

	testTable("testSimpleCommand", `echo hello`, expected, t)

	expected = []Token{
		{typ: token.Ident, val: "echo"},
		{typ: token.Arg, val: "hello-world"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.EOF},
	}

	testTable("testSimpleCommand", `echo hello-world`, expected, t)

	expected = []Token{
		{typ: token.Ident, val: "echo"},
		{typ: token.Comment, val: "#hello-world"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.EOF},
	}

	testTable("testSimpleCommand", `echo #hello-world`, expected, t)
}

func TestLexerDashedCommand(t *testing.T) {
	expected := []Token{
		{typ: token.Arg, val: "google-chrome"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.EOF},
	}

	testTable("testDashedCommand", `google-chrome`, expected, t)
}

func TestLexerPathCommand(t *testing.T) {
	expected := []Token{
		{typ: token.Arg, val: "/bin/echo"},
		{typ: token.String, val: "hello world"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.EOF},
	}

	testTable("testPathCommand", `/bin/echo "hello world"`, expected, t)
}

func TestLexerInvalidBlock(t *testing.T) {
	expected := []Token{
		{typ: token.LBrace, val: "{"},
		{typ: token.EOF},
	}

	testTable("testInvalidBlock", "{", expected, t)
}

func TestLexerQuotedStringNotFinished(t *testing.T) {
	expected := []Token{
		{
			typ: token.Ident,
			val: "echo",
		},
		{
			typ: token.Illegal,
			val: "testQuotedstringnotfinished:1:17: Quoted string not finished: hello world",
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
		{typ: token.Ident, val: "echo"},
		{typ: token.String, val: "hello world"},
		{typ: token.Semicolon, val: ";"},

		{typ: token.Ident, val: "mount"},
		{typ: token.Arg, val: "-t"},
		{typ: token.Ident, val: "proc"},
		{typ: token.Ident, val: "proc"},
		{typ: token.Arg, val: "/proc"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.EOF},
	}

	testTable("testVariouscommands", content, expected, t)
}

func TestLexerRfork(t *testing.T) {
	expected := []Token{
		{typ: token.Rfork, val: "rfork"},
		{typ: token.Ident, val: "u"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.EOF},
	}

	testTable("testRfork", "rfork u\n", expected, t)

	expected = []Token{
		{typ: token.Rfork, val: "rfork"},
		{typ: token.Ident, val: "usnm"},
		{typ: token.LBrace, val: "{"},
		{typ: token.Ident, val: "echo"},
		{typ: token.String, val: "inside namespace :)"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.RBrace, val: "}"},
		{typ: token.EOF},
	}

	testTable("testRforkWithBlock", `
            rfork usnm {
                echo "inside namespace :)"
            }
        `, expected, t)

}

func TestLexerSomethingIdontcareanymore(t *testing.T) {
	// maybe oneliner rfork isnt a good idea
	expected := []Token{
		{typ: token.Rfork, val: "rfork"},
		{typ: token.Ident, val: "u"},
		{typ: token.LBrace, val: "{"},
		{typ: token.Ident, val: "ls"},
		{typ: token.RBrace, val: "}"},
		{typ: token.EOF},
	}

	testTable("test whatever", "rfork u { ls }", expected, t)
}

func TestLexerBuiltinCd(t *testing.T) {
	expected := []Token{
		{typ: token.Ident, val: "cd"},
		{typ: token.String, val: "some place"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.EOF},
	}

	testTable("testBuiltinCd", `cd "some place"`, expected, t)

	expected = []Token{
		{typ: token.Ident, val: "cd"},
		{typ: token.Arg, val: "/proc"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.EOF},
	}

	testTable("testBuiltinCdNoQuote", `cd /proc`, expected, t)

	expected = []Token{
		{typ: token.Ident, val: "cd"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.EOF},
	}

	testTable("testBuiltincd home", `cd`, expected, t)

	expected = []Token{
		{typ: token.Ident, val: "HOME"},
		{typ: token.Assign, val: "="},
		{typ: token.String, val: "/"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.SetEnv, val: "setenv"},
		{typ: token.Ident, val: "HOME"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.Ident, val: "cd"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.Ident, val: "pwd"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.EOF},
	}

	testTable("testBuiltin cd bug", `
	               HOME="/"
                       setenv HOME
	               cd
	               pwd
	           `, expected, t)

	expected = []Token{
		{typ: token.Ident, val: "cd"},
		{typ: token.Variable, val: "$GOPATH"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.EOF},
	}

	testTable("test builtin cd into variable", `cd $GOPATH`, expected, t)

	expected = []Token{
		{typ: token.Ident, val: "cd"},
		{typ: token.Variable, val: "$GOPATH"},
		{typ: token.Plus, val: "+"},
		{typ: token.String, val: "/src/github.com"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.EOF},
	}

	testTable("test cd with concat", `cd $GOPATH+"/src/github.com"`, expected, t)
}

func TestLexerRedirectSimple(t *testing.T) {
	expected := []Token{
		{typ: token.Ident, val: "cmd"},
		{typ: token.Gt, val: ">"},
		{typ: token.Arg, val: "file.out"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.EOF},
	}

	testTable("test simple redirect", "cmd > file.out", expected, t)

	expected = []Token{
		{typ: token.Ident, val: "cmd"},
		{typ: token.Gt, val: ">"},
		{typ: token.String, val: "tcp://localhost:8888"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.EOF},
	}

	testTable("test simple redirect", `cmd > "tcp://localhost:8888"`, expected, t)

	expected = []Token{
		{typ: token.Ident, val: "cmd"},
		{typ: token.Gt, val: ">"},
		{typ: token.String, val: "udp://localhost:8888"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.EOF},
	}

	testTable("test simple redirect", `cmd > "udp://localhost:8888"`, expected, t)

	expected = []Token{
		{typ: token.Ident, val: "cmd"},
		{typ: token.Gt, val: ">"},
		{typ: token.String, val: "unix:///tmp/sock.txt"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.EOF},
	}

	testTable("test simple redirect", `cmd > "unix:///tmp/sock.txt"`, expected, t)

}

func TestLexerRedirectMap(t *testing.T) {

	// Suppress stderr output
	expected := []Token{
		{typ: token.Ident, val: "cmd"},
		{typ: token.Gt, val: ">"},
		{typ: token.LBrack, val: "["},
		{typ: token.Number, val: "2"},
		{typ: token.Assign, val: "="},
		{typ: token.RBrack, val: "]"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.EOF},
	}

	testTable("test suppress stderr", "cmd >[2=]", expected, t)

	// points stderr to stdout
	expected = []Token{
		{typ: token.Ident, val: "cmd"},
		{typ: token.Gt, val: ">"},
		{typ: token.LBrack, val: "["},
		{typ: token.Number, val: "2"},
		{typ: token.Assign, val: "="},
		{typ: token.Number, val: "1"},
		{typ: token.RBrack, val: "]"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.EOF},
	}

	testTable("test stderr=stdout", "cmd >[2=1]", expected, t)
}

func TestLexerRedirectMapToLocation(t *testing.T) {
	// Suppress stderr output
	expected := []Token{
		{typ: token.Ident, val: "cmd"},
		{typ: token.Gt, val: ">"},
		{typ: token.LBrack, val: "["},
		{typ: token.Number, val: "2"},
		{typ: token.Assign, val: "="},
		{typ: token.RBrack, val: "]"},
		{typ: token.Arg, val: "file.out"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.EOF},
	}

	testTable("test suppress stderr", "cmd >[2=] file.out", expected, t)

	// points stderr to stdout
	expected = []Token{
		{typ: token.Ident, val: "cmd"},
		{typ: token.Gt, val: ">"},
		{typ: token.LBrack, val: "["},
		{typ: token.Number, val: "2"},
		{typ: token.Assign, val: "="},
		{typ: token.Number, val: "1"},
		{typ: token.RBrack, val: "]"},
		{typ: token.Arg, val: "file.out"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.EOF},
	}

	testTable("test stderr=stdout", "cmd >[2=1] file.out", expected, t)

	expected = []Token{
		{typ: token.Ident, val: "cmd"},
		{typ: token.Gt, val: ">"},
		{typ: token.LBrack, val: "["},
		{typ: token.Number, val: "2"},
		{typ: token.RBrack, val: "]"},
		{typ: token.Arg, val: "/var/log/service.log"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.EOF},
	}

	testTable("test suppress stderr", `cmd >[2] /var/log/service.log`, expected, t)
}

func TestLexerRedirectMultipleMaps(t *testing.T) {
	expected := []Token{
		{typ: token.Ident, val: "cmd"},
		{typ: token.Gt, val: ">"},
		{typ: token.LBrack, val: "["},
		{typ: token.Number, val: "1"},
		{typ: token.RBrack, val: "]"},
		{typ: token.Arg, val: "file.out"},
		{typ: token.Gt, val: ">"},
		{typ: token.LBrack, val: "["},
		{typ: token.Number, val: "2"},
		{typ: token.RBrack, val: "]"},
		{typ: token.Arg, val: "file.err"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.EOF},
	}

	testTable("test suppress stderr", `cmd >[1] file.out >[2] file.err`, expected, t)

	expected = []Token{
		{typ: token.Ident, val: "cmd"},
		{typ: token.Gt, val: ">"},
		{typ: token.LBrack, val: "["},
		{typ: token.Number, val: "2"},
		{typ: token.Assign, val: "="},
		{typ: token.Number, val: "1"},
		{typ: token.RBrack, val: "]"},
		{typ: token.Gt, val: ">"},
		{typ: token.LBrack, val: "["},
		{typ: token.Number, val: "1"},
		{typ: token.RBrack, val: "]"},
		{typ: token.Arg, val: "/var/log/service.log"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.EOF},
	}

	testTable("test suppress stderr", `cmd >[2=1] >[1] /var/log/service.log`, expected, t)

	expected = []Token{
		{typ: token.Ident, val: "cmd"},
		{typ: token.Gt, val: ">"},
		{typ: token.LBrack, val: "["},
		{typ: token.Number, val: "1"},
		{typ: token.Assign, val: "="},
		{typ: token.Number, val: "2"},
		{typ: token.RBrack, val: "]"},
		{typ: token.Gt, val: ">"},
		{typ: token.LBrack, val: "["},
		{typ: token.Number, val: "2"},
		{typ: token.Assign, val: "="},
		{typ: token.RBrack, val: "]"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.EOF},
	}

	testTable("test suppress stderr", `cmd >[1=2] >[2=]`, expected, t)
}

func TestLexerImport(t *testing.T) {
	expected := []Token{
		{typ: token.Import, val: "import"},
		{typ: token.Arg, val: "env.sh"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.EOF},
	}

	testTable("test import", `import env.sh`, expected, t)
}

func TestLexerSimpleIf(t *testing.T) {
	expected := []Token{
		{typ: token.If, val: "if"},
		{typ: token.String, val: "test"},
		{typ: token.Equal, val: "=="},
		{typ: token.String, val: "other"},
		{typ: token.LBrace, val: "{"},
		{typ: token.Ident, val: "rm"},
		{typ: token.Arg, val: "-rf"},
		{typ: token.Arg, val: "/"},
		{typ: token.RBrace, val: "}"},
		{typ: token.EOF},
	}

	testTable("test simple if", `if "test" == "other" { rm -rf / }`, expected, t)

	expected = []Token{
		{typ: token.If, val: "if"},
		{typ: token.String, val: "test"},
		{typ: token.NotEqual, val: "!="},
		{typ: token.Variable, val: "$test"},
		{typ: token.LBrace, val: "{"},
		{typ: token.Ident, val: "rm"},
		{typ: token.Arg, val: "-rf"},
		{typ: token.Arg, val: "/"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.RBrace, val: "}"},
		{typ: token.EOF},
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
		{typ: token.If, val: "if"},
		{typ: token.Variable, val: "$test"},
		{typ: token.Plus, val: "+"},
		{typ: token.String, val: "001"},
		{typ: token.NotEqual, val: "!="},
		{typ: token.String, val: "value001"},
		{typ: token.LBrace, val: "{"},
		{typ: token.Ident, val: "rm"},
		{typ: token.Arg, val: "-rf"},
		{typ: token.Arg, val: "/"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.RBrace, val: "}"},
		{typ: token.EOF},
	}

	testTable("test if concat", `if $test + "001" != "value001" {
        rm -rf /
}`, expected, t)
}

func TestLexerIfWithFuncInvocation(t *testing.T) {
	expected := []Token{
		{typ: token.If, val: "if"},
		{typ: token.Ident, val: "test"},
		{typ: token.LParen, val: "("},
		{typ: token.String, val: "some val"},
		{typ: token.RParen, val: ")"},
		{typ: token.NotEqual, val: "!="},
		{typ: token.String, val: "value001"},
		{typ: token.LBrace, val: "{"},
		{typ: token.Ident, val: "rm"},
		{typ: token.Arg, val: "-rf"},
		{typ: token.Arg, val: "/"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.RBrace, val: "}"},
		{typ: token.EOF},
	}

	testTable("test if concat", `if test("some val") != "value001" {
        rm -rf /
}`, expected, t)
}

func TestLexerIfElse(t *testing.T) {
	expected := []Token{
		{typ: token.If, val: "if"},
		{typ: token.String, val: "test"},
		{typ: token.Equal, val: "=="},
		{typ: token.String, val: "other"},
		{typ: token.LBrace, val: "{"},
		{typ: token.Ident, val: "rm"},
		{typ: token.Arg, val: "-rf"},
		{typ: token.Arg, val: "/"},
		{typ: token.RBrace, val: "}"},
		{typ: token.Else, val: "else"},
		{typ: token.LBrace, val: "{"},
		{typ: token.Ident, val: "pwd"},
		{typ: token.RBrace, val: "}"},
		{typ: token.EOF},
	}

	testTable("test simple if", `if "test" == "other" { rm -rf / } else { pwd }`, expected, t)
}

func TestLexerIfElseIf(t *testing.T) {
	expected := []Token{
		{typ: token.If, val: "if"},
		{typ: token.String, val: "test"},
		{typ: token.Equal, val: "=="},
		{typ: token.String, val: "other"},
		{typ: token.LBrace, val: "{"},
		{typ: token.Ident, val: "rm"},
		{typ: token.Arg, val: "-rf"},
		{typ: token.Arg, val: "/"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.RBrace, val: "}"},
		{typ: token.Else, val: "else"},
		{typ: token.If, val: "if"},
		{typ: token.String, val: "test"},
		{typ: token.Equal, val: "=="},
		{typ: token.Variable, val: "$var"},
		{typ: token.LBrace, val: "{"},
		{typ: token.Ident, val: "pwd"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.RBrace, val: "}"},
		{typ: token.Else, val: "else"},
		{typ: token.LBrace, val: "{"},
		{typ: token.Ident, val: "exit"},
		{typ: token.Number, val: "1"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.RBrace, val: "}"},
		{typ: token.EOF},
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
		{typ: token.Fn, val: "fn"},
		{typ: token.Ident, val: "build"},
		{typ: token.LParen, val: "("},
		{typ: token.RParen, val: ")"},
		{typ: token.LBrace, val: "{"},
		{typ: token.RBrace, val: "}"},
		{typ: token.EOF},
	}

	testTable("test empty fn", `fn build() {}`, expected, t)

	// lambda
	expected = []Token{
		{typ: token.Fn, val: "fn"},
		{typ: token.LParen, val: "("},
		{typ: token.RParen, val: ")"},
		{typ: token.LBrace, val: "{"},
		{typ: token.RBrace, val: "}"},
		{typ: token.EOF},
	}

	testTable("test empty fn", `fn () {}`, expected, t)

	// IIFE
	expected = []Token{
		{typ: token.Fn, val: "fn"},
		{typ: token.LParen, val: "("},
		{typ: token.RParen, val: ")"},
		{typ: token.LBrace, val: "{"},
		{typ: token.RBrace, val: "}"},
		{typ: token.LParen, val: "("},
		{typ: token.RParen, val: ")"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.EOF},
	}

	testTable("test empty fn", `fn () {}()`, expected, t)

	expected = []Token{
		{typ: token.Fn, val: "fn"},
		{typ: token.Ident, val: "build"},
		{typ: token.LParen, val: "("},
		{typ: token.Ident, val: "image"},
		{typ: token.Comma, val: ","},
		{typ: token.Ident, val: "debug"},
		{typ: token.RParen, val: ")"},
		{typ: token.LBrace, val: "{"},
		{typ: token.RBrace, val: "}"},
		{typ: token.EOF},
	}

	testTable("test empty fn with args", `fn build(image, debug) {}`, expected, t)

	expected = []Token{
		{typ: token.Fn, val: "fn"},
		{typ: token.Ident, val: "build"},
		{typ: token.LParen, val: "("},
		{typ: token.Ident, val: "image"},
		{typ: token.Comma, val: ","},
		{typ: token.Ident, val: "debug"},
		{typ: token.RParen, val: ")"},
		{typ: token.LBrace, val: "{"},
		{typ: token.Ident, val: "ls"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.Ident, val: "tar"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.RBrace, val: "}"},
		{typ: token.EOF},
	}

	testTable("test empty fn with args and body", `fn build(image, debug) {
            ls
            tar
        }`, expected, t)

	expected = []Token{
		{typ: token.Fn, val: "fn"},
		{typ: token.Ident, val: "cd"},
		{typ: token.LParen, val: "("},
		{typ: token.Ident, val: "path"},
		{typ: token.RParen, val: ")"},
		{typ: token.LBrace, val: "{"},
		{typ: token.Ident, val: "cd"},
		{typ: token.Variable, val: "$path"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.Ident, val: "PROMPT"},
		{typ: token.Assign, val: "="},
		{typ: token.String, val: "("},
		{typ: token.Plus, val: "+"},
		{typ: token.Variable, val: "$path"},
		{typ: token.Plus, val: "+"},
		{typ: token.String, val: ")"},
		{typ: token.Plus, val: "+"},
		{typ: token.Variable, val: "$PROMPT"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.SetEnv, val: "setenv"},
		{typ: token.Ident, val: "PROMPT"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.RBrace, val: "}"},
		{typ: token.EOF},
	}

	testTable("test cd fn with PROMPT update", `fn cd(path) {
    cd $path
    PROMPT="(" + $path + ")"+$PROMPT
    setenv PROMPT
}`, expected, t)
}

func TestLexerFnInvocation(t *testing.T) {
	expected := []Token{
		{typ: token.Ident, val: "build"},
		{typ: token.LParen, val: "("},
		{typ: token.RParen, val: ")"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.EOF},
	}

	testTable("test fn invocation", `build()`, expected, t)

	expected = []Token{
		{typ: token.Ident, val: "build"},
		{typ: token.LParen, val: "("},
		{typ: token.String, val: "ubuntu"},
		{typ: token.RParen, val: ")"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.EOF},
	}

	testTable("test fn invocation", `build("ubuntu")`, expected, t)

	expected = []Token{
		{typ: token.Ident, val: "build"},
		{typ: token.LParen, val: "("},
		{typ: token.String, val: "ubuntu"},
		{typ: token.Comma, val: ","},
		{typ: token.Variable, val: "$debug"},

		{typ: token.RParen, val: ")"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.EOF},
	}

	testTable("test fn invocation", `build("ubuntu", $debug)`, expected, t)

	expected = []Token{
		{typ: token.Ident, val: "build"},
		{typ: token.LParen, val: "("},
		{typ: token.Variable, val: "$debug"},

		{typ: token.RParen, val: ")"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.EOF},
	}

	testTable("test fn invocation", `build($debug)`, expected, t)

	expected = []Token{
		{typ: token.Ident, val: "a"},
		{typ: token.LParen, val: "("},
		{typ: token.Ident, val: "b"},
		{typ: token.LParen, val: "("},
		{typ: token.RParen, val: ")"},
		{typ: token.RParen, val: ")"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.EOF},
	}

	testTable("test fn composition", `a(b())`, expected, t)
}

func TestLexerAssignCmdOut(t *testing.T) {
	expected := []Token{
		{typ: token.Ident, val: "ipaddr"},
		{typ: token.AssignCmd, val: "<="},
		{typ: token.Ident, val: "someprogram"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.EOF},
	}

	testTable("test assignCmdOut", `ipaddr <= someprogram`, expected, t)
}

func TestLexerBindFn(t *testing.T) {
	expected := []Token{
		{typ: token.BindFn, val: "bindfn"},
		{typ: token.Ident, val: "cd"},
		{typ: token.Ident, val: "cd2"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.EOF},
	}

	testTable("test bindfn", `bindfn cd cd2`, expected, t)

}

func TestLexerRedirectionNetwork(t *testing.T) {
	expected := []Token{
		{typ: token.Ident, val: "echo"},
		{typ: token.String, val: "hello world"},
		{typ: token.Gt, val: ">"},
		{typ: token.LBrack, val: "["},
		{typ: token.Number, val: "1"},
		{typ: token.RBrack, val: "]"},
		{typ: token.String, val: "tcp://localhost:6667"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.EOF},
	}

	testTable("test redirection network", `echo "hello world" >[1] "tcp://localhost:6667"`, expected, t)
}

func TestLexerDump(t *testing.T) {
	expected := []Token{
		{typ: token.Dump, val: "dump"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.EOF},
	}

	testTable("test dump", `dump`, expected, t)

	expected = []Token{
		{typ: token.Dump, val: "dump"},
		{typ: token.Ident, val: "out"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.EOF},
	}

	testTable("test dump", `dump out`, expected, t)

	expected = []Token{
		{typ: token.Dump, val: "dump"},
		{typ: token.Variable, val: "$out"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.EOF},
	}

	testTable("test dump", `dump $out`, expected, t)
}

func TestLexerReturn(t *testing.T) {
	expected := []Token{
		{typ: token.Return, val: "return"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.EOF},
	}

	testTable("test return", "return", expected, t)

	expected = []Token{
		{typ: token.Fn, val: "fn"},
		{typ: token.Ident, val: "test"},
		{typ: token.LParen, val: "("},
		{typ: token.RParen, val: ")"},
		{typ: token.LBrace, val: "{"},
		{typ: token.Return, val: "return"},
		{typ: token.RBrace, val: "}"},
		{typ: token.EOF},
	}

	testTable("test return", "fn test() { return }", expected, t)

	expected = []Token{
		{typ: token.Fn, val: "fn"},
		{typ: token.Ident, val: "test"},
		{typ: token.LParen, val: "("},
		{typ: token.RParen, val: ")"},
		{typ: token.LBrace, val: "{"},
		{typ: token.Return, val: "return"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.RBrace, val: "}"},
		{typ: token.EOF},
	}

	testTable("test return", `fn test() {
	return
}`, expected, t)

	expected = []Token{
		{typ: token.Fn, val: "fn"},
		{typ: token.Ident, val: "test"},
		{typ: token.LParen, val: "("},
		{typ: token.RParen, val: ")"},
		{typ: token.LBrace, val: "{"},
		{typ: token.Return, val: "return"},
		{typ: token.String, val: "some value"},
		{typ: token.RBrace, val: "}"},
		{typ: token.EOF},
	}

	testTable("test return", `fn test() { return "some value"}`, expected, t)

	expected = []Token{
		{typ: token.Fn, val: "fn"},
		{typ: token.Ident, val: "test"},
		{typ: token.LParen, val: "("},
		{typ: token.RParen, val: ")"},
		{typ: token.LBrace, val: "{"},
		{typ: token.Return, val: "return"},
		{typ: token.String, val: "some value"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.RBrace, val: "}"},
		{typ: token.EOF},
	}

	testTable("test return", `fn test() {
	return "some value"
}`, expected, t)

	expected = []Token{
		{typ: token.Fn, val: "fn"},
		{typ: token.Ident, val: "test"},
		{typ: token.LParen, val: "("},
		{typ: token.RParen, val: ")"},
		{typ: token.LBrace, val: "{"},
		{typ: token.Ident, val: "value"},
		{typ: token.Assign, val: "="},
		{typ: token.String, val: "some value"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.Return, val: "return"},
		{typ: token.Variable, val: "$value"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.RBrace, val: "}"},
		{typ: token.EOF},
	}

	testTable("test return", `fn test() {
	value = "some value"
	return $value
}`, expected, t)

	expected = []Token{
		{typ: token.Fn, val: "fn"},
		{typ: token.Ident, val: "test"},
		{typ: token.LParen, val: "("},
		{typ: token.RParen, val: ")"},
		{typ: token.LBrace, val: "{"},
		{typ: token.Return, val: "return"},
		{typ: token.LParen, val: "("},
		{typ: token.String, val: "test"},
		{typ: token.String, val: "test2"},
		{typ: token.RParen, val: ")"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.RBrace, val: "}"},
		{typ: token.EOF},
	}

	testTable("test return", `fn test() {
	return ("test" "test2")
}`, expected, t)

	expected = []Token{
		{typ: token.Fn, val: "fn"},
		{typ: token.Ident, val: "test"},
		{typ: token.LParen, val: "("},
		{typ: token.RParen, val: ")"},
		{typ: token.LBrace, val: "{"},
		{typ: token.Return, val: "return"},
		{typ: token.Variable, val: "$PWD"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.RBrace, val: "}"},
		{typ: token.EOF},
	}

	testTable("test return", `fn test() {
	return $PWD
}`, expected, t)
}

func TestLexerFor(t *testing.T) {
	expected := []Token{
		{typ: token.For, val: "for"},
		{typ: token.LBrace, val: "{"},
		{typ: token.RBrace, val: "}"},
		{typ: token.EOF},
	}

	testTable("test inf loop", `for {}`, expected, t)

	expected = []Token{
		{typ: token.For, val: "for"},
		{typ: token.Ident, val: "f"},
		{typ: token.Ident, val: "in"},
		{typ: token.Variable, val: "$files"},
		{typ: token.LBrace, val: "{"},
		{typ: token.RBrace, val: "}"},
		{typ: token.EOF},
	}

	testTable("test inf loop", `for f in $files {}`, expected, t)

	expected = []Token{
		{typ: token.For, val: "for"},
		{typ: token.Ident, val: "f"},
		{typ: token.Ident, val: "in"},
		{typ: token.Ident, val: "getfiles"},
		{typ: token.LParen, val: "("},
		{typ: token.String, val: "/"},
		{typ: token.RParen, val: ")"},
		{typ: token.LBrace, val: "{"},
		{typ: token.RBrace, val: "}"},
		{typ: token.EOF},
	}

	testTable("test inf loop", `for f in getfiles("/") {}`, expected, t)

	expected = []Token{
		{typ: token.For, val: "for"},
		{typ: token.Ident, val: "f"},
		{typ: token.Ident, val: "in"},
		{typ: token.LParen, val: "("},
		{typ: token.Number, val: "1"},
		{typ: token.Number, val: "2"},
		{typ: token.Number, val: "3"},
		{typ: token.Number, val: "4"},
		{typ: token.Number, val: "5"},
		{typ: token.RParen, val: ")"},
		{typ: token.LBrace, val: "{"},
		{typ: token.RBrace, val: "}"},
		{typ: token.EOF},
	}

	testTable("test inf loop", `for f in (1 2 3 4 5) {}`, expected, t)
}

func TestLexerFnAsFirstClass(t *testing.T) {
	expected := []Token{
		{typ: token.Fn, val: "fn"},
		{typ: token.Ident, val: "printer"},
		{typ: token.LParen, val: "("},
		{typ: token.Ident, val: "val"},
		{typ: token.RParen, val: ")"},
		{typ: token.LBrace, val: "{"},
		{typ: token.Ident, val: "echo"},
		{typ: token.Arg, val: "-n"},
		{typ: token.Variable, val: "$val"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.RBrace, val: "}"},
		{typ: token.Fn, val: "fn"},
		{typ: token.Ident, val: "success"},
		{typ: token.LParen, val: "("},
		{typ: token.Ident, val: "print"},
		{typ: token.Comma, val: ","},
		{typ: token.Ident, val: "val"},
		{typ: token.RParen, val: ")"},
		{typ: token.LBrace, val: "{"},
		{typ: token.Variable, val: "$print"},
		{typ: token.LParen, val: "("},
		{typ: token.String, val: "[SUCCESS] "},
		{typ: token.Plus, val: "+"},
		{typ: token.Variable, val: "$val"},
		{typ: token.RParen, val: ")"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.RBrace, val: "}"},
		{typ: token.Ident, val: "success"},
		{typ: token.LParen, val: "("},
		{typ: token.Variable, val: "$printer"},
		{typ: token.Comma, val: ","},
		{typ: token.String, val: "Command executed!"},
		{typ: token.RParen, val: ")"},
		{typ: token.Semicolon, val: ";"},

		{typ: token.EOF},
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
		{typ: token.Ident, val: "cmd"},
		{typ: token.Assign, val: "="},
		{typ: token.Variable, val: "$commands"},
		{typ: token.LBrack, val: "["},
		{typ: token.Number, val: "0"},
		{typ: token.RBrack, val: "]"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.EOF},
	}

	for i := 0; i < 1000; i++ {
		expected[4] = Token{
			typ: token.Number,
			val: strconv.Itoa(i),
		}

		testTable("test variable indexing", `cmd = $commands[`+strconv.Itoa(i)+`]`, expected, t)
	}

	expected = []Token{
		{typ: token.Ident, val: "cmd"},
		{typ: token.Assign, val: "="},
		{typ: token.Variable, val: "$commands"},
		{typ: token.LBrack, val: "["},
		{typ: token.Arg, val: "a"},
		{typ: token.RBrack, val: "]"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.EOF},
	}

	testTable("test invalid number", `cmd = $commands[a]`, expected, t)

	expected = []Token{
		{typ: token.Ident, val: "cmd"},
		{typ: token.Assign, val: "="},
		{typ: token.Variable, val: "$commands"},
		{typ: token.LBrack, val: "["},
		{typ: token.RBrack, val: "]"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.EOF},
	}

	testTable("test invalid number", `cmd = $commands[]`, expected, t)

	expected = []Token{
		{typ: token.Ident, val: "echo"},
		{typ: token.Ident, val: "test"},
		{typ: token.Variable, val: "$names"},
		{typ: token.LBrack, val: "["},
		{typ: token.Number, val: "666"},
		{typ: token.RBrack, val: "]"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.EOF},
	}

	testTable("test variable index on commands", `echo test $names[666]`, expected, t)

	expected = []Token{
		{typ: token.If, val: "if"},
		{typ: token.Variable, val: "$crazies"},
		{typ: token.LBrack, val: "["},
		{typ: token.Number, val: "0"},
		{typ: token.RBrack, val: "]"},
		{typ: token.Equal, val: "=="},
		{typ: token.String, val: "patito"},
		{typ: token.LBrace, val: "{"},
		{typ: token.Ident, val: "echo"},
		{typ: token.String, val: ":D"},
		{typ: token.RBrace, val: "}"},
		{typ: token.EOF},
	}

	testTable("test if with indexing", `if $crazies[0] == "patito" { echo ":D" }`, expected, t)
}

func TestLexerMultilineCmdExecution(t *testing.T) {
	expected := []Token{
		{typ: token.LParen, val: "("},
		{typ: token.RParen, val: ")"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.EOF},
	}

	testTable("multiline", `()`, expected, t)

	expected = []Token{
		{typ: token.LParen, val: "("},
		{typ: token.Ident, val: "echo"},
		{typ: token.RParen, val: ")"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.EOF},
	}

	testTable("multiline", `(echo)`, expected, t)

	expected = []Token{
		{typ: token.LParen, val: "("},
		{typ: token.Ident, val: "echo"},
		{typ: token.Ident, val: "AAAAA"},
		{typ: token.Ident, val: "AAAAA"},
		{typ: token.Ident, val: "AAAAA"},
		{typ: token.Ident, val: "AAAAA"},
		{typ: token.Ident, val: "AAAAA"},
		{typ: token.Ident, val: "AAAAA"},
		{typ: token.Ident, val: "BBBBB"},
		{typ: token.Ident, val: "BBBBB"},
		{typ: token.RParen, val: ")"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.EOF},
	}

	testTable("test multiline cmd execution", `(echo AAAAA AAAAA
	AAAAA AAAAA
	AAAAA AAAAA
	BBBBB BBBBB)`, expected, t)
}

func TestLexerMultilineCmdAssign(t *testing.T) {
	expected := []Token{
		{typ: token.Ident, val: "some"},
		{typ: token.AssignCmd, val: "<="},
		{typ: token.LParen, val: "("},
		{typ: token.RParen, val: ")"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.EOF},
	}

	testTable("multiline", `some <= ()`, expected, t)

	expected = []Token{
		{typ: token.Ident, val: "some"},
		{typ: token.AssignCmd, val: "<="},
		{typ: token.LParen, val: "("},
		{typ: token.Ident, val: "echo"},
		{typ: token.RParen, val: ")"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.EOF},
	}

	testTable("multiline", `some <= (echo)`, expected, t)

	expected = []Token{
		{typ: token.Ident, val: "some"},
		{typ: token.AssignCmd, val: "<="},
		{typ: token.LParen, val: "("},
		{typ: token.Ident, val: "echo"},
		{typ: token.Ident, val: "AAAAA"},
		{typ: token.Ident, val: "AAAAA"},
		{typ: token.Ident, val: "AAAAA"},
		{typ: token.Ident, val: "AAAAA"},
		{typ: token.Ident, val: "AAAAA"},
		{typ: token.Ident, val: "AAAAA"},
		{typ: token.Ident, val: "BBBBB"},
		{typ: token.Ident, val: "BBBBB"},
		{typ: token.RParen, val: ")"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.EOF},
	}

	testTable("multiline", `some <= (echo AAAAA AAAAA
	AAAAA AAAAA
	AAAAA AAAAA
	BBBBB BBBBB)`, expected, t)

	testTable("multiline", `some <= (
	echo AAAAA AAAAA
	AAAAA AAAAA
	AAAAA AAAAA
	BBBBB BBBBB
)`, expected, t)
}

func TestLexerCommandDelimiter(t *testing.T) {
	expected := []Token{
		{typ: token.Ident, val: "echo"},
		{typ: token.String, val: "hello"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.Ident, val: "echo"},
		{typ: token.String, val: "world"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.EOF},
	}

	testTable("semicolons to separate commands",
		`echo "hello"; echo "world"`, expected, t)
}

func TestLexerLongAssignment(t *testing.T) {
	expected := []Token{
		{typ: token.Ident, val: "grpid"},
		{typ: token.AssignCmd, val: "<="},
		{typ: token.LParen, val: "("},
		{typ: token.Ident, val: "aws"},
		{typ: token.Ident, val: "ec2"},
		{typ: token.Arg, val: "create-security-group"},
		{typ: token.Arg, val: "--group-name"},
		{typ: token.Variable, val: "$name"},
		{typ: token.Arg, val: "--description"},
		{typ: token.Variable, val: "$desc"},
		{typ: token.Variable, val: "$vpcarg"},
		{typ: token.Pipe, val: "|"},
		{typ: token.Ident, val: "jq"},
		{typ: token.String, val: ".GroupId"},
		{typ: token.Pipe, val: "|"},
		{typ: token.Ident, val: "xargs"},
		{typ: token.Ident, val: "echo"},
		{typ: token.Arg, val: "-n"},
		{typ: token.RParen, val: ")"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.EOF},
	}

	testTable("test xxx", `grpid <= (
	aws ec2 create-security-group
				--group-name $name
				--description $desc
				$vpcarg |
	jq ".GroupId" |
	xargs echo -n)`, expected, t)
}
