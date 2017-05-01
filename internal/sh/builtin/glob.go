package builtin

import (
	"io"

	"github.com/NeowayLabs/nash/sh"
)

type (
	globFn struct {
		fmt  string
		args []interface{}
	}
)

func newGlob() *globFn {
	return &globFn{}
}

func (p *globFn) ArgNames() []string {
	return []string{"fmt", "args..."}
}

func (p *globFn) Run(in io.Reader, out io.Writer, err io.Writer) ([]sh.Obj, error) {
	listobjs := []sh.Obj{}
	return []sh.Obj{sh.NewListObj(listobjs)}, nil
}

func (p *globFn) SetArgs(args []sh.Obj) error {
	//if len(args) == 0 {
	//return errors.NewError("glob expects at least 1 argument")
	//}

	// TODO check for string parameter
	return nil
}
