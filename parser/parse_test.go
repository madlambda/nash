package parser

import (
	"strings"
	"testing"

	"github.com/NeowayLabs/nash/ast"
)

func parserTestTable(name, content string, expected *ast.Tree, t *testing.T, enableReverse bool) *ast.Tree {
	parser := NewParser(name, content)
	tr, err := parser.Parse()

	if err != nil {
		t.Error(err)
		return nil
	}

	if tr == nil {
		t.Errorf("Failed to parse")
		return nil
	}

	if !expected.IsEqual(tr) {
		t.Errorf("Expected: %s\n\nResult: %s\n", expected, tr)
		return tr
	}

	if !enableReverse {
		return tr
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
	}

	return tr
}

func TestParseSimple(t *testing.T) {
	expected := ast.NewTree("parser simple")
	ln := ast.NewListNode()
	cmd := ast.NewCommandNode(0, "echo")
	cmd.AddArg(ast.NewStringExpr(6, "hello world", true))
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
	first.AddArg(ast.NewStringExpr(6, "hello world", true))

	second := ast.NewCommandNode(21, "awk")
	second.AddArg(ast.NewStringExpr(26, "{print $1}", true))

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
	set := ast.NewSetenvNode(0, "test")

	ln.Push(set)
	expected.Root = ln

	parserTestTable("simple set assignment", `setenv test`, expected, t, true)
}

func TestBasicAssignment(t *testing.T) {
	expected := ast.NewTree("simple assignment")
	ln := ast.NewListNode()
	assign := ast.NewAssignmentNode(0, "test", ast.NewStringExpr(8, "hello", true))
	ln.Push(assign)
	expected.Root = ln

	parserTestTable("simple assignment", `test = "hello"`, expected, t, true)

	// test concatenation of strings and variables

	ln = ast.NewListNode()

	concats := make([]ast.Expr, 2, 2)
	concats[0] = ast.NewStringExpr(8, "hello", true)
	concats[1] = ast.NewVarExpr(15, "$var")

	arg1 := ast.NewConcatExpr(8, concats)

	assign = ast.NewAssignmentNode(0, "test", arg1)

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

	values := make([]ast.Expr, 0, 4)

	values = append(values,
		ast.NewStringExpr(10, "plan9", false),
		ast.NewStringExpr(17, "from", false),
		ast.NewStringExpr(23, "bell", false),
		ast.NewStringExpr(29, "labs", false),
	)

	elem := ast.NewListExpr(7, values)

	assign := ast.NewAssignmentNode(0, "test", elem)

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

	cmd := ast.NewCommandNode(8, "ls")

	assign, err := ast.NewExecAssignNode(0, "test", cmd)

	if err != nil {
		t.Error(err)
		return
	}

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
	cmd.AddArg(ast.NewStringExpr(11, "hello world", true))
	ln.Push(cmd)

	expected.Root = ln

	parserTestTable("parser simple", `/bin/echo "hello world"`, expected, t, true)
}

