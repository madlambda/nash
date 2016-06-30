package parser

import (
	"fmt"
	"strings"
	"testing"
)

func parserTestTable(name, content string, expected *Tree, t *testing.T, enableReverse bool) {
	parser := NewParser(name, content)
	tr, err := parser.Parse()

	if err != nil {
		t.Error(err)
		return
	}

	if tr == nil {
		t.Errorf("Failed to parse")
		return
	}

	if ok, err := compare(expected, tr); !ok {
		fmt.Printf("Expected: %s\n\nResult: %s\n", expected.String(), tr.String())
		t.Error(err)
		return
	}

	if !enableReverse {
		return
	}

	// Test if the reverse of tree is the content again... *hard*
	trcontent := strings.TrimSpace(tr.String())
	content = strings.TrimSpace(content)

	if content != trcontent {
		t.Errorf(`Failed to reverse the tree.
Expected:
'%q'

But got:
'%q'
`, content, trcontent)
		return
	}
}

func TestParseShowEnv(t *testing.T) {
	expected := NewTree("parse showenv")
	ln := NewListNode()
	showenv := NewShowEnvNode(0)
	ln.Push(showenv)

	expected.Root = ln

	parserTestTable("parse showenv", `showenv`, expected, t, true)
}

func TestParseSimple(t *testing.T) {
	expected := NewTree("parser simple")
	ln := NewListNode()
	cmd := NewCommandNode(0, "echo")
	hello := NewArg(6, ArgQuoted)
	hello.SetString("hello world")
	cmd.AddArg(hello)
	ln.Push(cmd)

	expected.Root = ln

	parserTestTable("parser simple", `echo "hello world"`, expected, t, true)
}

func TestParseReverseGetSame(t *testing.T) {
	parser := NewParser("reverse simple", "echo \"hello world\"")

	tr, err := parser.Parse()

	if err != nil {
		t.Error(err)
		return
	}

	if tr.String() != "echo \"hello world\"" {
		t.Errorf("Failed to reverse tree: %s", tr.String())
		return
	}
}

func TestParsePipe(t *testing.T) {
	expected := NewTree("parser pipe")
	ln := NewListNode()
	first := NewCommandNode(0, "echo")
	hello := NewArg(6, ArgQuoted)
	hello.SetString("hello world")
	first.AddArg(hello)

	second := NewCommandNode(21, "awk")
	secondArg := NewArg(26, ArgQuoted)
	secondArg.SetString("{print $1}")

	second.AddArg(secondArg)

	pipe := NewPipeNode(19)
	pipe.AddCmd(first)
	pipe.AddCmd(second)

	ln.Push(pipe)

	expected.Root = ln

	parserTestTable("parser pipe", `echo "hello world" | awk "{print $1}"`, expected, t, true)
}

func TestBasicSetAssignment(t *testing.T) {
	expected := NewTree("simple set assignment")
	ln := NewListNode()
	set := NewSetAssignmentNode(0, "test")

	ln.Push(set)
	expected.Root = ln

	parserTestTable("simple set assignment", `setenv test`, expected, t, true)
}

func TestBasicAssignment(t *testing.T) {
	expected := NewTree("simple assignment")
	ln := NewListNode()
	assign := NewAssignmentNode(0)
	assign.SetVarName("test")

	elem := NewArg(8, ArgQuoted)
	elem.SetString("hello")

	assign.SetValue(elem)
	ln.Push(assign)
	expected.Root = ln

	parserTestTable("simple assignment", `test = "hello"`, expected, t, true)

	// test concatenation of strings and variables

	ln = NewListNode()
	assign = NewAssignmentNode(0)
	assign.SetVarName("test")

	concats := make([]*Arg, 2, 2)

	hello := NewArg(8, ArgQuoted)
	hello.SetString("hello")
	variable := NewArg(15, ArgVariable)
	variable.SetString("$var")

	concats[0] = hello
	concats[1] = variable
	arg1 := NewArg(8, ArgConcat)
	arg1.SetConcat(concats)

	assign.SetValue(arg1)
	ln.Push(assign)
	expected.Root = ln

	parserTestTable("test", `test = "hello"+$var`, expected, t, true)

	// invalid, requires quote
	// test=hello
	parser := NewParser("", `test = hello`)

	tr, err := parser.Parse()

	if err == nil {
		t.Error("Must fail")
		return
	}

	if tr != nil {
		t.Error("tr must be nil")
		return
	}
}

