package parser

import (
	"strings"
	"testing"

	"github.com/NeowayLabs/nash/ast"
	"github.com/NeowayLabs/nash/token"
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
	ln := ast.NewBlockNode(token.NewFileInfo(1, 0))
	cmd := ast.NewCommandNode(token.NewFileInfo(1, 0), "echo", false)
	cmd.AddArg(ast.NewStringExpr(token.NewFileInfo(1, 6), "hello world", true))
	ln.Push(cmd)

	expected.Root = ln

	parserTestTable("parser simple", `echo "hello world"`, expected, t, true)

	cmd1 := ast.NewCommandNode(token.NewFileInfo(1, 0), "cat", false)
	arg1 := ast.NewStringExpr(token.NewFileInfo(1, 4), "/etc/resolv.conf", false)
	arg2 := ast.NewStringExpr(token.NewFileInfo(1, 21), "/etc/hosts", false)
	cmd1.AddArg(arg1)
	cmd1.AddArg(arg2)

	ln = ast.NewBlockNode(token.NewFileInfo(1, 0))
	ln.Push(cmd1)
	expected.Root = ln

	parserTestTable("parser simple", `cat /etc/resolv.conf /etc/hosts`, expected, t, true)
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
	ln := ast.NewBlockNode(token.NewFileInfo(1, 0))
	first := ast.NewCommandNode(token.NewFileInfo(1, 0), "echo", false)
	first.AddArg(ast.NewStringExpr(token.NewFileInfo(1, 6), "hello world", true))

	second := ast.NewCommandNode(token.NewFileInfo(1, 21), "awk", false)
	second.AddArg(ast.NewStringExpr(token.NewFileInfo(1, 26), "{print $1}", true))

	pipe := ast.NewPipeNode(token.NewFileInfo(1, 19), false)
	pipe.AddCmd(first)
	pipe.AddCmd(second)

	ln.Push(pipe)

	expected.Root = ln

	parserTestTable("parser pipe", `echo "hello world" | awk "{print $1}"`, expected, t, true)
}

func TestBasicSetEnvAssignment(t *testing.T) {
	expected := ast.NewTree("simple set assignment")
	ln := ast.NewBlockNode(token.NewFileInfo(1, 0))
	set, err := ast.NewSetenvNode(token.NewFileInfo(1, 0), "test", nil)

	if err != nil {
		t.Fatal(err)
	}

	ln.Push(set)
	expected.Root = ln

	parserTestTable("simple set assignment", `setenv test`, expected, t, true)

	// setenv with assignment
	expected = ast.NewTree("setenv with simple assignment")
	ln = ast.NewBlockNode(token.NewFileInfo(1, 0))
	assign := ast.NewAssignmentNode(token.NewFileInfo(1, 7),
		ast.NewNameNode(token.NewFileInfo(1, 7), "test", nil),
		ast.NewStringExpr(token.NewFileInfo(1, 15), "hello", true))
	set, err = ast.NewSetenvNode(token.NewFileInfo(1, 0), "test", assign)

	if err != nil {
		t.Fatal(err)
	}

	ln.Push(set)
	expected.Root = ln

	parserTestTable("setenv with simple assignment", `setenv test = "hello"`, expected, t, true)

	expected = ast.NewTree("setenv with simple cmd assignment")
	ln = ast.NewBlockNode(token.NewFileInfo(1, 0))

	cmd := ast.NewCommandNode(token.NewFileInfo(1, 15), "ls", false)

	cmdAssign, err := ast.NewExecAssignNode(
		token.NewFileInfo(1, 7),
		ast.NewNameNode(token.NewFileInfo(1, 7), "test", nil),
		cmd,
	)

	if err != nil {
		t.Fatal(err)
	}

	set, err = ast.NewSetenvNode(token.NewFileInfo(1, 0), "test", cmdAssign)

	if err != nil {
		t.Fatal(err)
	}

	ln.Push(set)
	expected.Root = ln

	parserTestTable("simple assignment", `setenv test <= ls`, expected, t, true)
}

