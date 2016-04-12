package cnt

import (
	"errors"
	"fmt"
	"reflect"
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
		t.Error(err)
		return
	}
}

func TestParseSimple(t *testing.T) {
	expected := NewTree("parser simple")
	ln := NewListNode()
	cmd := NewCommandNode(0, "echo")
	cmd.AddArg(NewArg(6, "hello world"))
	ln.Push(cmd)

	expected.Root = ln

	parserTestTable("parser simple", `echo "hello world"`, expected, t)
}

func TestParseWithShebang(t *testing.T) {
	expected := NewTree("parser shebang")
	ln := NewListNode()
	cmt := NewCommentNode(0, "#!/bin/cnt")
	cmd := NewCommandNode(11, "echo")
	cmd.AddArg(NewArg(16, "bleh"))
	ln.Push(cmt)
	ln.Push(cmd)

	expected.Root = ln

	parserTestTable("parser shebang", `#!/bin/cnt
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

func comparePosition(expected Pos, value Pos) (bool, error) {
	if expected != value {
		return false, fmt.Errorf("Position mismatch: %d != %d", expected, value)
	}

	return true, nil
}

func compareArgs(expected *Arg, value *Arg) (bool, error) {
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

		valid, err := compareArgs(&ea, &va)

		if !valid {
			return valid, err
		}
	}

	return true, nil
}

func compareCommentNode(expected *CommentNode, value *CommentNode) (bool, error) {
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
	case *CommandNode:
		ec := expected.(*CommandNode)
		vc := value.(*CommandNode)
		valid, err = compareCommandNode(ec, vc)
	case *CommentNode:
		ec := expected.(*CommentNode)
		vc := value.(*CommentNode)
		valid, err = compareCommentNode(ec, vc)
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
