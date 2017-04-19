package builtin

import (
	"fmt"
	"io"

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
	//if len(args) != 1 {
	//return errors.NewError("splitfn expects 2 arguments")
	//}

	//if args[0].Type() != sh.StringType {
	//return errors.NewError("content must be of type string")
	//}

	s.fmt = args[0].String()
	for _, arg := range args[1:] {
		s.args = append(s.args, arg.String())
	}

	return nil
}