func TestParseListAssignment(t *testing.T) {
	expected := NewTree("list assignment")
	ln := NewListNode()
	assign := NewAssignmentNode(0)

	plan9 := NewArg(10, ArgUnquoted)
	plan9.SetString("plan9")
	from := NewArg(17, ArgUnquoted)
	from.SetString("from")
	bell := NewArg(23, ArgUnquoted)
	bell.SetString("bell")
	labs := NewArg(29, ArgUnquoted)
	labs.SetString("labs")

	values := make([]*Arg, 4)
	values[0] = plan9
	values[1] = from
	values[2] = bell
	values[3] = labs

	assign.SetVarName("test")

	elem := NewArg(7, ArgList)
	elem.SetList(values)
	assign.SetValue(elem)

	ln.Push(assign)
	expected.Root = ln

	parserTestTable("list assignment", `test = (
	plan9
	from
	bell
	labs
)`, expected, t, false)

}

func TestParseCmdAssignment(t *testing.T) {
	expected := NewTree("simple cmd assignment")
	ln := NewListNode()
	assign := NewCmdAssignmentNode(0, "test")

	cmd := NewCommandNode(8, "ls")
	assign.SetCommand(cmd)

	ln.Push(assign)
	expected.Root = ln

	parserTestTable("simple assignment", `test <= ls`, expected, t, true)
}

func TestParseInvalid(t *testing.T) {
	parser := NewParser("invalid", ";")

	_, err := parser.Parse()

	if err == nil {
		t.Error("Parse must fail")
		return
	}
}

func TestParsePathCommand(t *testing.T) {
	expected := NewTree("parser simple")
	ln := NewListNode()
	cmd := NewCommandNode(0, "/bin/echo")
	arg := NewArg(11, ArgQuoted)
	arg.SetString("hello world")
	cmd.AddArg(arg)
	ln.Push(cmd)

	expected.Root = ln

	parserTestTable("parser simple", `/bin/echo "hello world"`, expected, t, true)
}

func TestParseWithShebang(t *testing.T) {
	expected := NewTree("parser shebang")
	ln := NewListNode()
	cmt := NewCommentNode(0, "#!/bin/nash")
	cmd := NewCommandNode(12, "echo")
	arg := NewArg(17, ArgUnquoted)
	arg.SetString("bleh")
	cmd.AddArg(arg)
	ln.Push(cmt)
	ln.Push(cmd)

	expected.Root = ln

	parserTestTable("parser shebang", `#!/bin/nash
echo bleh
`, expected, t, true)
}

func TestParseEmptyFile(t *testing.T) {
	expected := NewTree("empty file")
	ln := NewListNode()
	expected.Root = ln

	parserTestTable("empty file", "", expected, t, true)
}

func TestParseSingleCommand(t *testing.T) {
	expected := NewTree("single command")
	expected.Root = NewListNode()
	expected.Root.Push(NewCommandNode(0, "bleh"))

	parserTestTable("single command", `bleh`, expected, t, true)
}

func TestParseRedirectSimple(t *testing.T) {
	expected := NewTree("redirect")
	ln := NewListNode()
	cmd := NewCommandNode(0, "cmd")
	redir := NewRedirectNode(0)
	redir.SetMap(2, redirMapSupress)
	cmd.AddRedirect(redir)
	ln.Push(cmd)

	expected.Root = ln

	parserTestTable("simple redirect", `cmd >[2=]`, expected, t, true)

	expected = NewTree("redirect2")
	ln = NewListNode()
	cmd = NewCommandNode(0, "cmd")
	redir = NewRedirectNode(0)
	redir.SetMap(2, 1)
	cmd.AddRedirect(redir)
	ln.Push(cmd)

	expected.Root = ln

	parserTestTable("simple redirect", `cmd >[2=1]`, expected, t, true)
}

