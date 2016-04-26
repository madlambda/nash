package nash

import "fmt"

type Error struct {
	format string
}

func NewError(format string) Error {
	return Error{
		format: format,
	}
}
func (e Error) Error() string { return e.format }
func (e Error) Params(args ...interface{}) error {
	e.format = fmt.Sprintf(e.format, args...)
	return e
}