func TestParseWithShebang(t *testing.T) {
	expected := ast.NewTree("parser shebang")
	ln := ast.NewListNode()
	cmt := ast.NewCommentNode(0, "#!/bin/nash")
	cmd := ast.NewCommandNode(12, "echo")
	cmd.AddArg(ast.NewStringExpr(17, "bleh", false))
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
	redir.SetLocation(ast.NewStringExpr(0, "/var/log/service.log", false))
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
	cmd1.AddArg(ast.NewStringExpr(5, "hello", false))
	cmd2.AddArg(ast.NewStringExpr(17, "hello", true))

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
	cd := ast.NewCdNode(0, ast.NewStringExpr(3, "/tmp", false))
	ln.Push(cd)
	expected.Root = ln

	parserTestTable("test cd", "cd /tmp", expected, t, true)

	// test cd into home
	expected = ast.NewTree("test cd into home")
	ln = ast.NewListNode()
	cd = ast.NewCdNode(0, nil)
	ln.Push(cd)
	expected.Root = ln

	parserTestTable("test cd into home", "cd", expected, t, true)

	expected = ast.NewTree("cd into HOME by setenv")
	ln = ast.NewListNode()

	assign := ast.NewAssignmentNode(0, "HOME", ast.NewStringExpr(8, "/", true))

	set := ast.NewSetenvNode(11, "HOME")
	cd = ast.NewCdNode(23, nil)
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

	arg := ast.NewStringExpr(10, "/home/i4k/gopath", true)

	assign = ast.NewAssignmentNode(0, "GOPATH", arg)
	cd = ast.NewCdNode(28, ast.NewVarExpr(31, "$GOPATH"))

	ln.Push(assign)
	ln.Push(cd)

	expected.Root = ln

	parserTestTable("test cd into variable value", `GOPATH = "/home/i4k/gopath"
cd $GOPATH`, expected, t, true)

	// Test cd into custom variable
	expected = ast.NewTree("cd into variable value with concat")
	ln = ast.NewListNode()

	arg = ast.NewStringExpr(10, "/home/i4k/gopath", true)

	assign = ast.NewAssignmentNode(0, "GOPATH", arg)

	concat := make([]ast.Expr, 0, 2)
	concat = append(concat, ast.NewVarExpr(31, "$GOPATH"))
	concat = append(concat, ast.NewStringExpr(40, "/src/github.com", true))

	cd = ast.NewCdNode(28, ast.NewConcatExpr(31, concat))

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
	f1 := ast.NewStringExpr(6, "u", false)
	cmd1.SetFlags(f1)
	ln.Push(cmd1)
	expected.Root = ln

	parserTestTable("test rfork", "rfork u", expected, t, true)
}

func TestParseRforkWithBlock(t *testing.T) {
	expected := ast.NewTree("rfork with block")
	ln := ast.NewListNode()
	rfork := ast.NewRforkNode(0)
	arg := ast.NewStringExpr(6, "u", false)
	rfork.SetFlags(arg)

	insideFork := ast.NewCommandNode(11, "mount")
	insideFork.AddArg(ast.NewStringExpr(17, "-t", false))
	insideFork.AddArg(ast.NewStringExpr(20, "proc", false))
	insideFork.AddArg(ast.NewStringExpr(25, "proc", false))
	insideFork.AddArg(ast.NewStringExpr(30, "/proc", false))

	bln := ast.NewListNode()
	bln.Push(insideFork)
	subtree := ast.NewTree("rfork")
	subtree.Root = bln

	rfork.SetTree(subtree)

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
	importStmt := ast.NewImportNode(0, ast.NewStringExpr(7, "env.sh", false))
	ln.Push(importStmt)
	expected.Root = ln

	parserTestTable("test import", "import env.sh", expected, t, true)

	expected = ast.NewTree("test import with quotes")
	ln = ast.NewListNode()
	importStmt = ast.NewImportNode(0, ast.NewStringExpr(8, "env.sh", true))
	ln.Push(importStmt)
	expected.Root = ln

	parserTestTable("test import", `import "env.sh"`, expected, t, true)
}