func TestParseRedirectWithLocation(t *testing.T) {
	expected := NewTree("redirect with location")
	ln := NewListNode()
	cmd := NewCommandNode(0, "cmd")
	redir := NewRedirectNode(0)
	redir.SetMap(2, redirMapNoValue)
	out := NewArg(0, ArgUnquoted)
	out.SetString("/var/log/service.log")
	redir.SetLocation(out)
	cmd.AddRedirect(redir)
	ln.Push(cmd)

	expected.Root = ln

	parserTestTable("simple redirect", `cmd >[2] /var/log/service.log`, expected, t, true)
}

func TestParseRedirectMultiples(t *testing.T) {
	expected := NewTree("redirect multiples")
	ln := NewListNode()
	cmd := NewCommandNode(0, "cmd")
	redir1 := NewRedirectNode(0)
	redir2 := NewRedirectNode(0)

	redir1.SetMap(1, 2)
	redir2.SetMap(2, redirMapSupress)

	cmd.AddRedirect(redir1)
	cmd.AddRedirect(redir2)
	ln.Push(cmd)

	expected.Root = ln

	parserTestTable("multiple redirects", `cmd >[1=2] >[2=]`, expected, t, true)
}

func TestParseCommandWithStringsEqualsNot(t *testing.T) {
	expected := NewTree("strings works as expected")
	ln := NewListNode()
	cmd1 := NewCommandNode(0, "echo")
	cmd2 := NewCommandNode(11, "echo")
	hello := NewArg(5, ArgUnquoted)
	hello.SetString("hello")
	cmd1.AddArg(hello)

	hello2 := NewArg(17, ArgQuoted)
	hello2.SetString("hello")

	cmd2.AddArg(hello2)

	ln.Push(cmd1)
	ln.Push(cmd2)
	expected.Root = ln

	parserTestTable("strings works as expected", `echo hello
echo "hello"
`, expected, t, true)
}

func TestParseStringNotFinished(t *testing.T) {
	parser := NewParser("string not finished", `echo "hello world`)
	tr, err := parser.Parse()

	if err == nil {
		t.Error("Error: should fail")
		return
	}

	if tr != nil {
		t.Errorf("Failed to parse")
		return
	}
}

func TestParseCd(t *testing.T) {
	expected := NewTree("test cd")
	ln := NewListNode()
	cd := NewCdNode(0)
	arg := NewArg(3, ArgUnquoted)
	arg.SetString("/tmp")
	cd.SetDir(arg)
	ln.Push(cd)
	expected.Root = ln

	parserTestTable("test cd", "cd /tmp", expected, t, true)

	// test cd into home
	expected = NewTree("test cd into home")
	ln = NewListNode()
	cd = NewCdNode(0)
	ln.Push(cd)
	expected.Root = ln

	parserTestTable("test cd into home", "cd", expected, t, true)

	expected = NewTree("cd into HOME by setenv")
	ln = NewListNode()
	assign := NewAssignmentNode(0)
	assign.SetVarName("HOME")
	arg = NewArg(8, ArgQuoted)
	arg.SetString("/")

	assign.SetValue(arg)

	set := NewSetAssignmentNode(11, "HOME")
	cd = NewCdNode(23)
	pwd := NewCommandNode(26, "pwd")

	ln.Push(assign)
	ln.Push(set)
	ln.Push(cd)
	ln.Push(pwd)

	expected.Root = ln

	parserTestTable("test cd into HOME by setenv", `HOME = "/"
setenv HOME
cd
pwd`, expected, t, true)

	// Test cd into custom variable
	expected = NewTree("cd into variable value")
	ln = NewListNode()
	assign = NewAssignmentNode(0)
	assign.SetVarName("GOPATH")
	arg = NewArg(10, ArgQuoted)
	arg.SetString("/home/i4k/gopath")

	assign.SetValue(arg)
	cd = NewCdNode(28)
	path := NewArg(31, ArgVariable)
	path.SetString("$GOPATH")
	cd.SetDir(path)

	ln.Push(assign)
	ln.Push(cd)

	expected.Root = ln

	parserTestTable("test cd into variable value", `GOPATH = "/home/i4k/gopath"
cd $GOPATH`, expected, t, true)

	// Test cd into custom variable
	expected = NewTree("cd into variable value with concat")
	ln = NewListNode()
	assign = NewAssignmentNode(0)
	assign.SetVarName("GOPATH")
	arg = NewArg(10, ArgQuoted)
	arg.SetString("/home/i4k/gopath")

	assign.SetValue(arg)
	cd = NewCdNode(28)
	path = NewArg(31, ArgConcat)
	varg := NewArg(31, ArgVariable)
	varg.SetString("$GOPATH")
	src := NewArg(40, ArgQuoted)
	src.SetString("/src/github.com")
	concat := make([]*Arg, 0, 2)
	concat = append(concat, varg)
	concat = append(concat, src)
	path.SetConcat(concat)
	cd.SetDir(path)

	ln.Push(assign)
	ln.Push(cd)

	expected.Root = ln

	parserTestTable("test cd into variable value", `GOPATH = "/home/i4k/gopath"
cd $GOPATH+"/src/github.com"`, expected, t, true)

}

