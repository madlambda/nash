package builtin

import (
	"io"

	"github.com/NeowayLabs/nash/sh"
)

// Fn is the contract of a built in function, that is simpler
// than the core nash Fn.
type (
	Fn interface {
		ArgNames() []sh.FnArg
		SetArgs(args []sh.Obj) error
		Run(
			stdin io.Reader,
			stdout io.Writer,
			stderr io.Writer,
		) ([]sh.Obj, error)
	}

	Constructor func() Fn
)

// Constructors returns a map of the builtin function name and its constructor
func Constructors() map[string]Constructor {
	return map[string]Constructor{
		"glob":   func() Fn { return newGlob() },
		"print":  func() Fn { return newPrint() },
		"format": func() Fn { return newFormat() },
		"split":  func() Fn { return newSplit() },
		"len":    func() Fn { return newLen() },
		"chdir":  func() Fn { return newChdir() },
		"append": func() Fn { return newAppend() },
		"exit":   func() Fn { return newExit() },
	}
}
