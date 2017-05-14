package sh

import "io"

type (
	Envs map[string]Obj
	Vars map[string]Obj

	Runner interface {
		Start() error
		Wait() error
		Results() []Obj

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

	FnArg struct {
		Name       string
		IsVariadic bool
	}

	Fn interface {
		Name() string
		ArgNames() []FnArg

		Runner

		String() string
	}
)

func NewFnArg(name string, isVariadic bool) FnArg {
	return FnArg{
		Name:       name,
		IsVariadic: isVariadic,
	}
}