func TestParseRfork(t *testing.T) {
	expected := NewTree("test rfork")
	ln := NewListNode()
	cmd1 := NewRforkNode(0)
	f1 := NewArg(6, ArgUnquoted)
	f1.SetString("u")
	cmd1.SetFlags(f1)
	ln.Push(cmd1)
	expected.Root = ln

	parserTestTable("test rfork", "rfork u", expected, t, true)
}

func TestParseRforkWithBlock(t *testing.T) {
	expected := NewTree("rfork with block")
	ln := NewListNode()
	rfork := NewRforkNode(0)
	arg := NewArg(6, ArgUnquoted)
	arg.SetString("u")
	rfork.SetFlags(arg)

	insideFork := NewCommandNode(11, "mount")
	insideFork.AddArg(newSimpleArg(17, "-t", ArgUnquoted))
	insideFork.AddArg(newSimpleArg(20, "proc", ArgUnquoted))
	insideFork.AddArg(newSimpleArg(25, "proc", ArgUnquoted))
	insideFork.AddArg(newSimpleArg(30, "/proc", ArgUnquoted))

	bln := NewListNode()
	bln.Push(insideFork)
	subtree := NewTree("rfork")
	subtree.Root = bln

	rfork.SetBlock(subtree)

	ln.Push(rfork)
	expected.Root = ln

	parserTestTable("rfork with block", `rfork u {
	mount -t proc proc /proc
}
`, expected, t, true)

}

func TestUnpairedRforkBlocks(t *testing.T) {
	parser := NewParser("unpaired", "rfork u {")

	_, err := parser.Parse()

	if err == nil {
		t.Errorf("Should fail because of unpaired open/close blocks")
		return
	}
}

func TestParseImport(t *testing.T) {
	expected := NewTree("test import")
	ln := NewListNode()
	importStmt := NewImportNode(0)
	importStmt.SetPath(newSimpleArg(7, "env.sh", ArgUnquoted))
	ln.Push(importStmt)
	expected.Root = ln

	parserTestTable("test import", "import env.sh", expected, t, true)

	expected = NewTree("test import with quotes")
	ln = NewListNode()
	importStmt = NewImportNode(0)
	importStmt.SetPath(newSimpleArg(8, "env.sh", ArgQuoted))
	ln.Push(importStmt)
	expected.Root = ln

	parserTestTable("test import", `import "env.sh"`, expected, t, true)
}

func TestParseIf(t *testing.T) {
	expected := NewTree("test if")
	ln := NewListNode()
	ifDecl := NewIfNode(0)
	ifDecl.SetLvalue(newSimpleArg(4, "test", ArgQuoted))
	ifDecl.SetRvalue(newSimpleArg(14, "other", ArgQuoted))
	ifDecl.SetOp("==")

	subBlock := NewListNode()
	cmd := NewCommandNode(24, "pwd")
	subBlock.Push(cmd)

	ifTree := NewTree("if block")
	ifTree.Root = subBlock

	ifDecl.SetIfTree(ifTree)

	ln.Push(ifDecl)
	expected.Root = ln

	parserTestTable("test if", `if "test" == "other" {
	pwd
}`, expected, t, true)

	expected = NewTree("test if")
	ln = NewListNode()
	ifDecl = NewIfNode(0)
	ifDecl.SetLvalue(newSimpleArg(4, "", ArgQuoted))
	ifDecl.SetRvalue(newSimpleArg(10, "other", ArgQuoted))
	ifDecl.SetOp("!=")

	subBlock = NewListNode()
	cmd = NewCommandNode(20, "pwd")
	subBlock.Push(cmd)

	ifTree = NewTree("if block")
	ifTree.Root = subBlock

	ifDecl.SetIfTree(ifTree)

	ln.Push(ifDecl)
	expected.Root = ln

	parserTestTable("test if", `if "" != "other" {
	pwd
}`, expected, t, true)
}

