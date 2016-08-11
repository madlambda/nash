package ast

import (
	"fmt"
	"reflect"

	"github.com/NeowayLabs/nash/token"
)

func cmpPosition(expected token.Pos, value token.Pos) (bool, error) {
	if expected != value {
		return false, fmt.Errorf("Position mismatch: %d != %d", expected, value)
	}

	return true, nil
}

func cmpCommon(expected, value Node) (bool, error) {
	if expected == value {
		return true, nil
	}

	if ok, err := cmpPosition(expected.Position(), value.Position()); !ok {
		return ok, fmt.Errorf(" CompareIfNode (%v, %v) -> %s", expected, value, err.Error())
	}

	return true, nil
}

func cmpExpr(expected Expr, value Expr) (bool, error) {
	if ok, err := cmpCommon(expected, value); !ok {
		return ok, err
	}

	if expected.Type() != value.Type() {
		return false, fmt.Errorf("  Type differs: %d != %d", expected.Type(), value.Type())
	}

	if expected.Value() != value.Value() {
		return false, fmt.Errorf("  Argument value differs: '%s' != '%s'", expected.Value(), value.Value())
	}

	ev := expected
	vv := value

	if ev.Type() != vv.Type() {
		return false, fmt.Errorf("  ArgType differs: %v != %v", ev.Type(), vv.Type())
	}


	eitype := reflect.TypeOf(expected)
	vitype := reflect.TypeOf(value)

	if eitype.Kind() != vitype.Kind() {
		return false, fmt.Errorf("Node type differs: %v != %v", eitype.Kind(), vitype.Kind())
	}

	switch v := expected.(type) {
	case *StringExpr:
		ok, err = cmpStringExpr(expected.(*StringExpr), value.(*StringExpr))
	case *VarExpr:
		ok, err = cmpVarExpr(expected.(*VarExpr), value.(*VarExpr))
	case *IndexExpr:
		ok, err = cmpIndexExpr(expected.(*IndexExpr), value.(*IndexExpr))
	case *ListExpr:
		ok, err = cmpListExpr(expected.(*ListExpr), value.(*ListExpr))
	case *ConcatExpr:
		ok, err = cmpConcatExpr(expected.(*ConcatExpr), value.(*ConcatExpr))
	default:
		return false, fmt.Errorf("Unexpected node: %s", expected)
	}

	if !ok {
		return ok, err
	}

	return true, nil
}

func cmpConcat

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

func cmpImport(expected, value *ImportNode) (bool, error) {
	if ok, err := cmpCommon(expected, value); !ok {
		return ok, err
	}

	if ok, err := cmpArg(expected.Path(), value.Path()); !ok {
		return false, err
	}

	return true, nil
}

func cmpCd(expected, value *CdNode) (bool, error) {
	if ok, err := cmpCommon(expected, value); !ok {
		return ok, err
	}

	if ok, err := cmpArg(expected.Dir(), value.Dir()); !ok {
		return false, err
	}

	return true, nil
}

