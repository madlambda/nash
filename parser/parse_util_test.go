package parser

import (
	"fmt"
	"reflect"

	"github.com/NeowayLabs/nash/ast"
	"github.com/NeowayLabs/nash/token"
)

func newSimpleArg(pos token.Pos, n string, typ ast.ArgType) *ast.Arg {
	arg := ast.NewArg(pos, typ)
	arg.SetString(n)
	return arg
}

func comparePosition(expected token.Pos, value token.Pos) (bool, error) {
	if expected != value {
		return false, fmt.Errorf("Position mismatch: %d != %d", expected, value)
	}

	return true, nil
}

func compareArg(expected *ast.Arg, value *ast.Arg) (bool, error) {
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

	if expected.Value() != value.Value() {
		return false, fmt.Errorf("Argument value differs: '%s' != '%s'", expected.Value(), value.Value())
	}

	ev := expected
	vv := value

	if ev.IsQuoted() != vv.IsQuoted() {
		return false, fmt.Errorf("Variable differs in IsQuoted: (%v, %v)", ev.IsQuoted(), vv.IsQuoted())
	}

	if ev.IsConcat() != vv.IsConcat() ||
		ev.IsVariable() != vv.IsVariable() ||
		ev.IsList() != vv.IsList() {
		return false, fmt.Errorf("Variable differs in isConcat(%v, %v) || isVariable(%v, %v) || isList(%v, %v)\nExpected Node(%s) = %v\nParsed node(%s): %v", ev.IsConcat(), vv.IsConcat(), ev.IsVariable(), vv.IsVariable(),
			ev.IsList(), vv.IsList(), ev.ArgType(), ev, vv.ArgType(), vv)
	}

	if len(ev.Concat()) != len(vv.Concat()) {
		return false, fmt.Errorf("Variable list concats length differs (%v, %v). %d != %d", ev, vv, len(ev.Concat()), len(vv.Concat()))
	}

	econcat := ev.Concat()
	vconcat := vv.Concat()

	for j := 0; j < len(econcat); j++ {
		ce := econcat[j]
		cv := vconcat[j]

		ok, err := compareArg(ce, cv)

		if !ok {
			return ok, err
		}

	}

	if len(ev.List()) != len(vv.List()) {
		return false, fmt.Errorf("Variable list length differs (%v, %v). %d != %d", ev, vv, len(ev.List()), len(vv.List()))
	}

	elist := ev.List()
	vlist := vv.List()

	for j := 0; j < len(ev.List()); j++ {
		ce := elist[j]
		cv := vlist[j]

		ok, err := compareArg(ce, cv)

		if !ok {
			return ok, err
		}

	}

	return true, nil
}

