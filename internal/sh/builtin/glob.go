package builtin

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/madlambda/nash/errors"
	"github.com/madlambda/nash/sh"
)

type (
	globFn struct {
		pattern string
	}
)

func newGlob() *globFn {
	return &globFn{}
}

func (p *globFn) ArgNames() []sh.FnArg {
	return []sh.FnArg{sh.NewFnArg("pattern", false)}
}

func (g *globFn) Run(in io.Reader, out io.Writer, e io.Writer) ([]sh.Obj, error) {
	listobjs := []sh.Obj{}
	matches, err := filepath.Glob(g.pattern)
	if err != nil {
		return []sh.Obj{
			sh.NewListObj([]sh.Obj{}),
			sh.NewStrObj(fmt.Sprintf("glob:error: %q", err)),
		}, nil
	}
	for _, match := range matches {
		listobjs = append(listobjs, sh.NewStrObj(match))
	}
	return []sh.Obj{sh.NewListObj(listobjs), sh.NewStrObj("")}, nil
}

func (g *globFn) SetArgs(args []sh.Obj) error {
	if len(args) != 1 {
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
