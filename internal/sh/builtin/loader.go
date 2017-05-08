package builtin

import (
	"io"

	"github.com/NeowayLabs/nash/sh"
)

// Fn is the contract of a built in function, that is simpler
// than the core nash Fn.
type Fn interface {
	ArgNames() []string
	SetArgs(args []sh.Obj) error
	Run(
		stdin io.Reader,
		stdout io.Writer,
		stderr io.Writer,
	) ([]sh.Obj, error)
}

//Load loads all available builtin functions. The return is a map
//of the builtin function name and its implementation.
func Load() map[string]Fn {
	return map[string]Fn{
		"glob":   newGlob(),
		"print":  newPrint(),
		"format": newFormat(),
		"split":  newSplit(),
		"len":    newLen(),
		"chdir":  newChdir(),
		"append": newAppend(),
		"exit":   newExit(),
	}
}
