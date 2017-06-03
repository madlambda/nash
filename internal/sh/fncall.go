package sh

import (
	"fmt"
	"io"
	"os"

	"github.com/NeowayLabs/nash/ast"
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

		name     string // debugging purposes
		parent   *Shell
		subshell *Shell

		environ []string

		stdin          io.Reader
		stdout, stderr io.Writer

		tree           *ast.Tree
		repr           string
		closeAfterWait []io.Closer
	}
)

func NewUserFn(name string, parent *Shell) *UserFn {
	return &UserFn{
		name:   name,
		done:   make(chan error),
		parent: parent,
		stdin:  parent.Stdin(),
		stdout: parent.Stdout(),
		stderr: parent.Stderr(),
	}
}

func (fn *UserFn) setup() error {
	subshell, err := NewSubShell(fn.name, fn.parent)
	if err != nil {
		return err
	}

	subshell.SetTree(fn.tree)
	subshell.SetRepr(fn.repr)
	subshell.SetDebug(fn.parent.debug)
	subshell.SetStdout(fn.stdout)
	subshell.SetStderr(fn.stderr)
	subshell.SetStdin(fn.stdin)
	subshell.SetEnviron(fn.environ)

	fn.subshell = subshell
	return nil
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

	if err := fn.setup(); err != nil {
		return err
	}

	for i, argName := range fn.argNames {
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
			fn.subshell.Setvar(fn.argNames[0].Name, sh.NewListObj([]sh.Obj{}))
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
			fn.subshell.Setvar(argName, valistarg)
		} else {
			fn.subshell.Setvar(argName, arg)
		}
	}

	// set remaining (variadic) list
	if len(fn.argNames) > 0 && i < len(fn.argNames) {
		last := fn.argNames[len(fn.argNames)-1]
		if !last.IsVariadic {
			return errors.NewError("internal error: optional arguments only for variadic parameter")
		}
		fn.subshell.Setvar(last.Name, sh.NewListObj([]sh.Obj{}))
	}

	return nil
}

func (fn *UserFn) Name() string { return fn.name }

func (fn *UserFn) SetRepr(repr string) {
	fn.repr = repr
}

func (fn *UserFn) SetTree(t *ast.Tree) {
	fn.tree = t
}

func (fn *UserFn) closeDescriptors(closers []io.Closer) {
	for _, fd := range closers {
		fd.Close()
	}
}

func (fn *UserFn) execute() ([]sh.Obj, error) {
	if fn.tree != nil {
		return fn.subshell.ExecuteTree(fn.tree)
	}

	return nil, fmt.Errorf("fn not properly created")
}

func (fn *UserFn) Start() error {
	if fn.subshell == nil {
		err := fn.setup()
		if err != nil {
			return err
		}
	}

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
	fn.subshell = nil
	return err
}

func (fn *UserFn) SetEnviron(env []string) {
	fn.environ = env
}

func (fn *UserFn) SetStderr(w io.Writer) {
	fn.stderr = w
}

func (fn *UserFn) SetStdout(w io.Writer) {
	fn.stdout = w
}

func (fn *UserFn) SetStdin(r io.Reader) {
	fn.stdin = r
}

func (fn *UserFn) Stdin() io.Reader  { return fn.stdin }
func (fn *UserFn) Stdout() io.Writer { return fn.stdout }
func (fn *UserFn) Stderr() io.Writer { return fn.stderr }

func (fn *UserFn) String() string {
	if fn.tree != nil {
		return fn.tree.String()
	}
	panic("fn not initialized")
}

func (fn *UserFn) StdoutPipe() (io.ReadCloser, error) {
	pr, pw, err := os.Pipe()

	if err != nil {
		return nil, err
	}

	fn.subshell.SetStdout(pw)

	// As fn doesn't fork, both fd can be closed after wait is called
	fn.closeAfterWait = append(fn.closeAfterWait, pw, pr)
	return pr, nil
}
