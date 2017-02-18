package builtin

import (
	"github.com/NeowayLabs/nash/errors"
	"github.com/NeowayLabs/nash/sh"
)

type (
	appendFn struct {
		obj []sh.Obj
		arg sh.Obj
	}
)

func newAppend() *appendFn {
	return &appendFn{}
}

func (appendfn *appendFn) ArgNames() []string {
	return []string{"list"}
}

func (appendfn *appendFn) Run() (sh.Obj, error) {
	newobj := append(appendfn.obj, appendfn.arg)
	return sh.NewListObj(newobj), nil
}

func (appendfn *appendFn) SetArgs(args []sh.Obj) error {
	if len(args) != 2 {
		return errors.NewError("appendfn expects two arguments")
	}

	obj := args[0]

	if obj.Type() != sh.ListType {
		return errors.NewError("appendfn expects a list as first argument, but a %s[%s] was provided", obj, obj.Type())
	}

	arg := args[1]

	if objlist, ok := obj.(*sh.ListObj); ok {
		appendfn.obj = objlist.List()
		appendfn.arg = arg
		return nil
	}

	return errors.NewError("internal error: object of wrong type")
}
