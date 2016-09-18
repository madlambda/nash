package sh

import (
	"io"

	"github.com/NeowayLabs/nash/ast"
	"github.com/NeowayLabs/nash/errors"
)

type (
	AppendFn struct {
		stdin          io.Reader
		stdout, stderr io.Writer

		done    chan struct{}
		err     error
		results *Obj

		obj []*Obj
		arg *Obj
	}
)

func NewAppendFn(env *Shell) *AppendFn {
	return &AppendFn{
		stdin:  env.stdin,
		stdout: env.stdout,
		stderr: env.stderr,
	}
}

func (appendfn *AppendFn) Name() string {
	return "len"
}

func (appendfn *AppendFn) ArgNames() []string {
	return append(make([]string, 0, 1), "list")
}

func (appendfn *AppendFn) run() error {
	newobj := append(appendfn.obj, appendfn.arg)
	appendfn.results = NewListObj(newobj)
	return nil
}

func (appendfn *AppendFn) Start() error {
	appendfn.done = make(chan struct{})

	go func() {
		appendfn.err = appendfn.run()
		appendfn.done <- struct{}{}
	}()

	return nil
}

func (appendfn *AppendFn) Wait() error {
	<-appendfn.done
	return appendfn.err
}

func (appendfn *AppendFn) Results() *Obj {
	return appendfn.results
}

func (appendfn *AppendFn) SetArgs(args []ast.Expr, envShell *Shell) error {
	if len(args) != 2 {
		return errors.NewError("appendfn expects two arguments")
	}

	obj, err := envShell.evalExpr(args[0])

	if err != nil {
		return err
	}

	if obj.Type() != ListType {
		return errors.NewError("appendfn expects a list as first argument, but a %s was provided", obj.Type())
	}

	arg, err := envShell.evalExpr(args[1])

	if err != nil {
		return err
	}

	appendfn.obj = obj.List()
	appendfn.arg = arg
	return nil
}

func (appendfn *AppendFn) SetEnviron(env []string) {
	// do nothing
}

func (appendfn *AppendFn) SetStdin(r io.Reader)  { appendfn.stdin = r }
func (appendfn *AppendFn) SetStderr(w io.Writer) { appendfn.stderr = w }
func (appendfn *AppendFn) SetStdout(w io.Writer) { appendfn.stdout = w }
func (appendfn *AppendFn) StdoutPipe() (io.ReadCloser, error) {
	return nil, errors.NewError("appendfn doesn't works with pipes")
}
func (appendfn *AppendFn) Stdin() io.Reader  { return appendfn.stdin }
func (appendfn *AppendFn) Stdout() io.Writer { return appendfn.stdout }
func (appendfn *AppendFn) Stderr() io.Writer { return appendfn.stderr }

func (appendfn *AppendFn) String() string { return "<builtin fn append>" }
