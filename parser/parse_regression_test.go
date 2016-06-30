package parser

import "testing"

func TestParseIssue38(t *testing.T) {
	expected := NewTree("parse issue38")

	ln := NewListNode()

	fnInv := NewFnInvNode(0, "cd")

	arg := NewArg(0, ArgConcat)

	args := make([]*Arg, 3)

	arg1 := NewArg(0, ArgVariable)
	arg1.SetString("$GOPATH")

	arg2 := NewArg(0, ArgQuoted)
	arg2.SetString("/src/")

	arg3 := NewArg(0, ArgVariable)
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

	expected := NewTree("parse issue 41")
	ln := NewListNode()

	fnDecl := NewFnDeclNode(0, "gpull")
	fnTree := NewTree("fn")
	fnBlock := NewListNode()

	branchAssign := NewCmdAssignmentNode(14, "branch")
	gitRevParse := NewCommandNode(24, "git")
	arg1 := NewArg(28, ArgUnquoted)
	arg1.SetString("rev-parse")

	arg2 := NewArg(38, ArgUnquoted)
	arg2.SetString("--abbrev-ref")

	arg3 := NewArg(51, ArgUnquoted)
	arg3.SetString("HEAD")

	gitRevParse.AddArg(arg1)
	gitRevParse.AddArg(arg2)
	gitRevParse.AddArg(arg3)

	xargs := NewCommandNode(58, "xargs")

	xarg1 := NewArg(64, ArgUnquoted)
	xarg1.SetString("echo")

	xarg2 := NewArg(69, ArgUnquoted)
	xarg2.SetString("-n")

	xargs.AddArg(xarg1)
	xargs.AddArg(xarg2)

	pipe := NewPipeNode(56)
	pipe.AddCmd(gitRevParse)
	pipe.AddCmd(xargs)

	branchAssign.SetCommand(pipe)

	fnBlock.Push(branchAssign)

	gitPull := NewCommandNode(73, "git")

	pullArg1 := NewArg(77, ArgUnquoted)
	pullArg1.SetString("pull")

	pullArg2 := NewArg(82, ArgUnquoted)
	pullArg2.SetString("origin")

	pullArg3 := NewArg(89, ArgVariable)
	pullArg3.SetString("$branch")

	gitPull.AddArg(pullArg1)
	gitPull.AddArg(pullArg2)
	gitPull.AddArg(pullArg3)

	fnBlock.Push(gitPull)

	fnInv := NewFnInvNode(98, "refreshPrompt")
	fnBlock.Push(fnInv)
	fnTree.Root = fnBlock

	fnDecl.SetTree(fnTree)
	ln.Push(fnDecl)

	expected.Root = ln

	parserTestTable("parse issue 41", content, expected, t, true)
}
