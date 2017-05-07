package sh

import (
	"fmt"
	"io"
	"os"

	"github.com/NeowayLabs/nash/errors"
	"github.com/NeowayLabs/nash/sh"
)

type (
	FnArg struct {
		Name       string
		IsVariadic bool
	}

	UserFn struct {
		argNames []sh.FnArg // argNames store parameter name
		done     chan error // for async execution
		results  []sh.Obj

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

func (fn *UserFn) ArgNames() []sh.FnArg { return fn.argNames }

func (fn *UserFn) AddArgName(arg sh.FnArg) {
	fn.argNames = append(fn.argNames, arg)
}

func (fn *UserFn) SetArgs(args []sh.Obj) error {
	var (
		isVariadic      bool
		countNormalArgs int
	)

	for i := 0; i < len(fn.argNames); i++ {
		argName := fn.argNames[i]
		if argName.IsVariadic {
			if i != len(fn.argNames)-1 {
				return errors.NewError("variadic expansion must be last argument")
			}
			isVariadic = true
		} else {
			countNormalArgs++
		}
	}

	if !isVariadic && len(args) != len(fn.argNames) {
		return errors.NewError("Wrong number of arguments for function %s. "+
			"Expected %d but found %d",
			fn.name, len(fn.argNames), len(args))
	}

	if isVariadic {
		if len(args) < countNormalArgs {
			return errors.NewError("Wrong number of arguments for function %s. "+
				"Expected at least %d arguments but found %d", fn.name,
				countNormalArgs, len(args))
		}

		if len(args) == 0 {
			// there's only a variadic (optional) argument
			// and user supplied no argument...
			// then only initialize the variadic variable to
			// empty list
			fn.Setvar(fn.argNames[0].Name, sh.NewListObj([]sh.Obj{}))
			return nil
		}
	}

	var i int
	for i = 0; i < len(fn.argNames) && i < len(args); i++ {
		arg := args[i]
		argName := fn.argNames[i].Name
		isVariadic := fn.argNames[i].IsVariadic

		if isVariadic {
			var valist []sh.Obj
			for ; i < len(args); i++ {
				arg = args[i]
				valist = append(valist, arg)
			}
			valistarg := sh.NewListObj(valist)
			fn.Setvar(argName, valistarg)
		} else {
			fn.Setvar(argName, arg)
		}
	}

	// set remaining (variadic) list
	if len(fn.argNames) > 0 && i < len(fn.argNames) {
		last := fn.argNames[len(fn.argNames)-1]
		if !last.IsVariadic {
			return errors.NewError("internal error: optional arguments only for variadic parameter")
		}
		fn.Setvar(last.Name, sh.NewListObj([]sh.Obj{}))
	}

	return nil
}

func (fn *UserFn) closeDescriptors(closers []io.Closer) {
	for _, fd := range closers {
		fd.Close()
	}
}

func (fn *UserFn) execute() ([]sh.Obj, error) {
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

func (fn *UserFn) Results() []sh.Obj { return fn.results }

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
