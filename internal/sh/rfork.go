// +build !linux,!plan9

//
package sh

import (
	"github.com/madlambda/nash/ast"
	"github.com/madlambda/nash/errors"
)

func (sh *Shell) executeRfork(rfork *ast.RforkNode) error {
	return errors.NewError("rfork only supported on Linux and Plan9")
}
