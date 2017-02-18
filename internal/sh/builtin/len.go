package builtin

import (
	"strconv"

	"github.com/NeowayLabs/nash/errors"
	"github.com/NeowayLabs/nash/sh"
)

type (
	LenFn struct {
		arg sh.Obj
	}
)

func newLenFn() *LenFn {
	return &LenFn{}
}

func (lenfn *LenFn) ArgNames() []string {
	return []string{"list"}
}

func (lenfn *LenFn) lenresult(res int) sh.Obj {
	return sh.NewStrObj(strconv.Itoa(res))
}

func (lenfn *LenFn) Run() (sh.Obj, error) {
	if lenfn.arg.Type() == sh.ListType {
		arglist := lenfn.arg.(*sh.ListObj)
		return lenresult(len(arglist.List())), nil
	}
	argstr := lenfn.arg.(*sh.StrObj)
	return lenresult(len(argstr.Str())), nil
}

func (lenfn *LenFn) SetArgs(args []sh.Obj) error {
	if len(args) != 1 {
		return errors.NewError("lenfn expects one argument")
	}

	obj := args[0]

	if obj.Type() != sh.ListType && obj.Type() != sh.StringType {
		return errors.NewError("lenfn expects a list or a string, but a %s was provided", obj.Type())
	}

	lenfn.arg = obj
	return nil
}
