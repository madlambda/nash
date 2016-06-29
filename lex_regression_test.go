package nash

import "testing"

func TestLexerIssue34(t *testing.T) {
	expected := []item{
		{
			typ: itemCommand,
			val: "cat",
		},
		{
			typ: itemArg,
			val: "/etc/passwd",
		},
		{
			typ: itemRedirRight,
			val: ">",
		},
		{
			typ: itemArg,
			val: "/dev/null",
		},
		{
			typ: itemError,
			val: "Expected end of line or redirection, but found 'e'",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("test issue #34", `cat /etc/passwd > /dev/null echo "hello world"`, expected, t)
}

func TestLexerIssue21(t *testing.T) {
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
			typ: itemVariable,
			val: "$outFname",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("test redirection variable", `cmd > $outFname`, expected, t)
}

func TestLexerIssue22(t *testing.T) {
	expected := []item{
		{
			typ: itemFnDecl,
			val: "fn",
		},
		{
			typ: itemIdentifier,
			val: "gocd",
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
			typ: itemIf,
			val: "if",
		},
		{
			typ: itemVariable,
			val: "$path",
		},
		{
			typ: itemComparison,
			val: "==",
		},
		{
			typ: itemString,
			val: "",
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
			val: "$GOPATH",
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
			val: "/src/",
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
			typ: itemBracesClose,
			val: "}",
		},
		{
			typ: itemBracesClose,
			val: "}",
		},
		{
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
		{
			typ: itemIdentifier,
			val: "version",
		},
		{
			typ: itemAssign,
			val: "=",
		},
		{
			typ: itemString,
			val: "4.5.6",
		},
		{
			typ: itemIdentifier,
			val: "canonName",
		},
		{
			typ: itemAssignCmd,
			val: "<=",
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
			val: "$version",
		},
		{
			typ: itemPipe,
			val: "|",
		},
		{
			typ: itemCommand,
			val: "sed",
		},
		{
			typ: itemString,
			val: "s/\\.//g",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("test issue 19", line, expected, t)
}

func TestLexerIssue38(t *testing.T) {
	expected := []item{
		{
			typ: itemFnInv,
			val: "cd",
		},
		{
			typ: itemParenOpen,
			val: "(",
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
			val: "/src/",
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
			typ: itemParenClose,
			val: ")",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("test issue38", `cd($GOPATH + "/src/" + $path)`, expected, t)
}

func TestLexerIssue43(t *testing.T) {
	expected := []item{
		{
			typ: itemFnDecl,
			val: "fn",
		},
		{
			typ: itemIdentifier,
			val: "gpull",
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
			val: "branch",
		},
		{
			typ: itemAssignCmd,
			val: "<=",
		},
		{
			typ: itemCommand,
			val: "git",
		},
		{
			typ: itemArg,
			val: "rev-parse",
		},
		{
			typ: itemArg,
			val: "--abbrev-ref",
		},
		{
			typ: itemArg,
			val: "HEAD",
		},
		{
			typ: itemPipe,
			val: "|",
		},
		{
			typ: itemCommand,
			val: "xargs",
		},
		{
			typ: itemArg,
			val: "echo",
		},
		{
			typ: itemArg,
			val: "-n",
		},
		{
			typ: itemCommand,
			val: "git",
		},
		{
			typ: itemArg,
			val: "pull",
		},
		{
			typ: itemArg,
			val: "origin",
		},
		{
			typ: itemVariable,
			val: "$branch",
		},
		{
			typ: itemFnInv,
			val: "refreshPrompt",
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
			typ: itemBracesClose,
			val: "}",
		},
		{
			typ: itemEOF,
		},
	}

	testTable("test issue #41", `fn gpull() {
        branch <= git rev-parse --abbrev-ref HEAD | xargs echo -n

        git pull origin $branch
        refreshPrompt()
}`, expected, t)
}
