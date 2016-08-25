package parser

import (
	"testing"

	"github.com/NeowayLabs/nash/ast"
)

func init() {
	ast.DebugCmp = true
}

func TestParseIssue38(t *testing.T) {
	expected := ast.NewTree("parse issue38")

	ln := ast.NewListNode()

	fnInv := ast.NewFnInvNode(0, "cd")

	args := make([]ast.Expr, 3)

	args[0] = ast.NewVarExpr(3, "$GOPATH")
	args[1] = ast.NewStringExpr(11, "/src/", true)
	args[2] = ast.NewVarExpr(19, "$path")

	arg := ast.NewConcatExpr(0, args)

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
	ln := ast.NewListNode()

	fnDecl := ast.NewFnDeclNode(0, "gpull")
	fnTree := ast.NewTree("fn")
	fnBlock := ast.NewListNode()

	gitRevParse := ast.NewCommandNode(24, "git")

	gitRevParse.AddArg(ast.NewStringExpr(28, "rev-parse", true))
	gitRevParse.AddArg(ast.NewStringExpr(38, "--abbrev-ref", false))
	gitRevParse.AddArg(ast.NewStringExpr(51, "HEAD", false))

	branchAssign, err := ast.NewExecAssignNode(14, "branch", gitRevParse)

	if err != nil {
		t.Error(err)
		return
	}

	xargs := ast.NewCommandNode(58, "xargs")
	xargs.AddArg(ast.NewStringExpr(64, "echo", false))
	xargs.AddArg(ast.NewStringExpr(69, "-n", false))

	pipe := ast.NewPipeNode(56)
	pipe.AddCmd(gitRevParse)
	pipe.AddCmd(xargs)

	branchAssign.SetCommand(pipe)

	fnBlock.Push(branchAssign)

	gitPull := ast.NewCommandNode(73, "git")

	gitPull.AddArg(ast.NewStringExpr(77, "pull", false))
	gitPull.AddArg(ast.NewStringExpr(82, "origin", false))
	gitPull.AddArg(ast.NewVarExpr(89, "$branch"))

	fnBlock.Push(gitPull)

	fnInv := ast.NewFnInvNode(98, "refreshPrompt")
	fnBlock.Push(fnInv)
	fnTree.Root = fnBlock

	fnDecl.SetTree(fnTree)
	ln.Push(fnDecl)

	expected.Root = ln

	parserTestTable("parse issue 41", content, expected, t, true)
}

func TestParseIssue68(t *testing.T) {
	expected := ast.NewTree("parse issue #68")
	ln := ast.NewListNode()

	catCmd := ast.NewCommandNode(0, "cat")

	catArg := ast.NewStringExpr(4, "PKGBUILD", false)
	catCmd.AddArg(catArg)

	sedCmd := ast.NewCommandNode(15, "sed")
	sedArg := ast.NewStringExpr(20, `s#\$pkgdir#/home/i4k/alt#g`, true)
	sedCmd.AddArg(sedArg)

	sedRedir := ast.NewRedirectNode(0)
	sedRedirArg := ast.NewStringExpr(0, "PKGBUILD2", false)
	sedRedir.SetLocation(sedRedirArg)
	sedCmd.AddRedirect(sedRedir)

	pipe := ast.NewPipeNode(13)
	pipe.AddCmd(catCmd)
	pipe.AddCmd(sedCmd)

	ln.Push(pipe)
	expected.Root = ln

	parserTestTable("parse issue #68", `cat PKGBUILD | sed "s#\\$pkgdir#/home/i4k/alt#g" > PKGBUILD2`, expected, t, false)
}

func TestParseIssue69(t *testing.T) {
	expected := ast.NewTree("parse-issue-69")
	ln := ast.NewListNode()

	parts := make([]ast.Expr, 2)

	parts[0] = ast.NewVarExpr(0, "$a")
	parts[1] = ast.NewStringExpr(0, "b", true)

	concat := ast.NewConcatExpr(0, parts)

	listValues := make([]ast.Expr, 1)
	listValues[0] = concat

	list := ast.NewListExpr(0, listValues)

	assign := ast.NewAssignmentNode(0, "a", list)
	ln.Push(assign)
	expected.Root = ln

	parserTestTable("parse-issue-69", `a = ($a+"b")`, expected, t, true)
}
