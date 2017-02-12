package parser

import (
	"testing"

	"github.com/NeowayLabs/nash/ast"
	"github.com/NeowayLabs/nash/token"
)

func init() {
	ast.DebugCmp = true
}

func TestParseIssue22(t *testing.T) {
	expected := ast.NewTree("issue 22")
	ln := ast.NewBlockNode(token.NewFileInfo(1, 0))

	fn := ast.NewFnDeclNode(token.NewFileInfo(1, 0), "gocd")
	fn.AddArg("path")

	fnTree := ast.NewTree("fn")
	fnBlock := ast.NewBlockNode(token.NewFileInfo(1, 0))

	ifDecl := ast.NewIfNode(token.NewFileInfo(2, 1))
	ifDecl.SetLvalue(ast.NewVarExpr(token.NewFileInfo(2, 4), "$path"))
	ifDecl.SetOp("==")

	ifDecl.SetRvalue(ast.NewStringExpr(token.NewFileInfo(2, 13), "", true))

	ifTree := ast.NewTree("if")
	ifBlock := ast.NewBlockNode(token.NewFileInfo(2, 1))

	cdNode := ast.NewCommandNode(token.NewFileInfo(3, 2), "cd", false)
	arg := ast.NewVarExpr(token.NewFileInfo(3, 5), "$GOPATH")
	cdNode.AddArg(arg)

	ifBlock.Push(cdNode)
	ifTree.Root = ifBlock
	ifDecl.SetIfTree(ifTree)

	elseTree := ast.NewTree("else")
	elseBlock := ast.NewBlockNode(token.NewFileInfo(4, 9))

	args := make([]ast.Expr, 3)
	args[0] = ast.NewVarExpr(token.NewFileInfo(5, 5), "$GOPATH")
	args[1] = ast.NewStringExpr(token.NewFileInfo(5, 12), "/src/", true)
	args[2] = ast.NewVarExpr(token.NewFileInfo(5, 20), "$path")

	cdNodeElse := ast.NewCommandNode(token.NewFileInfo(5, 2), "cd", false)
	carg := ast.NewConcatExpr(token.NewFileInfo(5, 5), args)
	cdNodeElse.AddArg(carg)

	elseBlock.Push(cdNodeElse)
	elseTree.Root = elseBlock

	ifDecl.SetElseTree(elseTree)

	fnBlock.Push(ifDecl)
	fnTree.Root = fnBlock
	fn.SetTree(fnTree)

	ln.Push(fn)
	expected.Root = ln

	parserTestTable("issue 22", `fn gocd(path) {
	if $path == "" {
		cd $GOPATH
	} else {
		cd $GOPATH+"/src/"+$path
	}
}`, expected, t, true)

}

func TestParseIssue38(t *testing.T) {
	expected := ast.NewTree("parse issue38")

	ln := ast.NewBlockNode(token.NewFileInfo(1, 0))

	fnInv := ast.NewFnInvNode(token.NewFileInfo(1, 0), "cd")

	args := make([]ast.Expr, 3)

	args[0] = ast.NewVarExpr(token.NewFileInfo(1, 3), "$GOPATH")
	args[1] = ast.NewStringExpr(token.NewFileInfo(1, 12), "/src/", true)
	args[2] = ast.NewVarExpr(token.NewFileInfo(1, 19), "$path")

	arg := ast.NewConcatExpr(token.NewFileInfo(1, 3), args)

	fnInv.AddArg(arg)

	ln.Push(fnInv)
	expected.Root = ln

	parserTestTable("parse issue38", `cd($GOPATH+"/src/"+$path)`, expected, t, true)
}

func TestParseIssue43(t *testing.T) {
	content := `fn gpull() {
	branch <= git rev-parse --abbrev-ref HEAD | xargs echo -n

	git pull origin $branch

	refreshPrompt()
}`

	expected := ast.NewTree("parse issue 41")
	ln := ast.NewBlockNode(token.NewFileInfo(1, 0))

	fnDecl := ast.NewFnDeclNode(token.NewFileInfo(1, 0), "gpull")
	fnTree := ast.NewTree("fn")
	fnBlock := ast.NewBlockNode(token.NewFileInfo(1, 0))

	gitRevParse := ast.NewCommandNode(token.NewFileInfo(2, 11), "git", false)

	gitRevParse.AddArg(ast.NewStringExpr(token.NewFileInfo(2, 15), "rev-parse", true))
	gitRevParse.AddArg(ast.NewStringExpr(token.NewFileInfo(2, 25), "--abbrev-ref", false))
	gitRevParse.AddArg(ast.NewStringExpr(token.NewFileInfo(2, 38), "HEAD", false))

	branchAssign, err := ast.NewExecAssignNode(token.NewFileInfo(2, 1), ast.NewNameNode(
		token.NewFileInfo(2, 1), "branch", nil), gitRevParse)

	if err != nil {
		t.Error(err)
		return
	}

	xargs := ast.NewCommandNode(token.NewFileInfo(2, 45), "xargs", false)
	xargs.AddArg(ast.NewStringExpr(token.NewFileInfo(2, 51), "echo", false))
	xargs.AddArg(ast.NewStringExpr(token.NewFileInfo(2, 56), "-n", false))

	pipe := ast.NewPipeNode(token.NewFileInfo(2, 43), false)
	pipe.AddCmd(gitRevParse)
	pipe.AddCmd(xargs)

	branchAssign.SetCommand(pipe)

	fnBlock.Push(branchAssign)

	gitPull := ast.NewCommandNode(token.NewFileInfo(1, 0), "git", false)

	gitPull.AddArg(ast.NewStringExpr(token.NewFileInfo(1, 0), "pull", false))
	gitPull.AddArg(ast.NewStringExpr(token.NewFileInfo(1, 0), "origin", false))
	gitPull.AddArg(ast.NewVarExpr(token.NewFileInfo(1, 0), "$branch"))

	fnBlock.Push(gitPull)

	fnInv := ast.NewFnInvNode(token.NewFileInfo(1, 0), "refreshPrompt")
	fnBlock.Push(fnInv)
	fnTree.Root = fnBlock

	fnDecl.SetTree(fnTree)
	ln.Push(fnDecl)

	expected.Root = ln

	parserTestTable("parse issue 41", content, expected, t, true)
}