func TestParseIfLvariable(t *testing.T) {
	expected := NewTree("test if with variable")
	ln := NewListNode()
	ifDecl := NewIfNode(0)
	ifDecl.SetLvalue(newSimpleArg(3, "$test", ArgVariable))
	ifDecl.SetRvalue(newSimpleArg(13, "other", ArgQuoted))
	ifDecl.SetOp("==")

	subBlock := NewListNode()
	cmd := NewCommandNode(23, "pwd")
	subBlock.Push(cmd)

	ifTree := NewTree("if block")
	ifTree.Root = subBlock

	ifDecl.SetIfTree(ifTree)

	ln.Push(ifDecl)
	expected.Root = ln

	parserTestTable("test if", `if $test == "other" {
	pwd
}`, expected, t, true)
}

func TestParseIfElse(t *testing.T) {
	expected := NewTree("test if else with variable")
	ln := NewListNode()
	ifDecl := NewIfNode(0)
	ifDecl.SetLvalue(newSimpleArg(3, "$test", ArgVariable))
	ifDecl.SetRvalue(newSimpleArg(13, "other", ArgQuoted))
	ifDecl.SetOp("==")

	subBlock := NewListNode()
	cmd := NewCommandNode(23, "pwd")
	subBlock.Push(cmd)

	ifTree := NewTree("if block")
	ifTree.Root = subBlock

	ifDecl.SetIfTree(ifTree)

	elseBlock := NewListNode()
	exitCmd := NewCommandNode(0, "exit")
	elseBlock.Push(exitCmd)

	elseTree := NewTree("else block")
	elseTree.Root = elseBlock

	ifDecl.SetElseTree(elseTree)

	ln.Push(ifDecl)
	expected.Root = ln

	parserTestTable("test if", `if $test == "other" {
	pwd
} else {
	exit
}`, expected, t, true)
}

func TestParseIfElseIf(t *testing.T) {
	expected := NewTree("test if else with variable")
	ln := NewListNode()
	ifDecl := NewIfNode(0)
	ifDecl.SetLvalue(newSimpleArg(3, "$test", ArgVariable))
	ifDecl.SetRvalue(newSimpleArg(13, "other", ArgQuoted))
	ifDecl.SetOp("==")

	subBlock := NewListNode()
	cmd := NewCommandNode(23, "pwd")
	subBlock.Push(cmd)

	ifTree := NewTree("if block")
	ifTree.Root = subBlock

	ifDecl.SetIfTree(ifTree)

	elseIfDecl := NewIfNode(0)

	elseIfDecl.SetLvalue(newSimpleArg(4, "$test", ArgVariable))
	elseIfDecl.SetRvalue(newSimpleArg(15, "others", ArgQuoted))
	elseIfDecl.SetOp("==")

	elseIfBlock := NewListNode()
	elseifCmd := NewCommandNode(25, "ls")
	elseIfBlock.Push(elseifCmd)

	elseIfTree := NewTree("if block")
	elseIfTree.Root = elseIfBlock

	elseIfDecl.SetIfTree(elseIfTree)

	elseBlock := NewListNode()
	exitCmd := NewCommandNode(0, "exit")
	elseBlock.Push(exitCmd)

	elseTree := NewTree("else block")
	elseTree.Root = elseBlock

	elseIfDecl.SetElseTree(elseTree)

	elseBlock2 := NewListNode()
	elseBlock2.Push(elseIfDecl)

	elseTree2 := NewTree("first else tree")
	elseTree2.Root = elseBlock2
	ifDecl.SetElseTree(elseTree2)

	ln.Push(ifDecl)
	expected.Root = ln

	parserTestTable("test if", `if $test == "other" {
	pwd
} else if $test == "others" {
	ls
} else {
	exit
}`, expected, t, true)
}

