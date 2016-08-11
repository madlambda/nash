package parser

import (
	"fmt"
	"strings"
	"testing"

	"github.com/NeowayLabs/nash/ast"
)

func parserTestTable(name, content string, expected *ast.Tree, t *testing.T, enableReverse bool) {
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
'%s'

But got:
'%s'
`, content, trcontent)
		return
	}
}

func TestParseSimple(t *testing.T) {
	expected := ast.NewTree("parser simple")
	ln := ast.NewListNode()
	cmd := ast.NewCommandNode(0, "echo")
	hello := ast.NewArg(6, ast.ArgQuoted)
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
	expected := ast.NewTree("parser pipe")
	ln := ast.NewListNode()
	first := ast.NewCommandNode(0, "echo")
	hello := ast.NewArg(6, ast.ArgQuoted)
	hello.SetString("hello world")
	first.AddArg(hello)

	second := ast.NewCommandNode(21, "awk")
	secondArg := ast.NewArg(26, ast.ArgQuoted)
	secondArg.SetString("{print $1}")

	second.AddArg(secondArg)

	pipe := ast.NewPipeNode(19)
	pipe.AddCmd(first)
	pipe.AddCmd(second)

	ln.Push(pipe)

	expected.Root = ln

	parserTestTable("parser pipe", `echo "hello world" | awk "{print $1}"`, expected, t, true)
}

func TestBasicSetAssignment(t *testing.T) {
	expected := ast.NewTree("simple set assignment")
	ln := ast.NewListNode()
	set := ast.NewSetAssignmentNode(0, "test")

	ln.Push(set)
	expected.Root = ln

	parserTestTable("simple set assignment", `setenv test`, expected, t, true)
}

func TestBasicAssignment(t *testing.T) {
	expected := ast.NewTree("simple assignment")
	ln := ast.NewListNode()
	assign := ast.NewAssignmentNode(0)
	assign.SetIdentifier("test")

	elem := ast.NewArg(8, ast.ArgQuoted)
	elem.SetString("hello")

	assign.SetValue(elem)
	ln.Push(assign)
	expected.Root = ln

	parserTestTable("simple assignment", `test = "hello"`, expected, t, true)

	// test concatenation of strings and variables

	ln = ast.NewListNode()
	assign = ast.NewAssignmentNode(0)
	assign.SetIdentifier("test")

	concats := make([]*ast.Arg, 2, 2)

	hello := ast.NewArg(8, ast.ArgQuoted)
	hello.SetString("hello")
	variable := ast.NewArg(15, ast.ArgVariable)
	variable.SetString("$var")

	concats[0] = hello
	concats[1] = variable
	arg1 := ast.NewArg(8, ast.ArgConcat)
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
	expected := ast.NewTree("list assignment")
	ln := ast.NewListNode()
	assign := ast.NewAssignmentNode(0)

	plan9 := ast.NewArg(10, ast.ArgUnquoted)
	plan9.SetString("plan9")
	from := ast.NewArg(17, ast.ArgUnquoted)
	from.SetString("from")
	bell := ast.NewArg(23, ast.ArgUnquoted)
	bell.SetString("bell")
	labs := ast.NewArg(29, ast.ArgUnquoted)
	labs.SetString("labs")

	values := make([]*ast.Arg, 4)
	values[0] = plan9
	values[1] = from
	values[2] = bell
	values[3] = labs

	assign.SetIdentifier("test")

	elem := ast.NewArg(7, ast.ArgList)
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
	expected := ast.NewTree("simple cmd assignment")
	ln := ast.NewListNode()
	assign := ast.NewCmdAssignmentNode(0, "test")

	cmd := ast.NewCommandNode(8, "ls")
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
	expected := ast.NewTree("parser simple")
	ln := ast.NewListNode()
	cmd := ast.NewCommandNode(0, "/bin/echo")
	arg := ast.NewArg(11, ast.ArgQuoted)
	arg.SetString("hello world")
	cmd.AddArg(arg)
	ln.Push(cmd)

	expected.Root = ln

	parserTestTable("parser simple", `/bin/echo "hello world"`, expected, t, true)
}

func TestParseWithShebang(t *testing.T) {
	expected := ast.NewTree("parser shebang")
	ln := ast.NewListNode()
	cmt := ast.NewCommentNode(0, "#!/bin/nash")
	cmd := ast.NewCommandNode(12, "echo")
	arg := ast.NewArg(17, ast.ArgUnquoted)
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
	expected := ast.NewTree("empty file")
	ln := ast.NewListNode()
	expected.Root = ln

	parserTestTable("empty file", "", expected, t, true)
}

func TestParseSingleCommand(t *testing.T) {
	expected := ast.NewTree("single command")
	expected.Root = ast.NewListNode()
	expected.Root.Push(ast.NewCommandNode(0, "bleh"))

	parserTestTable("single command", `bleh`, expected, t, true)
}

func TestParseRedirectSimple(t *testing.T) {
	expected := ast.NewTree("redirect")
	ln := ast.NewListNode()
	cmd := ast.NewCommandNode(0, "cmd")
	redir := ast.NewRedirectNode(0)
	redir.SetMap(2, ast.RedirMapSupress)
	cmd.AddRedirect(redir)
	ln.Push(cmd)

	expected.Root = ln

	parserTestTable("simple redirect", `cmd >[2=]`, expected, t, true)

	expected = ast.NewTree("redirect2")
	ln = ast.NewListNode()
	cmd = ast.NewCommandNode(0, "cmd")
	redir = ast.NewRedirectNode(0)
	redir.SetMap(2, 1)
	cmd.AddRedirect(redir)
	ln.Push(cmd)

	expected.Root = ln

	parserTestTable("simple redirect", `cmd >[2=1]`, expected, t, true)
}

func TestParseRedirectWithLocation(t *testing.T) {
	expected := ast.NewTree("redirect with location")
	ln := ast.NewListNode()
	cmd := ast.NewCommandNode(0, "cmd")
	redir := ast.NewRedirectNode(0)
	redir.SetMap(2, ast.RedirMapNoValue)
	out := ast.NewArg(0, ast.ArgUnquoted)
	out.SetString("/var/log/service.log")
	redir.SetLocation(out)
	cmd.AddRedirect(redir)
	ln.Push(cmd)

	expected.Root = ln

	parserTestTable("simple redirect", `cmd >[2] /var/log/service.log`, expected, t, true)
}

func TestParseRedirectMultiples(t *testing.T) {
	expected := ast.NewTree("redirect multiples")
	ln := ast.NewListNode()
	cmd := ast.NewCommandNode(0, "cmd")
	redir1 := ast.NewRedirectNode(0)
	redir2 := ast.NewRedirectNode(0)

	redir1.SetMap(1, 2)
	redir2.SetMap(2, ast.RedirMapSupress)

	cmd.AddRedirect(redir1)
	cmd.AddRedirect(redir2)
	ln.Push(cmd)

	expected.Root = ln

	parserTestTable("multiple redirects", `cmd >[1=2] >[2=]`, expected, t, true)
}

func TestParseCommandWithStringsEqualsNot(t *testing.T) {
	expected := ast.NewTree("strings works as expected")
	ln := ast.NewListNode()
	cmd1 := ast.NewCommandNode(0, "echo")
	cmd2 := ast.NewCommandNode(11, "echo")
	hello := ast.NewArg(5, ast.ArgUnquoted)
	hello.SetString("hello")
	cmd1.AddArg(hello)

	hello2 := ast.NewArg(17, ast.ArgQuoted)
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
	expected := ast.NewTree("test cd")
	ln := ast.NewListNode()
	cd := ast.NewCdNode(0)
	arg := ast.NewArg(3, ast.ArgUnquoted)
	arg.SetString("/tmp")
	cd.SetDir(arg)
	ln.Push(cd)
	expected.Root = ln

	parserTestTable("test cd", "cd /tmp", expected, t, true)

	// test cd into home
	expected = ast.NewTree("test cd into home")
	ln = ast.NewListNode()
	cd = ast.NewCdNode(0)
	ln.Push(cd)
	expected.Root = ln

	parserTestTable("test cd into home", "cd", expected, t, true)

	expected = ast.NewTree("cd into HOME by setenv")
	ln = ast.NewListNode()
	assign := ast.NewAssignmentNode(0)
	assign.SetIdentifier("HOME")
	arg = ast.NewArg(8, ast.ArgQuoted)
	arg.SetString("/")

	assign.SetValue(arg)

	set := ast.NewSetAssignmentNode(11, "HOME")
	cd = ast.NewCdNode(23)
	pwd := ast.NewCommandNode(26, "pwd")

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
	expected = ast.NewTree("cd into variable value")
	ln = ast.NewListNode()
	assign = ast.NewAssignmentNode(0)
	assign.SetIdentifier("GOPATH")
	arg = ast.NewArg(10, ast.ArgQuoted)
	arg.SetString("/home/i4k/gopath")

	assign.SetValue(arg)
	cd = ast.NewCdNode(28)
	path := ast.NewArg(31, ast.ArgVariable)
	path.SetString("$GOPATH")
	cd.SetDir(path)

	ln.Push(assign)
	ln.Push(cd)

	expected.Root = ln

	parserTestTable("test cd into variable value", `GOPATH = "/home/i4k/gopath"
cd $GOPATH`, expected, t, true)

	// Test cd into custom variable
	expected = ast.NewTree("cd into variable value with concat")
	ln = ast.NewListNode()
	assign = ast.NewAssignmentNode(0)
	assign.SetIdentifier("GOPATH")
	arg = ast.NewArg(10, ast.ArgQuoted)
	arg.SetString("/home/i4k/gopath")

	assign.SetValue(arg)
	cd = ast.NewCdNode(28)
	path = ast.NewArg(31, ast.ArgConcat)
	varg := ast.NewArg(31, ast.ArgVariable)
	varg.SetString("$GOPATH")
	src := ast.NewArg(40, ast.ArgQuoted)
	src.SetString("/src/github.com")
	concat := make([]*ast.Arg, 0, 2)
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
	expected := ast.NewTree("test rfork")
	ln := ast.NewListNode()
	cmd1 := ast.NewRforkNode(0)
	f1 := ast.NewArg(6, ast.ArgUnquoted)
	f1.SetString("u")
	cmd1.SetFlags(f1)
	ln.Push(cmd1)
	expected.Root = ln

	parserTestTable("test rfork", "rfork u", expected, t, true)
}

func TestParseRforkWithBlock(t *testing.T) {
	expected := ast.NewTree("rfork with block")
	ln := ast.NewListNode()
	rfork := ast.NewRforkNode(0)
	arg := ast.NewArg(6, ast.ArgUnquoted)
	arg.SetString("u")
	rfork.SetFlags(arg)

	insideFork := ast.NewCommandNode(11, "mount")
	insideFork.AddArg(ast.NewSimpleArg(17, "-t", ast.ArgUnquoted))
	insideFork.AddArg(ast.NewSimpleArg(20, "proc", ast.ArgUnquoted))
	insideFork.AddArg(ast.NewSimpleArg(25, "proc", ast.ArgUnquoted))
	insideFork.AddArg(ast.NewSimpleArg(30, "/proc", ast.ArgUnquoted))

	bln := ast.NewListNode()
	bln.Push(insideFork)
	subtree := ast.NewTree("rfork")
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
	expected := ast.NewTree("test import")
	ln := ast.NewListNode()
	importStmt := ast.NewImportNode(0)
	importStmt.SetPath(ast.NewSimpleArg(7, "env.sh", ast.ArgUnquoted))
	ln.Push(importStmt)
	expected.Root = ln

	parserTestTable("test import", "import env.sh", expected, t, true)

	expected = ast.NewTree("test import with quotes")
	ln = ast.NewListNode()
	importStmt = ast.NewImportNode(0)
	importStmt.SetPath(ast.NewSimpleArg(8, "env.sh", ast.ArgQuoted))
	ln.Push(importStmt)
	expected.Root = ln

	parserTestTable("test import", `import "env.sh"`, expected, t, true)
}

func TestParseIf(t *testing.T) {
	expected := ast.NewTree("test if")
	ln := ast.NewListNode()
	ifDecl := ast.NewIfNode(0)
	ifDecl.SetLvalue(ast.NewSimpleArg(4, "test", ast.ArgQuoted))
	ifDecl.SetRvalue(ast.NewSimpleArg(14, "other", ast.ArgQuoted))
	ifDecl.SetOp("==")

	subBlock := ast.NewListNode()
	cmd := ast.NewCommandNode(24, "pwd")
	subBlock.Push(cmd)

	ifTree := ast.NewTree("if block")
	ifTree.Root = subBlock

	ifDecl.SetIfTree(ifTree)

	ln.Push(ifDecl)
	expected.Root = ln

	parserTestTable("test if", `if "test" == "other" {
	pwd
}`, expected, t, true)

	expected = ast.NewTree("test if")
	ln = ast.NewListNode()
	ifDecl = ast.NewIfNode(0)
	ifDecl.SetLvalue(ast.NewSimpleArg(4, "", ast.ArgQuoted))
	ifDecl.SetRvalue(ast.NewSimpleArg(10, "other", ast.ArgQuoted))
	ifDecl.SetOp("!=")

	subBlock = ast.NewListNode()
	cmd = ast.NewCommandNode(20, "pwd")
	subBlock.Push(cmd)

	ifTree = ast.NewTree("if block")
	ifTree.Root = subBlock

	ifDecl.SetIfTree(ifTree)

	ln.Push(ifDecl)
	expected.Root = ln

	parserTestTable("test if", `if "" != "other" {
	pwd
}`, expected, t, true)
}

func TestParseIfLvariable(t *testing.T) {
	expected := ast.NewTree("test if with variable")
	ln := ast.NewListNode()
	ifDecl := ast.NewIfNode(0)
	ifDecl.SetLvalue(ast.NewSimpleArg(3, "$test", ast.ArgVariable))
	ifDecl.SetRvalue(ast.NewSimpleArg(13, "other", ast.ArgQuoted))
	ifDecl.SetOp("==")

	subBlock := ast.NewListNode()
	cmd := ast.NewCommandNode(23, "pwd")
	subBlock.Push(cmd)

	ifTree := ast.NewTree("if block")
	ifTree.Root = subBlock

	ifDecl.SetIfTree(ifTree)

	ln.Push(ifDecl)
	expected.Root = ln

	parserTestTable("test if", `if $test == "other" {
	pwd
}`, expected, t, true)
}

func TestParseIfElse(t *testing.T) {
	expected := ast.NewTree("test if else with variable")
	ln := ast.NewListNode()
	ifDecl := ast.NewIfNode(0)
	ifDecl.SetLvalue(ast.NewSimpleArg(3, "$test", ast.ArgVariable))
	ifDecl.SetRvalue(ast.NewSimpleArg(13, "other", ast.ArgQuoted))
	ifDecl.SetOp("==")

	subBlock := ast.NewListNode()
	cmd := ast.NewCommandNode(23, "pwd")
	subBlock.Push(cmd)

	ifTree := ast.NewTree("if block")
	ifTree.Root = subBlock

	ifDecl.SetIfTree(ifTree)

	elseBlock := ast.NewListNode()
	exitCmd := ast.NewCommandNode(0, "exit")
	elseBlock.Push(exitCmd)

	elseTree := ast.NewTree("else block")
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
	expected := ast.NewTree("test if else with variable")
	ln := ast.NewListNode()
	ifDecl := ast.NewIfNode(0)
	ifDecl.SetLvalue(ast.NewSimpleArg(3, "$test", ast.ArgVariable))
	ifDecl.SetRvalue(ast.NewSimpleArg(13, "other", ast.ArgQuoted))
	ifDecl.SetOp("==")

	subBlock := ast.NewListNode()
	cmd := ast.NewCommandNode(23, "pwd")
	subBlock.Push(cmd)

	ifTree := ast.NewTree("if block")
	ifTree.Root = subBlock

	ifDecl.SetIfTree(ifTree)

	elseIfDecl := ast.NewIfNode(0)

	elseIfDecl.SetLvalue(ast.NewSimpleArg(4, "$test", ast.ArgVariable))
	elseIfDecl.SetRvalue(ast.NewSimpleArg(15, "others", ast.ArgQuoted))
	elseIfDecl.SetOp("==")

	elseIfBlock := ast.NewListNode()
	elseifCmd := ast.NewCommandNode(25, "ls")
	elseIfBlock.Push(elseifCmd)

	elseIfTree := ast.NewTree("if block")
	elseIfTree.Root = elseIfBlock

	elseIfDecl.SetIfTree(elseIfTree)

	elseBlock := ast.NewListNode()
	exitCmd := ast.NewCommandNode(0, "exit")
	elseBlock.Push(exitCmd)

	elseTree := ast.NewTree("else block")
	elseTree.Root = elseBlock

	elseIfDecl.SetElseTree(elseTree)

	elseBlock2 := ast.NewListNode()
	elseBlock2.Push(elseIfDecl)

	elseTree2 := ast.NewTree("first else tree")
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
	expected := ast.NewTree("fn")
	ln := ast.NewListNode()

	// fn
	fn := ast.NewFnDeclNode(0, "build")
	tree := ast.NewTree("fn body")
	lnBody := ast.NewListNode()
	tree.Root = lnBody
	fn.SetTree(tree)

	// root
	ln.Push(fn)
	expected.Root = ln

	parserTestTable("fn", `fn build() {
}`, expected, t, true)

	// root
	expected = ast.NewTree("fn")
	ln = ast.NewListNode()

	// fn
	fn = ast.NewFnDeclNode(0, "build")
	tree = ast.NewTree("fn body")
	lnBody = ast.NewListNode()
	cmd := ast.NewCommandNode(14, "ls")
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
	expected = ast.NewTree("fn")
	ln = ast.NewListNode()

	// fn
	fn = ast.NewFnDeclNode(0, "build")
	fn.AddArg("image")
	tree = ast.NewTree("fn body")
	lnBody = ast.NewListNode()
	cmd = ast.NewCommandNode(19, "ls")
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
	expected = ast.NewTree("fn")
	ln = ast.NewListNode()

	// fn
	fn = ast.NewFnDeclNode(0, "build")
	fn.AddArg("image")
	fn.AddArg("debug")
	tree = ast.NewTree("fn body")
	lnBody = ast.NewListNode()
	cmd = ast.NewCommandNode(26, "ls")
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
	expected := ast.NewTree("bindfn")
	ln := ast.NewListNode()

	bindFn := ast.NewBindFnNode(0, "cd", "cd2")
	ln.Push(bindFn)
	expected.Root = ln

	parserTestTable("bindfn", `bindfn cd cd2`, expected, t, true)
}

func TestParseRedirectionVariable(t *testing.T) {
	expected := ast.NewTree("redirection var")
	ln := ast.NewListNode()

	cmd := ast.NewCommandNode(0, "cmd")
	redir := ast.NewRedirectNode(0)
	redirArg := ast.NewArg(0, ast.ArgVariable)
	redirArg.SetString("$outFname")
	redir.SetLocation(redirArg)
	cmd.AddRedirect(redir)
	ln.Push(cmd)
	expected.Root = ln

	parserTestTable("redir var", `cmd > $outFname`, expected, t, true)
}

func TestParseIssue22(t *testing.T) {
	expected := ast.NewTree("issue 22")
	ln := ast.NewListNode()

	fn := ast.NewFnDeclNode(0, "gocd")
	fn.AddArg("path")

	fnTree := ast.NewTree("fn")
	fnBlock := ast.NewListNode()

	ifDecl := ast.NewIfNode(17)
	patharg := ast.NewArg(20, ast.ArgVariable)
	patharg.SetString("$path")
	ifDecl.SetLvalue(patharg)
	ifDecl.SetOp("==")
	emptyArg := ast.NewArg(30, ast.ArgQuoted)
	emptyArg.SetString("")
	ifDecl.SetRvalue(emptyArg)

	ifTree := ast.NewTree("if")
	ifBlock := ast.NewListNode()

	cdNode := ast.NewCdNode(36)
	arg1 := ast.NewArg(39, ast.ArgVariable)
	arg1.SetString("$GOPATH")
	cdNode.SetDir(arg1)
	ifBlock.Push(cdNode)
	ifTree.Root = ifBlock
	ifDecl.SetIfTree(ifTree)

	elseTree := ast.NewTree("else")
	elseBlock := ast.NewListNode()

	cdNodeElse := ast.NewCdNode(0)
	arg2 := ast.NewArg(0, ast.ArgConcat)

	arg21 := ast.NewArg(0, ast.ArgVariable)
	arg21.SetString("$GOPATH")
	arg22 := ast.NewArg(0, ast.ArgQuoted)
	arg22.SetString("/src/")
	arg23 := ast.NewArg(0, ast.ArgVariable)
	arg23.SetString("$path")

	args := make([]*ast.Arg, 3)
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
	expected := ast.NewTree("dump")
	ln := ast.NewListNode()

	dump := ast.NewDumpNode(0)
	arg := ast.NewArg(5, ast.ArgUnquoted)
	arg.SetString("./init")
	dump.SetFilename(arg)
	ln.Push(dump)
	expected.Root = ln

	parserTestTable("dump", `dump ./init`, expected, t, true)
}

func TestParseReturn(t *testing.T) {
	expected := ast.NewTree("return")
	ln := ast.NewListNode()

	ret := ast.NewReturnNode(0)
	ln.Push(ret)
	expected.Root = ln

	parserTestTable("return", `return`, expected, t, true)

	expected = ast.NewTree("return list")
	ln = ast.NewListNode()

	ret = ast.NewReturnNode(0)

	listvalues := make([]*ast.Arg, 2)
	arg1 := ast.NewArg(9, ast.ArgQuoted)
	arg1.SetString("val1")
	arg2 := ast.NewArg(16, ast.ArgQuoted)
	arg2.SetString("val2")
	listvalues[0] = arg1
	listvalues[1] = arg2

	retReturn := ast.NewArg(7, ast.ArgList)
	retReturn.SetList(listvalues)

	ret.SetReturn(retReturn)

	ln.Push(ret)
	expected.Root = ln

	parserTestTable("return", `return ("val1" "val2")`, expected, t, true)

	expected = ast.NewTree("return variable")
	ln = ast.NewListNode()

	ret = ast.NewReturnNode(0)
	arg1 = ast.NewArg(7, ast.ArgVariable)
	arg1.SetString("$var")

	ret.SetReturn(arg1)

	ln.Push(ret)
	expected.Root = ln

	parserTestTable("return", `return $var`, expected, t, true)

	expected = ast.NewTree("return string")
	ln = ast.NewListNode()

	ret = ast.NewReturnNode(0)
	arg1 = ast.NewArg(8, ast.ArgQuoted)
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
	expected := ast.NewTree("for")

	forStmt := ast.NewForNode(0)
	forTree := ast.NewTree("for block")
	forBlock := ast.NewListNode()
	forTree.Root = forBlock
	forStmt.SetTree(forTree)

	ln := ast.NewListNode()
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
	expected := ast.NewTree("variable indexing")
	ln := ast.NewListNode()

	assignment := ast.NewAssignmentNode(0)
	assignment.SetIdentifier("test")

	variable := ast.NewArg(7, ast.ArgVariable)
	variable.SetString("$values")

	index := ast.NewArg(0, ast.ArgNumber)
	index.SetString("0")
	variable.SetIndex(index)
	assignment.SetValue(variable)
	ln.Push(assignment)
	expected.Root = ln

	parserTestTable("variable indexing", `test = $values[0]`, expected, t, true)

	ln = ast.NewListNode()

	ifDecl := ast.NewIfNode(0)
	lvalue := ast.NewArg(3, ast.ArgVariable)
	lvalue.SetString("$values")
	index = ast.NewArg(0, ast.ArgNumber)
	index.SetString("0")
	lvalue.SetIndex(index)
	ifDecl.SetLvalue(lvalue)
	ifDecl.SetOp("==")
	rvalue := ast.NewArg(18, ast.ArgQuoted)
	rvalue.SetString("1")
	ifDecl.SetRvalue(rvalue)

	ifBlock := ast.NewTree("if")
	lnBody := ast.NewListNode()
	ifBlock.Root = lnBody
	ifDecl.SetIfTree(ifBlock)

	ln.Push(ifDecl)
	expected.Root = ln

	parserTestTable("variable indexing", `if $values[0] == "1" {

}`, expected, t, true)
}
