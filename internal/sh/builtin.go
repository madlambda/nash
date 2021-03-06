package sh

import (
	"fmt"
	"io"

	"github.com/madlambda/nash/errors"
	"github.com/madlambda/nash/internal/sh/builtin"
	"github.com/madlambda/nash/sh"
)

type (
	// builtinFn maps a builtin function to a nash sh.FnDef
	// avoiding a lot of duplicated code and decoupling the
	// builtin functions of some unnecessary details on how
	// the sh.Fn works (lots of complexity to provide features of
	// other kinds of runners/functions).
	builtinFn struct {
		stdin          io.Reader
		stdout, stderr io.Writer

		done    chan struct{}
		err     error
		results []sh.Obj

		name string
		fn   builtin.Fn
	}
)

func NewBuiltinFn(
	name string,
	fn builtin.Fn,
	in io.Reader,
	out io.Writer,
	outerr io.Writer,
) *builtinFn {
	return &builtinFn{
		name:   name,
		fn:     fn,
		stdin:  in,
		stdout: out,
		stderr: outerr,
	}
}

func (f *builtinFn) Name() string {
	return f.name
}

func (f *builtinFn) ArgNames() []sh.FnArg {
	return f.fn.ArgNames()
}

func (f *builtinFn) Start() error {
	f.done = make(chan struct{})

	go func() {
		f.results, f.err = f.fn.Run(f.stdin, f.stdout, f.stderr)
		f.done <- struct{}{}
	}()

	return nil
}

func (f *builtinFn) Wait() error {
	<-f.done
	return f.err
}

func (f *builtinFn) Results() []sh.Obj {
	return f.results
}

func (f *builtinFn) String() string {
	return fmt.Sprintf("<builtin function %q>", f.Name())
}

func (f *builtinFn) SetArgs(args []sh.Obj) error {
	return f.fn.SetArgs(args)
}

func (f *builtinFn) SetEnviron(env []string) {
	// do nothing
	// terrible design smell having functions that do nothing =/
}

func (f *builtinFn) SetStdin(r io.Reader) {
	f.stdin = r
}
func (f *builtinFn) SetStderr(w io.Writer) {
	f.stderr = w
}
func (f *builtinFn) SetStdout(w io.Writer) {
	f.stdout = w
}
func (f *builtinFn) StdoutPipe() (io.ReadCloser, error) {
	return nil, errors.NewError("builtin functions doesn't works with pipes")
}
func (f *builtinFn) Stdin() io.Reader  { return f.stdin }
func (f *builtinFn) Stdout() io.Writer { return f.stdout }
func (f *builtinFn) Stderr() io.Writer { return f.stderr }
