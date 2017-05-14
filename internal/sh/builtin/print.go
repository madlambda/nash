package builtin

import (
	"fmt"
	"io"

	"github.com/NeowayLabs/nash/errors"
	"github.com/NeowayLabs/nash/sh"
)

type (
	printFn struct {
		fmt  string
		args []interface{}
	}
)

func newPrint() *printFn {
	return &printFn{}
}

func (p *printFn) ArgNames() []sh.FnArg {
	return []sh.FnArg{
		sh.NewFnArg("fmt", false),
		sh.NewFnArg("arg...", true),
	}
}

func (p *printFn) Run(in io.Reader, out io.Writer, err io.Writer) ([]sh.Obj, error) {
	fmt.Fprintf(out, p.fmt, p.args...)
	return nil, nil
}

func (p *printFn) SetArgs(args []sh.Obj) error {
	if len(args) == 0 {
		return errors.NewError("print expects at least 1 argument")
	}

	p.fmt = args[0].String()
	p.args = nil
	for _, arg := range args[1:] {
		p.args = append(p.args, arg.String())
	}

	return nil
}
