package sh

import "io"

type (
	Runner interface {
		Start() error
		Wait() error
		Results() Obj

		SetArgs([]Obj) error
		SetEnviron([]string)
		SetStdin(io.Reader)
		SetStdout(io.Writer)
		SetStderr(io.Writer)

		StdoutPipe() (io.ReadCloser, error)

		Stdin() io.Reader
		Stdout() io.Writer
		Stderr() io.Writer
	}

	Fn interface {
		Name() string
		ArgNames() []string

		Runner

		String() string
	}
)
