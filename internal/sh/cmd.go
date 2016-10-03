package sh

import (
	"io"
	"os/exec"

	"github.com/NeowayLabs/nash/ast"
	"github.com/NeowayLabs/nash/errors"
	"github.com/NeowayLabs/nash/sh"
)

type (
	// Cmd is a nash command. It has maps of input and output file
	// descriptors that can be set by SetInputfd and SetOutputfd.
	// This can be used to pipe execution of Cmd commands.
	Cmd struct {
		*exec.Cmd

		argExprs []ast.Expr
	}

	// errCmdNotFound is an error indicating the command wasn't found.
	errCmdNotFound struct {
		*errors.NashError
	}
)

func newCmdNotFound(format string, arg ...interface{}) error {
	e := &errCmdNotFound{
		NashError: errors.NewError(format, arg...),
	}

	return e
}

func (e *errCmdNotFound) NotFound() bool {
	return true
}

func NewCmd(name string) (*Cmd, error) {
	var (
		err     error
		cmdPath = name
	)

	cmd := Cmd{}

	if name[0] != '/' {
		cmdPath, err = exec.LookPath(name)

		if err != nil {
			return nil, newCmdNotFound(err.Error())
		}
	}

	cmd.Cmd = &exec.Cmd{
		Path: cmdPath,
	}

	return &cmd, nil
}

func (c *Cmd) Stdin() io.Reader  { return c.Cmd.Stdin }
func (c *Cmd) Stdout() io.Writer { return c.Cmd.Stdout }
func (c *Cmd) Stderr() io.Writer { return c.Cmd.Stderr }

func (c *Cmd) SetStdin(in io.Reader)   { c.Cmd.Stdin = in }
func (c *Cmd) SetStdout(out io.Writer) { c.Cmd.Stdout = out }
func (c *Cmd) SetStderr(err io.Writer) { c.Cmd.Stderr = err }

func (c *Cmd) SetArgs(nodeArgs []sh.Obj) error {
	args := make([]string, 1, len(nodeArgs)+1)
	args[0] = c.Path

	for _, obj := range nodeArgs {
		if obj.Type() == sh.StringType {
			objstr := obj.(*sh.StrObj)
			args = append(args, objstr.Str())
		} else if obj.Type() == sh.ListType {
			objlist := obj.(*sh.ListObj)
			values := objlist.List()

			for _, l := range values {
				if l.Type() != sh.StringType {
					return errors.NewError("Command arguments requires string or list of strings. But received '%v'", l.String())
				}

				lstr := l.(*sh.StrObj)
				args = append(args, lstr.Str())
			}
		} else if obj.Type() == sh.FnType {
			return errors.NewError("Function cannot be passed as argument to commands.")
		} else {
			return errors.NewError("Invalid command argument '%v'", obj)
		}
	}

	c.Cmd.Args = args
	return nil
}

func (c *Cmd) Args() []ast.Expr { return c.argExprs }

func (c *Cmd) SetEnviron(env []string) {
	c.Cmd.Env = env
}

func (c *Cmd) Wait() error {
	err := c.Cmd.Wait()

	if err != nil {
		return err
	}

	return nil
}

func (c *Cmd) Start() error {
	err := c.Cmd.Start()

	if err != nil {
		return err
	}

	return nil
}

func (c *Cmd) Results() sh.Obj { return nil }

func cmdArgs(nodeArgs []ast.Expr, envShell *Shell) ([]sh.Obj, error) {
	args := make([]sh.Obj, 0, len(nodeArgs))

	for i := 0; i < len(nodeArgs); i++ {
		carg := nodeArgs[i]

		obj, err := envShell.evalExpr(carg)

		if err != nil {
			return nil, err
		}

		args = append(args, obj)

	}

	return args, nil
}
