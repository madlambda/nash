package errors

import (
	"fmt"

	"github.com/NeowayLabs/nash/scanner"
)

type (
	NashError struct {
		reason string
		format string
	}

	unfinishedBlockError struct {
		*NashError
	}

	unfinishedListError struct {
		*NashError
	}
)

func NewError(format string, arg ...interface{}) *NashError {
	e := &NashError{}
	e.SetReason(format, arg...)
	return e
}

func (e *NashError) SetReason(format string, arg ...interface{}) {
	e.reason = fmt.Sprintf(format, arg...)
}

func (e *NashError) Error() string { return e.reason }

func NewUnfinishedBlockError(name string, it scanner.Token) error {
	return &unfinishedBlockError{
		NashError: NewError("%s:%d:%d: Statement's block '{' not finished", name, it.Line(), it.Column()),
	}
}

func (e *unfinishedBlockError) Unfinished() bool { return true }

func NewUnfinishedListError(name string, it scanner.Token) error {
	return &unfinishedListError{
		NashError: NewError("%s:%d:%d: List assignment not finished", name, it.Line(), it.Column()),
	}
}

func (e *unfinishedListError) Unfinished() bool { return true }
