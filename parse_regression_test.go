package nash

import "testing"

func TestParseIssue38(t *testing.T) {
	expected := NewTree("parse issue38")

	ln := NewListNode()

	fnInv := NewFnInvNode(0, "cd")

	arg := NewArg(0, ArgConcat)

	args := make([]*Arg, 3)

	arg1 := NewArg(0, ArgVariable)
	arg1.SetString("$GOPATH")

	arg2 := NewArg(0, ArgQuoted)
	arg2.SetString("/src/")

	arg3 := NewArg(0, ArgVariable)
	arg3.SetString("$path")

	args[0] = arg1
	args[1] = arg2
	args[2] = arg3

	arg.SetConcat(args)

	fnInv.AddArg(arg)

	ln.Push(fnInv)
	expected.Root = ln

	parserTestTable("parse issue38", `cd($GOPATH+"/src/"+$path)`, expected, t, true)
}
