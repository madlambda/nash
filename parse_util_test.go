package nash

import (
	"fmt"
	"reflect"
)

func newSimpleArg(pos Pos, n string, typ ArgType) *Arg {
	arg := NewArg(pos, typ)
	arg.SetString(n)
	return arg
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

	ev := expected
	vv := value

	if ev.IsQuoted() != vv.IsQuoted() {
		return false, fmt.Errorf("Variable differs in IsQuoted: (%v, %v)", ev.IsQuoted(), vv.IsQuoted())
	}

	if ev.IsConcat() != vv.IsConcat() ||
		ev.IsVariable() != vv.IsVariable() {
		return false, fmt.Errorf("Variable differs in isConcat(%v, %v) || isVariable(%v, %v)\nExpected Node(%s) = %v\nParsed node(%s): %v", ev.IsConcat(), vv.IsConcat(), ev.IsVariable(), vv.IsVariable(), ev.argType, ev, vv.argType, vv)
	}

	if len(ev.concat) != len(vv.concat) {
		return false, fmt.Errorf("Variable list concats length differs (%v, %v). %d != %d", ev, vv, len(ev.concat), len(vv.concat))
	}

	for j := 0; j < len(ev.concat); j++ {
		ce := ev.concat[j]
		cv := vv.concat[j]

		ok, err := compareArg(ce, cv)

		if !ok {
			return ok, err
		}

	}

	return true, nil
}

func compareShowEnvNode(expected, value *ShowEnvNode) (bool, error) {
	if expected == nil && value == nil {
		return true, nil
	}

	if (expected == nil) != (value == nil) {
		return false, fmt.Errorf("One of the ShowEnvNode are nil")
	}

	if ok, err := comparePosition(expected.Position(), value.Position()); !ok {
		return ok, fmt.Errorf(" CompareShowEnvNode (%v, %v)-> %s", expected, value, err.Error())
	}

	return true, nil
}

