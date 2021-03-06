package builtin

import (
	"fmt"
	"io"

	"github.com/madlambda/nash/errors"
	"github.com/madlambda/nash/sh"
)

type (
	formatFn struct {
		fmt  string
		args []interface{}
	}
)

func newFormat() *formatFn {
	return &formatFn{}
}

func (f *formatFn) ArgNames() []sh.FnArg {
	return []sh.FnArg{
		sh.NewFnArg("fmt", false),
		sh.NewFnArg("arg...", true),
	}
}

func (f *formatFn) Run(in io.Reader, out io.Writer, err io.Writer) ([]sh.Obj, error) {
	return []sh.Obj{sh.NewStrObj(fmt.Sprintf(f.fmt, f.args...))}, nil
}

func (f *formatFn) SetArgs(args []sh.Obj) error {
	if len(args) == 0 {
		return errors.NewError("format expects at least 1 argument")
	}

	f.fmt = args[0].String()
	f.args = nil

	for _, arg := range args[1:] {
		f.args = append(f.args, arg.String())
	}

	return nil
}
