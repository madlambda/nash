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

		ignoreError bool

		stdinDone, stdoutDone, stderrDone chan bool
	}
)

var (
	ErrVarNotSet          = NewError("Variable '%s' not set")
	ErrInvalidFD          = NewError("Invalid file descriptor redirection: fd=%d")
	ErrMissingFile        = NewError("Missing file in redirection: >[%d] <??>")
	ErrInvalidSuppressMap = NewError("Suppressing a file descriptor and redirecting to file")
	ErrInvalidMap         = NewError("Invalid redirect mapping: %d -> %d")
	ErrInvalidStdin       = NewError("Stdin requires a reader stream in redirect")
	ErrInvalidStdout      = NewError("Stdout requires a writer stream in redirect")
	ErrInvalidStderr      = NewError("Stderr requires a writer stream in redirect")
)

func NewCommand(name string, sh *Shell) (*Command, error) {
	var (
		ignoreError bool
		err         error
	)

	cmdPath := name

	if len(name) > 1 && name[0] == '-' {
		ignoreError = true
		name = name[1:]

		sh.log("Ignoring error\n")
	}

	if name[0] != '/' {
		cmdPath, err = exec.LookPath(name)

		if err != nil {
			return nil, err
		}
	}

	sh.log("Executing: %s\n", cmdPath)

	cmd := &Command{
		name:        name,
		sh:          sh,
		ignoreError: ignoreError,
		Cmd: &exec.Cmd{
			Path:   cmdPath,
			Stdin:  os.Stdin,
			Stdout: os.Stdout,
			Stderr: os.Stderr,
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

func (cmd *Command) SetArgs(cargs []Arg) error {
	sh := cmd.sh
	args := make([]string, len(cargs)+1)
	args[0] = cmd.name

	for i := 0; i < len(cargs); i++ {
		argval := cargs[i].val

		// variable substitution
		if len(argval) > 0 && argval[0] == '$' {
			if sh.env[argval[1:]] != nil {
				arglist := sh.env[argval[1:]]

				if len(arglist) == 1 {
					args[i+1] = arglist[0]
				} else if len(arglist) > 1 {
					args[i+1] = strings.Join(arglist, " ")
				}
			} else {
				return ErrVarNotSet
			}
		} else {
			args[i+1] = argval
		}
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
		return ErrInvalidFD.Params(redirDecl.rmap.lfd)
	}

	if redirDecl.rmap.rfd > 2 || redirDecl.rmap.rfd < redirMapSupress {
		return ErrInvalidFD.Params(redirDecl.rmap.rfd)
	}

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
				return ErrMissingFile.Params(redirDecl.rmap.lfd)
			}

			file, err := os.OpenFile(redirDecl.location, os.O_RDWR|os.O_CREATE, 0644)

			if err != nil {
				return err
			}

			cmd.fdMap[0] = file
		case redirMapSupress:
			if redirDecl.location != "" {
				return ErrInvalidMap.Params(redirDecl.rmap.lfd,
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
			return ErrInvalidMap.Params(1, 0)
		case 1: // do nothing
		case 2:
			cmd.fdMap[1] = cmd.fdMap[2]
		case redirMapNoValue:
			if redirDecl.location == "" {
				return ErrMissingFile.Params(redirDecl.rmap.lfd)
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
			return ErrInvalidMap.Params(2, 1)
		case 1:
			cmd.fdMap[2] = cmd.fdMap[1]
		case 2: // do nothing
		case redirMapNoValue:
			if redirDecl.location == "" {
				return ErrMissingFile.Params(redirDecl.rmap.lfd)
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
			return ErrMissingFile.Params(redirDecl.rmap.lfd)
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
		return ErrInvalidStdin
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
		return ErrInvalidStdout
	}

	switch wc {
	case os.Stdin:
		return ErrInvalidMap.Params(1, 0)
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
		return ErrInvalidStderr
	}

	switch wc {
	case os.Stdin:
		return ErrInvalidMap.Params(2, 1)
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
			return ErrInvalidFD.Params(k)
		}
	}

	return nil
}

func (cmd *Command) Execute() error {
	err := cmd.Start()

	if err != nil {
		if cmd.ignoreError {
			return nil
		}

		return err
	}

	<-cmd.stdinDone
	<-cmd.stdoutDone
	<-cmd.stderrDone

	err = cmd.Wait()

	if err != nil {
		if cmd.ignoreError {
			return nil
		}

		return err
	}

	return nil
}