func TestParseIssue68(t *testing.T) {
	expected := ast.NewTree("parse issue #68")
	ln := ast.NewBlockNode(token.NewFileInfo(1, 0))

	catCmd := ast.NewCommandNode(token.NewFileInfo(1, 0), "cat", false)

	catArg := ast.NewStringExpr(token.NewFileInfo(1, 4), "PKGBUILD", false)
	catCmd.AddArg(catArg)

	sedCmd := ast.NewCommandNode(token.NewFileInfo(1, 15), "sed", false)
	sedArg := ast.NewStringExpr(token.NewFileInfo(1, 20), `s#\$pkgdir#/home/i4k/alt#g`, true)
	sedCmd.AddArg(sedArg)

	sedRedir := ast.NewRedirectNode(token.NewFileInfo(1, 49))
	sedRedirArg := ast.NewStringExpr(token.NewFileInfo(1, 51), "PKGBUILD2", false)
	sedRedir.SetLocation(sedRedirArg)
	sedCmd.AddRedirect(sedRedir)

	pipe := ast.NewPipeNode(token.NewFileInfo(1, 13), false)
	pipe.AddCmd(catCmd)
	pipe.AddCmd(sedCmd)

	ln.Push(pipe)
	expected.Root = ln

	parserTestTable("parse issue #68", `cat PKGBUILD | sed "s#\\$pkgdir#/home/i4k/alt#g" > PKGBUILD2`, expected, t, false)
}

func TestParseIssue69(t *testing.T) {
	expected := ast.NewTree("parse-issue-69")
	ln := ast.NewBlockNode(token.NewFileInfo(1, 0))

	parts := make([]ast.Expr, 2)

	parts[0] = ast.NewVarExpr(token.NewFileInfo(1, 5), "$a")
	parts[1] = ast.NewStringExpr(token.NewFileInfo(1, 9), "b", true)

	concat := ast.NewConcatExpr(token.NewFileInfo(1, 5), parts)

	listValues := make([]ast.Expr, 1)
	listValues[0] = concat

	list := ast.NewListExpr(token.NewFileInfo(1, 4), listValues)

	assign := ast.NewAssignmentNode(token.NewFileInfo(1, 0),
		ast.NewNameNode(token.NewFileInfo(1, 0), "a", nil), list,
	)

	ln.Push(assign)
	expected.Root = ln

	parserTestTable("parse-issue-69", `a = ($a+"b")`, expected, t, true)
}

func TestParseImportIssue94(t *testing.T) {
	expected := ast.NewTree("test import")
	ln := ast.NewBlockNode(token.NewFileInfo(1, 0))
	importStmt := ast.NewImportNode(token.NewFileInfo(1, 0), ast.NewStringExpr(token.NewFileInfo(1, 7), "common", false))
	ln.Push(importStmt)
	expected.Root = ln

	parserTestTable("test import", "import common", expected, t, true)
}

func TestParseIssue108(t *testing.T) {
	// keywords cannot be used as command arguments

	expected := ast.NewTree("parse issue #108")
	ln := ast.NewBlockNode(token.NewFileInfo(1, 0))

	catCmd := ast.NewCommandNode(token.NewFileInfo(1, 0), "cat", false)

	catArg := ast.NewStringExpr(token.NewFileInfo(1, 4), "spec.ebnf", false)
	catCmd.AddArg(catArg)

	grepCmd := ast.NewCommandNode(token.NewFileInfo(1, 16), "grep", false)
	grepArg := ast.NewStringExpr(token.NewFileInfo(1, 21), `-i`, false)
	grepArg2 := ast.NewStringExpr(token.NewFileInfo(1, 24), "rfork", false)

	grepCmd.AddArg(grepArg)
	grepCmd.AddArg(grepArg2)

	pipe := ast.NewPipeNode(token.NewFileInfo(1, 14), false)
	pipe.AddCmd(catCmd)
	pipe.AddCmd(grepCmd)

	ln.Push(pipe)
	expected.Root = ln

	parserTestTable("parse issue #108", `cat spec.ebnf | grep -i rfork`, expected, t, false)
}

func TestParseIssue123(t *testing.T) {
	parser := NewParser("invalid cmd assignment", `IFS <= ("\n")`)

	_, err := parser.Parse()

	if err == nil {
		t.Errorf("Must fail...")
		return
	}

	expected := "invalid cmd assignment:1:9: Unexpected token STRING. Expecting IDENT or ARG"
	if err.Error() != expected {
		t.Fatalf("Error string differs. Expecting '%s' but got '%s'",
			expected, err.Error())
	}
}