func TestParseFnBasic(t *testing.T) {
	// root
	expected := NewTree("fn")
	ln := NewListNode()

	// fn
	fn := NewFnDeclNode(0, "build")
	tree := NewTree("fn body")
	lnBody := NewListNode()
	tree.Root = lnBody
	fn.SetTree(tree)

	// root
	ln.Push(fn)
	expected.Root = ln

	parserTestTable("fn", `fn build() {
}`, expected, t, true)

	// root
	expected = NewTree("fn")
	ln = NewListNode()

	// fn
	fn = NewFnDeclNode(0, "build")
	tree = NewTree("fn body")
	lnBody = NewListNode()
	cmd := NewCommandNode(14, "ls")
	lnBody.Push(cmd)
	tree.Root = lnBody
	fn.SetTree(tree)

	// root
	ln.Push(fn)
	expected.Root = ln

	parserTestTable("fn", `fn build() {
	ls
}`, expected, t, true)

	// root
	expected = NewTree("fn")
	ln = NewListNode()

	// fn
	fn = NewFnDeclNode(0, "build")
	fn.AddArg("image")
	tree = NewTree("fn body")
	lnBody = NewListNode()
	cmd = NewCommandNode(19, "ls")
	lnBody.Push(cmd)
	tree.Root = lnBody
	fn.SetTree(tree)

	// root
	ln.Push(fn)
	expected.Root = ln

	parserTestTable("fn", `fn build(image) {
	ls
}`, expected, t, true)

	// root
	expected = NewTree("fn")
	ln = NewListNode()

	// fn
	fn = NewFnDeclNode(0, "build")
	fn.AddArg("image")
	fn.AddArg("debug")
	tree = NewTree("fn body")
	lnBody = NewListNode()
	cmd = NewCommandNode(26, "ls")
	lnBody.Push(cmd)
	tree.Root = lnBody
	fn.SetTree(tree)

	// root
	ln.Push(fn)
	expected.Root = ln

	parserTestTable("fn", `fn build(image, debug) {
	ls
}`, expected, t, true)
}

func TestParseBindFn(t *testing.T) {
	expected := NewTree("bindfn")
	ln := NewListNode()

	bindFn := NewBindFnNode(0, "cd", "cd2")
	ln.Push(bindFn)
	expected.Root = ln

	parserTestTable("bindfn", `bindfn cd cd2`, expected, t, true)
}

func TestParseRedirectionVariable(t *testing.T) {
	expected := NewTree("redirection var")
	ln := NewListNode()

	cmd := NewCommandNode(0, "cmd")
	redir := NewRedirectNode(0)
	redirArg := NewArg(0, ArgVariable)
	redirArg.SetString("$outFname")
	redir.SetLocation(redirArg)
	cmd.AddRedirect(redir)
	ln.Push(cmd)
	expected.Root = ln

	parserTestTable("redir var", `cmd > $outFname`, expected, t, true)
}

