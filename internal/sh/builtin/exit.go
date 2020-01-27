package builtin

import (
	"io"
	"os"
	"strconv"

	"github.com/madlambda/nash/errors"
	"github.com/madlambda/nash/sh"
)

type (
	exitFn struct {
		status int
	}
)

func newExit() Fn {
	return &exitFn{}
}

func (e *exitFn) ArgNames() []sh.FnArg {
	return []sh.FnArg{
		sh.NewFnArg("status", false),
	}
}

func (e *exitFn) Run(in io.Reader, out io.Writer, err io.Writer) ([]sh.Obj, error) {
	os.Exit(e.status)
	return nil, nil //Unrecheable code
}

func (e *exitFn) SetArgs(args []sh.Obj) error {
	if len(args) != 1 {
		return errors.NewError("exit expects 1 argument")
	}

	obj := args[0]
	if obj.Type() != sh.StringType {
		return errors.NewError(
			"exit expects a status string, but a %s was provided",
			obj.Type(),
		)
	}
	statusstr := obj.(*sh.StrObj).Str()
	status, err := strconv.Atoi(statusstr)
	if err != nil {
		return errors.NewError(
			"exit:error[%s] converting status[%s] to int",
			err,
			statusstr,
		)

	}
	e.status = status
	return nil
}
