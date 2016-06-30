package nash

import (
	"io"
	"net"
	"os"
	"os/exec"
	"strings"

	"github.com/NeowayLabs/nash/ast"
	"github.com/NeowayLabs/nash/errors"
)

type (
	FDMap map[int]interface{}

	Command struct {
		name string
		*exec.Cmd
		sh    *Shell
		fdMap FDMap

		stdinDone, stdoutDone, stderrDone chan bool
		passDone                          bool
	}

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

func NewCommand(name string, sh *Shell) (*Command, error) {
	var (
		err error
	)

	cmdPath := name

	if name[0] != '/' {
		cmdPath, err = exec.LookPath(name)

		if err != nil {
			return nil, newCmdNotFound("Command '%s' not found on PATH=%s: %s",
				name,
				os.Getenv("PATH"),
				err.Error())
		}
	}

	sh.logf("Executing: %s\n", cmdPath)

	envVars := buildenv(sh.Environ())

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

		// if set to false, you need to sinchronize by hand
		// be careful with deadlocks
		passDone: true,
	}

	cmd.fdMap[0] = os.Stdin
	cmd.fdMap[1] = os.Stdout
	cmd.fdMap[2] = os.Stderr

	return cmd, nil
}

func (cmd *Command) SetPassDone(b bool) {
	cmd.passDone = b
}

func (cmd *Command) SetFDMap(id int, value interface{}) {
	cmd.fdMap[id] = value
}

func (cmd *Command) SetArgs(cargs []*ast.Arg) error {
	sh := cmd.sh
	args := make([]string, len(cargs)+1)
	args[0] = cmd.name

	for i := 0; i < len(cargs); i++ {
		var argVal string

		carg := cargs[i]

		obj, err := sh.evalArg(carg)

		if err != nil {
			return err
		}

		if obj.Type() == StringType {
			argVal = obj.Str()
		} else if obj.Type() == ListType {
			argVal = strings.Join(obj.List(), " ")
		} else if obj.Type() == FnType {
			return errors.NewError("Impossible to pass function to command as argument.")
		} else {
			return errors.NewError("Invalid argument '%v'", carg)
		}

		args[i+1] = argVal
	}

	cmd.Cmd.Args = args
	return nil
}

