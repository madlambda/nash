// +build !darwin,!dragonfly,!freebsd,!linux,!nacl,!netbsd,!openbsd,!solaris,!windows

package builtin

import (
	"fmt"
	"runtime"

	"github.com/NeowayLabs/nash/sh"
)

type (
	exitNotSupportedFn struct {
		err error
	}
)

func newExit() *exitNotSupportedFn {
	return &exitNotSupportedFn{
		err: fmt.Errorf("exit is not implemented on OS: %s", runtime.GOOS),
	}
}

func (e *exitNotSupportedFn) ArgNames() []string {
	panic(e.err)
	return []string{}
}

func (e *exitNotSupportedFn) Run() (sh.Obj, error) {
	panic(e.err)
	return nil, nil //Unrecheable code
}

func (e *exitNotSupportedFn) SetArgs(args []sh.Obj) error {
	panic(e.err)
	return nil
}