func TestParseIf(t *testing.T) {
	expected := ast.NewTree("test if")
	ln := ast.NewListNode()
	ifDecl := ast.NewIfNode(0)
	ifDecl.SetLvalue(ast.NewStringExpr(4, "test", true))
	ifDecl.SetRvalue(ast.NewStringExpr(14, "other", true))
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
	ifDecl.SetLvalue(ast.NewStringExpr(4, "", true))
	ifDecl.SetRvalue(ast.NewStringExpr(10, "other", true))
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
	ifDecl.SetLvalue(ast.NewVarExpr(3, "$test"))
	ifDecl.SetRvalue(ast.NewStringExpr(13, "other", true))
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
	ifDecl.SetLvalue(ast.NewVarExpr(3, "$test"))
	ifDecl.SetRvalue(ast.NewStringExpr(13, "other", true))
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
	ifDecl.SetLvalue(ast.NewVarExpr(3, "$test"))
	ifDecl.SetRvalue(ast.NewStringExpr(13, "other", true))
	ifDecl.SetOp("==")

	subBlock := ast.NewListNode()
	cmd := ast.NewCommandNode(23, "pwd")
	subBlock.Push(cmd)

	ifTree := ast.NewTree("if block")
	ifTree.Root = subBlock

	ifDecl.SetIfTree(ifTree)

	elseIfDecl := ast.NewIfNode(0)

	elseIfDecl.SetLvalue(ast.NewVarExpr(4, "$test"))
	elseIfDecl.SetRvalue(ast.NewStringExpr(15, "others", true))
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
	redirArg := ast.NewVarExpr(0, "$outFname")
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
	ifDecl.SetLvalue(ast.NewVarExpr(20, "$path"))
	ifDecl.SetOp("==")

	ifDecl.SetRvalue(ast.NewStringExpr(30, "", true))

	ifTree := ast.NewTree("if")
	ifBlock := ast.NewListNode()

	cdNode := ast.NewCdNode(36, ast.NewVarExpr(39, "$GOPATH"))

	ifBlock.Push(cdNode)
	ifTree.Root = ifBlock
	ifDecl.SetIfTree(ifTree)

	elseTree := ast.NewTree("else")
	elseBlock := ast.NewListNode()

	args := make([]ast.Expr, 3)
	args[0] = ast.NewVarExpr(0, "$GOPATH")
	args[1] = ast.NewStringExpr(0, "/src/", true)
	args[2] = ast.NewVarExpr(0, "$path")

	cdNodeElse := ast.NewCdNode(0, ast.NewConcatExpr(0, args))

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
	dump.SetFilename(ast.NewStringExpr(5, "./init", false))
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

	listvalues := make([]ast.Expr, 2)

	listvalues[0] = ast.NewStringExpr(9, "val1", true)
	listvalues[1] = ast.NewStringExpr(16, "val2", true)

	retReturn := ast.NewListExpr(7, listvalues)

	ret.SetReturn(retReturn)

	ln.Push(ret)
	expected.Root = ln

	parserTestTable("return", `return ("val1" "val2")`, expected, t, true)

	expected = ast.NewTree("return variable")
	ln = ast.NewListNode()

	ret = ast.NewReturnNode(0)

	ret.SetReturn(ast.NewVarExpr(7, "$var"))

	ln.Push(ret)
	expected.Root = ln

	parserTestTable("return", `return $var`, expected, t, true)

	expected = ast.NewTree("return string")
	ln = ast.NewListNode()

	ret = ast.NewReturnNode(0)

	ret.SetReturn(ast.NewStringExpr(8, "value", true))

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

	indexedVar := ast.NewIndexExpr(
		7,
		ast.NewVarExpr(7, "$values"),
		ast.NewIntExpr(0, 0),
	)

	assignment := ast.NewAssignmentNode(0, "test", indexedVar)
	ln.Push(assignment)
	expected.Root = ln

	parserTestTable("variable indexing", `test = $values[0]`, expected, t, true)

	ln = ast.NewListNode()

	ifDecl := ast.NewIfNode(0)
	lvalue := ast.NewVarExpr(3, "$values")

	indexedVar = ast.NewIndexExpr(3, lvalue, ast.NewIntExpr(0, 0))

	ifDecl.SetLvalue(indexedVar)
	ifDecl.SetOp("==")
	ifDecl.SetRvalue(ast.NewStringExpr(18, "1", true))

	ifBlock := ast.NewTree("if")
	lnBody := ast.NewListNode()
	ifBlock.Root = lnBody
	ifDecl.SetIfTree(ifBlock)

	ln.Push(ifDecl)
	expected.Root = ln

	parserTestTable("variable indexing", `if $values[0] == "1" {

}`, expected, t, true)
}