func compareShowEnvNode(expected, value *ast.ShowEnvNode) (bool, error) {
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

func compareImportNode(expected, value *ast.ImportNode) (bool, error) {
	if expected == nil && value == nil {
		return true, nil
	}

	if (expected == nil) != (value == nil) {
		return false, fmt.Errorf("One of the nodecommand are nil")
	}

	if ok, err := compareArg(expected.Path(), value.Path()); !ok {
		return false, err
	}

	if ok, err := comparePosition(expected.Position(), value.Position()); !ok {
		return ok, fmt.Errorf(" CompareImportNode (%v, %v)-> %s", expected, value, err.Error())
	}

	return true, nil
}

func compareCdNode(expected, value *ast.CdNode) (bool, error) {
	if expected == nil && value == nil {
		return true, nil
	}

	if (expected == nil) != (value == nil) {
		return false, fmt.Errorf("One of the nodecommand are nil")
	}

	if ok, err := compareArg(expected.Dir(), value.Dir()); !ok {
		return false, err
	}

	if ok, err := comparePosition(expected.Position(), value.Position()); !ok {
		return ok, fmt.Errorf(" CompareCdNode (%v, %v)-> %s", expected, value, err.Error())
	}

	return true, nil
}

func comparePipeNode(expected, value *ast.PipeNode) (bool, error) {
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

func compareCommandNode(expected, value *ast.CommandNode) (bool, error) {
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

func compareCommentNode(expected, value *ast.CommentNode) (bool, error) {
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

func compareSetAssignmentNode(expected, value *ast.SetAssignmentNode) (bool, error) {
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

func compareAssignmentNode(expected, value *ast.AssignmentNode) (bool, error) {
	if expected == nil && value == nil {
		return true, nil
	}

	if (expected == nil) != (value == nil) {
		return false, fmt.Errorf("Only one of the nodes are nil. %v != %v", expected, value)
	}

	if expected.name != value.name {
		return false, fmt.Errorf("Variable name differs. '%s' != '%s'", expected.name, value.name)
	}

	if ok, err := compareArg(expected.Value(), value.Value()); !ok {
		return false, err
	}

	return true, nil
}

func compareRforkNode(expected, value *ast.RforkNode) (bool, error) {
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

func compareDefault(expected, value ast.Node) (bool, error) {
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

func compareFnInvNode(expected, value *ast.FnInvNode) (bool, error) {
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

func compareFnDeclNode(expected, value *ast.FnDeclNode) (bool, error) {
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
			expected, value, len(eargs), len(vargs))
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

func compareForNode(expected, value *ast.ForNode) (bool, error) {
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

func compareReturnNode(expected, value *ast.ReturnNode) (bool, error) {
	if ok, err := compareDefault(expected, value); !ok {
		return ok, err
	}

	erets := expected.Return()
	vrets := value.Return()

	if ok, err := compareArg(erets, vrets); !ok {
		return false, err
	}

	return true, nil
}

func compareDumpNode(expected, value *ast.DumpNode) (bool, error) {
	if ok, err := compareDefault(expected, value); !ok {
		return ok, err
	}

	efname := expected.Filename()
	vfname := value.Filename()

	return compareArg(efname, vfname)
}

func compareBindFnNode(expected, value *ast.BindFnNode) (bool, error) {
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

func compareCmdAssignmentNode(expected, value *ast.CmdAssignmentNode) (bool, error) {
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
		return compareCommandNode(ecmd.(*ast.CommandNode), vcmd.(*CommandNode))
	case NodePipe:
		return comparePipeNode(ecmd.(*ast.PipeNode), vcmd.(*PipeNode))
	}

	return false, fmt.Errorf("Unexpected type %s", ecmd.Type())
}

func compareIfNode(expected, value *ast.IfNode) (bool, error) {
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

func compareNodes(expected ast.Node, value ast.Node) (bool, error) {
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
	case *ast.ImportNode:
		ec := expected.(*ast.ImportNode)
		vc := value.(*ast.ImportNode)

		valid, err = compareImportNode(ec, vc)
	case *ast.ShowEnvNode:
		ec := expected.(*ast.ShowEnvNode)
		vc := value.(*ast.ShowEnvNode)

		valid, err = compareShowEnvNode(ec, vc)
	case *ast.SetAssignmentNode:
		ec := expected.(*ast.SetAssignmentNode)
		vc := value.(*ast.SetAssignmentNode)

		valid, err = compareSetAssignmentNode(ec, vc)
	case *ast.AssignmentNode:
		ec := expected.(*ast.AssignmentNode)
		vc := value.(*ast.AssignmentNode)
		valid, err = compareAssignmentNode(ec, vc)
	case *ast.CdNode:
		ec := expected.(*ast.CdNode)
		vc := value.(*ast.CdNode)
		valid, err = compareCdNode(ec, vc)
	case *ast.CommandNode:
		ec := expected.(*ast.CommandNode)
		vc := value.(*ast.CommandNode)
		valid, err = compareCommandNode(ec, vc)
	case *ast.PipeNode:
		ec := expected.(*ast.PipeNode)
		vc := value.(*ast.PipeNode)
		valid, err = comparePipeNode(ec, vc)
	case *ast.CommentNode:
		ec := expected.(*ast.CommentNode)
		vc := value.(*ast.CommentNode)
		valid, err = compareCommentNode(ec, vc)
	case *ast.RforkNode:
		er := expected.(*ast.RforkNode)
		vr := value.(*ast.RforkNode)
		valid, err = compareRforkNode(er, vr)
	case *ast.IfNode:
		ec := expected.(*ast.IfNode)
		vc := value.(*ast.IfNode)
		valid, err = compareIfNode(ec, vc)
	case *ast.FnDeclNode:
		ec := expected.(*ast.FnDeclNode)
		vc := value.(*ast.FnDeclNode)
		valid, err = compareFnDeclNode(ec, vc)
	case *ast.FnInvNode:
		ec := expected.(*ast.FnInvNode)
		vc := value.(*ast.FnInvNode)
		valid, err = compareFnInvNode(ec, vc)
	case *ast.CmdAssignmentNode:
		ec := expected.(*ast.CmdAssignmentNode)
		vc := value.(*ast.CmdAssignmentNode)
		valid, err = compareCmdAssignmentNode(ec, vc)
	case *ast.BindFnNode:
		ec := expected.(*ast.BindFnNode)
		vc := value.(*ast.BindFnNode)
		valid, err = compareBindFnNode(ec, vc)

	case *ast.DumpNode:
		ec := expected.(*ast.DumpNode)
		vc := value.(*ast.DumpNode)
		valid, err = compareDumpNode(ec, vc)
	case *ast.ReturnNode:
		ec := expected.(*ast.ReturnNode)
		vc := value.(*ast.ReturnNode)
		valid, err = compareReturnNode(ec, vc)
	case *ast.ForNode:
		ec := expected.(*ast.ForNode)
		vc := value.(*ast.ForNode)
		valid, err = compareForNode(ec, vc)
	default:
		return false, fmt.Errorf("Type %v not comparable yet", v)
	}

	if !valid {
		return valid, err
	}

	return true, nil
}

func compare(expected *ast.Tree, tr *ast.Tree) (bool, error) {
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
