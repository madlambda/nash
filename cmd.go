package nash

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"sort"

	"github.com/NeowayLabs/nash/errors"
)

type (
	// Cmd is a nash command. It has maps of input and output file
	// descriptors that can be set by SetInputfd and SetOutputfd.
	// This can be used to pipe execution of Cmd commands.
	Cmd struct {
		*exec.Cmd

		fdIn  map[uint]io.Reader
		fdOut map[uint]io.Writer

		goroutines      []func() error // goroutines copying data pipes
		errch           chan error     // one send per goroutine
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

	cmd := Cmd{
		fdIn:  make(map[uint]io.Reader),
		fdOut: make(map[uint]io.Writer),
	}

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

// SetInputfd sets an input file descriptor into fork'ed process.
// It receives an io.Reader, but if the underlying type is
// an *os.File, then its fd is used, otherwise a pipe will be created
// with an additional goroutine to copy data from the reader.
func (c *Cmd) SetInputfd(n uint, in io.Reader) error {
	if n == 1 || n == 2 {
		return fmt.Errorf("File descriptors 1 and 2 must be writable")
	}

	c.fdIn[n] = in
	return nil
}

func (c *Cmd) GetInputfd(n uint) (io.Reader, bool) {
	r, ok := c.fdIn[n]

	return r, ok
}

// SetOutputfd sets an output file descriptor into fork'ed process.
// It receives an io.Writer, but if the underlying type is
// an *os.File, then its fd is used, otherwise a pipe will be created
// with an additional goroutine to copy data to writer.
func (c *Cmd) SetOutputfd(n uint, out io.Writer) error {
	if n == 0 {
		return fmt.Errorf("File descriptor 0 must be read-only")
	}

	c.fdOut[n] = out
	return nil
}

func (c *Cmd) GetOutputfd(n uint) (io.Writer, bool) {
	r, ok := c.fdOut[n]

	return r, ok
}

func (c *Cmd) addExtraFile(value *os.File) {
	if c.Cmd.ExtraFiles == nil {
		c.Cmd.ExtraFiles = make([]*os.File, 0, 8)
	}

	c.Cmd.ExtraFiles = append(c.Cmd.ExtraFiles, value)
}

func (c *Cmd) applyInputfd() error {
	for fd, in := range c.fdIn {
		if fd == 0 {
			// Golang os/exec already handles the case of non-file reader
			return nil
		}

		file, ok := in.(*os.File)

		if ok {
			c.fdIn[fd] = file
			return nil
		}

		// for non-standard file descriptors we need to get proper pipes
		pr, pw, err := os.Pipe()

		if err != nil {
			return err
		}

		c.closeAfterStart = append(c.closeAfterStart, pr)
		c.closeAfterWait = append(c.closeAfterWait, pw)
		c.goroutines = append(c.goroutines, func() error {
			_, err := io.Copy(pw, in)

			if err1 := pw.Close(); err == nil {
				err = err1
			}

			return err
		})

		c.fdIn[fd] = pr
	}

	return nil
}

func (c *Cmd) applyOutputfd() error {
	for fd, out := range c.fdOut {
		if fd == 1 || fd == 2 {
			return nil
		}

		file, ok := out.(*os.File)

		if ok {
			c.fdOut[fd] = file
			return nil
		}

		// for non-standard file descriptors we need to get proper pipes
		pr, pw, err := os.Pipe()

		if err != nil {
			return err
		}

		c.closeAfterStart = append(c.closeAfterStart, pw)
		c.closeAfterWait = append(c.closeAfterWait, pr)
		c.goroutines = append(c.goroutines, func() error {
			_, err := io.Copy(out, pr)
			pr.Close() // in case io.Copy stopped due to write error
			return err
		})

		c.fdOut[fd] = pw
	}

	return nil

}

func (c *Cmd) applyfd() error {
	err := c.applyInputfd()

	if err != nil {
		return err
	}

	err = c.applyOutputfd()

	if err != nil {
		return err
	}

	fds := make([]int, 0, len(c.fdIn)+len(c.fdOut))

	for fd, _ := range c.fdIn {
		fds = append(fds, int(fd))
	}

	for fd, _ := range c.fdOut {
		fds = append(fds, int(fd))
	}

	sort.Ints(fds)

	for _, fd := range fds {
		switch fd {
		case 0:
			c.Cmd.Stdin = c.fdIn[0]
		case 1:
			c.Cmd.Stdout = c.fdOut[1]
		case 2:
			c.Cmd.Stderr = c.fdOut[2]
		default:
			var value interface{}

			value, ok := c.fdIn[uint(fd)]

			if !ok {
				value, ok = c.fdOut[uint(fd)]

				if !ok {
					return fmt.Errorf("internal error: race cond applying fd")
				}
			}

			file, ok := value.(*os.File)

			if !ok {
				return fmt.Errorf("File descriptors > 2 must be open files.")
			}

			c.addExtraFile(file)
		}
	}

	return nil
}

func (c *Cmd) closeDescriptors(closers []io.Closer) {
	for _, fd := range closers {
		fd.Close()
	}
}

func (c *Cmd) Wait() error {
	defer c.closeDescriptors(c.closeAfterWait)

	err := c.Cmd.Wait()

	var copyError error
	for range c.goroutines {
		if err := <-c.errch; err != nil && copyError == nil {
			copyError = err
		}
	}

	if err != nil {
		return err
	} else if copyError != nil {
		return copyError
	}

	return nil
}

func (c *Cmd) Start() error {
	if err := c.applyfd(); err != nil {
		c.closeDescriptors(c.closeAfterStart)
		c.closeDescriptors(c.closeAfterWait)

		return err
	}

	err := c.Cmd.Start()

	if err != nil {
		c.closeDescriptors(c.closeAfterStart)
		c.closeDescriptors(c.closeAfterWait)
		return err
	}

	c.closeDescriptors(c.closeAfterStart)

	c.errch = make(chan error, len(c.goroutines))

	for _, fn := range c.goroutines {
		go func(fn func() error) {
			c.errch <- fn()
		}(fn)
	}

	return nil
}
