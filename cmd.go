package nash

import (
	"io"
	"os"
	"os/exec"
	"strings"
)

type (
	FDMap map[int]interface{}

	Command struct {
		name string
		*exec.Cmd
		sh    *Shell
		fdMap FDMap

		stdinDone, stdoutDone, stderrDone chan bool
	}

	errCmdNotFound struct {
		*nashError
	}
)

func newCmdNotFound(format string, arg ...interface{}) error {
	e := &errCmdNotFound{
		nashError: newError(format, arg...),
	}

	return e
}

func (e *errCmdNotFound) NotFound() bool {
	return true
}

func NewCommand(name string, sh *Shell) (*Command, error) {
	var (
		err error
	)

	cmdPath := name

	if name[0] != '/' {
		cmdPath, err = exec.LookPath(name)

		if err != nil {
			return nil, newCmdNotFound("Command '%s' not found: %s", name, err.Error())
		}
	}

	sh.log("Executing: %s\n", cmdPath)

	envVars := buildenv(sh.Environ())

	if sh.debug {
		for _, ev := range envVars {
			sh.log("ENV %s", ev)
		}
	}

	cmd := &Command{
		name: name,
		sh:   sh,
		Cmd: &exec.Cmd{
			Path:   cmdPath,
			Stdin:  os.Stdin,
			Stdout: os.Stdout,
			Stderr: os.Stderr,
			Env:    envVars,
		},
		fdMap:      make(FDMap),
		stdinDone:  make(chan bool, 1),
		stdoutDone: make(chan bool, 1),
		stderrDone: make(chan bool, 1),
	}

	cmd.fdMap[0] = sh.stdin
	cmd.fdMap[1] = sh.stdout
	cmd.fdMap[2] = sh.stderr

	return cmd, nil
}

func (cmd *Command) SetArgs(cargs []*Arg) error {
	var err error

	sh := cmd.sh
	args := make([]string, len(cargs)+1)
	args[0] = cmd.name

	for i := 0; i < len(cargs); i++ {
		var argVal string

		carg := cargs[i]

		if carg.IsConcat() {
			argVal, err = sh.executeConcat(carg)

			if err != nil {
				return err
			}
		} else if carg.IsVariable() {
			arglist, err := sh.evalVariable(carg.Value())

			if err != nil {
				return err
			}

			if len(arglist) == 0 {
				return newError("Variable '%s' not set", carg.Value())
			} else if len(arglist) == 1 {
				argVal = arglist[0]
			} else if len(arglist) > 1 {
				argVal = strings.Join(arglist, " ")
			}
		} else {
			argVal = carg.Value()
		}

		args[i+1] = argVal
	}

	cmd.Cmd.Args = args
	return nil
}

func (cmd *Command) SetRedirects(redirDecls []*RedirectNode) error {
	var err error

	for _, r := range redirDecls {
		err = cmd.buildRedirect(r)

		if err != nil {
			return err
		}
	}

	err = cmd.setupRedirects()

	if err != nil {
		return err
	}

	return nil
}

