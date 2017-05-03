package builtin

import (
	"io"
	"path/filepath"

	"github.com/NeowayLabs/nash/errors"
	"github.com/NeowayLabs/nash/sh"
)

type (
	globFn struct {
		pattern string
	}
)

func newGlob() *globFn {
	return &globFn{}
}

func (p *globFn) ArgNames() []string {
	return []string{"fmt", "args..."}
}

func (g *globFn) Run(in io.Reader, out io.Writer, e io.Writer) ([]sh.Obj, error) {
	listobjs := []sh.Obj{}
	matches, err := filepath.Glob(g.pattern)
	if err != nil {
		return nil, errors.NewError("glob:error: %q", err)
	}
	for _, match := range matches {
		listobjs = append(listobjs, sh.NewStrObj(match))
	}
	return []sh.Obj{sh.NewListObj(listobjs)}, nil
}

func (g *globFn) SetArgs(args []sh.Obj) error {
	if len(args) == 0 {
		return errors.NewError("glob expects 1 string argument (the pattern)")
	}

	obj := args[0]
	if obj.Type() != sh.StringType {
		return errors.NewError(
			"glob expects a pattern string, but a %s was provided",
			obj.Type(),
		)
	}
	g.pattern = obj.(*sh.StrObj).Str()
	return nil
}