func TestParseIssue22(t *testing.T) {
	expected := NewTree("issue 22")
	ln := NewListNode()

	fn := NewFnDeclNode(0, "gocd")
	fn.AddArg("path")

	fnTree := NewTree("fn")
	fnBlock := NewListNode()

	ifDecl := NewIfNode(17)
	patharg := NewArg(20, ArgVariable)
	patharg.SetString("$path")
	ifDecl.SetLvalue(patharg)
	ifDecl.SetOp("==")
	emptyArg := NewArg(30, ArgQuoted)
	emptyArg.SetString("")
	ifDecl.SetRvalue(emptyArg)

	ifTree := NewTree("if")
	ifBlock := NewListNode()

	cdNode := NewCdNode(36)
	arg1 := NewArg(39, ArgVariable)
	arg1.SetString("$GOPATH")
	cdNode.SetDir(arg1)
	ifBlock.Push(cdNode)
	ifTree.Root = ifBlock
	ifDecl.SetIfTree(ifTree)

	elseTree := NewTree("else")
	elseBlock := NewListNode()

	cdNodeElse := NewCdNode(0)
	arg2 := NewArg(0, ArgConcat)

	arg21 := NewArg(0, ArgVariable)
	arg21.SetString("$GOPATH")
	arg22 := NewArg(0, ArgQuoted)
	arg22.SetString("/src/")
	arg23 := NewArg(0, ArgVariable)
	arg23.SetString("$path")

	args := make([]*Arg, 3)
	args[0] = arg21
	args[1] = arg22
	args[2] = arg23

	arg2.SetConcat(args)
	cdNodeElse.SetDir(arg2)
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

func TestParseDump(t *testing.T) {
	expected := NewTree("dump")
	ln := NewListNode()

	dump := NewDumpNode(0)
	arg := NewArg(5, ArgUnquoted)
	arg.SetString("./init")
	dump.SetFilename(arg)
	ln.Push(dump)
	expected.Root = ln

	parserTestTable("dump", `dump ./init`, expected, t, true)
}

func TestParseReturn(t *testing.T) {
	expected := NewTree("return")
	ln := NewListNode()

	ret := NewReturnNode(0)
	ln.Push(ret)
	expected.Root = ln

	parserTestTable("return", `return`, expected, t, true)

	expected = NewTree("return list")
	ln = NewListNode()

	ret = NewReturnNode(0)

	listvalues := make([]*Arg, 2)
	arg1 := NewArg(9, ArgQuoted)
	arg1.SetString("val1")
	arg2 := NewArg(16, ArgQuoted)
	arg2.SetString("val2")
	listvalues[0] = arg1
	listvalues[1] = arg2

	retReturn := NewArg(7, ArgList)
	retReturn.SetList(listvalues)

	ret.SetReturn(retReturn)

	ln.Push(ret)
	expected.Root = ln

	parserTestTable("return", `return ("val1" "val2")`, expected, t, true)

	expected = NewTree("return variable")
	ln = NewListNode()

	ret = NewReturnNode(0)
	arg1 = NewArg(7, ArgVariable)
	arg1.SetString("$var")

	ret.SetReturn(arg1)

	ln.Push(ret)
	expected.Root = ln

	parserTestTable("return", `return $var`, expected, t, true)

	expected = NewTree("return string")
	ln = NewListNode()

	ret = NewReturnNode(0)
	arg1 = NewArg(8, ArgQuoted)
	arg1.SetString("value")

	ret.SetReturn(arg1)

	ln.Push(ret)
	expected.Root = ln

	parserTestTable("return", `return "value"`, expected, t, true)
}

func TestParseIfInvalid(t *testing.T) {
	parser := NewParser("if invalid", `if a == b { pwd }`)
	_, err := parser.Parse()

	if err == nil {
		t.Error("Must fail. Only quoted strings and variables on if clauses.")
		return
	}
}

func TestParseFor(t *testing.T) {
	expected := NewTree("for")

	forStmt := NewForNode(0)
	forTree := NewTree("for block")
	forBlock := NewListNode()
	forTree.Root = forBlock
	forStmt.SetTree(forTree)

	ln := NewListNode()
	ln.Push(forStmt)
	expected.Root = ln

	parserTestTable("for", `for {

}`, expected, t, true)

	forStmt.SetIdentifier("f")
	forStmt.SetInVar("$files")

	parserTestTable("for", `for f in $files {

}`, expected, t, true)
}

func TestParseVariableIndexing(t *testing.T) {
	expected := NewTree("variable indexing")
	ln := NewListNode()

	assignment := NewAssignmentNode(0)
	assignment.SetVarName("test")

	variable := NewArg(7, ArgVariable)
	variable.SetString("$values")

	index := NewArg(0, ArgNumber)
	index.SetString("0")
	variable.SetIndex(index)
	assignment.SetValue(variable)
	ln.Push(assignment)
	expected.Root = ln

	parserTestTable("variable indexing", `test = $values[0]`, expected, t, true)

	ln = NewListNode()

	ifDecl := NewIfNode(0)
	lvalue := NewArg(3, ArgVariable)
	lvalue.SetString("$values")
	index = NewArg(0, ArgNumber)
	index.SetString("0")
	lvalue.SetIndex(index)
	ifDecl.SetLvalue(lvalue)
	ifDecl.SetOp("==")
	rvalue := NewArg(18, ArgQuoted)
	rvalue.SetString("1")
	ifDecl.SetRvalue(rvalue)

	ifBlock := NewTree("if")
	lnBody := NewListNode()
	ifBlock.Root = lnBody
	ifDecl.SetIfTree(ifBlock)

	ln.Push(ifDecl)
	expected.Root = ln

	parserTestTable("variable indexing", `if $values[0] == "1" {

}`, expected, t, true)
}
