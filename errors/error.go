package errors

import (
	"fmt"

	"github.com/NeowayLabs/nash/ast"
	"github.com/NeowayLabs/nash/scanner"
)

type (
	NashError struct {
		reason string
		format string
	}

	unfinished struct{}

	unfinishedBlockError struct {
		*NashError
		unfinished
	}

	unfinishedListError struct {
		*NashError
		unfinished
	}

	unfinishedCmdError struct {
		*NashError
		unfinished
	}
)

func NewError(format string, arg ...interface{}) *NashError {
	e := &NashError{}
	e.SetReason(format, arg...)
	return e
}

func NewEvalError(path string, node ast.Node, format string, arg ...interface{}) *NashError {
	linenum := fmt.Sprintf("%s:%d:%d: ", path, node.Line(), node.Column())
	return NewError(linenum+format, arg...)
}

func (e *NashError) SetReason(format string, arg ...interface{}) {
	e.reason = fmt.Sprintf(format, arg...)
}

func (e *NashError) Error() string { return e.reason }

func (e unfinished) Unfinished() bool { return true }

func NewUnfinishedBlockError(name string, it scanner.Token) error {
	return &unfinishedBlockError{
		NashError: NewError("%s:%d:%d: Statement's block '{' not finished",
			name, it.Line(), it.Column()),
	}
}

func NewUnfinishedListError(name string, it scanner.Token) error {
	return &unfinishedListError{
		NashError: NewError("%s:%d:%d: List assignment not finished. Found %v",
			name, it.Line(), it.Column(), it),
	}
}

func NewUnfinishedCmdError(name string, it scanner.Token) error {
	return &unfinishedCmdError{
		NashError: NewError("%s:%d:%d: Multi-line command not finished. Found %v but expect ')'",
			name, it.Line(), it.Column(), it),
	}
}
