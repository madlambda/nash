package nash

import "fmt"

type (
	nashError struct {
		reason string
		format string
	}

	unfinishedBlockError struct {
		*nashError
	}

	unfinishedListError struct {
		*nashError
	}
)

func newError(format string, arg ...interface{}) *nashError {
	e := &nashError{}
	e.SetReason(format, arg...)
	return e
}

func (e *nashError) SetReason(format string, arg ...interface{}) {
	e.reason = fmt.Sprintf(format, arg...)
}

func (e *nashError) Error() string { return e.reason }

func newUnfinishedBlockError() *unfinishedBlockError {
	return &unfinishedBlockError{
		nashError: newError("Statement's block '{' not finished"),
	}
}

func (e *unfinishedBlockError) Unfinished() bool { return true }

func newUnfinishedListError() *unfinishedListError {
	return &unfinishedListError{
		nashError: newError("List assignment not finished"),
	}
}

func (e *unfinishedListError) Unfinished() bool { return true }
