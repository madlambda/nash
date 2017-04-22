package builtin

import (
	"fmt"
	"io"

	"github.com/NeowayLabs/nash/errors"
	"github.com/NeowayLabs/nash/sh"
)

type (
	sprintFn struct {
		fmt  string
		args []interface{}
	}
)

func newSprint() *sprintFn {
	return &sprintFn{}
}

func (s *sprintFn) ArgNames() []string {
	return []string{"fmt", "args..."}
}

func (s *sprintFn) Run(in io.Reader, out io.Writer, err io.Writer) ([]sh.Obj, error) {
	return []sh.Obj{sh.NewStrObj(fmt.Sprintf(s.fmt, s.args...))}, nil
}

func (s *sprintFn) SetArgs(args []sh.Obj) error {
	if len(args) == 0 {
		return errors.NewError("sprint expects at least 1 argument")
	}

	s.fmt = args[0].String()
	for _, arg := range args[1:] {
		s.args = append(s.args, arg.String())
	}

	return nil
}
