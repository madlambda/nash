package sh

import (
	"io"
	"os"

	"github.com/NeowayLabs/nash/errors"
	"github.com/NeowayLabs/nash/sh"
)

type (
	ChdirFn struct {
		stdin          io.Reader
		stdout, stderr io.Writer

		done chan struct{}
		err  error

		arg string
	}
)

func NewChdir(env *Shell) *ChdirFn {
	return &ChdirFn{
		stdin:  env.stdin,
		stdout: env.stdout,
		stderr: env.stderr,
	}
}

func (chdir *ChdirFn) Name() string {
	return "chdir"
}

func (chdir *ChdirFn) ArgNames() []string {
	return append(make([]string, 0, 1), "dir")
}

func (chdir *ChdirFn) run() error {
	return os.Chdir(chdir.arg)
}

func (chdir *ChdirFn) Start() error {
	chdir.done = make(chan struct{})

	go func() {
		chdir.err = chdir.run()
		chdir.done <- struct{}{}
	}()

	return nil
}

func (chdir *ChdirFn) Wait() error {
	<-chdir.done
	return chdir.err
}

func (chdir *ChdirFn) Results() sh.Obj { return nil }

func (chdir *ChdirFn) SetArgs(args []sh.Obj) error {
	if len(args) != 1 {
		return errors.NewError("chdir expects one argument")
	}

	obj := args[0]

	if obj.Type() != sh.StringType {
		return errors.NewError("chdir expects a string, but a %s was provided", obj.Type())
	}

	if objstr, ok := obj.(*sh.StrObj); ok {
		chdir.arg = objstr.Str()
		return nil
	}

	return errors.NewError("internal error: object of wrong type")
}

func (chdir *ChdirFn) SetEnviron(env []string) {
	// do nothing
}

func (chdir *ChdirFn) SetStdin(r io.Reader)  { chdir.stdin = r }
func (chdir *ChdirFn) SetStderr(w io.Writer) { chdir.stderr = w }
func (chdir *ChdirFn) SetStdout(w io.Writer) { chdir.stdout = w }
func (chdir *ChdirFn) StdoutPipe() (io.ReadCloser, error) {
	return nil, errors.NewError("chdir doesn't works with pipes")
}
func (chdir *ChdirFn) Stdin() io.Reader  { return chdir.stdin }
func (chdir *ChdirFn) Stdout() io.Writer { return chdir.stdout }
func (chdir *ChdirFn) Stderr() io.Writer { return chdir.stderr }

func (chdir *ChdirFn) String() string { return "<builtin fn chdir>" }
