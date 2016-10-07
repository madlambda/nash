package scanner

import (
	"testing"

	"github.com/NeowayLabs/nash/token"
)

func TestLexerIssue34(t *testing.T) {
	expected := []Token{
		{typ: token.Ident, val: "cat"},
		{typ: token.Arg, val: "/etc/passwd"},
		{typ: token.Gt, val: ">"},
		{typ: token.Arg, val: "/dev/null"},
		{typ: token.Ident, val: "echo"},
		{typ: token.String, val: "hello world"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.EOF},
	}

	testTable("test-issue-34", `cat /etc/passwd > /dev/null echo "hello world"`, expected, t)
}

func TestLexerIssue21(t *testing.T) {
	expected := []Token{
		{typ: token.Ident, val: "cmd"},
		{typ: token.Gt, val: ">"},
		{typ: token.Variable, val: "$outFname"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.EOF},
	}

	testTable("test redirection variable", `cmd > $outFname`, expected, t)
}

func TestLexerIssue22(t *testing.T) {
	expected := []Token{
		{typ: token.Fn, val: "fn"},
		{typ: token.Ident, val: "gocd"},
		{typ: token.LParen, val: "("},
		{typ: token.Ident, val: "path"},
		{typ: token.RParen, val: ")"},
		{typ: token.LBrace, val: "{"},
		{typ: token.If, val: "if"},
		{typ: token.Variable, val: "$path"},
		{typ: token.Equal, val: "=="},
		{typ: token.String, val: ""},
		{typ: token.LBrace, val: "{"},
		{typ: token.Ident, val: "cd"},
		{typ: token.Variable, val: "$GOPATH"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.RBrace, val: "}"},
		{typ: token.Else, val: "else"},
		{typ: token.LBrace, val: "{"},
		{typ: token.Ident, val: "cd"},
		{typ: token.Variable, val: "$GOPATH"},
		{typ: token.Plus, val: "+"},
		{typ: token.String, val: "/src/"},
		{typ: token.Plus, val: "+"},
		{typ: token.Variable, val: "$path"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.RBrace, val: "}"},
		{typ: token.RBrace, val: "}"},
		{typ: token.EOF},
	}

	testTable("test issue 22", `fn gocd(path) {
    if $path == "" {
        cd $GOPATH
    } else {
        cd $GOPATH + "/src/" + $path
    }
}`, expected, t)
}

func TestLexerIssue19(t *testing.T) {
	line := `version = "4.5.6"
canonName <= echo -n $version | sed "s/\\.//g"`

	expected := []Token{
		{typ: token.Ident, val: "version"},
		{typ: token.Assign, val: "="},
		{typ: token.String, val: "4.5.6"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.Ident, val: "canonName"},
		{typ: token.AssignCmd, val: "<="},
		{typ: token.Ident, val: "echo"},
		{typ: token.Arg, val: "-n"},
		{typ: token.Variable, val: "$version"},
		{typ: token.Pipe, val: "|"},
		{typ: token.Ident, val: "sed"},
		{typ: token.String, val: "s/\\.//g"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.EOF}}

	testTable("test issue 19", line, expected, t)
}

func TestLexerIssue38(t *testing.T) {
	expected := []Token{
		{typ: token.Ident, val: "cd"},
		{typ: token.LParen, val: "("},
		{typ: token.Variable, val: "$GOPATH"},
		{typ: token.Plus, val: "+"},
		{typ: token.String, val: "/src/"},
		{typ: token.Plus, val: "+"},
		{typ: token.Variable, val: "$path"},
		{typ: token.RParen, val: ")"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.EOF},
	}

	testTable("test issue38", `cd($GOPATH + "/src/" + $path)`, expected, t)
}

func TestLexerIssue43(t *testing.T) {
	expected := []Token{
		{typ: token.Fn, val: "fn"},
		{typ: token.Ident, val: "gpull"},
		{typ: token.LParen, val: "("},
		{typ: token.RParen, val: ")"},
		{typ: token.LBrace, val: "{"},
		{typ: token.Ident, val: "branch"},
		{typ: token.AssignCmd, val: "<="},
		{typ: token.Ident, val: "git"},
		{typ: token.Arg, val: "rev-parse"},
		{typ: token.Arg, val: "--abbrev-ref"},
		{typ: token.Ident, val: "HEAD"},
		{typ: token.Pipe, val: "|"},
		{typ: token.Ident, val: "xargs"},
		{typ: token.Ident, val: "echo"},
		{typ: token.Arg, val: "-n"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.Ident, val: "git"},
		{typ: token.Ident, val: "pull"},
		{typ: token.Ident, val: "origin"},
		{typ: token.Variable, val: "$branch"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.Ident, val: "refreshPrompt"},
		{typ: token.LParen, val: "("},
		{typ: token.RParen, val: ")"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.RBrace, val: "}"},
		{typ: token.EOF},
	}

	testTable("test issue #41", `fn gpull() {
        branch <= git rev-parse --abbrev-ref HEAD | xargs echo -n

        git pull origin $branch
        refreshPrompt()
}`, expected, t)
}

func TestLexerIssue68(t *testing.T) {
	expected := []Token{
		{typ: token.Ident, val: "cat"},
		{typ: token.Ident, val: "PKGBUILD"},
		{typ: token.Pipe, val: "|"},
		{typ: token.Ident, val: "sed"},
		{typ: token.String, val: "s#\\\\$pkgdir#/home/i4k/alt#g"},
		{typ: token.Gt, val: ">"},
		{typ: token.Ident, val: "PKGBUILD2"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.EOF},
	}

	testTable("test issue #68", `cat PKGBUILD | sed "s#\\\\$pkgdir#/home/i4k/alt#g" > PKGBUILD2`, expected, t)
}

func TestLexerIssue85(t *testing.T) {
	expected := []Token{
		{typ: token.Ident, val: "a"},
		{typ: token.AssignCmd, val: "<="},
		{typ: token.Arg, val: "-echo"},
		{typ: token.Ident, val: "hello"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.EOF},
	}

	testTable("test issue 85", `a <= -echo hello`, expected, t)
}

func TestLexerIssue69(t *testing.T) {
	expected := []Token{
		{typ: token.Ident, val: "a"},
		{typ: token.Assign, val: "="},
		{typ: token.LParen, val: "("},
		{typ: token.Variable, val: "$a"},
		{typ: token.Plus, val: "+"},
		{typ: token.String, val: "b"},
		{typ: token.RParen, val: ")"},
		{typ: token.Semicolon, val: ";"},
		{typ: token.EOF},
	}

	testTable("test69", `a = ($a + "b")`, expected, t)

}

func TestLexerIssue127(t *testing.T) {
	expected := []Token{
		{typ: token.Ident, val: "rm"},
		{typ: token.Arg, val: "-rf"},
		{typ: token.Illegal, val: "test127:1:12: Unrecognized character in action: U+002F '/'"},
		{typ: token.EOF},
	}

	testTable("test127", `rm -rf $HOME/.vim`, expected, t)
}
