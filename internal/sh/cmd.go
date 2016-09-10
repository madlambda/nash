package sh

import (
	"fmt"
	"io"
	"os/exec"

	"github.com/NeowayLabs/nash/ast"
	"github.com/NeowayLabs/nash/errors"
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

func (c *Cmd) processArgs(cmd string, nodeArgs []ast.Expr, envShell *Shell) ([]string, error) {
	args := make([]string, 1, len(nodeArgs)+1)
	args[0] = cmd

	for i := 0; i < len(nodeArgs); i++ {
		carg := nodeArgs[i]

		obj, err := envShell.evalExpr(carg)

		if err != nil {
			return nil, err
		}

		if obj.Type() == StringType {
			args = append(args, obj.Str())
		} else if obj.Type() == ListType {
			objlist := obj.List()

			for _, l := range objlist {
				if l.Type() != StringType {
					return nil, errors.NewError("Command arguments requires string or list of strings. But received '%v'", l.String())
				}

				args = append(args, l.Str())
			}
		} else if obj.Type() == FnType {
			return nil, errors.NewError("Function cannot be passed as argument to commands.")
		} else {
			return nil, errors.NewError("Invalid command argument '%v'", carg)
		}
	}

	return args, nil
}

func (c *Cmd) SetArgs(nodeArgs []ast.Expr, envShell *Shell) error {
	args, err := c.processArgs(c.Path, nodeArgs, envShell)

	if err != nil {
		return err
	}

	if len(args) < 1 {
		return fmt.Errorf("Require at least the argument name")
	}

	if args[0] != c.Path {
		return fmt.Errorf("Require first argument equals command name")
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

func (c *Cmd) Results() *Obj { return nil }
