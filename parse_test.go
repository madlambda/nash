package nash

import (
	"fmt"
	"strings"
	"testing"
)

func parserTestTable(name, content string, expected *Tree, t *testing.T) {
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

	// Test if the reverse of tree is the content again... *hard*
	trcontent := tr.String()
	content = strings.Trim(content, "\n ")

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

func TestParseShowEnv(t *testing.T) {
	expected := NewTree("parse showenv")
	ln := NewListNode()
	showenv := NewShowEnvNode(0)
	ln.Push(showenv)

	expected.Root = ln

	parserTestTable("parse showenv", `showenv`, expected, t)
}

func TestParseSimple(t *testing.T) {
	expected := NewTree("parser simple")
	ln := NewListNode()
	cmd := NewCommandNode(0, "echo")
	cmd.AddArg(NewArg(6, "hello world", true))
	ln.Push(cmd)

	expected.Root = ln

	parserTestTable("parser simple", `echo "hello world"`, expected, t)
}

func TestParseReverseGetSame(t *testing.T) {
	parser := NewParser("reverse simple", "echo \"hello world\"")

	tr, err := parser.Parse()

	if err != nil {
		t.Error(err)
		return
	}

	if tr.String() != "echo \"hello world\"" {
		t.Error("Failed to reverse tree: %s", tr.String())
		return
	}
}

func TestBasicSetAssignment(t *testing.T) {
	expected := NewTree("simple set assignment")
	ln := NewListNode()
	set := NewSetAssignmentNode(0, "test")

	ln.Push(set)
	expected.Root = ln

	parserTestTable("simple set assignment", `setenv test`, expected, t)
}

func TestBasicAssignment(t *testing.T) {
	expected := NewTree("simple assignment")
	ln := NewListNode()
	assign := NewAssignmentNode(0)
	assign.SetVarName("test")
	elems := make([]ElemNode, 1, 1)
	elems[0] = ElemNode{
		elem: "hello",
	}

	assign.SetValueList(elems)
	ln.Push(assign)
	expected.Root = ln

	parserTestTable("simple assignment", `test="hello"`, expected, t)

	// test concatenation of strings and variables

	ln = NewListNode()
	assign = NewAssignmentNode(0)
	assign.SetVarName("test")
	elems = make([]ElemNode, 1, 1)
	concats := make([]string, 2, 2)
	concats[0] = "hello"
	concats[1] = "$var"
	elems[0] = ElemNode{
		concats: concats,
	}

	assign.SetValueList(elems)
	ln.Push(assign)
	expected.Root = ln

	parserTestTable("test", `test="hello" + $var`, expected, t)

	// invalid, requires quote
	// test=hello
	parser := NewParser("", `test=hello`)

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
	cmd.AddArg(NewArg(11, "hello world", true))
	ln.Push(cmd)

	expected.Root = ln

	parserTestTable("parser simple", `/bin/echo "hello world"`, expected, t)
}

func TestParseWithShebang(t *testing.T) {
	expected := NewTree("parser shebang")
	ln := NewListNode()
	cmt := NewCommentNode(0, "#!/bin/nash")
	cmd := NewCommandNode(12, "echo")
	cmd.AddArg(NewArg(17, "bleh", false))
	ln.Push(cmt)
	ln.Push(cmd)

	expected.Root = ln

	parserTestTable("parser shebang", `#!/bin/nash
echo bleh
`, expected, t)
}

func TestParseEmptyFile(t *testing.T) {
	expected := NewTree("empty file")
	ln := NewListNode()
	expected.Root = ln

	parserTestTable("empty file", "", expected, t)
}

func TestParseSingleCommand(t *testing.T) {
	expected := NewTree("single command")
	expected.Root = NewListNode()
	expected.Root.Push(NewCommandNode(0, "bleh"))

	parserTestTable("single command", `bleh`, expected, t)
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

	parserTestTable("simple redirect", `cmd >[2=]`, expected, t)

	expected = NewTree("redirect2")
	ln = NewListNode()
	cmd = NewCommandNode(0, "cmd")
	redir = NewRedirectNode(0)
	redir.SetMap(2, 1)
	cmd.AddRedirect(redir)
	ln.Push(cmd)

	expected.Root = ln

	parserTestTable("simple redirect", `cmd >[2=1]`, expected, t)
}

func TestParseRedirectWithLocation(t *testing.T) {
	expected := NewTree("redirect with location")
	ln := NewListNode()
	cmd := NewCommandNode(0, "cmd")
	redir := NewRedirectNode(0)
	redir.SetMap(2, redirMapNoValue)
	redir.SetLocation("/var/log/service.log")
	cmd.AddRedirect(redir)
	ln.Push(cmd)

	expected.Root = ln

	parserTestTable("simple redirect", `cmd >[2] /var/log/service.log`, expected, t)
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

	parserTestTable("multiple redirects", `cmd >[1=2] >[2=]`, expected, t)
}

func TestParseCommandWithStringsEqualsNot(t *testing.T) {
	expected := NewTree("strings works as expected")
	ln := NewListNode()
	cmd1 := NewCommandNode(0, "echo")
	cmd2 := NewCommandNode(11, "echo")
	cmd1.AddArg(NewArg(5, "hello", false))
	cmd2.AddArg(NewArg(17, "hello", true))

	ln.Push(cmd1)
	ln.Push(cmd2)
	expected.Root = ln

	parserTestTable("strings works as expected", `echo hello
echo "hello"
`, expected, t)
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
	cd.SetDir(NewArg(0, "/tmp", false))
	ln.Push(cd)
	expected.Root = ln

	parserTestTable("test cd", "cd /tmp", expected, t)

	// test cd into home
	expected = NewTree("test cd into home")
	ln = NewListNode()
	cd = NewCdNode(0)
	ln.Push(cd)
	expected.Root = ln

	parserTestTable("test cd into home", "cd", expected, t)

	expected = NewTree("cd into HOME by setenv")
	ln = NewListNode()
	assign := NewAssignmentNode(0)
	assign.SetVarName("HOME")
	assign.SetValueList(append(make([]ElemNode, 0, 1), ElemNode{
		elem: "/",
	}))
	set := NewSetAssignmentNode(9, "HOME")
	cd = NewCdNode(21)
	pwd := NewCommandNode(24, "pwd")

	ln.Push(assign)
	ln.Push(set)
	ln.Push(cd)
	ln.Push(pwd)

	expected.Root = ln

	parserTestTable("test cd into HOME by setenv", `HOME="/"
setenv HOME
cd
pwd`, expected, t)

	// Test cd into custom variable
	expected = NewTree("cd into variable value")
	ln = NewListNode()
	assign = NewAssignmentNode(0)
	assign.SetVarName("GOPATH")
	assign.SetValueList(append(make([]ElemNode, 0, 1), ElemNode{
		elem: "/home/i4k/gopath",
	}))
	cd = NewCdNode(26)
	cd.SetDir(NewArg(0, "$GOPATH", false))

	ln.Push(assign)
	ln.Push(cd)

	expected.Root = ln

	parserTestTable("test cd into variable value", `GOPATH="/home/i4k/gopath"
cd $GOPATH`, expected, t)

}

func TestParseRfork(t *testing.T) {
	expected := NewTree("test rfork")
	ln := NewListNode()
	cmd1 := NewRforkNode(0)
	cmd1.SetFlags(NewArg(6, "u", false))
	ln.Push(cmd1)
	expected.Root = ln

	parserTestTable("test rfork", "rfork u", expected, t)
}

func TestParseRforkWithBlock(t *testing.T) {
	expected := NewTree("rfork with block")
	ln := NewListNode()
	rfork := NewRforkNode(0)
	rfork.SetFlags(NewArg(6, "u", false))

	insideFork := NewCommandNode(11, "mount")
	insideFork.AddArg(NewArg(17, "-t", false))
	insideFork.AddArg(NewArg(20, "proc", false))
	insideFork.AddArg(NewArg(25, "proc", false))
	insideFork.AddArg(NewArg(30, "/proc", false))
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
`, expected, t)

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
	importStmt.SetPath(NewArg(0, "env.sh", false))
	ln.Push(importStmt)
	expected.Root = ln

	parserTestTable("test import", "import env.sh", expected, t)

	expected = NewTree("test import with quotes")
	ln = NewListNode()
	importStmt = NewImportNode(0)
	importStmt.SetPath(NewArg(0, "env.sh", true))
	ln.Push(importStmt)
	expected.Root = ln

	parserTestTable("test import", `import "env.sh"`, expected, t)
}

func TestParseIf(t *testing.T) {
	expected := NewTree("test if")
	ln := NewListNode()
	ifDecl := NewIfNode(0)
	ifDecl.SetLvalue(NewArg(4, "test", true))
	ifDecl.SetRvalue(NewArg(14, "other", true))
	ifDecl.SetOp("==")

	subBlock := NewListNode()
	cmd := NewCommandNode(24, "pwd")
	subBlock.Push(cmd)

	ifTree := NewTree("if block")
	ifTree.Root = subBlock

	ifDecl.tree = ifTree

	ln.Push(ifDecl)
	expected.Root = ln

	parserTestTable("test if", `if "test" == "other" {
	pwd
}`, expected, t)

	expected = NewTree("test if")
	ln = NewListNode()
	ifDecl = NewIfNode(0)
	ifDecl.SetLvalue(NewArg(4, "", true))
	ifDecl.SetRvalue(NewArg(10, "other", true))
	ifDecl.SetOp("!=")

	subBlock = NewListNode()
	cmd = NewCommandNode(20, "pwd")
	subBlock.Push(cmd)

	ifTree = NewTree("if block")
	ifTree.Root = subBlock

	ifDecl.tree = ifTree

	ln.Push(ifDecl)
	expected.Root = ln

	parserTestTable("test if", `if "" != "other" {
	pwd
}`, expected, t)
}

func TestParseIfLvariable(t *testing.T) {
	expected := NewTree("test if with variable")
	ln := NewListNode()
	ifDecl := NewIfNode(0)
	ifDecl.SetLvalue(NewArg(4, "$test", false))
	ifDecl.SetRvalue(NewArg(15, "other", true))
	ifDecl.SetOp("==")

	subBlock := NewListNode()
	cmd := NewCommandNode(25, "pwd")
	subBlock.Push(cmd)

	ifTree := NewTree("if block")
	ifTree.Root = subBlock

	ifDecl.tree = ifTree

	ln.Push(ifDecl)
	expected.Root = ln

	parserTestTable("test if", `if "$test" == "other" {
	pwd
}`, expected, t)
}
