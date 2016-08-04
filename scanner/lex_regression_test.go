package scanner

import (
	"testing"

	"github.com/NeowayLabs/nash/token"
)

func TestLexerIssue34(t *testing.T) {
	expected := []Token{
		{
			typ: token.Command,
			val: "cat",
		},
		{
			typ: token.Arg,
			val: "/etc/passwd",
		},
		{
			typ: token.RedirRight,
			val: ">",
		},
		{
			typ: token.Arg,
			val: "/dev/null",
		},
		{
			typ: token.Illegal,
			val: "Expected end of line or redirection, but found 'e'",
		},
		{
			typ: token.EOF,
		},
	}

	testTable("test issue #34", `cat /etc/passwd > /dev/null echo "hello world"`, expected, t)
}

func TestLexerIssue21(t *testing.T) {
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
			typ: token.Variable,
			val: "$outFname",
		},
		{
			typ: token.EOF,
		},
	}

	testTable("test redirection variable", `cmd > $outFname`, expected, t)
}

func TestLexerIssue22(t *testing.T) {
	expected := []Token{
		{
			typ: token.FnDecl,
			val: "fn",
		},
		{
			typ: token.Ident,
			val: "gocd",
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
			typ: token.If,
			val: "if",
		},
		{
			typ: token.Variable,
			val: "$path",
		},
		{
			typ: token.Equal,
			val: "==",
		},
		{
			typ: token.String,
			val: "",
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
			val: "$GOPATH",
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
			val: "/src/",
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
			typ: token.RBrace,
			val: "}",
		},
		{
			typ: token.RBrace,
			val: "}",
		},
		{
			typ: token.EOF,
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

func TestLexerIssue19(t *testing.T) {
	line := `version = "4.5.6"
canonName <= echo -n $version | sed "s/\\.//g"`

	expected := []Token{
		{
			typ: token.Ident,
			val: "version",
		},
		{
			typ: token.Assign,
			val: "=",
		},
		{
			typ: token.String,
			val: "4.5.6",
		},
		{
			typ: token.Ident,
			val: "canonName",
		},
		{
			typ: token.AssignCmd,
			val: "<=",
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
			val: "$version",
		},
		{
			typ: token.Pipe,
			val: "|",
		},
		{
			typ: token.Command,
			val: "sed",
		},
		{
			typ: token.String,
			val: "s/\\.//g",
		},
		{
			typ: token.EOF,
		},
	}

	testTable("test issue 19", line, expected, t)
}

func TestLexerIssue38(t *testing.T) {
	expected := []Token{
		{
			typ: token.FnInv,
			val: "cd",
		},
		{
			typ: token.LParen,
			val: "(",
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
			val: "/src/",
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
			typ: token.RParen,
			val: ")",
		},
		{
			typ: token.EOF,
		},
	}

	testTable("test issue38", `cd($GOPATH + "/src/" + $path)`, expected, t)
}

func TestLexerIssue43(t *testing.T) {
	expected := []Token{
		{
			typ: token.FnDecl,
			val: "fn",
		},
		{
			typ: token.Ident,
			val: "gpull",
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
			val: "branch",
		},
		{
			typ: token.AssignCmd,
			val: "<=",
		},
		{
			typ: token.Command,
			val: "git",
		},
		{
			typ: token.Arg,
			val: "rev-parse",
		},
		{
			typ: token.Arg,
			val: "--abbrev-ref",
		},
		{
			typ: token.Arg,
			val: "HEAD",
		},
		{
			typ: token.Pipe,
			val: "|",
		},
		{
			typ: token.Command,
			val: "xargs",
		},
		{
			typ: token.Arg,
			val: "echo",
		},
		{
			typ: token.Arg,
			val: "-n",
		},
		{
			typ: token.Command,
			val: "git",
		},
		{
			typ: token.Arg,
			val: "pull",
		},
		{
			typ: token.Arg,
			val: "origin",
		},
		{
			typ: token.Variable,
			val: "$branch",
		},
		{
			typ: token.FnInv,
			val: "refreshPrompt",
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
			typ: token.RBrace,
			val: "}",
		},
		{
			typ: token.EOF,
		},
	}

	testTable("test issue #41", `fn gpull() {
        branch <= git rev-parse --abbrev-ref HEAD | xargs echo -n

        git pull origin $branch
        refreshPrompt()
}`, expected, t)
}

func TestLexerIssue68(t *testing.T) {
	expected := []Token{
		{typ: token.Command, val: "cat"},
		{typ: token.Arg, val: "PKGBUILD"},
		{typ: token.Pipe, val: "|"},
		{typ: token.Command, val: "sed"},
		{typ: token.String, val: "s#\\\\$pkgdir#/home/i4k/alt#g"},
		{typ: token.RedirRight, val: ">"},
		{typ: token.Arg, val: "PKGBUILD2"},
		{typ: token.EOF},
	}

	testTable("test issue #68", `cat PKGBUILD | sed "s#\\\\$pkgdir#/home/i4k/alt#g" > PKGBUILD2`, expected, t)
}
