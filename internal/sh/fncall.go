package sh

import (
	"github.com/NeowayLabs/nash/sh"
)

type (
	FnArg struct {
		Name       string
		IsVariadic bool
	}

	UserFn struct {
		*Interpreter // sub-shell
	}
)

func NewUserFn(args []sh.FnArg, parent *Interpreter) *UserFn {
	return &UserFn{}
}