func compareImportNode(expected, value *ImportNode) (bool, error) {
	if expected == nil && value == nil {
		return true, nil
	}

	if (expected == nil) != (value == nil) {
		return false, fmt.Errorf("One of the nodecommand are nil")
	}

	if expected.path.val != value.path.val {
		return false, fmt.Errorf("Expected.path.val (%v) != value.path.val (%v)", expected.path.val, value.path.val)
	}

	if ok, err := comparePosition(expected.Position(), value.Position()); !ok {
		return ok, fmt.Errorf(" CompareImportNode (%v, %v)-> %s", expected, value, err.Error())
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

	if expected.dir != nil && value.dir != nil && expected.dir.val != value.dir.val {
		return false, fmt.Errorf("Expected.dir.val (%v) != value.dir.val (%v)", expected.dir.val, value.dir.val)
	}

	if ok, err := comparePosition(expected.Position(), value.Position()); !ok {
		return ok, fmt.Errorf(" CompareCdNode (%v, %v)-> %s", expected, value, err.Error())
	}

	return true, nil
}

func comparePipeNode(expected, value *PipeNode) (bool, error) {
	if ok, err := compareDefault(expected, value); !ok {
		return ok, err
	}

	ecmds := expected.Commands()
	vcmds := value.Commands()

	if len(ecmds) != len(vcmds) {
		return false, fmt.Errorf(" comparePipeNode - length differs: %d != %d", len(ecmds), len(vcmds))
	}

	for i := 0; i < len(ecmds); i++ {
		ecmd := ecmds[i]
		vcmd := vcmds[i]

		ok, err := compareCommandNode(ecmd, vcmd)

		if !ok {
			return ok, err
		}
	}

	if expected.String() != value.String() {
		return false, fmt.Errorf("String differs: '%s' != '%s'", expected.String(), value.String())
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

		valid, err := compareArg(ea, va)

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

func compareSetAssignmentNode(expected, value *SetAssignmentNode) (bool, error) {
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

func compareAssignmentNode(expected, value *AssignmentNode) (bool, error) {
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

		ok, err := compareArg(ev, vv)

		if !ok {
			return ok, err
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

	if ok, err := compareArg(expected.arg, value.arg); !ok {
		return ok, fmt.Errorf("CompareRforkNode (%v, %v) -> %s", expected, value, err.Error())
	}

	expectedTree := expected.Tree()
	valueTree := value.Tree()

	return compare(expectedTree, valueTree)
}

func compareDefault(expected, value Node) (bool, error) {
	if expected == nil && value == nil {
		return true, nil
	}

	if (expected == nil) != (value == nil) {
		return false, fmt.Errorf("Only one of the nodes are nil. %v != %v", expected, value)
	}

	if ok, err := comparePosition(expected.Position(), value.Position()); !ok {
		return ok, fmt.Errorf(" CompareIfNode (%v, %v) -> %s", expected, value, err.Error())
	}

	return true, nil
}

func compareFnInvNode(expected, value *FnInvNode) (bool, error) {
	if ok, err := compareDefault(expected, value); !ok {
		return ok, err
	}

	ename := expected.Name()
	vname := value.Name()

	if ename != vname {
		return false, fmt.Errorf(" CompareFnInvNode(%v, %v) -> Name differs: '%s' != '%s'",
			expected, value, ename, vname)
	}

	if expected.String() != value.String() {
		return false, fmt.Errorf("Reverse failed: '%s' != '%s'", expected.String(), value.String())
	}

	return true, nil
}

func compareFnDeclNode(expected, value *FnDeclNode) (bool, error) {
	if ok, err := compareDefault(expected, value); !ok {
		return ok, err
	}

	ename := expected.Name()
	vname := value.Name()

	if ename != vname {
		return false, fmt.Errorf(" CompareFnDeclNode (%v, %v) -> '%s' != '%s'. name differs.",
			expected, value, ename, vname)
	}

	eargs := expected.Args()
	vargs := value.Args()

	if len(eargs) != len(vargs) {
		return false, fmt.Errorf(" CompareFnDeclNode (%v, %v) -> '%d' != '%d'. Length differs.",
			expected, value, ename, vname)
	}

	for i := 0; i < len(eargs); i++ {
		earg := eargs[i]
		varg := vargs[i]

		if earg != varg {
			return false, fmt.Errorf(" CompareFnDeclNode (%v, %v) -> '%s' != '%s'. arg differs.",
				expected, value, earg, varg)
		}
	}

	return compare(expected.Tree(), value.Tree())
}

func compareForNode(expected, value *ForNode) (bool, error) {
	if ok, err := compareDefault(expected, value); !ok {
		return ok, err
	}

	eid := expected.Identifier()
	vid := value.Identifier()

	if eid != vid {
		return false, fmt.Errorf("for identifier differs. '%s' != '%s'", eid, vid)
	}

	evar := expected.InVar()
	vvar := value.InVar()

	if evar != vvar {
		return false, fmt.Errorf("for in variable differ. '%s' != '%s'", evar, vvar)
	}

	return compare(expected.Tree(), value.Tree())
}

func compareReturnNode(expected, value *ReturnNode) (bool, error) {
	if ok, err := compareDefault(expected, value); !ok {
		return ok, err
	}

	erets := expected.Return()
	vrets := value.Return()

	if len(erets) != len(vrets) {
		return false, fmt.Errorf("compareReturnNode(%v, %v) -> Length differs (%d) != (%d)", expected, value, len(erets), len(vrets))
	}

	for i := 0; i < len(erets); i++ {
		eret := erets[i]
		vret := vrets[i]

		if ok, err := compareArg(eret, vret); !ok {
			return ok, err
		}
	}

	return true, nil
}

func compareDumpNode(expected, value *DumpNode) (bool, error) {
	if ok, err := compareDefault(expected, value); !ok {
		return ok, err
	}

	efname := expected.Filename()
	vfname := value.Filename()

	return compareArg(efname, vfname)
}

func compareBindFnNode(expected, value *BindFnNode) (bool, error) {
	if ok, err := compareDefault(expected, value); !ok {
		return ok, err
	}

	ename := expected.Name()
	vname := value.Name()

	if ename != vname {
		return false, fmt.Errorf(" CompareBindFnNode (%v, %v) -> '%s' != '%s'", expected, value, ename, vname)
	}

	cmdename := expected.CmdName()
	cmdvname := value.CmdName()

	if cmdename != cmdvname {
		return false, fmt.Errorf(" CompareBindFnNode (%v, %v) -> '%s' != '%s'", expected, value, cmdename, cmdvname)
	}

	return true, nil
}

func compareCmdAssignmentNode(expected, value *CmdAssignmentNode) (bool, error) {
	if ok, err := compareDefault(expected, value); !ok {
		return ok, err
	}

	ename := expected.Name()
	vname := value.Name()

	if ename != vname {
		return false, fmt.Errorf(" CompareCmdAssignmentnode (%v, %v) -> '%s' != '%s'", expected, value, ename, vname)
	}

	ecmd := expected.Command()
	vcmd := value.Command()

	if ecmd.Type() != vcmd.Type() {
		return false, fmt.Errorf("Node type differs: %v != %v", ecmd.Type(), vcmd.Type())
	}

	switch ecmd.Type() {
	case NodeCommand:
		return compareCommandNode(ecmd.(*CommandNode), vcmd.(*CommandNode))
	case NodePipe:
		return comparePipeNode(ecmd.(*PipeNode), vcmd.(*PipeNode))
	}

	return false, fmt.Errorf("Unexpected type %s", ecmd.Type())
}

func compareIfNode(expected, value *IfNode) (bool, error) {
	if ok, err := compareDefault(expected, value); !ok {
		return ok, err
	}

	elvalue := expected.Lvalue()
	ervalue := expected.Rvalue()
	vlvalue := value.Lvalue()
	vrvalue := value.Rvalue()

	if ok, err := compareArg(elvalue, vlvalue); !ok {
		return ok, fmt.Errorf("CompareIfNode (%v, %v) -> %s", expected, value, err.Error())
	}

	if ok, err := compareArg(ervalue, vrvalue); !ok {
		return ok, fmt.Errorf("CompareIfNode (%v, %v) -> %s", expected, value, err.Error())
	}

	if expected.Op() != value.Op() {
		return false, fmt.Errorf("CompareIfNode (%v, %v) -> Operation differ: %s != %s", expected, value, expected.Op(), value.Op())
	}

	expectedTree := expected.IfTree()
	valueTree := value.IfTree()

	ok, err := compare(expectedTree, valueTree)

	if !ok {
		return ok, err
	}

	expectedTree = expected.ElseTree()
	valueTree = expected.ElseTree()

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
		return false, fmt.Errorf("Node type differs: %v != %v", etype, vtype)
	}

	eitype := reflect.TypeOf(expected)
	vitype := reflect.TypeOf(value)

	if eitype.Kind() != vitype.Kind() {
		return false, fmt.Errorf("Node type differs: %v != %v", eitype.Kind(), vitype.Kind())
	}

	switch v := expected.(type) {
	case *ImportNode:
		ec := expected.(*ImportNode)
		vc := value.(*ImportNode)

		valid, err = compareImportNode(ec, vc)
	case *ShowEnvNode:
		ec := expected.(*ShowEnvNode)
		vc := value.(*ShowEnvNode)

		valid, err = compareShowEnvNode(ec, vc)
	case *SetAssignmentNode:
		ec := expected.(*SetAssignmentNode)
		vc := value.(*SetAssignmentNode)

		valid, err = compareSetAssignmentNode(ec, vc)
	case *AssignmentNode:
		ec := expected.(*AssignmentNode)
		vc := value.(*AssignmentNode)
		valid, err = compareAssignmentNode(ec, vc)
	case *CdNode:
		ec := expected.(*CdNode)
		vc := value.(*CdNode)
		valid, err = compareCdNode(ec, vc)
	case *CommandNode:
		ec := expected.(*CommandNode)
		vc := value.(*CommandNode)
		valid, err = compareCommandNode(ec, vc)
	case *PipeNode:
		ec := expected.(*PipeNode)
		vc := value.(*PipeNode)
		valid, err = comparePipeNode(ec, vc)
	case *CommentNode:
		ec := expected.(*CommentNode)
		vc := value.(*CommentNode)
		valid, err = compareCommentNode(ec, vc)
	case *RforkNode:
		er := expected.(*RforkNode)
		vr := value.(*RforkNode)
		valid, err = compareRforkNode(er, vr)
	case *IfNode:
		ec := expected.(*IfNode)
		vc := value.(*IfNode)
		valid, err = compareIfNode(ec, vc)
	case *FnDeclNode:
		ec := expected.(*FnDeclNode)
		vc := value.(*FnDeclNode)
		valid, err = compareFnDeclNode(ec, vc)
	case *FnInvNode:
		ec := expected.(*FnInvNode)
		vc := value.(*FnInvNode)
		valid, err = compareFnInvNode(ec, vc)
	case *CmdAssignmentNode:
		ec := expected.(*CmdAssignmentNode)
		vc := value.(*CmdAssignmentNode)
		valid, err = compareCmdAssignmentNode(ec, vc)
	case *BindFnNode:
		ec := expected.(*BindFnNode)
		vc := value.(*BindFnNode)
		valid, err = compareBindFnNode(ec, vc)

	case *DumpNode:
		ec := expected.(*DumpNode)
		vc := value.(*DumpNode)
		valid, err = compareDumpNode(ec, vc)
	case *ReturnNode:
		ec := expected.(*ReturnNode)
		vc := value.(*ReturnNode)
		valid, err = compareReturnNode(ec, vc)
	case *ForNode:
		ec := expected.(*ForNode)
		vc := value.(*ForNode)
		valid, err = compareForNode(ec, vc)
	default:
		return false, fmt.Errorf("Type %v not comparable yet", v)
	}

	if !valid {
		return valid, err
	}

	return true, nil
}

func compare(expected *Tree, tr *Tree) (bool, error) {
	if expected == nil && tr == nil {
		return true, nil
	}

	if (expected == nil) != (tr == nil) {
		return false, fmt.Errorf("only one of the expected and tree are nil (%v, %v)", expected, tr)
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
