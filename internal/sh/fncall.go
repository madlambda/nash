package sh

import (
	"fmt"
	"io"
	"os"

	"github.com/NeowayLabs/nash/errors"
	"github.com/NeowayLabs/nash/sh"
)

type (
	UserFn struct {
		argNames []string   // argNames store parameter name
		done     chan error // for async execution
		results  sh.Obj

		closeAfterWait []io.Closer

		*Shell // sub-shell
	}
)

func NewUserFn(name string, parent *Shell) (*UserFn, error) {
	fn := UserFn{
		done: make(chan error),
	}

	subshell, err := NewSubShell(name, parent)

	if err != nil {
		return nil, err
	}

	fn.Shell = subshell
	fn.SetDebug(parent.debug)
	fn.SetStdout(parent.stdout)
	fn.SetStderr(parent.stderr)
	fn.SetStdin(parent.stdin)

	return &fn, nil
}

func (fn *UserFn) ArgNames() []string { return fn.argNames }

func (fn *UserFn) AddArgName(name string) {
	fn.argNames = append(fn.argNames, name)
}

func (fn *UserFn) SetArgs(args []sh.Obj) error {
	if len(fn.argNames) != len(args) {
		return errors.NewError("Wrong number of arguments for function %s. Expected %d but found %d",
			fn.name, len(fn.argNames), len(args))
	}

	for i := 0; i < len(args); i++ {
		arg := args[i]
		argName := fn.argNames[i]
		fn.Setvar(argName, arg)
	}

	return nil
}

func (fn *UserFn) closeDescriptors(closers []io.Closer) {
	for _, fd := range closers {
		fd.Close()
	}
}

func (fn *UserFn) execute() (sh.Obj, error) {
	if fn.root != nil {
		return fn.ExecuteTree(fn.root)
	}

	return nil, fmt.Errorf("fn not properly created")
}

func (fn *UserFn) Start() error {
	// TODO: what we'll do with fn return values in case of pipes?

	go func() {
		var err error
		fn.results, err = fn.execute()
		fn.done <- err
	}()

	return nil
}

func (fn *UserFn) Results() sh.Obj { return fn.results }

func (fn *UserFn) Wait() error {
	err := <-fn.done

	fn.closeDescriptors(fn.closeAfterWait)
	return err
}

func (fn *UserFn) StdoutPipe() (io.ReadCloser, error) {
	pr, pw, err := os.Pipe()

	if err != nil {
		return nil, err
	}

	fn.SetStdout(pw)

	// As fn doesn't fork, both fd can be closed after wait is called
	fn.closeAfterWait = append(fn.closeAfterWait, pw, pr)
	return pr, nil
}
