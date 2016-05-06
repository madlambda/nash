package nash

import "fmt"

type nashError struct {
	reason string
	format string
}

func newError(format string, arg ...interface{}) *nashError {
	e := &nashError{}
	e.SetReason(format, arg...)
	return e
}

func (e *nashError) SetReason(format string, arg ...interface{}) {
	e.reason = fmt.Sprintf(format, arg...)
}

func (e *nashError) Error() string { return e.reason }
