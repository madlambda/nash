package nash

import (
	"fmt"
	"io"
	"os/exec"

	"github.com/NeowayLabs/nash/errors"
)

type (
	// Cmd is a nash command. It has maps of input and output file
	// descriptors that can be set by SetInputfd and SetOutputfd.
	// This can be used to pipe execution of Cmd commands.
	Cmd struct {
		*exec.Cmd

		closeAfterStart []io.Closer
		closeAfterWait  []io.Closer
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
			return nil, newCmdNotFound("Command '%s' not found on PATH: %s",
				name,
				err.Error())
		}
	}

	cmd.Cmd = &exec.Cmd{
		Path: cmdPath,
	}

	return &cmd, nil
}

func (c *Cmd) SetArgs(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("Require at least the argument name")
	}

	if args[0] != c.Path {
		return fmt.Errorf("Require first argument equals command name")
	}

	c.Cmd.Args = args
	return nil
}

func (c *Cmd) SetEnviron(env []string) {
	c.Cmd.Env = env
}

func (c *Cmd) closeDescriptors(closers []io.Closer) {
	for _, fd := range closers {
		fd.Close()
	}
}

func (c *Cmd) AddCloseAfterWait(closer io.Closer) {
	c.closeAfterWait = append(c.closeAfterWait, closer)
}

func (c *Cmd) AddCloseAfterStart(closer io.Closer) {
	c.closeAfterStart = append(c.closeAfterStart, closer)
}

func (c *Cmd) Wait() error {
	defer c.closeDescriptors(c.closeAfterWait)

	err := c.Cmd.Wait()

	if err != nil {
		return err
	}

	return nil
}

func (c *Cmd) Start() error {
	err := c.Cmd.Start()

	if err != nil {
		c.closeDescriptors(c.closeAfterStart)
		c.closeDescriptors(c.closeAfterWait)
		return err
	}

	c.closeDescriptors(c.closeAfterStart)

	return nil
}