func (cmd *Command) SetRedirects(redirDecls []*ast.RedirectNode) error {
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

func (cmd *Command) openRedirectLocation(location *ast.Arg) (io.WriteCloser, error) {
	var (
		protocol, locationStr string
	)

	if !location.IsVariable() && !location.IsQuoted() && !location.IsUnquoted() {
		return nil, errors.NewError("Invalid argument of type %v in redirection", location.ArgType())
	}

	if location.IsQuoted() || location.IsUnquoted() {
		locationStr = location.Value()
	} else {
		obj, err := cmd.sh.evalVariable(location)

		if err != nil {
			return nil, err
		}

		if obj.Type() != StringType {
			return nil, errors.NewError("Invalid object type in redirection: %+v", obj.Type())
		}

		locationStr = obj.Str()
	}

	if len(locationStr) > 6 {
		if locationStr[0:6] == "tcp://" {
			protocol = "tcp"
		} else if locationStr[0:6] == "udp://" {
			protocol = "udp"
		} else if len(locationStr) > 7 && locationStr[0:7] == "unix://" {
			protocol = "unix"
		}
	}

	if protocol == "" {
		return os.OpenFile(locationStr, os.O_RDWR|os.O_CREATE, 0644)
	}

	switch protocol {
	case "tcp", "udp":
		netParts := strings.Split(locationStr[6:], ":")

		if len(netParts) != 2 {
			return nil, errors.NewError("Invalid tcp/udp address: %s", locationStr)
		}

		url := netParts[0] + ":" + netParts[1]

		return net.Dial(protocol, url)
	case "unix":
		return net.Dial(protocol, locationStr[7:])
	}

	return nil, errors.NewError("Unexpected redirection value: %s", locationStr)
}

func (cmd *Command) buildRedirect(redirDecl *ast.RedirectNode) error {
	if redirDecl.LeftFD() > 2 || redirDecl.LeftFD() < ast.RedirMapSupress {
		return errors.NewError("Invalid file descriptor redirection: fd=%d", redirDecl.LeftFD())
	}

	if redirDecl.RightFD() > 2 || redirDecl.RightFD() < ast.RedirMapSupress {
		return errors.NewError("Invalid file descriptor redirection: fd=%d", redirDecl.RightFD())
	}

	// Note(i4k): We need to remove the repetitive code in some smarter way
	switch redirDecl.LeftFD() {
	case 0:
		switch redirDecl.RightFD() {
		case 0: // does nothing
		case 1:
			cmd.fdMap[0] = cmd.fdMap[1]
		case 2:
			cmd.fdMap[0] = cmd.fdMap[2]
		case ast.RedirMapNoValue:
			if redirDecl.Location() == nil {
				return errors.NewError("Missing file in redirection: >[%d] <??>", redirDecl.LeftFD())
			}

			file, err := cmd.openRedirectLocation(redirDecl.Location())

			if err != nil {
				return err
			}

			cmd.fdMap[0] = file
		case ast.RedirMapSupress:
			if redirDecl.Location() != nil {
				return errors.NewError("Invalid redirect mapping: %d -> %d",
					redirDecl.LeftFD(),
					redirDecl.RightFD())
			}

			file, err := os.OpenFile("/dev/null", os.O_RDWR, 0644)

			if err != nil {
				return err
			}

			cmd.fdMap[0] = file
		}
	case 1:
		switch redirDecl.RightFD() {
		case 0:
			return errors.NewError("Invalid redirect mapping: %d -> %d", 1, 0)
		case 1: // do nothing
		case 2:
			cmd.fdMap[1] = cmd.fdMap[2]
		case ast.RedirMapNoValue:
			if redirDecl.Location() == nil {
				return errors.NewError("Missing file in redirection: >[%d] <??>", redirDecl.LeftFD())
			}

			file, err := cmd.openRedirectLocation(redirDecl.Location())

			if err != nil {
				return err
			}

			cmd.fdMap[1] = file
		case ast.RedirMapSupress:
			file, err := os.OpenFile("/dev/null", os.O_RDWR, 0644)

			if err != nil {
				return err
			}

			cmd.fdMap[1] = file
		}
	case 2:
		switch redirDecl.RightFD() {
		case 0:
			return errors.NewError("Invalid redirect mapping: %d -> %d", 2, 1)
		case 1:
			cmd.fdMap[2] = cmd.fdMap[1]
		case 2: // do nothing
		case ast.RedirMapNoValue:
			if redirDecl.Location() == nil {
				return errors.NewError("Missing file in redirection: >[%d] <??>", redirDecl.LeftFD())
			}

			file, err := cmd.openRedirectLocation(redirDecl.Location())

			if err != nil {
				return err
			}

			cmd.fdMap[2] = file
		case ast.RedirMapSupress:
			file, err := os.OpenFile("/dev/null", os.O_RDWR, 0644)

			if err != nil {
				return err
			}

			cmd.fdMap[2] = file
		}
	case ast.RedirMapNoValue:
		if redirDecl.Location() == nil {
			return errors.NewError("Missing file in redirection: >[%d] <??>", redirDecl.LeftFD())
		}

		file, err := cmd.openRedirectLocation(redirDecl.Location())

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
		return errors.NewError("Stdin requires a reader stream in redirect")
	}

	if rc == os.Stdin {
		cmd.Stdin = rc

		if cmd.passDone {
			cmd.stdinDone <- true
		}
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
		return errors.NewError("Stdout requires a writer stream in redirect")
	}

	switch wc {
	case os.Stdin:
		return errors.NewError("Invalid redirect mapping: %d -> %d", 1, 0)
	case os.Stdout:
		cmd.Stdout = os.Stdout

		if cmd.passDone {
			cmd.stdoutDone <- true
		}
	case os.Stderr:
		cmd.Stdout = cmd.Stderr
		if cmd.passDone {
			cmd.stdoutDone <- true
		}
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
		return errors.NewError("Stderr requires a writer stream in redirect")
	}

	switch wc {
	case os.Stdin:
		return errors.NewError("Invalid redirect mapping: %d -> %d", 2, 1)
	case os.Stdout:
		cmd.Stderr = cmd.Stdout

		if cmd.passDone {
			cmd.stderrDone <- true
		}
	case os.Stderr:
		cmd.Stderr = os.Stderr

		if cmd.passDone {
			cmd.stderrDone <- true
		}
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
			return errors.NewError("Invalid file descriptor redirection: fd=%d", k)
		}
	}

	return nil
}

func (cmd *Command) Wait() error {
	<-cmd.stdinDone
	<-cmd.stdoutDone
	<-cmd.stderrDone

	return cmd.Cmd.Wait()
}

func (cmd *Command) CloseNetDescriptors() {
	for _, fd := range cmd.fdMap {
		if fdc, ok := fd.(*net.TCPConn); ok {
			fdc.Close()
		}
	}
}
