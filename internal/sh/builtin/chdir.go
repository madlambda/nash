package builtin

import (
	"os"

	"github.com/NeowayLabs/nash/errors"
	"github.com/NeowayLabs/nash/sh"
)

type (
	chdirFn struct {
		arg string
	}
)

func newChdir() *chdirFn {
	return &chdirFn{}
}

func (chdir *chdirFn) ArgNames() []string {
	return []string{"dir"}
}

func (chdir *chdirFn) Run() (sh.Obj, error) {
	return nil, os.Chdir(chdir.arg)
}

func (chdir *chdirFn) SetArgs(args []sh.Obj) error {
	if len(args) != 1 {
		return errors.NewError("chdir expects one argument")
	}

	obj := args[0]

	if obj.Type() != sh.StringType {
		return errors.NewError("chdir expects a string, but a %s was provided", obj.Type())
	}

	if objstr, ok := obj.(*sh.StrObj); ok {
		chdir.arg = objstr.Str()
		return nil
	}

	return errors.NewError("internal error: object of wrong type")
}
