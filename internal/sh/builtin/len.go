package builtin

import (
	"strconv"

	"github.com/NeowayLabs/nash/errors"
	"github.com/NeowayLabs/nash/sh"
)

type (
	lenFn struct {
		arg sh.Obj
	}
)

func newLen() *lenFn {
	return &lenFn{}
}

func (l *lenFn) ArgNames() []string {
	return []string{"list"}
}

func lenresult(res int) sh.Obj {
	return sh.NewStrObj(strconv.Itoa(res))
}

func (l *lenFn) Run() (sh.Obj, error) {
	if l.arg.Type() == sh.ListType {
		arglist := l.arg.(*sh.ListObj)
		return lenresult(len(arglist.List())), nil
	}
	argstr := l.arg.(*sh.StrObj)
	return lenresult(len(argstr.Str())), nil
}

func (l *lenFn) SetArgs(args []sh.Obj) error {
	if len(args) != 1 {
		return errors.NewError("lenfn expects one argument")
	}

	obj := args[0]

	if obj.Type() != sh.ListType && obj.Type() != sh.StringType {
		return errors.NewError("lenfn expects a list or a string, but a %s was provided", obj.Type())
	}

	l.arg = obj
	return nil
}
