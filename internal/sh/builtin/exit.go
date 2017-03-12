package builtin

import (
	"fmt"
	"os"
	"strconv"

	"github.com/NeowayLabs/nash/errors"
	"github.com/NeowayLabs/nash/sh"
)

type (
	exitFn struct {
		status int
	}
)

func newExit() *exitFn {
	return &exitFn{}
}

func (e *exitFn) ArgNames() []string {
	return []string{"status"}
}

func (e *exitFn) Run() (sh.Obj, error) {
	os.Exit(e.status)
	return nil, nil //Unrecheable code
}

func (e *exitFn) SetArgs(args []sh.Obj) error {
	if len(args) != 1 {
		return errors.NewError("exit expects one argument")
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
		return fmt.Errorf(
			"exit:error[%s] converting status[%s] to int",
			err,
			statusstr,
		)

	}
	e.status = status
	return nil
}