func cmpPipe(expected, value *PipeNode) (bool, error) {
	if ok, err := cmpCommon(expected, value); !ok {
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

func cmpCommand(expected, value *CommandNode) (bool, error) {
	if ok, err := cmpCommon(expected, value); !ok {
		return ok, err
	}

	ename := expected.Name()
	vname := value.Name()

	if ename != vname {
		return false, fmt.Errorf("CommandNode: expected.name('%s') != value.name('%s')",
			ename, vname)
	}

	eargs := expected.Args()
	vargs := value.Args()

	if len(eargs) != len(vargs) {
		return false, fmt.Errorf("CommandNode: length of expected.args and value.args differs: %d != %d", len(eargs), len(vargs))
	}

	for i := 0; i < len(eargs); i++ {
		ea := eargs[i]
		va := vargs[i]

		valid, err := cmpArg(ea, va)

		if !valid {
			return valid, err
		}
	}

	return true, nil
}

func cmpComment(expected, value *CommentNode) (bool, error) {
	if ok, err := cmpCommon(expected, value); !ok {
		return ok, err
	}

	if expected.String() != value.String() {
		return false, fmt.Errorf("Comment val differ: '%s' != '%s'", expected.String(), value.String())
	}

	return true, nil
}

func cmpSetenv(expected, value *SetenvNode) (bool, error) {
	if ok, err := cmpCommon(expected, value); !ok {
		return ok, err
	}

	if expected.Identifier() != value.Identifier() {
		return false, fmt.Errorf("Set identifier mismatch. %s != %s", expected.Identifier(), value.Identifier())
	}

	return true, nil
}

func cmpAssignment(expected, value *AssignmentNode) (bool, error) {
	if ok, err := cmpCommon(expected, value); !ok {
		return ok, err
	}

	if expected.Identifier() != value.Identifier() {
		return false, fmt.Errorf("Variable name differs. '%s' != '%s'", expected.Identifier(), value.Identifier())
	}

	if ok, err := compareArg(expected.Value(), value.Value()); !ok {
		return false, err
	}

	return true, nil
}

func cmpRfork(expected, value *RforkNode) (bool, error) {
	if ok, err := cmpCommon(expected, value); !ok {
		return ok, err
	}

	if ok, err := compareArg(expected.Arg(), value.Arg()); !ok {
		return ok, fmt.Errorf("CompareRforkNode (%v, %v) -> %s", expected, value, err.Error())
	}

	expectedTree := expected.Tree()
	valueTree := value.Tree()

	return Compare(expectedTree, valueTree)
}

func cmpFnInv(expected, value *FnInvNode) (bool, error) {
	if ok, err := cmpCommon(expected, value); !ok {
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

func cmpFnDecl(expected, value *FnDeclNode) (bool, error) {
	if ok, err := cmpCommon(expected, value); !ok {
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

	return Cmp(expected.Tree(), value.Tree())
}

func cmpFor(expected, value *ForNode) (bool, error) {
	if ok, err := cmpCommon(expected, value); !ok {
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

	return Cmp(expected.Tree(), value.Tree())
}

func cmpReturn(expected, value *ReturnNode) (bool, error) {
	if ok, err := cmpCommon(expected, value); !ok {
		return ok, err
	}

	erets := expected.Return()
	vrets := value.Return()

	if ok, err := cmpArg(erets, vrets); !ok {
		return false, err
	}

	return true, nil
}

func cmpDump(expected, value *DumpNode) (bool, error) {
	if ok, err := cmpCommon(expected, value); !ok {
		return ok, err
	}

	efname := expected.Filename()
	vfname := value.Filename()

	return cmpArg(efname, vfname)
}

func cmpBindFn(expected, value *BindFnNode) (bool, error) {
	if ok, err := cmpCommon(expected, value); !ok {
		return ok, err
	}

	ename := expected.Name()
	vname := value.Name()

	if ename != vname {
		return false, fmt.Errorf(" CompareBindFnNode (%v, %v) -> '%s' != '%s'", expected,
			value, ename, vname)
	}

	cmdename := expected.CmdName()
	cmdvname := value.CmdName()

	if cmdename != cmdvname {
		return false, fmt.Errorf(" CompareBindFnNode (%v, %v) -> '%s' != '%s'", expected,
			value, cmdename, cmdvname)
	}

	return true, nil
}

func cmpCmdAssignment(expected, value *CmdAssignmentNode) (bool, error) {
	if ok, err := cmpCommon(expected, value); !ok {
		return ok, err
	}

	ename := expected.Name()
	vname := value.Name()

	if ename != vname {
		return false, fmt.Errorf(" CompareCmdAssignmentnode (%v, %v) -> '%s' != '%s'",
			expected, value, ename, vname)
	}

	ecmd := expected.Command()
	vcmd := value.Command()

	if ecmd.Type() != vcmd.Type() {
		return false, fmt.Errorf("Node type differs: %v != %v", ecmd.Type(), vcmd.Type())
	}

	switch ecmd.Type() {
	case NodeCommand:
		return cmpCommand(ecmd.(*CommandNode), vcmd.(*CommandNode))
	case NodePipe:
		return cmpPipe(ecmd.(*PipeNode), vcmd.(*PipeNode))
	case NodeFnInv:
		return cmpFnInv(ecmd.(*FnInvNode), vcmd.(*FnInvNode))
	}

	return false, fmt.Errorf("Unexpected type %s", ecmd.Type())
}

func CmpNode(expected Node, value Node) (bool, error) {
	var (
		valid = true
		err   error
	)

	if ok, err := cmpCommon(expected, value); !ok {
		return ok, err
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

		valid, err = cmpImport(ec, vc)
	case *SetAssignmentNode:
		ec := expected.(*SetAssignmentNode)
		vc := value.(*SetAssignmentNode)

		valid, err = cmpSetenv(ec, vc)
	case *AssignmentNode:
		ec := expected.(*AssignmentNode)
		vc := value.(*AssignmentNode)
		valid, err = cmpAssignment(ec, vc)
	case *CdNode:
		ec := expected.(*CdNode)
		vc := value.(*CdNode)
		valid, err = cmpCd(ec, vc)
	case *CommandNode:
		ec := expected.(*CommandNode)
		vc := value.(*CommandNode)
		valid, err = cmpCommand(ec, vc)
	case *PipeNode:
		ec := expected.(*PipeNode)
		vc := value.(*PipeNode)
		valid, err = cmpPipe(ec, vc)
	case *CommentNode:
		ec := expected.(*CommentNode)
		vc := value.(*CommentNode)
		valid, err = cmpComment(ec, vc)
	case *RforkNode:
		er := expected.(*RforkNode)
		vr := value.(*RforkNode)
		valid, err = cmpRfork(er, vr)
	case *IfNode:
		ec := expected.(*IfNode)
		vc := value.(*IfNode)
		valid, err = cmpIf(ec, vc)
	case *FnDeclNode:
		ec := expected.(*FnDeclNode)
		vc := value.(*FnDeclNode)
		valid, err = cmpFnDecl(ec, vc)
	case *FnInvNode:
		ec := expected.(*FnInvNode)
		vc := value.(*FnInvNode)
		valid, err = cmpFnInv(ec, vc)
	case *CmdAssignmentNode:
		ec := expected.(*CmdAssignmentNode)
		vc := value.(*CmdAssignmentNode)
		valid, err = cmpCmdAssignment(ec, vc)
	case *BindFnNode:
		ec := expected.(*BindFnNode)
		vc := value.(*BindFnNode)
		valid, err = cmpBindFn(ec, vc)

	case *DumpNode:
		ec := expected.(*DumpNode)
		vc := value.(*DumpNode)
		valid, err = cmpDump(ec, vc)
	case *ReturnNode:
		ec := expected.(*ReturnNode)
		vc := value.(*ReturnNode)
		valid, err = cmpReturn(ec, vc)
	case *ForNode:
		ec := expected.(*ForNode)
		vc := value.(*ForNode)
		valid, err = cmpFor(ec, vc)
	default:
		return false, fmt.Errorf("Type %v not comparable yet", v)
	}

	if !valid {
		return valid, err
	}

	return true, nil
}

func Cmp(expected *Tree, tr *Tree) (bool, error) {
	if expected == tr {
		return true, nil
	}

	en := expected.Name
	tn := expected.Name

	if en != tn {
		return false, fmt.Errorf("expected.Name != tr.Name. rxpected.Name='%s' and tr.Name='%s'",
			en, tn)
	}

	eroot := expected.Root
	troot := tr.Root

	if eroot == troot {
		return true, nil
	}

	if len(eroot.Nodes) != len(troot.Nodes) {
		return false, fmt.Errorf("Length differs. len(expected.Root) == %d and len(tr.Root) = %d",
			len(eroot.Nodes), len(troot.Nodes))
	}

	for i := 0; i < len(eroot.Nodes); i++ {
		e := eroot.Nodes[i]
		t := troot.Nodes[i]

		valid, err := CmpNode(e, t)

		if !valid {
			return valid, err
		}
	}

	return true, nil
}
