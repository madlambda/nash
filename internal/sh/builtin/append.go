package builtin

import (
	"io"

	"github.com/NeowayLabs/nash/errors"
	"github.com/NeowayLabs/nash/sh"
)

type (
	appendFn struct {
		obj  []sh.Obj
		args []sh.Obj
	}
)

func newAppend() *appendFn {
	return &appendFn{}
}

func (appendfn *appendFn) ArgNames() []sh.FnArg {
	return []sh.FnArg{
		sh.NewFnArg("list", false),
		sh.NewFnArg("value...", true), // variadic
	}
}

func (appendfn *appendFn) Run(in io.Reader, out io.Writer, err io.Writer) ([]sh.Obj, error) {
	newobj := append(appendfn.obj, appendfn.args...)
	return []sh.Obj{sh.NewListObj(newobj)}, nil
}

func (appendfn *appendFn) SetArgs(args []sh.Obj) error {
	if len(args) < 2 {
		return errors.NewError("appendfn expects at least two arguments")
	}

	obj := args[0]
	if obj.Type() != sh.ListType {
		return errors.NewError("appendfn expects a list as first argument, but a %s was provided",
			obj.Type())
	}

	values := args[1:]
	if objlist, ok := obj.(*sh.ListObj); ok {
		appendfn.obj = objlist.List()
		appendfn.args = values
		return nil
	}

	return errors.NewError("internal error: object of wrong type")
}
