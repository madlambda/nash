package builtin

import (
	"fmt"
	"io"

	"github.com/NeowayLabs/nash/errors"
	"github.com/NeowayLabs/nash/sh"
)

type (
	printfFn struct {
		fmt  string
		args []interface{}
	}
)

func newPrintf() *printfFn {
	return &printfFn{}
}

func (s *printfFn) ArgNames() []string {
	return []string{"fmt", "args"}
}

func (s *printfFn) Run(
	stdin io.Reader,
	stdout io.Writer,
	stderr io.Writer,
) ([]sh.Obj, error) {
	fmt.Fprintf(stdout, s.fmt, s.args...)
	return nil, nil
}

func (s *printfFn) SetArgs(args []sh.Obj) error {
	if len(args) == 0 {
		return errors.NewError("printf expects at least 1 argument")
	}

	s.fmt = args[0].String()
	for _, arg := range args[1:] {
		s.args = append(s.args, arg.String())
	}

	return nil
}
