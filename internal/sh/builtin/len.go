package builtin

import (
	"io"
	"strconv"

	"github.com/NeowayLabs/nash/errors"
	"github.com/NeowayLabs/nash/sh"
)

type (
	lenFn struct {
		arg sh.Enumerable
	}
)

func newLen() *lenFn {
	return &lenFn{}
}

func (l *lenFn) ArgNames() []sh.FnArg {
	return []sh.FnArg{
		sh.NewFnArg("list", false),
	}
}

func lenresult(res int) []sh.Obj {
	return []sh.Obj{sh.NewStrObj(strconv.Itoa(res))}
}

func (l *lenFn) Run(in io.Reader, out io.Writer, err io.Writer) ([]sh.Obj, error) {
	return lenresult(l.arg.Len()), nil
}

func (l *lenFn) SetArgs(args []sh.Obj) error {
	if len(args) != 1 {
		return errors.NewError("lenfn expects one argument")
	}

	obj := args[0]
	indexer, err := sh.NewEnumerable(obj)
	if err != nil {
		return errors.NewError("len:error[%s]", err)
	}

	l.arg = indexer
	return nil
}