func TestBasicAssignment(t *testing.T) {
	expected := ast.NewTree("simple assignment")
	ln := ast.NewBlockNode(token.NewFileInfo(1, 0))
	assign := ast.NewAssignmentNode(token.NewFileInfo(1, 0),
		ast.NewNameNode(token.NewFileInfo(1, 0), "test", nil),
		ast.NewStringExpr(token.NewFileInfo(1, 8), "hello", true))
	ln.Push(assign)
	expected.Root = ln

	parserTestTable("simple assignment", `test = "hello"`, expected, t, true)

	// test concatenation of strings and variables

	ln = ast.NewBlockNode(token.NewFileInfo(1, 0))

	concats := make([]ast.Expr, 2, 2)
	concats[0] = ast.NewStringExpr(token.NewFileInfo(1, 8), "hello", true)
	concats[1] = ast.NewVarExpr(token.NewFileInfo(1, 15), "$var")

	arg1 := ast.NewConcatExpr(token.NewFileInfo(1, 8), concats)

	assign = ast.NewAssignmentNode(token.NewFileInfo(1, 0),
		ast.NewNameNode(token.NewFileInfo(1, 0), "test", nil),
		arg1,
	)

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

func TestParseInvalidIndexing(t *testing.T) {
	// test indexed assignment
	parser := NewParser("invalid", `test[a] = "a"`)

	_, err := parser.Parse()

	if err == nil {
		t.Error("Parse must fail")
		return
	} else if err.Error() != "invalid:1:5: Expected number or variable in index. Found ARG" {
		t.Error("Invalid err msg")
		return
	}

	parser = NewParser("invalid", `test[] = "a"`)

	_, err = parser.Parse()

	if err == nil {
		t.Error("Parse must fail")
		return
	} else if err.Error() != "invalid:1:5: Expected number or variable in index. Found ]" {
		t.Error("Invalid err msg")
		return
	}

	parser = NewParser("invalid", `test[10.0] = "a"`)

	_, err = parser.Parse()

	if err == nil {
		t.Error("Parse must fail")
		return
	} else if err.Error() != "invalid:1:5: Expected number or variable in index. Found ARG" {
		t.Error("Invalid err msg")
		return
	}
}

func TestParseListAssignment(t *testing.T) {
	expected := ast.NewTree("list assignment")
	ln := ast.NewBlockNode(token.NewFileInfo(1, 0))

	values := make([]ast.Expr, 0, 4)

	values = append(values,
		ast.NewStringExpr(token.NewFileInfo(2, 1), "plan9", false),
		ast.NewStringExpr(token.NewFileInfo(3, 1), "from", false),
		ast.NewStringExpr(token.NewFileInfo(4, 1), "bell", false),
		ast.NewStringExpr(token.NewFileInfo(5, 1), "labs", false),
	)

	elem := ast.NewListExpr(token.NewFileInfo(1, 7), values)

	assign := ast.NewAssignmentNode(token.NewFileInfo(1, 0),
		ast.NewNameNode(token.NewFileInfo(1, 0), "test", nil),
		elem,
	)

	ln.Push(assign)
	expected.Root = ln

	parserTestTable("list assignment", `test = (
	plan9
	from
	bell
	labs
)`, expected, t, false)
}

func TestParseListOfListsAssignment(t *testing.T) {
	expected := ast.NewTree("list assignment")
	ln := ast.NewBlockNode(token.NewFileInfo(1, 0))

	plan9 := make([]ast.Expr, 0, 4)
	plan9 = append(plan9,
		ast.NewStringExpr(token.NewFileInfo(2, 2), "plan9", false),
		ast.NewStringExpr(token.NewFileInfo(2, 8), "from", false),
		ast.NewStringExpr(token.NewFileInfo(2, 13), "bell", false),
		ast.NewStringExpr(token.NewFileInfo(2, 18), "labs", false),
	)

	elem1 := ast.NewListExpr(token.NewFileInfo(2, 1), plan9)

	linux := make([]ast.Expr, 0, 2)
	linux = append(linux, ast.NewStringExpr(token.NewFileInfo(3, 2), "linux", false))
	linux = append(linux, ast.NewStringExpr(token.NewFileInfo(3, 8), "kernel", false))

	elem2 := ast.NewListExpr(token.NewFileInfo(3, 1), linux)

	values := make([]ast.Expr, 2)
	values[0] = elem1
	values[1] = elem2

	elem := ast.NewListExpr(token.NewFileInfo(1, 7), values)

	assign := ast.NewAssignmentNode(token.NewFileInfo(1, 0),
		ast.NewNameNode(token.NewFileInfo(1, 0), "test", nil),
		elem,
	)

	ln.Push(assign)
	expected.Root = ln

	parserTestTable("list assignment", `test = (
	(plan9 from bell labs)
	(linux kernel)
	)`, expected, t, false)
}

func TestParseCmdAssignment(t *testing.T) {
	expected := ast.NewTree("simple cmd assignment")
	ln := ast.NewBlockNode(token.NewFileInfo(1, 0))

	cmd := ast.NewCommandNode(token.NewFileInfo(1, 8), "ls", false)

	assign, err := ast.NewExecAssignNode(token.NewFileInfo(1, 0),
		ast.NewNameNode(token.NewFileInfo(1, 0), "test", nil),
		cmd,
	)

	if err != nil {
		t.Error(err)
		return
	}

	ln.Push(assign)
	expected.Root = ln

	parserTestTable("simple assignment", `test <= ls`, expected, t, true)
}

func TestParseInvalidEmpty(t *testing.T) {
	parser := NewParser("invalid", ";")

	_, err := parser.Parse()

	if err == nil {
		t.Error("Parse must fail")
		return
	}
}

func TestParsePathCommand(t *testing.T) {
	expected := ast.NewTree("parser simple")
	ln := ast.NewBlockNode(token.NewFileInfo(1, 0))
	cmd := ast.NewCommandNode(token.NewFileInfo(1, 0), "/bin/echo", false)
	cmd.AddArg(ast.NewStringExpr(token.NewFileInfo(1, 11), "hello world", true))
	ln.Push(cmd)

	expected.Root = ln

	parserTestTable("parser simple", `/bin/echo "hello world"`, expected, t, true)
}

func TestParseWithShebang(t *testing.T) {
	expected := ast.NewTree("parser shebang")
	ln := ast.NewBlockNode(token.NewFileInfo(1, 0))
	cmt := ast.NewCommentNode(token.NewFileInfo(1, 0), "#!/bin/nash")
	cmd := ast.NewCommandNode(token.NewFileInfo(3, 0), "echo", false)
	cmd.AddArg(ast.NewStringExpr(token.NewFileInfo(3, 5), "bleh", false))
	ln.Push(cmt)
	ln.Push(cmd)

	expected.Root = ln

	parserTestTable("parser shebang", `#!/bin/nash

echo bleh
`, expected, t, true)
}

func TestParseEmptyFile(t *testing.T) {
	expected := ast.NewTree("empty file")
	ln := ast.NewBlockNode(token.NewFileInfo(1, 0))
	expected.Root = ln

	parserTestTable("empty file", "", expected, t, true)
}

func TestParseSingleCommand(t *testing.T) {
	expected := ast.NewTree("single command")
	expected.Root = ast.NewBlockNode(token.NewFileInfo(1, 0))
	expected.Root.Push(ast.NewCommandNode(token.NewFileInfo(1, 0), "bleh", false))

	parserTestTable("single command", `bleh`, expected, t, true)
}

func TestParseRedirectSimple(t *testing.T) {
	expected := ast.NewTree("redirect")
	ln := ast.NewBlockNode(token.NewFileInfo(1, 0))
	cmd := ast.NewCommandNode(token.NewFileInfo(1, 0), "cmd", false)
	redir := ast.NewRedirectNode(token.NewFileInfo(1, 4))
	redir.SetMap(2, ast.RedirMapSupress)
	cmd.AddRedirect(redir)
	ln.Push(cmd)

	expected.Root = ln

	parserTestTable("simple redirect", `cmd >[2=]`, expected, t, true)

	expected = ast.NewTree("redirect2")
	ln = ast.NewBlockNode(token.NewFileInfo(1, 0))
	cmd = ast.NewCommandNode(token.NewFileInfo(1, 0), "cmd", false)
	redir = ast.NewRedirectNode(token.NewFileInfo(1, 4))
	redir.SetMap(2, 1)
	cmd.AddRedirect(redir)
	ln.Push(cmd)

	expected.Root = ln

	parserTestTable("simple redirect", `cmd >[2=1]`, expected, t, true)
}

func TestParseRedirectWithLocation(t *testing.T) {
	expected := ast.NewTree("redirect with location")
	ln := ast.NewBlockNode(token.NewFileInfo(1, 0))
	cmd := ast.NewCommandNode(token.NewFileInfo(1, 0), "cmd", false)
	redir := ast.NewRedirectNode(token.NewFileInfo(1, 4))
	redir.SetMap(2, ast.RedirMapNoValue)
	redir.SetLocation(ast.NewStringExpr(token.NewFileInfo(1, 9), "/var/log/service.log", false))
	cmd.AddRedirect(redir)
	ln.Push(cmd)

	expected.Root = ln

	parserTestTable("simple redirect", `cmd >[2] /var/log/service.log`, expected, t, true)
}

func TestParseRedirectMultiples(t *testing.T) {
	expected := ast.NewTree("redirect multiples")
	ln := ast.NewBlockNode(token.NewFileInfo(1, 0))
	cmd := ast.NewCommandNode(token.NewFileInfo(1, 0), "cmd", false)
	redir1 := ast.NewRedirectNode(token.NewFileInfo(1, 4))
	redir2 := ast.NewRedirectNode(token.NewFileInfo(1, 11))

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
	ln := ast.NewBlockNode(token.NewFileInfo(1, 0))
	cmd1 := ast.NewCommandNode(token.NewFileInfo(1, 0), "echo", false)
	cmd2 := ast.NewCommandNode(token.NewFileInfo(2, 0), "echo", false)
	cmd1.AddArg(ast.NewStringExpr(token.NewFileInfo(1, 5), "hello", false))
	cmd2.AddArg(ast.NewStringExpr(token.NewFileInfo(2, 6), "hello", true))

	ln.Push(cmd1)
	ln.Push(cmd2)
	expected.Root = ln

	parserTestTable("strings works as expected", `echo hello
echo "hello"
`, expected, t, true)
}

func TestParseCommandSeparatedBySemicolon(t *testing.T) {
	expected := ast.NewTree("semicolon")
	ln := ast.NewBlockNode(token.NewFileInfo(1, 0))
	cmd1 := ast.NewCommandNode(token.NewFileInfo(1, 0), "echo", false)
	cmd2 := ast.NewCommandNode(token.NewFileInfo(1, 11), "echo", false)
	cmd1.AddArg(ast.NewStringExpr(token.NewFileInfo(1, 5), "hello", false))
	cmd2.AddArg(ast.NewStringExpr(token.NewFileInfo(1, 16), "world", false))

	ln.Push(cmd1)
	ln.Push(cmd2)
	expected.Root = ln

	parserTestTable("strings works as expected", `echo hello;echo world`, expected, t, false)
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
	ln := ast.NewBlockNode(token.NewFileInfo(1, 0))
	cd := ast.NewCommandNode(token.NewFileInfo(1, 0), "cd", false)
	arg := ast.NewStringExpr(token.NewFileInfo(1, 3), "/tmp", false)
	cd.AddArg(arg)
	ln.Push(cd)
	expected.Root = ln

	parserTestTable("test cd", "cd /tmp", expected, t, true)

	// test cd into home
	expected = ast.NewTree("test cd into home")
	ln = ast.NewBlockNode(token.NewFileInfo(1, 0))
	cd = ast.NewCommandNode(token.NewFileInfo(1, 0), "cd", false)
	ln.Push(cd)
	expected.Root = ln

	parserTestTable("test cd into home", "cd", expected, t, true)

	expected = ast.NewTree("cd into HOME by setenv")
	ln = ast.NewBlockNode(token.NewFileInfo(1, 0))

	assign := ast.NewAssignmentNode(token.NewFileInfo(1, 0),
		ast.NewNameNode(token.NewFileInfo(1, 0), "HOME", nil),
		ast.NewStringExpr(token.NewFileInfo(1, 8), "/", true),
	)

	set, err := ast.NewSetenvNode(token.NewFileInfo(3, 0), "HOME", nil)

	if err != nil {
		t.Fatal(err)
	}

	cd = ast.NewCommandNode(token.NewFileInfo(5, 0), "cd", false)
	pwd := ast.NewCommandNode(token.NewFileInfo(6, 0), "pwd", false)

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
	ln = ast.NewBlockNode(token.NewFileInfo(1, 0))

	arg = ast.NewStringExpr(token.NewFileInfo(1, 10), "/home/i4k/gopath", true)

	assign = ast.NewAssignmentNode(token.NewFileInfo(1, 0),
		ast.NewNameNode(token.NewFileInfo(1, 0), "GOPATH", nil),
		arg,
	)

	cd = ast.NewCommandNode(token.NewFileInfo(3, 0), "cd", false)
	arg2 := ast.NewVarExpr(token.NewFileInfo(3, 3), "$GOPATH")
	cd.AddArg(arg2)

	ln.Push(assign)
	ln.Push(cd)

	expected.Root = ln

	parserTestTable("test cd into variable value", `GOPATH = "/home/i4k/gopath"

cd $GOPATH`, expected, t, true)

	// Test cd into custom variable
	expected = ast.NewTree("cd into variable value with concat")
	ln = ast.NewBlockNode(token.NewFileInfo(1, 0))

	arg = ast.NewStringExpr(token.NewFileInfo(1, 10), "/home/i4k/gopath", true)

	assign = ast.NewAssignmentNode(token.NewFileInfo(1, 0),
		ast.NewNameNode(token.NewFileInfo(1, 0), "GOPATH", nil),
		arg,
	)

	concat := make([]ast.Expr, 0, 2)
	concat = append(concat, ast.NewVarExpr(token.NewFileInfo(3, 3), "$GOPATH"))
	concat = append(concat, ast.NewStringExpr(token.NewFileInfo(3, 12), "/src/github.com", true))

	cd = ast.NewCommandNode(token.NewFileInfo(3, 0), "cd", false)
	carg := ast.NewConcatExpr(token.NewFileInfo(3, 3), concat)
	cd.AddArg(carg)

	ln.Push(assign)
	ln.Push(cd)

	expected.Root = ln

	parserTestTable("test cd into variable value", `GOPATH = "/home/i4k/gopath"

cd $GOPATH+"/src/github.com"`, expected, t, true)

}

func TestParseConcatOfIndexedVar(t *testing.T) {
	expected := ast.NewTree("concat indexed var")
	ln := ast.NewBlockNode(token.NewFileInfo(1, 0))
	arg1 := ast.NewStringExpr(token.NewFileInfo(1, 4), "ec2", false)
	arg2 := ast.NewStringExpr(token.NewFileInfo(1, 8), "create-tags", false)
	arg3 := ast.NewStringExpr(token.NewFileInfo(1, 20), "--resources", false)
	arg4 := ast.NewVarExpr(token.NewFileInfo(1, 32), "$resource")
	arg5 := ast.NewStringExpr(token.NewFileInfo(1, 42), "--tags", false)

	c1 := ast.NewStringExpr(token.NewFileInfo(1, 50), "Key=", true)
	c2 := ast.NewIndexExpr(token.NewFileInfo(1, 56),
		ast.NewVarExpr(token.NewFileInfo(1, 56), "$tag"),
		ast.NewIntExpr(token.NewFileInfo(1, 61), 0))
	c3 := ast.NewStringExpr(token.NewFileInfo(1, 65), ",Value=", true)
	c4 := ast.NewIndexExpr(token.NewFileInfo(1, 74),
		ast.NewVarExpr(token.NewFileInfo(1, 74), "$tag"),
		ast.NewIntExpr(token.NewFileInfo(1, 79), 1))
	cvalues := make([]ast.Expr, 4)
	cvalues[0] = c1
	cvalues[1] = c2
	cvalues[2] = c3
	cvalues[3] = c4

	arg6 := ast.NewConcatExpr(token.NewFileInfo(1, 50), cvalues)

	cmd := ast.NewCommandNode(token.NewFileInfo(1, 0), "aws", false)
	cmd.AddArg(arg1)
	cmd.AddArg(arg2)
	cmd.AddArg(arg3)
	cmd.AddArg(arg4)
	cmd.AddArg(arg5)
	cmd.AddArg(arg6)

	ln.Push(cmd)
	expected.Root = ln

	parserTestTable("concat indexed var",
		`aws ec2 create-tags --resources $resource --tags "Key="+$tag[0]+",Value="+$tag[1]`,
		expected, t, true)
}

func TestParseRfork(t *testing.T) {
	expected := ast.NewTree("test rfork")
	ln := ast.NewBlockNode(token.NewFileInfo(1, 0))
	cmd1 := ast.NewRforkNode(token.NewFileInfo(1, 0))
	f1 := ast.NewStringExpr(token.NewFileInfo(1, 6), "u", false)
	cmd1.SetFlags(f1)
	ln.Push(cmd1)
	expected.Root = ln

	parserTestTable("test rfork", "rfork u", expected, t, true)
}

func TestParseRforkWithBlock(t *testing.T) {
	expected := ast.NewTree("rfork with block")
	ln := ast.NewBlockNode(token.NewFileInfo(1, 0))
	rfork := ast.NewRforkNode(token.NewFileInfo(1, 0))
	arg := ast.NewStringExpr(token.NewFileInfo(1, 6), "u", false)
	rfork.SetFlags(arg)

	insideFork := ast.NewCommandNode(token.NewFileInfo(2, 1), "mount", false)
	insideFork.AddArg(ast.NewStringExpr(token.NewFileInfo(2, 7), "-t", false))
	insideFork.AddArg(ast.NewStringExpr(token.NewFileInfo(2, 10), "proc", false))
	insideFork.AddArg(ast.NewStringExpr(token.NewFileInfo(2, 15), "proc", false))
	insideFork.AddArg(ast.NewStringExpr(token.NewFileInfo(2, 20), "/proc", false))

	bln := ast.NewBlockNode(token.NewFileInfo(1, 8))
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
	ln := ast.NewBlockNode(token.NewFileInfo(1, 0))
	importStmt := ast.NewImportNode(token.NewFileInfo(1, 0),
		ast.NewStringExpr(token.NewFileInfo(1, 7), "env.sh", false))
	ln.Push(importStmt)
	expected.Root = ln

	parserTestTable("test import", "import env.sh", expected, t, true)

	expected = ast.NewTree("test import with quotes")
	ln = ast.NewBlockNode(token.NewFileInfo(1, 0))
	importStmt = ast.NewImportNode(token.NewFileInfo(1, 0),
		ast.NewStringExpr(token.NewFileInfo(1, 8), "env.sh", true))
	ln.Push(importStmt)
	expected.Root = ln

	parserTestTable("test import", `import "env.sh"`, expected, t, true)
}

func TestParseIf(t *testing.T) {
	expected := ast.NewTree("test if")
	ln := ast.NewBlockNode(token.NewFileInfo(1, 0))
	ifDecl := ast.NewIfNode(token.NewFileInfo(1, 0))
	ifDecl.SetLvalue(ast.NewStringExpr(token.NewFileInfo(1, 4), "test", true))
	ifDecl.SetRvalue(ast.NewStringExpr(token.NewFileInfo(1, 14), "other", true))
	ifDecl.SetOp("==")

	subBlock := ast.NewBlockNode(token.NewFileInfo(1, 21))
	cmd := ast.NewCommandNode(token.NewFileInfo(2, 1), "pwd", false)
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
	ln = ast.NewBlockNode(token.NewFileInfo(1, 0))
	ifDecl = ast.NewIfNode(token.NewFileInfo(1, 0))
	ifDecl.SetLvalue(ast.NewStringExpr(token.NewFileInfo(1, 4), "", true))
	ifDecl.SetRvalue(ast.NewStringExpr(token.NewFileInfo(1, 10), "other", true))
	ifDecl.SetOp("!=")

	subBlock = ast.NewBlockNode(token.NewFileInfo(1, 17))
	cmd = ast.NewCommandNode(token.NewFileInfo(2, 1), "pwd", false)
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

func TestParseFnInv(t *testing.T) {
	expected := ast.NewTree("fn inv")
	ln := ast.NewBlockNode(token.NewFileInfo(1, 0))
	aFn := ast.NewFnInvNode(token.NewFileInfo(1, 0), "a")
	ln.Push(aFn)
	expected.Root = ln

	parserTestTable("test basic fn inv", `a()`, expected, t, true)

	expected = ast.NewTree("fn inv")
	ln = ast.NewBlockNode(token.NewFileInfo(1, 0))
	aFn = ast.NewFnInvNode(token.NewFileInfo(1, 0), "a")
	bFn := ast.NewFnInvNode(token.NewFileInfo(1, 2), "b")
	aFn.AddArg(bFn)
	ln.Push(aFn)
	expected.Root = ln

	parserTestTable("test fn composition", `a(b())`, expected, t, true)

	expected = ast.NewTree("fn inv")
	ln = ast.NewBlockNode(token.NewFileInfo(1, 0))
	aFn = ast.NewFnInvNode(token.NewFileInfo(1, 0), "a")
	bFn = ast.NewFnInvNode(token.NewFileInfo(1, 2), "b")
	b2Fn := ast.NewFnInvNode(token.NewFileInfo(1, 7), "b")
	aFn.AddArg(bFn)
	aFn.AddArg(b2Fn)
	ln.Push(aFn)
	expected.Root = ln

	parserTestTable("test fn composition", `a(b(), b())`, expected, t, true)

	expected = ast.NewTree("fn inv")
	ln = ast.NewBlockNode(token.NewFileInfo(1, 0))
	aFn = ast.NewFnInvNode(token.NewFileInfo(1, 0), "a")
	bFn = ast.NewFnInvNode(token.NewFileInfo(1, 2), "b")
	b2Fn = ast.NewFnInvNode(token.NewFileInfo(1, 4), "b")
	bFn.AddArg(b2Fn)
	aFn.AddArg(bFn)
	ln.Push(aFn)
	expected.Root = ln

	parserTestTable("test fn composition", `a(b(b()))`, expected, t, true)
}

func TestParseIfFnInv(t *testing.T) {
	expected := ast.NewTree("test if")
	ln := ast.NewBlockNode(token.NewFileInfo(1, 0))
	ifDecl := ast.NewIfNode(token.NewFileInfo(1, 0))
	ifDecl.SetLvalue(ast.NewFnInvNode(token.NewFileInfo(1, 3), "test"))
	ifDecl.SetRvalue(ast.NewStringExpr(token.NewFileInfo(1, 14), "other", true))
	ifDecl.SetOp("==")

	subBlock := ast.NewBlockNode(token.NewFileInfo(1, 21))
	cmd := ast.NewCommandNode(token.NewFileInfo(2, 1), "pwd", false)
	subBlock.Push(cmd)

	ifTree := ast.NewTree("if block")
	ifTree.Root = subBlock

	ifDecl.SetIfTree(ifTree)

	ln.Push(ifDecl)
	expected.Root = ln

	parserTestTable("test if", `if test() == "other" {
	pwd
}`, expected, t, true)

	expected = ast.NewTree("test if")
	ln = ast.NewBlockNode(token.NewFileInfo(1, 0))
	ifDecl = ast.NewIfNode(token.NewFileInfo(1, 0))

	fnInv := ast.NewFnInvNode(token.NewFileInfo(1, 3), "test")
	fnInv.AddArg(ast.NewStringExpr(token.NewFileInfo(1, 9), "bleh", true))
	ifDecl.SetLvalue(fnInv)
	ifDecl.SetRvalue(ast.NewStringExpr(token.NewFileInfo(1, 20), "other", true))
	ifDecl.SetOp("!=")

	subBlock = ast.NewBlockNode(token.NewFileInfo(1, 27))
	cmd = ast.NewCommandNode(token.NewFileInfo(2, 1), "pwd", false)
	subBlock.Push(cmd)

	ifTree = ast.NewTree("if block")
	ifTree.Root = subBlock

	ifDecl.SetIfTree(ifTree)

	ln.Push(ifDecl)
	expected.Root = ln

	parserTestTable("test if", `if test("bleh") != "other" {
	pwd
}`, expected, t, true)
}

func TestParseIfLvariable(t *testing.T) {
	expected := ast.NewTree("test if with variable")
	ln := ast.NewBlockNode(token.NewFileInfo(1, 0))
	ifDecl := ast.NewIfNode(token.NewFileInfo(1, 0))
	ifDecl.SetLvalue(ast.NewVarExpr(token.NewFileInfo(1, 3), "$test"))
	ifDecl.SetRvalue(ast.NewStringExpr(token.NewFileInfo(1, 13), "other", true))
	ifDecl.SetOp("==")

	subBlock := ast.NewBlockNode(token.NewFileInfo(1, 20))
	cmd := ast.NewCommandNode(token.NewFileInfo(2, 1), "pwd", false)
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

func TestParseIfRvariable(t *testing.T) {
	expected := ast.NewTree("test if with variable")
	ln := ast.NewBlockNode(token.NewFileInfo(1, 0))
	ifDecl := ast.NewIfNode(token.NewFileInfo(1, 0))
	ifDecl.SetLvalue(ast.NewVarExpr(token.NewFileInfo(1, 3), "$test"))
	ifDecl.SetRvalue(ast.NewVarExpr(token.NewFileInfo(1, 12), "$other"))
	ifDecl.SetOp("==")

	subBlock := ast.NewBlockNode(token.NewFileInfo(1, 19))
	cmd := ast.NewCommandNode(token.NewFileInfo(2, 1), "pwd", false)
	subBlock.Push(cmd)

	ifTree := ast.NewTree("if block")
	ifTree.Root = subBlock

	ifDecl.SetIfTree(ifTree)

	ln.Push(ifDecl)
	expected.Root = ln

	parserTestTable("test if", `if $test == $other {
	pwd
}`, expected, t, true)
}

func TestParseIfElse(t *testing.T) {
	expected := ast.NewTree("test if else with variable")
	ln := ast.NewBlockNode(token.NewFileInfo(1, 0))
	ifDecl := ast.NewIfNode(token.NewFileInfo(1, 0))
	ifDecl.SetLvalue(ast.NewVarExpr(token.NewFileInfo(1, 3), "$test"))
	ifDecl.SetRvalue(ast.NewStringExpr(token.NewFileInfo(1, 13), "other", true))
	ifDecl.SetOp("==")

	subBlock := ast.NewBlockNode(token.NewFileInfo(1, 20))
	cmd := ast.NewCommandNode(token.NewFileInfo(2, 1), "pwd", false)
	subBlock.Push(cmd)

	ifTree := ast.NewTree("if block")
	ifTree.Root = subBlock

	ifDecl.SetIfTree(ifTree)

	elseBlock := ast.NewBlockNode(token.NewFileInfo(3, 7))
	exitCmd := ast.NewCommandNode(token.NewFileInfo(4, 1), "exit", false)
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
	ln := ast.NewBlockNode(token.NewFileInfo(1, 0))
	ifDecl := ast.NewIfNode(token.NewFileInfo(1, 0))
	ifDecl.SetLvalue(ast.NewVarExpr(token.NewFileInfo(1, 3), "$test"))
	ifDecl.SetRvalue(ast.NewStringExpr(token.NewFileInfo(1, 13), "other", true))
	ifDecl.SetOp("==")

	subBlock := ast.NewBlockNode(token.NewFileInfo(1, 20))
	cmd := ast.NewCommandNode(token.NewFileInfo(2, 1), "pwd", false)
	subBlock.Push(cmd)

	ifTree := ast.NewTree("if block")
	ifTree.Root = subBlock

	ifDecl.SetIfTree(ifTree)

	elseIfDecl := ast.NewIfNode(token.NewFileInfo(3, 7))

	elseIfDecl.SetLvalue(ast.NewVarExpr(token.NewFileInfo(3, 10), "$test"))
	elseIfDecl.SetRvalue(ast.NewStringExpr(token.NewFileInfo(3, 20), "others", true))
	elseIfDecl.SetOp("==")

	elseIfBlock := ast.NewBlockNode(token.NewFileInfo(3, 28))
	elseifCmd := ast.NewCommandNode(token.NewFileInfo(4, 1), "ls", false)
	elseIfBlock.Push(elseifCmd)

	elseIfTree := ast.NewTree("if block")
	elseIfTree.Root = elseIfBlock

	elseIfDecl.SetIfTree(elseIfTree)

	elseBlock := ast.NewBlockNode(token.NewFileInfo(5, 7))
	exitCmd := ast.NewCommandNode(token.NewFileInfo(6, 1), "exit", false)
	elseBlock.Push(exitCmd)

	elseTree := ast.NewTree("else block")
	elseTree.Root = elseBlock

	elseIfDecl.SetElseTree(elseTree)

	elseBlock2 := ast.NewBlockNode(token.NewFileInfo(3, 7))
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
	ln := ast.NewBlockNode(token.NewFileInfo(1, 0))

	// fn
	fn := ast.NewFnDeclNode(token.NewFileInfo(1, 0), "build")
	tree := ast.NewTree("fn body")
	lnBody := ast.NewBlockNode(token.NewFileInfo(1, 0))
	tree.Root = lnBody
	fn.SetTree(tree)

	// root
	ln.Push(fn)
	expected.Root = ln

	parserTestTable("fn", `fn build() {

}`, expected, t, true)

	// root
	expected = ast.NewTree("fn")
	ln = ast.NewBlockNode(token.NewFileInfo(1, 0))

	// fn
	fn = ast.NewFnDeclNode(token.NewFileInfo(1, 0), "build")
	tree = ast.NewTree("fn body")
	lnBody = ast.NewBlockNode(token.NewFileInfo(1, 0))
	cmd := ast.NewCommandNode(token.NewFileInfo(1, 0), "ls", false)
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
	ln = ast.NewBlockNode(token.NewFileInfo(1, 0))

	// fn
	fn = ast.NewFnDeclNode(token.NewFileInfo(1, 0), "build")
	fn.AddArg("image")
	tree = ast.NewTree("fn body")
	lnBody = ast.NewBlockNode(token.NewFileInfo(1, 0))
	cmd = ast.NewCommandNode(token.NewFileInfo(1, 0), "ls", false)
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
	ln = ast.NewBlockNode(token.NewFileInfo(1, 0))

	// fn
	fn = ast.NewFnDeclNode(token.NewFileInfo(1, 0), "build")
	fn.AddArg("image")
	fn.AddArg("debug")
	tree = ast.NewTree("fn body")
	lnBody = ast.NewBlockNode(token.NewFileInfo(1, 0))
	cmd = ast.NewCommandNode(token.NewFileInfo(1, 0), "ls", false)
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

func TestParseInlineFnDecl(t *testing.T) {
	expected := ast.NewTree("fn")
	ln := ast.NewBlockNode(token.NewFileInfo(1, 0))

	fn := ast.NewFnDeclNode(token.NewFileInfo(1, 0), "cd")
	tree := ast.NewTree("fn body")
	lnBody := ast.NewBlockNode(token.NewFileInfo(1, 0))
	echo := ast.NewCommandNode(token.NewFileInfo(1, 11), "echo", false)
	echo.AddArg(ast.NewStringExpr(token.NewFileInfo(1, 16), "hello", true))
	lnBody.Push(echo)

	tree.Root = lnBody
	fn.SetTree(tree)

	// root
	ln.Push(fn)
	expected.Root = ln

	parserTestTable("inline fn", `fn cd() { echo "hello" }`,
		expected, t, false)

	test := ast.NewCommandNode(token.NewFileInfo(1, 26), "test", false)
	test.AddArg(ast.NewStringExpr(token.NewFileInfo(1, 32), "-d", false))
	test.AddArg(ast.NewStringExpr(token.NewFileInfo(1, 35), "/etc", false))

	pipe := ast.NewPipeNode(token.NewFileInfo(1, 11), false)
	pipe.AddCmd(echo)
	pipe.AddCmd(test)
	lnBody = ast.NewBlockNode(token.NewFileInfo(1, 0))
	lnBody.Push(pipe)
	tree.Root = lnBody
	fn.SetTree(tree)
	ln = ast.NewBlockNode(token.NewFileInfo(1, 0))
	ln.Push(fn)
	expected.Root = ln

	parserTestTable("inline fn", `fn cd() { echo "hello" | test -d /etc }`,
		expected, t, false)
}

func TestParseBindFn(t *testing.T) {
	expected := ast.NewTree("bindfn")
	ln := ast.NewBlockNode(token.NewFileInfo(1, 0))

	bindFn := ast.NewBindFnNode(token.NewFileInfo(1, 0), "cd", "cd2")
	ln.Push(bindFn)
	expected.Root = ln

	parserTestTable("bindfn", `bindfn cd cd2`, expected, t, true)
}

func TestParseRedirectionVariable(t *testing.T) {
	expected := ast.NewTree("redirection var")
	ln := ast.NewBlockNode(token.NewFileInfo(1, 0))

	cmd := ast.NewCommandNode(token.NewFileInfo(1, 0), "cmd", false)
	redir := ast.NewRedirectNode(token.NewFileInfo(1, 4))
	redirArg := ast.NewVarExpr(token.NewFileInfo(1, 6), "$outFname")
	redir.SetLocation(redirArg)
	cmd.AddRedirect(redir)
	ln.Push(cmd)
	expected.Root = ln

	parserTestTable("redir var", `cmd > $outFname`, expected, t, true)
}

func TestParseDump(t *testing.T) {
	expected := ast.NewTree("dump")
	ln := ast.NewBlockNode(token.NewFileInfo(1, 0))

	dump := ast.NewDumpNode(token.NewFileInfo(1, 0))
	dump.SetFilename(ast.NewStringExpr(token.NewFileInfo(1, 5), "./init", false))
	ln.Push(dump)
	expected.Root = ln

	parserTestTable("dump", `dump ./init`, expected, t, true)
}

func TestParseReturn(t *testing.T) {
	expected := ast.NewTree("return")
	ln := ast.NewBlockNode(token.NewFileInfo(1, 0))

	ret := ast.NewReturnNode(token.NewFileInfo(1, 0))
	ln.Push(ret)
	expected.Root = ln

	parserTestTable("return", `return`, expected, t, true)

	expected = ast.NewTree("return list")
	ln = ast.NewBlockNode(token.NewFileInfo(1, 0))

	ret = ast.NewReturnNode(token.NewFileInfo(1, 0))

	listvalues := make([]ast.Expr, 2)

	listvalues[0] = ast.NewStringExpr(token.NewFileInfo(1, 9), "val1", true)
	listvalues[1] = ast.NewStringExpr(token.NewFileInfo(1, 16), "val2", true)

	retReturn := ast.NewListExpr(token.NewFileInfo(1, 7), listvalues)

	ret.SetReturn(retReturn)

	ln.Push(ret)
	expected.Root = ln

	parserTestTable("return", `return ("val1" "val2")`, expected, t, true)

	expected = ast.NewTree("return variable")
	ln = ast.NewBlockNode(token.NewFileInfo(1, 0))

	ret = ast.NewReturnNode(token.NewFileInfo(1, 0))

	ret.SetReturn(ast.NewVarExpr(token.NewFileInfo(1, 7), "$var"))

	ln.Push(ret)
	expected.Root = ln

	parserTestTable("return", `return $var`, expected, t, true)

	expected = ast.NewTree("return string")
	ln = ast.NewBlockNode(token.NewFileInfo(1, 0))

	ret = ast.NewReturnNode(token.NewFileInfo(1, 0))

	ret.SetReturn(ast.NewStringExpr(token.NewFileInfo(1, 8), "value", true))

	ln.Push(ret)
	expected.Root = ln

	parserTestTable("return", `return "value"`, expected, t, true)

	expected = ast.NewTree("return funcall")
	ln = ast.NewBlockNode(token.NewFileInfo(1, 0))

	ret = ast.NewReturnNode(token.NewFileInfo(1, 0))

	aFn := ast.NewFnInvNode(token.NewFileInfo(1, 7), "a")

	ret.SetReturn(aFn)

	ln.Push(ret)
	expected.Root = ln

	parserTestTable("return", `return a()`, expected, t, true)
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

	forStmt := ast.NewForNode(token.NewFileInfo(1, 0))
	forTree := ast.NewTree("for block")
	forBlock := ast.NewBlockNode(token.NewFileInfo(1, 0))
	forTree.Root = forBlock
	forStmt.SetTree(forTree)

	ln := ast.NewBlockNode(token.NewFileInfo(1, 0))
	ln.Push(forStmt)
	expected.Root = ln

	parserTestTable("for", `for {

}`, expected, t, true)

	forStmt.SetIdentifier("f")
	forStmt.SetInExpr(ast.NewVarExpr(token.NewFileInfo(1, 9), "$files"))

	parserTestTable("for", `for f in $files {

}`, expected, t, true)

	forStmt.SetIdentifier("f")
	fnInv := ast.NewFnInvNode(token.NewFileInfo(1, 9), "getfiles")
	fnArg := ast.NewStringExpr(token.NewFileInfo(1, 19), "/", true)
	fnInv.AddArg(fnArg)
	forStmt.SetInExpr(fnInv)

	parserTestTable("for", `for f in getfiles("/") {

}`, expected, t, true)

	forStmt.SetIdentifier("f")
	value1 := ast.NewStringExpr(token.NewFileInfo(1, 10), "1", false)
	value2 := ast.NewStringExpr(token.NewFileInfo(1, 12), "2", false)
	value3 := ast.NewStringExpr(token.NewFileInfo(1, 14), "3", false)
	value4 := ast.NewStringExpr(token.NewFileInfo(1, 16), "4", false)
	value5 := ast.NewStringExpr(token.NewFileInfo(1, 18), "5", false)

	list := ast.NewListExpr(token.NewFileInfo(1, 9), []ast.Expr{
		value1, value2, value3, value4, value5,
	})

	forStmt.SetInExpr(list)

	parserTestTable("for", `for f in (1 2 3 4 5) {

}`, expected, t, true)
}

func TestParseVariableIndexing(t *testing.T) {
	expected := ast.NewTree("variable indexing")
	ln := ast.NewBlockNode(token.NewFileInfo(1, 0))

	indexedVar := ast.NewIndexExpr(
		token.NewFileInfo(1, 7),
		ast.NewVarExpr(token.NewFileInfo(1, 7), "$values"),
		ast.NewIntExpr(token.NewFileInfo(1, 15), 0),
	)

	assignment := ast.NewAssignmentNode(token.NewFileInfo(1, 0),
		ast.NewNameNode(token.NewFileInfo(1, 0), "test", nil),
		indexedVar,
	)

	ln.Push(assignment)
	expected.Root = ln

	parserTestTable("variable indexing", `test = $values[0]`, expected, t, true)

	ln = ast.NewBlockNode(token.NewFileInfo(1, 0))

	ifDecl := ast.NewIfNode(token.NewFileInfo(1, 0))
	lvalue := ast.NewVarExpr(token.NewFileInfo(1, 3), "$values")

	indexedVar = ast.NewIndexExpr(token.NewFileInfo(1, 3), lvalue,
		ast.NewIntExpr(token.NewFileInfo(1, 11), 0))

	ifDecl.SetLvalue(indexedVar)
	ifDecl.SetOp("==")
	ifDecl.SetRvalue(ast.NewStringExpr(token.NewFileInfo(1, 18), "1", true))

	ifBlock := ast.NewTree("if")
	lnBody := ast.NewBlockNode(token.NewFileInfo(1, 21))
	ifBlock.Root = lnBody
	ifDecl.SetIfTree(ifBlock)

	ln.Push(ifDecl)
	expected.Root = ln

	parserTestTable("variable indexing", `if $values[0] == "1" {

}`, expected, t, true)
}

func TestParseMultilineCmdExec(t *testing.T) {
	expected := ast.NewTree("parser simple")
	ln := ast.NewBlockNode(token.NewFileInfo(1, 0))
	cmd := ast.NewCommandNode(token.NewFileInfo(1, 1), "echo", true)
	cmd.AddArg(ast.NewStringExpr(token.NewFileInfo(1, 7), "hello world", true))
	ln.Push(cmd)

	expected.Root = ln

	parserTestTable("parser simple", `(echo "hello world")`, expected, t, true)

	expected = ast.NewTree("parser aws cmd")
	ln = ast.NewBlockNode(token.NewFileInfo(1, 0))
	cmd = ast.NewCommandNode(token.NewFileInfo(2, 1), "aws", true)
	cmd.AddArg(ast.NewStringExpr(token.NewFileInfo(2, 5), "ec2", false))
	cmd.AddArg(ast.NewStringExpr(token.NewFileInfo(2, 9), "run-instances", false))
	cmd.AddArg(ast.NewStringExpr(token.NewFileInfo(3, 3), "--image-id", false))
	cmd.AddArg(ast.NewStringExpr(token.NewFileInfo(3, 14), "ami-xxxxxxxx", false))
	cmd.AddArg(ast.NewStringExpr(token.NewFileInfo(4, 3), "--count", false))
	cmd.AddArg(ast.NewStringExpr(token.NewFileInfo(4, 11), "1", false))
	cmd.AddArg(ast.NewStringExpr(token.NewFileInfo(5, 3), "--instance-type", false))
	cmd.AddArg(ast.NewStringExpr(token.NewFileInfo(5, 19), "t1.micro", false))
	cmd.AddArg(ast.NewStringExpr(token.NewFileInfo(6, 3), "--key-name", false))
	cmd.AddArg(ast.NewStringExpr(token.NewFileInfo(6, 14), "MyKeyPair", false))
	cmd.AddArg(ast.NewStringExpr(token.NewFileInfo(7, 3), "--security-groups", false))
	cmd.AddArg(ast.NewStringExpr(token.NewFileInfo(7, 21), "my-sg", false))

	ln.Push(cmd)

	expected.Root = ln

	parserTestTable("parser simple", `(
	aws ec2 run-instances
			--image-id ami-xxxxxxxx
			--count 1
			--instance-type t1.micro
			--key-name MyKeyPair
			--security-groups my-sg
)`, expected, t, true)
}

func TestParseMultilineCmdAssign(t *testing.T) {
	expected := ast.NewTree("parser simple assign")
	ln := ast.NewBlockNode(token.NewFileInfo(1, 0))
	cmd := ast.NewCommandNode(token.NewFileInfo(1, 10), "echo", true)
	cmd.AddArg(ast.NewStringExpr(token.NewFileInfo(1, 16), "hello world", true))
	assign, err := ast.NewExecAssignNode(token.NewFileInfo(1, 0),
		ast.NewNameNode(token.NewFileInfo(1, 0), "hello", nil),
		cmd,
	)

	if err != nil {
		t.Error(err)
		return
	}

	ln.Push(assign)

	expected.Root = ln

	parserTestTable("parser simple", `hello <= (echo "hello world")`, expected, t, true)
}

func TestMultiPipe(t *testing.T) {
	expected := ast.NewTree("parser pipe")
	ln := ast.NewBlockNode(token.NewFileInfo(1, 0))
	first := ast.NewCommandNode(token.NewFileInfo(1, 1), "echo", false)
	first.AddArg(ast.NewStringExpr(token.NewFileInfo(1, 7), "hello world", true))

	second := ast.NewCommandNode(token.NewFileInfo(1, 22), "awk", false)
	second.AddArg(ast.NewStringExpr(token.NewFileInfo(1, 27), "{print $1}", true))

	pipe := ast.NewPipeNode(token.NewFileInfo(1, 20), true)
	pipe.AddCmd(first)
	pipe.AddCmd(second)

	ln.Push(pipe)

	expected.Root = ln

	parserTestTable("parser pipe", `(echo "hello world" | awk "{print $1}")`, expected, t, true)

	// get longer stringify
	expected = ast.NewTree("parser pipe")
	ln = ast.NewBlockNode(token.NewFileInfo(1, 0))
	first = ast.NewCommandNode(token.NewFileInfo(2, 1), "echo", false)
	first.AddArg(ast.NewStringExpr(token.NewFileInfo(2, 7), "hello world", true))

	second = ast.NewCommandNode(token.NewFileInfo(3, 1), "awk", false)
	second.AddArg(ast.NewStringExpr(token.NewFileInfo(3, 6), "{print AAAAAAAAAAAAAAAAAAAAAA}", true))

	pipe = ast.NewPipeNode(token.NewFileInfo(2, 20), true)
	pipe.AddCmd(first)
	pipe.AddCmd(second)

	ln.Push(pipe)

	expected.Root = ln

	parserTestTable("parser pipe", `(
	echo "hello world" |
	awk "{print AAAAAAAAAAAAAAAAAAAAAA}"
)`, expected, t, true)
}

func TestFunctionPipes(t *testing.T) {
	parser := NewParser("invalid pipe with functions",
		`echo "some thing" | replace(" ", "|")`)

	_, err := parser.Parse()

	if err == nil {
		t.Error("Must fail. Function must be bind'ed to command name to use in pipe.")
		return
	}
}
