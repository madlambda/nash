package nash

import "testing"

func TestLexerIssue34(t *testing.T) {
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

func TestLexerIssue21(t *testing.T) {
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

func TestLexerIssue22(t *testing.T) {
	expected := []item{
		item{
			typ: itemFnDecl,
			val: "fn",
		},
		item{
			typ: itemIdentifier,
			val: "gocd",
		},
		item{
			typ: itemParenOpen,
			val: "(",
		},
		item{
			typ: itemIdentifier,
			val: "path",
		},
		item{
			typ: itemParenClose,
			val: ")",
		},
		item{
			typ: itemBracesOpen,
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
			typ: itemBracesOpen,
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
			typ: itemBracesClose,
			val: "}",
		},
		item{
			typ: itemElse,
			val: "else",
		},
		item{
			typ: itemBracesOpen,
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
			typ: itemBracesClose,
			val: "}",
		},
		item{
			typ: itemBracesClose,
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

func TestLexerIssue19(t *testing.T) {
	line := `version = "4.5.6"
canonName <= echo -n $version | sed "s/\\.//g"`

	expected := []item{
		item{
			typ: itemIdentifier,
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
			typ: itemIdentifier,
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

func TestLexerIssue38(t *testing.T) {
	expected := []item{
		item{
			typ: itemFnInv,
			val: "cd",
		},
		item{
			typ: itemParenOpen,
			val: "(",
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
			typ: itemParenClose,
			val: ")",
		},
		item{
			typ: itemEOF,
		},
	}

	testTable("test issue38", `cd($GOPATH + "/src/" + $path)`, expected, t)
}

func TestLexerIssue43(t *testing.T) {
	expected := []item{
		item{
			typ: itemFnDecl,
			val: "fn",
		},
		item{
			typ: itemIdentifier,
			val: "gpull",
		},
		item{
			typ: itemParenOpen,
			val: "(",
		},
		item{
			typ: itemParenClose,
			val: ")",
		},
		item{
			typ: itemBracesOpen,
			val: "{",
		},
		item{
			typ: itemIdentifier,
			val: "branch",
		},
		item{
			typ: itemAssignCmd,
			val: "<=",
		},
		item{
			typ: itemCommand,
			val: "git",
		},
		item{
			typ: itemArg,
			val: "rev-parse",
		},
		item{
			typ: itemArg,
			val: "--abbrev-ref",
		},
		item{
			typ: itemArg,
			val: "HEAD",
		},
		item{
			typ: itemPipe,
			val: "|",
		},
		item{
			typ: itemCommand,
			val: "xargs",
		},
		item{
			typ: itemArg,
			val: "echo",
		},
		item{
			typ: itemArg,
			val: "-n",
		},
		item{
			typ: itemCommand,
			val: "git",
		},
		item{
			typ: itemArg,
			val: "pull",
		},
		item{
			typ: itemArg,
			val: "origin",
		},
		item{
			typ: itemVariable,
			val: "$branch",
		},
		item{
			typ: itemFnInv,
			val: "refreshPrompt",
		},
		item{
			typ: itemParenOpen,
			val: "(",
		},
		item{
			typ: itemParenClose,
			val: ")",
		},
		item{
			typ: itemBracesClose,
			val: "}",
		},
		item{
			typ: itemEOF,
		},
	}

	testTable("test issue #41", `fn gpull() {
        branch <= git rev-parse --abbrev-ref HEAD | xargs echo -n

        git pull origin $branch
        refreshPrompt()
}`, expected, t)
}
