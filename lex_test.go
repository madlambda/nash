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

func TestLexerShowEnv(t *testing.T) {
	expected := []item{
		item{
			typ: itemShowEnv,
			val: "showenv",
		},
		item{
			typ: itemEOF,
		},
	}

	testTable("testShowEnv", `showenv`, expected, t)

	expected = []item{
		item{
			typ: itemShowEnv,
			val: "showenv",
		},
		item{
			typ: itemError,
			val: "Unexpected character 'a' at pos 9. Showenv doesn't have arguments.",
		},
		item{
			typ: itemEOF,
		},
	}

	testTable("testShowEnv", `showenv a`, expected, t)
}

func TestLexerSimpleSetAssignment(t *testing.T) {
	expected := []item{
		item{
			typ: itemSetEnv,
			val: "setenv",
		},
		item{
			typ: itemVarName,
			val: "name",
		},
		item{
			typ: itemEOF,
		},
	}

	testTable("testSet", `setenv name`, expected, t)
}

func TestLexerSimpleAssignment(t *testing.T) {
	expected := []item{
		item{
			typ: itemVarName,
			val: "test",
		},
		item{
			typ: itemAssign,
			val: "=",
		},
		item{
			typ: itemString,
			val: "value",
		},
		item{
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
		item{
			typ: itemVarName,
			val: "test",
		},
		item{
			typ: itemAssign,
			val: "=",
		},
		item{
			typ: itemString,
			val: "value",
		},
		item{
			typ: itemVarName,
			val: "other",
		},
		item{
			typ: itemAssign,
			val: "=",
		},
		item{
			typ: itemString,
			val: "other",
		},
		item{
			typ: itemEOF,
		},
	}

	testTable("test multiple Assignments", `
        test="value"
        other="other"`, expected, t)

	expected = []item{
		item{
			typ: itemVarName,
			val: "test",
		},
		item{
			typ: itemAssign,
			val: "=",
		},
		item{
			typ: itemString,
			val: "value",
		},
		item{
			typ: itemVarName,
			val: "other",
		},
		item{
			typ: itemAssign,
			val: "=",
		},
		item{
			typ: itemVariable,
			val: "$test",
		},
		item{
			typ: itemCommand,
			val: "echo",
		},
		item{
			typ: itemVariable,
			val: "$other",
		},
		item{
			typ: itemEOF,
		},
	}

	testTable("test multiple Assignments", `
        test="value"
        other=$test
        echo $other`, expected, t)

	expected = []item{
		item{
			typ: itemVarName,
			val: "STALI_SRC",
		},
		item{
			typ: itemAssign,
			val: "=",
		},
		item{
			typ: itemVariable,
			val: "$PWD",
		},
		item{
			typ: itemConcat,
			val: "+",
		},
		item{
			typ: itemString,
			val: "/src",
		},
		item{
			typ: itemEOF,
		},
	}

	testTable("test underscore", `STALI_SRC=$PWD + "/src"`, expected, t)

	expected = []item{
		item{
			typ: itemVarName,
			val: "PROMPT",
		},
		item{
			typ: itemAssign,
			val: "=",
		},
		item{
			typ: itemString,
			val: "(",
		},
		item{
			typ: itemConcat,
			val: "+",
		},
		item{
			typ: itemVariable,
			val: "$path",
		},
		item{
			typ: itemConcat,
			val: "+",
		},
		item{
			typ: itemString,
			val: ")",
		},
		item{
			typ: itemConcat,
			val: "+",
		},
		item{
			typ: itemVariable,
			val: "$PROMPT",
		},
		item{
			typ: itemEOF,
		},
	}

	testTable("test concat with parenthesis", `PROMPT="("+$path+")"+$PROMPT`, expected, t)

}

func TestLexerListAssignment(t *testing.T) {
	expected := []item{
		item{
			typ: itemVarName,
			val: "test",
		},
		item{
			typ: itemAssign,
			val: "=",
		},
		item{
			typ: itemListOpen,
			val: "(",
		},
		item{
			typ: itemArg,
			val: "plan9",
		},
		item{
			typ: itemArg,
			val: "from",
		},
		item{
			typ: itemArg,
			val: "bell",
		},
		item{
			typ: itemArg,
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
	testTable("testListAssignment multiline", `test = (
	plan9
	from
	bell
	labs
)`, expected, t)

	expected = []item{
		item{
			typ: itemVarName,
			val: "test",
		},
		item{
			typ: itemAssign,
			val: "=",
		},
		item{
			typ: itemListOpen,
			val: "(",
		},
		item{
			typ: itemString,
			val: "plan9",
		},
		item{
			typ: itemArg,
			val: "from",
		},
		item{
			typ: itemString,
			val: "bell",
		},
		item{
			typ: itemArg,
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

	testTable("testListAssignment mixed args", `test=( "plan9" from "bell" labs )`, expected, t)
	testTable("testListAssignment mixed args", `test=("plan9" from "bell" labs)`, expected, t)
	testTable("testListAssignment mixed args", `test = (
        "plan9"
        from
        "bell"
        labs
)`, expected, t)

	expected = []item{
		item{
			typ: itemVarName,
			val: "test",
		},
		item{
			typ: itemAssign,
			val: "=",
		},
		item{
			typ: itemListOpen,
			val: "(",
		},
		item{
			typ: itemVariable,
			val: "$plan9",
		},
		item{
			typ: itemArg,
			val: "from",
		},
		item{
			typ: itemVariable,
			val: "$bell",
		},
		item{
			typ: itemArg,
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
		item{
			typ: itemVarName,
			val: "test",
		},
		item{
			typ: itemAssign,
			val: "=",
		},
		item{
			typ: itemString,
			val: "value",
		},
		item{
			typ: itemError,
			val: "Invalid assignment. Expected '+' or EOL, but found 'o' at pos '13'",
		},
		item{
			typ: itemEOF,
		},
	}

	testTable("testInvalidAssignments", `test="value" other`, expected, t)
}

func TestLexerSimpleCommand(t *testing.T) {
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

	expected = []item{
		item{
			typ: itemCommand,
			val: "echo",
		},
		item{
			typ: itemArg,
			val: "rootfs-x86_64",
		},
		item{
			typ: itemEOF,
		},
	}

	testTable("testSimpleCommand", `echo rootfs-x86_64`, expected, t)

	expected = []item{
		item{
			typ: itemCommand,
			val: "git",
		},
		item{
			typ: itemArg,
			val: "clone",
		},
		item{
			typ: itemArg,
			val: "--depth=1",
		},
		item{
			typ: itemArg,
			val: "http://git.sta.li/toolchain",
		},
		item{
			typ: itemEOF,
		},
	}

	testTable("testSimpleCommand", `git clone --depth=1 http://git.sta.li/toolchain`, expected, t)

	expected = []item{
		item{
			typ: itemCommand,
			val: "ls",
		},
		item{
			typ: itemVariable,
			val: "$GOPATH",
		},
		item{
			typ: itemEOF,
		},
	}

	testTable("testSimpleCommand", `ls $GOPATH`, expected, t)

	expected = []item{
		item{
			typ: itemCommand,
			val: "ls",
		},
		item{
			typ: itemVariable,
			val: "$GOPATH",
		},
		item{
			typ: itemConcat,
			val: "+",
		},
		item{
			typ: itemString,
			val: "/src/github.com",
		},
		item{
			typ: itemEOF,
		},
	}

	testTable("testSimpleCommand", `ls $GOPATH+"/src/github.com"`, expected, t)

	expected = []item{
		item{
			typ: itemCommand,
			val: "ls",
		},
		item{
			typ: itemString,
			val: "/src/github.com",
		},
		item{
			typ: itemConcat,
			val: "+",
		},
		item{
			typ: itemVariable,
			val: "$GOPATH",
		},
		item{
			typ: itemEOF,
		},
	}

	testTable("testSimpleCommand", `ls "/src/github.com"+$GOPATH`, expected, t)

	expected = []item{
		item{
			typ: itemCommand,
			val: "ls",
		},
		item{
			typ: itemString,
			val: "/home/user",
		},
		item{
			typ: itemConcat,
			val: "+",
		},
		item{
			typ: itemString,
			val: "/.gvm/pkgsets/global/src",
		},
		item{
			typ: itemEOF,
		},
	}

	testTable("testSimpleCommand", `ls "/home/user" + "/.gvm/pkgsets/global/src"`, expected, t)

}

func TestLexerPipe(t *testing.T) {
	expected := []item{
		item{
			typ: itemCommand,
			val: "ls",
		},
		item{
			typ: itemPipe,
			val: "|",
		},
		item{
			typ: itemCommand,
			val: "wc",
		},
		item{
			typ: itemArg,
			val: "-l",
		},
		item{
			typ: itemEOF,
		},
	}

	testTable("testPipe", `ls | wc -l`, expected, t)

	expected = []item{
		item{
			typ: itemCommand,
			val: "ls",
		},
		item{
			typ: itemArg,
			val: "-l",
		},
		item{
			typ: itemPipe,
			val: "|",
		},
		item{
			typ: itemCommand,
			val: "wc",
		},
		item{
			typ: itemPipe,
			val: "|",
		},
		item{
			typ: itemCommand,
			val: "awk",
		},
		item{
			typ: itemString,
			val: "{print $1}",
		},
		item{
			typ: itemEOF,
		},
	}

	testTable("testPipe", `ls -l | wc | awk "{print $1}"`, expected, t)

	expected = []item{
		item{
			typ: itemCommand,
			val: "go",
		},
		item{
			typ: itemArg,
			val: "tool",
		},
		item{
			typ: itemArg,
			val: "vet",
		},
		item{
			typ: itemArg,
			val: "-h",
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
			typ: itemPipe,
			val: "|",
		},
		item{
			typ: itemCommand,
			val: "grep",
		},
		item{
			typ: itemArg,
			val: "log",
		},
		item{
			typ: itemEOF,
		},
	}

	testTable("testPipe with redirection", `go tool vet -h >[2=1] | grep log`, expected, t)

	expected = []item{
		item{
			typ: itemCommand,
			val: "go",
		},
		item{
			typ: itemArg,
			val: "tool",
		},
		item{
			typ: itemArg,
			val: "vet",
		},
		item{
			typ: itemArg,
			val: "-h",
		},
		item{
			typ: itemRedirRight,
			val: ">",
		},
		item{
			typ: itemArg,
			val: "out.log",
		},
		item{
			typ: itemPipe,
			val: "|",
		},
		item{
			typ: itemCommand,
			val: "grep",
		},
		item{
			typ: itemArg,
			val: "log",
		},
		item{
			typ: itemEOF,
		},
	}

	testTable("testPipe with redirection", `go tool vet -h > out.log | grep log`, expected, t)
}

func TestLexerUnquoteArg(t *testing.T) {
	expected := []item{
		item{
			typ: itemCommand,
			val: "echo",
		},
		item{
			typ: itemArg,
			val: "hello",
		},
		item{
			typ: itemEOF,
		},
	}

	testTable("testSimpleCommand", `echo hello`, expected, t)

	expected = []item{
		item{
			typ: itemCommand,
			val: "echo",
		},
		item{
			typ: itemArg,
			val: "hello-world",
		},
		item{
			typ: itemEOF,
		},
	}

	testTable("testSimpleCommand", `echo hello-world`, expected, t)

	expected = []item{
		item{
			typ: itemCommand,
			val: "echo",
		},
		item{
			typ: itemComment,
			val: "#hello-world",
		},
		item{
			typ: itemEOF,
		},
	}

	testTable("testSimpleCommand", `echo #hello-world`, expected, t)
}

func TestLexerPathCommand(t *testing.T) {
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

func TestLexerInvalidBlock(t *testing.T) {
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

func TestLexerQuotedStringNotFinished(t *testing.T) {
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

func TestLexerRfork(t *testing.T) {
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

func TestLexerRforkInvalidArguments(t *testing.T) {
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

func TestLexerSomethingIdontcareanymore(t *testing.T) {
	// maybe oneliner rfork isnt a good idea
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

func TestLexerBuiltinCd(t *testing.T) {
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

	expected = []item{
		item{
			typ: itemCd,
			val: "cd",
		},
		item{
			typ: itemEOF,
		},
	}

	testTable("testBuiltincd home", `cd`, expected, t)

	expected = []item{
		item{
			typ: itemVarName,
			val: "HOME",
		},
		item{
			typ: itemAssign,
			val: "=",
		},
		item{
			typ: itemString,
			val: "/",
		},
		item{
			typ: itemSetEnv,
			val: "setenv",
		},
		item{
			typ: itemVarName,
			val: "HOME",
		},
		item{
			typ: itemCd,
			val: "cd",
		},
		item{
			typ: itemCommand,
			val: "pwd",
		},
		item{
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
		item{
			typ: itemCd,
			val: "cd",
		},
		item{
			typ: itemVariable,
			val: "$GOPATH",
		},
		item{
			typ: itemEOF,
		},
	}

	testTable("test builtin cd into variable", `cd $GOPATH`, expected, t)

	expected = []item{
		item{
			typ: itemCd,
			val: "cd",
		},
		item{
			typ: itemVariable,
			val: "$GOPATH",
		},
		item{
			typ: itemConcat,
			val: "+",
		},
		item{
			typ: itemString,
			val: "/src/github.com",
		},
		item{
			typ: itemEOF,
		},
	}

	testTable("test cd with concat", `cd $GOPATH+"/src/github.com"`, expected, t)
}

func TestLexerMinusAlone(t *testing.T) {
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

func TestLexerRedirectSimple(t *testing.T) {
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
			typ: itemArg,
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
			typ: itemString,
			val: "tcp://localhost:8888",
		},
		item{
			typ: itemEOF,
		},
	}

	testTable("test simple redirect", `cmd > "tcp://localhost:8888"`, expected, t)

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
			typ: itemString,
			val: "udp://localhost:8888",
		},
		item{
			typ: itemEOF,
		},
	}

	testTable("test simple redirect", `cmd > "udp://localhost:8888"`, expected, t)

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
			typ: itemString,
			val: "unix:///tmp/sock.txt",
		},
		item{
			typ: itemEOF,
		},
	}

	testTable("test simple redirect", `cmd > "unix:///tmp/sock.txt"`, expected, t)

}

func TestLexerRedirectMap(t *testing.T) {

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

func TestLexerRedirectMapToLocation(t *testing.T) {
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
			typ: itemArg,
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
			typ: itemArg,
			val: "file.out",
		},
		item{
			typ: itemEOF,
		},
	}

	testTable("test stderr=stdout", "cmd >[2=1] file.out", expected, t)

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
			typ: itemRedirRBracket,
			val: "]",
		},
		item{
			typ: itemArg,
			val: "/var/log/service.log",
		},
		item{
			typ: itemEOF,
		},
	}

	testTable("test suppress stderr", `cmd >[2] /var/log/service.log`, expected, t)
}

func TestLexerRedirectMultipleMaps(t *testing.T) {
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
			val: "1",
		},
		item{
			typ: itemRedirRBracket,
			val: "]",
		},
		item{
			typ: itemArg,
			val: "file.out",
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
			typ: itemRedirRBracket,
			val: "]",
		},
		item{
			typ: itemArg,
			val: "file.err",
		},
		item{
			typ: itemEOF,
		},
	}

	testTable("test suppress stderr", `cmd >[1] file.out >[2] file.err`, expected, t)

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
			typ: itemRedirRight,
			val: ">",
		},
		item{
			typ: itemRedirLBracket,
			val: "[",
		},
		item{
			typ: itemRedirMapLSide,
			val: "1",
		},
		item{
			typ: itemRedirRBracket,
			val: "]",
		},
		item{
			typ: itemArg,
			val: "/var/log/service.log",
		},
		item{
			typ: itemEOF,
		},
	}

	testTable("test suppress stderr", `cmd >[2=1] >[1] /var/log/service.log`, expected, t)

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
			val: "1",
		},
		item{
			typ: itemRedirMapEqual,
			val: "=",
		},
		item{
			typ: itemRedirMapRSide,
			val: "2",
		},
		item{
			typ: itemRedirRBracket,
			val: "]",
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

	testTable("test suppress stderr", `cmd >[1=2] >[2=]`, expected, t)
}

func TestLexerImport(t *testing.T) {
	expected := []item{
		item{
			typ: itemImport,
			val: "import",
		},
		item{
			typ: itemArg,
			val: "env.sh",
		},
		item{
			typ: itemEOF,
		},
	}

	testTable("test import", `import env.sh`, expected, t)
}

func TestLexerSimpleIf(t *testing.T) {
	expected := []item{
		item{
			typ: itemIf,
			val: "if",
		},
		item{
			typ: itemString,
			val: "test",
		},
		item{
			typ: itemComparison,
			val: "==",
		},
		item{
			typ: itemString,
			val: "other",
		},
		item{
			typ: itemLeftBlock,
			val: "{",
		},
		item{
			typ: itemCommand,
			val: "rm",
		},
		item{
			typ: itemArg,
			val: "-rf",
		},
		item{
			typ: itemArg,
			val: "/",
		},
		item{
			typ: itemRightBlock,
			val: "}",
		},
		item{
			typ: itemEOF,
		},
	}

	testTable("test simple if", `if "test" == "other" { rm -rf / }`, expected, t)

	expected = []item{
		item{
			typ: itemIf,
			val: "if",
		},
		item{
			typ: itemString,
			val: "test",
		},
		item{
			typ: itemComparison,
			val: "!=",
		},
		item{
			typ: itemVariable,
			val: "$test",
		},
		item{
			typ: itemLeftBlock,
			val: "{",
		},
		item{
			typ: itemCommand,
			val: "rm",
		},
		item{
			typ: itemArg,
			val: "-rf",
		},
		item{
			typ: itemArg,
			val: "/",
		},
		item{
			typ: itemRightBlock,
			val: "}",
		},
		item{
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

	/*	testTableFail("test simple if", `

		if "test" != $test
		{
		    rm -rf /
		}`, expected, t) */
}

func TestLexerIfElse(t *testing.T) {
	expected := []item{
		item{
			typ: itemIf,
			val: "if",
		},
		item{
			typ: itemString,
			val: "test",
		},
		item{
			typ: itemComparison,
			val: "==",
		},
		item{
			typ: itemString,
			val: "other",
		},
		item{
			typ: itemLeftBlock,
			val: "{",
		},
		item{
			typ: itemCommand,
			val: "rm",
		},
		item{
			typ: itemArg,
			val: "-rf",
		},
		item{
			typ: itemArg,
			val: "/",
		},
		item{
			typ: itemRightBlock,
			val: "}",
		},
		item{
			typ: itemElse,
			val: "else",
		},
		item{
			typ: itemLeftBlock,
			val: "{",
		},
		item{
			typ: itemCommand,
			val: "pwd",
		},
		item{
			typ: itemRightBlock,
			val: "}",
		},
		item{
			typ: itemEOF,
		},
	}

	testTable("test simple if", `if "test" == "other" { rm -rf / } else { pwd }`, expected, t)
}

func TestLexerIfElseIf(t *testing.T) {
	expected := []item{
		item{
			typ: itemIf,
			val: "if",
		},
		item{
			typ: itemString,
			val: "test",
		},
		item{
			typ: itemComparison,
			val: "==",
		},
		item{
			typ: itemString,
			val: "other",
		},
		item{
			typ: itemLeftBlock,
			val: "{",
		},
		item{
			typ: itemCommand,
			val: "rm",
		},
		item{
			typ: itemArg,
			val: "-rf",
		},
		item{
			typ: itemArg,
			val: "/",
		},
		item{
			typ: itemRightBlock,
			val: "}",
		},
		item{
			typ: itemElse,
			val: "else",
		},
		item{
			typ: itemIf,
			val: "if",
		},
		item{
			typ: itemString,
			val: "test",
		},
		item{
			typ: itemComparison,
			val: "==",
		},
		item{
			typ: itemVariable,
			val: "$var",
		},
		item{
			typ: itemLeftBlock,
			val: "{",
		},
		item{
			typ: itemCommand,
			val: "pwd",
		},
		item{
			typ: itemRightBlock,
			val: "}",
		},
		item{
			typ: itemElse,
			val: "else",
		},
		item{
			typ: itemLeftBlock,
			val: "{",
		},
		item{
			typ: itemCommand,
			val: "exit",
		},
		item{
			typ: itemArg,
			val: "1",
		},
		item{
			typ: itemRightBlock,
			val: "}",
		},
		item{
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
		item{
			typ: itemFnDecl,
			val: "fn",
		},
		item{
			typ: itemVarName,
			val: "build",
		},
		item{
			typ: itemLeftParen,
			val: "(",
		},
		item{
			typ: itemRightParen,
			val: ")",
		},
		item{
			typ: itemLeftBlock,
			val: "{",
		},
		item{
			typ: itemRightBlock,
			val: "}",
		},
		item{
			typ: itemEOF,
		},
	}

	testTable("test empty fn", `fn build() {}`, expected, t)

	// lambda
	expected = []item{
		item{
			typ: itemFnDecl,
			val: "fn",
		},
		item{
			typ: itemVarName,
			val: "",
		},
		item{
			typ: itemLeftParen,
			val: "(",
		},
		item{
			typ: itemRightParen,
			val: ")",
		},
		item{
			typ: itemLeftBlock,
			val: "{",
		},
		item{
			typ: itemRightBlock,
			val: "}",
		},
		item{
			typ: itemEOF,
		},
	}

	testTable("test empty fn", `fn () {}`, expected, t)

	// IIFE
	expected = []item{
		item{
			typ: itemFnDecl,
			val: "fn",
		},
		item{
			typ: itemVarName,
			val: "",
		},
		item{
			typ: itemLeftParen,
			val: "(",
		},
		item{
			typ: itemRightParen,
			val: ")",
		},
		item{
			typ: itemLeftBlock,
			val: "{",
		},
		item{
			typ: itemRightBlock,
			val: "}",
		},
		item{
			typ: itemLeftParen,
			val: "(",
		},
		item{
			typ: itemRightParen,
			val: ")",
		},
		item{
			typ: itemEOF,
		},
	}

	testTable("test empty fn", `fn () {}()`, expected, t)

	expected = []item{
		item{
			typ: itemFnDecl,
			val: "fn",
		},
		item{
			typ: itemVarName,
			val: "build",
		},
		item{
			typ: itemLeftParen,
			val: "(",
		},
		item{
			typ: itemVarName,
			val: "image",
		},
		item{
			typ: itemVarName,
			val: "debug",
		},
		item{
			typ: itemRightParen,
			val: ")",
		},
		item{
			typ: itemLeftBlock,
			val: "{",
		},
		item{
			typ: itemRightBlock,
			val: "}",
		},
		item{
			typ: itemEOF,
		},
	}

	testTable("test empty fn with args", `fn build(image, debug) {}`, expected, t)

	expected = []item{
		item{
			typ: itemFnDecl,
			val: "fn",
		},
		item{
			typ: itemVarName,
			val: "build",
		},
		item{
			typ: itemLeftParen,
			val: "(",
		},
		item{
			typ: itemVarName,
			val: "image",
		},
		item{
			typ: itemVarName,
			val: "debug",
		},
		item{
			typ: itemRightParen,
			val: ")",
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
			typ: itemCommand,
			val: "tar",
		},
		item{
			typ: itemRightBlock,
			val: "}",
		},
		item{
			typ: itemEOF,
		},
	}

	testTable("test empty fn with args and body", `fn build(image, debug) {
            ls
            tar
        }`, expected, t)

	expected = []item{
		item{
			typ: itemFnDecl,
			val: "fn",
		},
		item{
			typ: itemVarName,
			val: "cd",
		},
		item{
			typ: itemLeftParen,
			val: "(",
		},
		item{
			typ: itemVarName,
			val: "path",
		},
		item{
			typ: itemRightParen,
			val: ")",
		},
		item{
			typ: itemLeftBlock,
			val: "{",
		},
		item{
			typ: itemCd,
			val: "cd",
		},
		item{
			typ: itemVariable,
			val: "$path",
		},
		item{
			typ: itemVarName,
			val: "PROMPT",
		},
		item{
			typ: itemAssign,
			val: "=",
		},
		item{
			typ: itemString,
			val: "(",
		},
		item{
			typ: itemConcat,
			val: "+",
		},
		item{
			typ: itemVariable,
			val: "$path",
		},
		item{
			typ: itemConcat,
			val: "+",
		},
		item{
			typ: itemString,
			val: ")",
		},
		item{
			typ: itemConcat,
			val: "+",
		},
		item{
			typ: itemVariable,
			val: "$PROMPT",
		},
		item{
			typ: itemSetEnv,
			val: "setenv",
		},
		item{
			typ: itemVarName,
			val: "PROMPT",
		},
		item{
			typ: itemRightBlock,
			val: "}",
		},
		item{
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
		item{
			typ: itemFnInv,
			val: "build",
		},
		item{
			typ: itemLeftParen,
			val: "(",
		},
		item{
			typ: itemRightParen,
			val: ")",
		},
		item{
			typ: itemEOF,
		},
	}

	testTable("test fn invocation", `build()`, expected, t)

	expected = []item{
		item{
			typ: itemFnInv,
			val: "build",
		},
		item{
			typ: itemLeftParen,
			val: "(",
		},
		item{
			typ: itemString,
			val: "ubuntu",
		},

		item{
			typ: itemRightParen,
			val: ")",
		},
		item{
			typ: itemEOF,
		},
	}

	testTable("test fn invocation", `build("ubuntu")`, expected, t)

	expected = []item{
		item{
			typ: itemFnInv,
			val: "build",
		},
		item{
			typ: itemLeftParen,
			val: "(",
		},
		item{
			typ: itemString,
			val: "ubuntu",
		},
		item{
			typ: itemVariable,
			val: "$debug",
		},

		item{
			typ: itemRightParen,
			val: ")",
		},
		item{
			typ: itemEOF,
		},
	}

	testTable("test fn invocation", `build("ubuntu", $debug)`, expected, t)

	expected = []item{
		item{
			typ: itemFnInv,
			val: "build",
		},
		item{
			typ: itemLeftParen,
			val: "(",
		},
		item{
			typ: itemVariable,
			val: "$debug",
		},

		item{
			typ: itemRightParen,
			val: ")",
		},
		item{
			typ: itemEOF,
		},
	}

	testTable("test fn invocation", `build($debug)`, expected, t)
}

func TestLexerAssignCmdOut(t *testing.T) {
	expected := []item{
		item{
			typ: itemVarName,
			val: "ipaddr",
		},
		item{
			typ: itemAssignCmd,
			val: "<=",
		},
		item{
			typ: itemCommand,
			val: "someprogram",
		},
		item{
			typ: itemEOF,
		},
	}

	testTable("test assignCmdOut", `ipaddr <= someprogram`, expected, t)
}

func TestBindFn(t *testing.T) {
	expected := []item{
		item{
			typ: itemBindFn,
			val: "bindfn",
		},
		item{
			typ: itemVarName,
			val: "cd",
		},
		item{
			typ: itemVarName,
			val: "cd2",
		},
		item{
			typ: itemEOF,
		},
	}

	testTable("test bindfn", `bindfn cd cd2`, expected, t)

}

func TestIssue19CommandAssignment(t *testing.T) {
	line := `version = "4.5.6"
canonName <= echo -n $version | sed "s/\\.//g"`

	expected := []item{
		item{
			typ: itemVarName,
			val: "version",
		},
		item{
			typ: itemAssign,
			val: "=",
		},
		item{
			typ: itemString,
			val: "4.5.6",
		},
		item{
			typ: itemVarName,
			val: "canonName",
		},
		item{
			typ: itemAssignCmd,
			val: "<=",
		},
		item{
			typ: itemCommand,
			val: "echo",
		},
		item{
			typ: itemArg,
			val: "-n",
		},
		item{
			typ: itemVariable,
			val: "$version",
		},
		item{
			typ: itemPipe,
			val: "|",
		},
		item{
			typ: itemCommand,
			val: "sed",
		},
		item{
			typ: itemString,
			val: "s/\\.//g",
		},
		item{
			typ: itemEOF,
		},
	}

	testTable("test issue 19", line, expected, t)
}

func TestLexerRedirectionNetwork(t *testing.T) {
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
			typ: itemRedirRight,
			val: ">",
		},
		item{
			typ: itemRedirLBracket,
			val: "[",
		},
		item{
			typ: itemRedirMapLSide,
			val: "1",
		},
		item{
			typ: itemRedirRBracket,
			val: "]",
		},
		item{
			typ: itemString,
			val: "tcp://localhost:6667",
		},
		item{
			typ: itemEOF,
		},
	}

	testTable("test redirection network", `echo "hello world" >[1] "tcp://localhost:6667"`, expected, t)
}

func TestRedirectionIssue34(t *testing.T) {
	expected := []item{
		item{
			typ: itemCommand,
			val: "cat",
		},
		item{
			typ: itemArg,
			val: "/etc/passwd",
		},
		item{
			typ: itemRedirRight,
			val: ">",
		},
		item{
			typ: itemArg,
			val: "/dev/null",
		},
		item{
			typ: itemError,
			val: "Expected end of line or redirection, but found 'e'",
		},
		item{
			typ: itemEOF,
		},
	}

	testTable("test issue #34", `cat /etc/passwd > /dev/null echo "hello world"`, expected, t)
}

func TestLexerIssue21RedirectionWithVariables(t *testing.T) {
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
			typ: itemVariable,
			val: "$outFname",
		},
		item{
			typ: itemEOF,
		},
	}

	testTable("test redirection variable", `cmd > $outFname`, expected, t)
}

func TestLexerIssue22ConcatenationCd(t *testing.T) {
	expected := []item{
		item{
			typ: itemFnDecl,
			val: "fn",
		},
		item{
			typ: itemVarName,
			val: "gocd",
		},
		item{
			typ: itemLeftParen,
			val: "(",
		},
		item{
			typ: itemVarName,
			val: "path",
		},
		item{
			typ: itemRightParen,
			val: ")",
		},
		item{
			typ: itemLeftBlock,
			val: "{",
		},
		item{
			typ: itemIf,
			val: "if",
		},
		item{
			typ: itemVariable,
			val: "$path",
		},
		item{
			typ: itemComparison,
			val: "==",
		},
		item{
			typ: itemString,
			val: "",
		},
		item{
			typ: itemLeftBlock,
			val: "{",
		},
		item{
			typ: itemCd,
			val: "cd",
		},
		item{
			typ: itemVariable,
			val: "$GOPATH",
		},
		item{
			typ: itemRightBlock,
			val: "}",
		},
		item{
			typ: itemElse,
			val: "else",
		},
		item{
			typ: itemLeftBlock,
			val: "{",
		},
		item{
			typ: itemCd,
			val: "cd",
		},
		item{
			typ: itemVariable,
			val: "$GOPATH",
		},
		item{
			typ: itemConcat,
			val: "+",
		},
		item{
			typ: itemString,
			val: "/src/",
		},
		item{
			typ: itemConcat,
			val: "+",
		},
		item{
			typ: itemVariable,
			val: "$path",
		},
		item{
			typ: itemRightBlock,
			val: "}",
		},
		item{
			typ: itemRightBlock,
			val: "}",
		},
		item{
			typ: itemEOF,
		},
	}

	testTable("test issue 22", `fn gocd(path) {
    if $path == "" {
        cd $GOPATH
    } else {
        cd $GOPATH + "/src/" + $path
    }
}`, expected, t)
}

func TestLexerDump(t *testing.T) {
	expected := []item{
		item{
			typ: itemDump,
			val: "dump",
		},
		item{
			typ: itemEOF,
		},
	}

	testTable("test dump", `dump`, expected, t)

	expected = []item{
		item{
			typ: itemDump,
			val: "dump",
		},
		item{
			typ: itemArg,
			val: "out",
		},
		item{
			typ: itemEOF,
		},
	}

	testTable("test dump", `dump out`, expected, t)
}

func TestLexerReturn(t *testing.T) {
	expected := []item{
		item{
			typ: itemReturn,
			val: "return",
		},
		item{
			typ: itemEOF,
		},
	}

	testTable("test return", "return", expected, t)

	expected = []item{
		item{
			typ: itemFnDecl,
			val: "fn",
		},
		item{
			typ: itemVarName,
			val: "test",
		},
		item{
			typ: itemLeftParen,
			val: "(",
		},
		item{
			typ: itemRightParen,
			val: ")",
		},
		item{
			typ: itemLeftBlock,
			val: "{",
		},
		item{
			typ: itemReturn,
			val: "return",
		},
		item{
			typ: itemRightBlock,
			val: "}",
		},
		item{
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
		item{
			typ: itemFnDecl,
			val: "fn",
		},
		item{
			typ: itemVarName,
			val: "test",
		},
		item{
			typ: itemLeftParen,
			val: "(",
		},
		item{
			typ: itemRightParen,
			val: ")",
		},
		item{
			typ: itemLeftBlock,
			val: "{",
		},
		item{
			typ: itemReturn,
			val: "return",
		},
		item{
			typ: itemString,
			val: "some value",
		},
		item{
			typ: itemRightBlock,
			val: "}",
		},
		item{
			typ: itemEOF,
		},
	}

	testTable("test return", `fn test() { return "some value"}`, expected, t)
	testTable("test return", `fn test() {
	return "some value"
}`, expected, t)

	expected = []item{
		item{
			typ: itemFnDecl,
			val: "fn",
		},
		item{
			typ: itemVarName,
			val: "test",
		},
		item{
			typ: itemLeftParen,
			val: "(",
		},
		item{
			typ: itemRightParen,
			val: ")",
		},
		item{
			typ: itemLeftBlock,
			val: "{",
		},
		item{
			typ: itemVarName,
			val: "value",
		},
		item{
			typ: itemAssign,
			val: "=",
		},
		item{
			typ: itemString,
			val: "some value",
		},
		item{
			typ: itemReturn,
			val: "return",
		},
		item{
			typ: itemVariable,
			val: "$value",
		},
		item{
			typ: itemRightBlock,
			val: "}",
		},
		item{
			typ: itemEOF,
		},
	}

	testTable("test return", `fn test() {
	value = "some value"
	return $value
}`, expected, t)

	expected = []item{
		item{
			typ: itemFnDecl,
			val: "fn",
		},
		item{
			typ: itemVarName,
			val: "test",
		},
		item{
			typ: itemLeftParen,
			val: "(",
		},
		item{
			typ: itemRightParen,
			val: ")",
		},
		item{
			typ: itemLeftBlock,
			val: "{",
		},
		item{
			typ: itemReturn,
			val: "return",
		},
		item{
			typ: itemListOpen,
			val: "(",
		},
		item{
			typ: itemString,
			val: "test",
		},
		item{
			typ: itemString,
			val: "test2",
		},
		item{
			typ: itemListClose,
			val: ")",
		},
		item{
			typ: itemRightBlock,
			val: "}",
		},
		item{
			typ: itemEOF,
		},
	}

	testTable("test return", `fn test() {
	return ("test" "test2")
}`, expected, t)

	expected = []item{
		item{
			typ: itemFnDecl,
			val: "fn",
		},
		item{
			typ: itemVarName,
			val: "test",
		},
		item{
			typ: itemLeftParen,
			val: "(",
		},
		item{
			typ: itemRightParen,
			val: ")",
		},
		item{
			typ: itemLeftBlock,
			val: "{",
		},
		item{
			typ: itemReturn,
			val: "return",
		},
		item{
			typ: itemVariable,
			val: "$PWD",
		},
		item{
			typ: itemRightBlock,
			val: "}",
		},
		item{
			typ: itemEOF,
		},
	}

	testTable("test return", `fn test() {
	return $PWD
}`, expected, t)
}
