package parser

import (
	"testing"

	"github.com/NeowayLabs/nash/ast"
)

func TestParseIssue38(t *testing.T) {
	expected := ast.NewTree("parse issue38")

	ln := ast.NewListNode()

	fnInv := ast.NewFnInvNode(0, "cd")

	arg := ast.NewArg(0, ast.ArgConcat)

	args := make([]*ast.Arg, 3)

	arg1 := ast.NewArg(0, ast.ArgVariable)
	arg1.SetString("$GOPATH")

	arg2 := ast.NewArg(0, ast.ArgQuoted)
	arg2.SetString("/src/")

	arg3 := ast.NewArg(0, ast.ArgVariable)
	arg3.SetString("$path")

	args[0] = arg1
	args[1] = arg2
	args[2] = arg3

	arg.SetConcat(args)

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

	branchAssign := ast.NewCmdAssignmentNode(14, "branch")
	gitRevParse := ast.NewCommandNode(24, "git")
	arg1 := ast.NewArg(28, ast.ArgUnquoted)
	arg1.SetString("rev-parse")

	arg2 := ast.NewArg(38, ast.ArgUnquoted)
	arg2.SetString("--abbrev-ref")

	arg3 := ast.NewArg(51, ast.ArgUnquoted)
	arg3.SetString("HEAD")

	gitRevParse.AddArg(arg1)
	gitRevParse.AddArg(arg2)
	gitRevParse.AddArg(arg3)

	xargs := ast.NewCommandNode(58, "xargs")

	xarg1 := ast.NewArg(64, ast.ArgUnquoted)
	xarg1.SetString("echo")

	xarg2 := ast.NewArg(69, ast.ArgUnquoted)
	xarg2.SetString("-n")

	xargs.AddArg(xarg1)
	xargs.AddArg(xarg2)

	pipe := ast.NewPipeNode(56)
	pipe.AddCmd(gitRevParse)
	pipe.AddCmd(xargs)

	branchAssign.SetCommand(pipe)

	fnBlock.Push(branchAssign)

	gitPull := ast.NewCommandNode(73, "git")

	pullArg1 := ast.NewArg(77, ast.ArgUnquoted)
	pullArg1.SetString("pull")

	pullArg2 := ast.NewArg(82, ast.ArgUnquoted)
	pullArg2.SetString("origin")

	pullArg3 := ast.NewArg(89, ast.ArgVariable)
	pullArg3.SetString("$branch")

	gitPull.AddArg(pullArg1)
	gitPull.AddArg(pullArg2)
	gitPull.AddArg(pullArg3)

	fnBlock.Push(gitPull)

	fnInv := ast.NewFnInvNode(98, "refreshPrompt")
	fnBlock.Push(fnInv)
	fnTree.Root = fnBlock

	fnDecl.SetTree(fnTree)
	ln.Push(fnDecl)

	expected.Root = ln

	parserTestTable("parse issue 41", content, expected, t, true)
}
