package errors

import "fmt"

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

func NewUnfinishedBlockError() error {
	return &unfinishedBlockError{
		NashError: NewError("Statement's block '{' not finished"),
	}
}

func (e *unfinishedBlockError) Unfinished() bool { return true }

func NewUnfinishedListError() error {
	return &unfinishedListError{
		NashError: NewError("List assignment not finished"),
	}
}

func (e *unfinishedListError) Unfinished() bool { return true }
