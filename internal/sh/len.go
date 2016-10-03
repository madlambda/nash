package sh

import (
	"io"
	"strconv"

	"github.com/NeowayLabs/nash/errors"
	"github.com/NeowayLabs/nash/sh"
)

type (
	LenFn struct {
		stdin          io.Reader
		stdout, stderr io.Writer

		done    chan struct{}
		err     error
		results int

		arg sh.Obj
	}
)

func NewLenFn(env *Shell) *LenFn {
	return &LenFn{
		stdin:  env.stdin,
		stdout: env.stdout,
		stderr: env.stderr,
	}
}

func (lenfn *LenFn) Name() string {
	return "len"
}

func (lenfn *LenFn) ArgNames() []string {
	return append(make([]string, 0, 1), "list")
}

func (lenfn *LenFn) run() error {
	if lenfn.arg.Type() == sh.ListType {
		arglist := lenfn.arg.(*sh.ListObj)
		lenfn.results = len(arglist.List())
	} else {
		argstr := lenfn.arg.(*sh.StrObj)
		lenfn.results = len(argstr.Str())
	}

	return nil
}

func (lenfn *LenFn) Start() error {
	lenfn.done = make(chan struct{})

	go func() {
		lenfn.err = lenfn.run()
		lenfn.done <- struct{}{}
	}()

	return nil
}

func (lenfn *LenFn) Wait() error {
	<-lenfn.done
	return lenfn.err
}

func (lenfn *LenFn) Results() sh.Obj {
	return sh.NewStrObj(strconv.Itoa(lenfn.results))
}

func (lenfn *LenFn) SetArgs(args []sh.Obj) error {
	if len(args) != 1 {
		return errors.NewError("lenfn expects one argument")
	}

	obj := args[0]

	if obj.Type() != sh.ListType && obj.Type() != sh.StringType {
		return errors.NewError("lenfn expects a list or a string, but a %s was provided", obj.Type())
	}

	lenfn.arg = obj
	return nil
}

func (lenfn *LenFn) SetEnviron(env []string) {
	// do nothing
}

func (lenfn *LenFn) SetStdin(r io.Reader)  { lenfn.stdin = r }
func (lenfn *LenFn) SetStderr(w io.Writer) { lenfn.stderr = w }
func (lenfn *LenFn) SetStdout(w io.Writer) { lenfn.stdout = w }
func (lenfn *LenFn) StdoutPipe() (io.ReadCloser, error) {
	return nil, errors.NewError("lenfn doesn't works with pipes")
}
func (lenfn *LenFn) Stdin() io.Reader  { return lenfn.stdin }
func (lenfn *LenFn) Stdout() io.Writer { return lenfn.stdout }
func (lenfn *LenFn) Stderr() io.Writer { return lenfn.stderr }

func (lenfn *LenFn) String() string { return "<builtin fn len>" }
