package nash

import (
	"errors"
	"fmt"
	"reflect"
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
	set := NewSetAssignment(0, "test")

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

func TestCd(t *testing.T) {
	expected := NewTree("test cd")
	ln := NewListNode()
	cd := NewCdNode(0)
	cd.SetDir(NewArg(0, "/tmp", false))
	ln.Push(cd)
	expected.Root = ln

	parserTestTable("test cd", "cd /tmp", expected, t)
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

func comparePosition(expected Pos, value Pos) (bool, error) {
	if expected != value {
		return false, fmt.Errorf("Position mismatch: %d != %d", expected, value)
	}

	return true, nil
}

func compareArg(expected *Arg, value *Arg) (bool, error) {
	if expected == nil && value == nil {
		return true, nil
	}

	if (expected == nil) != (value == nil) {
		return false, fmt.Errorf("Value differ: %v != %v", expected, value)
	}

	if expected.Type() != value.Type() {
		return false, fmt.Errorf("Type differs: %d != %d", expected.Type(), value.Type())
	}

	if ok, err := comparePosition(expected.Position(), value.Position()); !ok {
		return ok, fmt.Errorf(" CompareArgs(%v, %v) -> %s", expected, value, err.Error())
	}

	if expected.val != value.val {
		return false, fmt.Errorf("Argument value differs: '%s' != '%s'", expected.val, value.val)
	}

	return true, nil
}

func compareCdNode(expected, value *CdNode) (bool, error) {
	if expected == nil && value == nil {
		return true, nil
	}

	if (expected == nil) != (value == nil) {
		return false, fmt.Errorf("One of the nodecommand are nil")
	}

	if expected.dir.val != value.dir.val {
		return false, fmt.Errorf("Expected.dir.val (%v) != value.dir.val (%v)", expected.dir.val, value.dir.val)
	}

	if ok, err := comparePosition(expected.Position(), value.Position()); !ok {
		return ok, fmt.Errorf(" CompareCommandNode (%v, %v)-> %s", expected, value, err.Error())
	}

	return true, nil
}

func compareCommandNode(expected *CommandNode, value *CommandNode) (bool, error) {
	if expected == nil && value == nil {
		return true, nil
	}

	if (expected == nil) != (value == nil) {
		return false, fmt.Errorf("One of the nodecommand are nil")
	}

	ename := expected.name
	vname := value.name

	if ename != vname {
		return false, fmt.Errorf("CommandNode: expected.name('%s') != value.name('%s')",
			ename, vname)
	}

	if ok, err := comparePosition(expected.Position(), value.Position()); !ok {
		return ok, fmt.Errorf(" CompareCommandNode (%v, %v)-> %s", expected, value, err.Error())
	}

	eargs := expected.args
	vargs := value.args

	if len(eargs) != len(vargs) {
		return false, fmt.Errorf("CommandNode: length of expected.args and value.args differs: %d != %d", len(eargs), len(vargs))
	}

	for i := 0; i < len(eargs); i++ {
		ea := eargs[i]
		va := vargs[i]

		valid, err := compareArg(&ea, &va)

		if !valid {
			return valid, err
		}
	}

	return true, nil
}

func compareCommentNode(expected, value *CommentNode) (bool, error) {
	if expected == nil && value == nil {
		return true, nil
	}

	if (expected == nil) != (value == nil) {
		return false, fmt.Errorf("Only one of the nodes are nil. %v != %v", expected, value)
	}

	if expected.val != value.val {
		return false, fmt.Errorf("Comment val differ: '%s' != '%s'", expected.val, value.val)
	}

	return true, nil
}

func compareSetAssignment(expected, value *SetAssignmentNode) (bool, error) {
	if expected == nil && value == nil {
		return true, nil
	}

	if (expected == nil) != (value == nil) {
		return false, fmt.Errorf("Only one of the nodes are nil. %v != %v", expected, value)
	}

	if ok, err := comparePosition(expected.Position(), value.Position()); !ok {
		return ok, fmt.Errorf(" CompareRforkNode (%v, %v) -> %s", expected, value, err.Error())
	}

	if expected.varName != value.varName {
		return false, fmt.Errorf("Set identifier mismatch. %s != %s", expected.varName, value.varName)
	}

	return true, nil
}

func compareAssignment(expected, value *AssignmentNode) (bool, error) {
	if expected == nil && value == nil {
		return true, nil
	}

	if (expected == nil) != (value == nil) {
		return false, fmt.Errorf("Only one of the nodes are nil. %v != %v", expected, value)
	}

	if expected.name != value.name {
		return false, fmt.Errorf("Variable name differs. '%s' != '%s'", expected.name, value.name)
	}

	if len(expected.list) != len(value.list) {
		return false, fmt.Errorf("Variable list value length differs. %d != %d", len(expected.list), len(value.list))
	}

	for i := 0; i < len(expected.list); i++ {
		ev := expected.list[i]
		vv := value.list[i]

		if ev.elem != vv.elem {
			return false, fmt.Errorf("Variable list differs at index %d. %s != %s", i, ev, vv)
		}

		if len(ev.concats) != len(vv.concats) {
			return false, fmt.Errorf("Variable list concats length differs. %d != %d", len(ev.concats), len(vv.concats))
		}

		for j := 0; j < len(ev.concats); j++ {
			if ev.concats[j] != vv.concats[j] {
				return false, fmt.Errorf("Variable concatenation entry differ. '%s' != '%s'", ev.concats[j], vv.concats[j])
			}
		}
	}

	return true, nil
}

func compareRforkNode(expected, value *RforkNode) (bool, error) {
	if expected == nil && value == nil {
		return true, nil
	}

	if (expected == nil) != (value == nil) {
		return false, fmt.Errorf("Only one of the nodes are nil. %v != %v", expected, value)
	}

	if ok, err := comparePosition(expected.Position(), value.Position()); !ok {
		return ok, fmt.Errorf(" CompareRforkNode (%v, %v) -> %s", expected, value, err.Error())
	}

	if ok, err := compareArg(&expected.arg, &value.arg); !ok {
		return ok, fmt.Errorf("CompareRforkNode (%v, %v) -> %s", expected, value, err.Error())
	}

	expectedTree := expected.Tree()
	valueTree := value.Tree()

	return compare(expectedTree, valueTree)
}

func compareNodes(expected Node, value Node) (bool, error) {
	var (
		valid = true
		err   error
	)

	if expected == nil && value == nil {
		return true, nil
	}

	if (expected == nil) != (value == nil) {
		return false, fmt.Errorf("Only one of the nodes are nil. %v != %v", expected, value)
	}

	if valid, err = comparePosition(expected.Position(), value.Position()); !valid {
		return valid, fmt.Errorf(" CompareNodes (%v, %v)-> %s", expected, value, err.Error())
	}

	etype := expected.Type()
	vtype := value.Type()

	if etype != vtype {
		return false, fmt.Errorf("Node type differs: %d != %d", etype, vtype)
	}

	eitype := reflect.TypeOf(expected)
	vitype := reflect.TypeOf(value)

	if eitype.Kind() != vitype.Kind() {
		return false, fmt.Errorf("Node type differs: %v != %v", eitype.Kind(), vitype.Kind())
	}

	switch v := expected.(type) {
	case *SetAssignmentNode:
		ec := expected.(*SetAssignmentNode)
		vc := value.(*SetAssignmentNode)

		valid, err = compareSetAssignment(ec, vc)
	case *AssignmentNode:
		ec := expected.(*AssignmentNode)
		vc := value.(*AssignmentNode)
		valid, err = compareAssignment(ec, vc)
	case *CdNode:
		ec := expected.(*CdNode)
		vc := value.(*CdNode)
		valid, err = compareCdNode(ec, vc)
	case *CommandNode:
		ec := expected.(*CommandNode)
		vc := value.(*CommandNode)
		valid, err = compareCommandNode(ec, vc)
	case *CommentNode:
		ec := expected.(*CommentNode)
		vc := value.(*CommentNode)
		valid, err = compareCommentNode(ec, vc)
	case *RforkNode:
		er := expected.(*RforkNode)
		vr := value.(*RforkNode)
		valid, err = compareRforkNode(er, vr)
	default:
		return false, fmt.Errorf("Type %v not comparable yet", v)
	}

	if !valid {
		return valid, err
	}

	return compare(expected.Tree(), value.Tree())
}

func compare(expected *Tree, tr *Tree) (bool, error) {
	if expected == nil && tr == nil {
		return true, nil
	}

	if (expected == nil) != (tr == nil) {
		return false, errors.New("only one of the expected and tr are nil")
	}

	en := expected.Name
	tn := expected.Name

	if en != tn {
		return false, fmt.Errorf("expected.Name != tr.Name. rxpected.Name='%s' and tr.Name='%s'",
			en, tn)
	}

	eroot := expected.Root
	troot := tr.Root

	if eroot == nil && troot == nil {
		return true, nil
	}

	if (eroot == nil) != (troot == nil) {
		return false, fmt.Errorf("Only one of the expected.Root and tr.Root is nil")
	}

	if len(eroot.Nodes) != len(troot.Nodes) {
		return false, fmt.Errorf("Length differs. len(expected.Root) == %d and len(tr.Root) = %d",
			len(eroot.Nodes), len(troot.Nodes))
	}

	for i := 0; i < len(eroot.Nodes); i++ {
		e := eroot.Nodes[i]
		t := troot.Nodes[i]

		valid, err := compareNodes(e, t)

		if !valid {
			return valid, err
		}
	}

	return true, nil
}