func (cmd *Command) buildRedirect(redirDecl *RedirectNode) error {
	if redirDecl.rmap.lfd > 2 || redirDecl.rmap.lfd < redirMapSupress {
		return newError("Invalid file descriptor redirection: fd=%d", redirDecl.rmap.lfd)
	}

	if redirDecl.rmap.rfd > 2 || redirDecl.rmap.rfd < redirMapSupress {
		return newError("Invalid file descriptor redirection: fd=%d", redirDecl.rmap.rfd)
	}

	// Note(i4k): We need to remove the repetitive code in some smarter way
	switch redirDecl.rmap.lfd {
	case 0:
		switch redirDecl.rmap.rfd {
		case 0: // does nothing
		case 1:
			cmd.fdMap[0] = cmd.fdMap[1]
		case 2:
			cmd.fdMap[0] = cmd.fdMap[2]
		case redirMapNoValue:
			if redirDecl.location == "" {
				return newError("Missing file in redirection: >[%d] <??>", redirDecl.rmap.lfd)
			}

			file, err := os.OpenFile(redirDecl.location, os.O_RDWR|os.O_CREATE, 0644)

			if err != nil {
				return err
			}

			cmd.fdMap[0] = file
		case redirMapSupress:
			if redirDecl.location != "" {
				return newError("Invalid redirect mapping: %d -> %d",
					redirDecl.rmap.lfd,
					redirDecl.location)
			}

			file, err := os.OpenFile("/dev/null", os.O_RDWR, 0644)

			if err != nil {
				return err
			}

			cmd.fdMap[0] = file
		}
	case 1:
		switch redirDecl.rmap.rfd {
		case 0:
			return newError("Invalid redirect mapping: %d -> %d", 1, 0)
		case 1: // do nothing
		case 2:
			cmd.fdMap[1] = cmd.fdMap[2]
		case redirMapNoValue:
			if redirDecl.location == "" {
				return newError("Missing file in redirection: >[%d] <??>", redirDecl.rmap.lfd)
			}

			file, err := os.OpenFile(redirDecl.location, os.O_RDWR|os.O_CREATE, 0644)

			if err != nil {
				return err
			}

			cmd.fdMap[1] = file
		case redirMapSupress:
			file, err := os.OpenFile("/dev/null", os.O_RDWR, 0644)

			if err != nil {
				return err
			}

			cmd.fdMap[1] = file
		}
	case 2:
		switch redirDecl.rmap.rfd {
		case 0:
			return newError("Invalid redirect mapping: %d -> %d", 2, 1)
		case 1:
			cmd.fdMap[2] = cmd.fdMap[1]
		case 2: // do nothing
		case redirMapNoValue:
			if redirDecl.location == "" {
				return newError("Missing file in redirection: >[%d] <??>", redirDecl.rmap.lfd)
			}

			file, err := os.OpenFile(redirDecl.location, os.O_RDWR|os.O_CREATE, 0644)

			if err != nil {
				return err
			}

			cmd.fdMap[2] = file
		case redirMapSupress:
			file, err := os.OpenFile("/dev/null", os.O_RDWR, 0644)

			if err != nil {
				return err
			}

			cmd.fdMap[2] = file
		}
	case redirMapNoValue:
		if redirDecl.location == "" {
			return newError("Missing file in redirection: >[%d] <??>", redirDecl.rmap.lfd)
		}

		file, err := os.OpenFile(redirDecl.location, os.O_RDWR|os.O_CREATE, 0644)

		if err != nil {
			return err
		}

		cmd.fdMap[1] = file
	}

	return nil
}

func (cmd *Command) setupStdin(value interface{}) error {
	rc, ok := value.(io.Reader)

	if !ok {
		return newError("Stdin requires a reader stream in redirect")
	}

	if rc == os.Stdin {
		cmd.Stdin = rc
		cmd.stdinDone <- true
	} else {
		cmd.Stdin = nil
		stdin, err := cmd.StdinPipe()

		if err != nil {
			return err
		}

		go func() {
			io.Copy(stdin, rc)
			cmd.stdinDone <- true
		}()
	}

	return nil
}

func (cmd *Command) setupStdout(value interface{}) error {
	wc, ok := value.(io.Writer)

	if !ok {
		return newError("Stdout requires a writer stream in redirect")
	}

	switch wc {
	case os.Stdin:
		return newError("Invalid redirect mapping: %d -> %d", 1, 0)
	case os.Stdout:
		cmd.Stdout = os.Stdout
		cmd.stdoutDone <- true
	case os.Stderr:
		cmd.Stdout = cmd.Stderr
		cmd.stdoutDone <- true
	default:
		cmd.Stdout = nil
		stdout, err := cmd.StdoutPipe()

		if err != nil {
			return err
		}

		go func() {
			io.Copy(wc, stdout)
			cmd.stdoutDone <- true
		}()
	}

	return nil
}

func (cmd *Command) setupStderr(value interface{}) error {
	wc, ok := value.(io.Writer)

	if !ok {
		return newError("Stderr requires a writer stream in redirect")
	}

	switch wc {
	case os.Stdin:
		return newError("Invalid redirect mapping: %d -> %d", 2, 1)
	case os.Stdout:
		cmd.Stderr = cmd.Stdout
		cmd.stderrDone <- true
	case os.Stderr:
		cmd.Stderr = os.Stderr
		cmd.stderrDone <- true
	default:
		cmd.Stderr = nil
		stderr, err := cmd.StderrPipe()

		if err != nil {
			return err
		}

		go func() {
			io.Copy(wc, stderr)
			cmd.stderrDone <- true
		}()
	}

	return nil
}

func (cmd *Command) setupRedirects() error {
	for k, v := range cmd.fdMap {
		switch k {
		case 0:
			err := cmd.setupStdin(v)

			if err != nil {
				return err
			}
		case 1:
			err := cmd.setupStdout(v)

			if err != nil {
				return err
			}
		case 2:
			err := cmd.setupStderr(v)

			if err != nil {
				return err
			}
		default:
			return newError("Invalid file descriptor redirection: fd=%d", k)
		}
	}

	return nil
}

func (cmd *Command) Execute() error {
	err := cmd.Start()

	if err != nil {
		return err
	}

	<-cmd.stdinDone
	<-cmd.stdoutDone
	<-cmd.stderrDone

	err = cmd.Wait()

	if err != nil {
		return err
	}

	return nil
}
