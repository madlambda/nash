package builtin

import (
	"fmt"
	"io"
	"os"

	"github.com/madlambda/nash/errors"
	"github.com/madlambda/nash/sh"
)

type (
	chdirFn struct {
		arg string
	}
)

func newChdir() *chdirFn {
	return &chdirFn{}
}

func (chdir *chdirFn) ArgNames() []sh.FnArg {
	return []sh.FnArg{
		sh.NewFnArg("dir", false),
	}
}

func (chdir *chdirFn) Run(in io.Reader, out io.Writer, ioerr io.Writer) ([]sh.Obj, error) {
	err := os.Chdir(chdir.arg)
	if err != nil {
		err = fmt.Errorf("builtin: chdir: error[%s] path[%s]", err, chdir.arg)
	}
	return nil, err
}

func (chdir *chdirFn) SetArgs(args []sh.Obj) error {
	if len(args) != 1 {
		return errors.NewError("chdir expects one argument, but received %q", args)
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
