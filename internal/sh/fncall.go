package sh

import (
	"fmt"
	"io"
	"os"

	"github.com/NeowayLabs/nash/ast"
	"github.com/NeowayLabs/nash/errors"
)

type (
	Fn struct {
		argNames []string   // argNames store parameter name
		done     chan error // for async execution

		closeAfterWait []io.Closer

		*Shell // sub-shell
	}
)

func NewFn(name string, parent *Shell) (*Fn, error) {
	fn := Fn{
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

func (fn *Fn) AddArgName(name string) {
	fn.argNames = append(fn.argNames, name)
}

func (fn *Fn) SetArgs(nodeArgs []*ast.Arg, envShell *Shell) error {
	if len(fn.argNames) != len(nodeArgs) {
		return errors.NewError("Wrong number of arguments for function %s. Expected %d but found %d",
			fn.name, len(fn.argNames), len(nodeArgs))
	}

	for i := 0; i < len(nodeArgs); i++ {
		arg := nodeArgs[i]
		argName := fn.argNames[i]

		obj, err := envShell.evalArg(arg)

		if err != nil {
			return err
		}

		fn.Setvar(argName, obj)
	}

	return nil
}

func (fn *Fn) closeDescriptors(closers []io.Closer) {
	for _, fd := range closers {
		fd.Close()
	}
}

func (fn *Fn) Execute() (*Obj, error) {
	if fn.root != nil {
		return fn.ExecuteTree(fn.root)
	}

	return nil, fmt.Errorf("fn not properly created")
}

func (fn *Fn) Start() error {
	// TODO: what we'll do with fn return values in case of pipes?

	go func() {
		_, err := fn.Execute()
		fn.done <- err
	}()

	return nil
}

func (fn *Fn) Wait() error {
	err := <-fn.done

	fn.closeDescriptors(fn.closeAfterWait)
	return err
}

func (fn *Fn) StdoutPipe() (io.ReadCloser, error) {
	pr, pw, err := os.Pipe()

	if err != nil {
		return nil, err
	}

	fn.SetStdout(pw)

	// As fn doesn't fork, both fd can be closed after wait is called
	fn.closeAfterWait = append(fn.closeAfterWait, pw, pr)
	return pr, nil
}
