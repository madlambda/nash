// +build linux

// nash provides the execution engine
package sh

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"strconv"
	"syscall"
	"time"

	"github.com/NeowayLabs/nash/ast"
)

func getProcAttrs(flags uintptr) *syscall.SysProcAttr {
	uid := os.Getuid()
	gid := os.Getgid()

	sysproc := &syscall.SysProcAttr{
		Cloneflags: flags,
	}

	if (flags & syscall.CLONE_NEWUSER) == syscall.CLONE_NEWUSER {
		sysproc.UidMappings = []syscall.SysProcIDMap{
			{
				ContainerID: 0,
				HostID:      uid,
				Size:        1,
			},
		}

		sysproc.GidMappings = []syscall.SysProcIDMap{
			{
				ContainerID: 0,
				HostID:      gid,
				Size:        1,
			},
		}
	}

	return sysproc
}

func dialRc(sockpath string) (net.Conn, error) {
	retries := 0

retryRforkDial:
	client, err := net.Dial("unix", sockpath)

	if err != nil {
		if retries < 3 {
			retries++
			time.Sleep(time.Duration(retries) * time.Second)
			goto retryRforkDial
		}
	}

	return client, err
}

// executeRfork executes the calling program again but passing
// a new name for the process on os.Args[0] and passing an unix
// socket file to communicate to.
func (sh *Shell) executeRfork(rfork *ast.RforkNode) error {
	var (
		tr               *ast.Tree
		i                int
		nashClient       net.Conn
		copyOut, copyErr bool
	)

	if sh.stdout != os.Stdout {
		copyOut = true
	}

	if sh.stderr != os.Stderr {
		copyErr = true
	}

	if sh.nashdPath == "" {
		return fmt.Errorf("Nashd not set")
	}

	unixfile := "/tmp/nash." + randRunes(4) + ".sock"

	cmd := exec.Cmd{
		Path: sh.nashdPath,
		Args: append([]string{"-nashd-"}, "-noinit", "-addr", unixfile),
		Env:  buildenv(sh.Environ()),
	}

	arg := rfork.Arg()

	forkFlags, err := getflags(arg.Value())

	if err != nil {
		return err
	}

	cmd.SysProcAttr = getProcAttrs(forkFlags)

	stdoutDone := make(chan bool)
	stderrDone := make(chan bool)

	var (
		stdout, stderr io.ReadCloser
	)

	if copyOut {
		stdout, err = cmd.StdoutPipe()

		if err != nil {
			return err
		}
	} else {
		cmd.Stdout = sh.stdout
		close(stdoutDone)
	}

	if copyErr {
		stderr, err = cmd.StderrPipe()

		if err != nil {
			return err
		}
	} else {
		cmd.Stderr = sh.stderr
		close(stderrDone)
	}

	cmd.Stdin = sh.stdin

	err = cmd.Start()

	if err != nil {
		return err
	}

	if copyOut {
		go func() {
			defer close(stdoutDone)

			io.Copy(sh.stdout, stdout)
		}()
	}

	if copyErr {
		go func() {
			defer close(stderrDone)

			io.Copy(sh.stderr, stderr)
		}()
	}

	nashClient, err = dialRc(unixfile)

	defer nashClient.Close()

	tr = rfork.Tree()

	if tr == nil || tr.Root == nil {
		return fmt.Errorf("Rfork with no sub block")
	}

	for i = 0; i < len(tr.Root.Nodes); i++ {
		var (
			n, status int
		)

		node := tr.Root.Nodes[i]
		data := []byte(node.String() + "\n")

		n, err = nashClient.Write(data)

		if err != nil || n != len(data) {
			return fmt.Errorf("RPC call failed: Err: %v, bytes written: %d", err, n)
		}

		// read response

		var response [1024]byte
		n, err = nashClient.Read(response[:])

		if err != nil {
			break
		}

		status, err = strconv.Atoi(string(response[0:n]))

		if err != nil {
			err = fmt.Errorf("Invalid status: %s", string(response[0:n]))
			break
		}

		if status != 0 {
			err = fmt.Errorf("nash: Exited with status %d", status)
			break
		}
	}

	// we're done with rfork daemon
	nashClient.Write([]byte("quit"))

	<-stdoutDone
	<-stderrDone

	err2 := cmd.Wait()

	if err != nil {
		return err
	}

	if err2 != nil {
		return err2
	}

	return nil
}

func getflags(flags string) (uintptr, error) {
	var (
		lflags uintptr
	)

	for i := 0; i < len(flags); i++ {
		switch flags[i] {
		case 'c':
			lflags |= (syscall.CLONE_NEWUSER |
				syscall.CLONE_NEWPID |
				syscall.CLONE_NEWNET |
				syscall.CLONE_NEWNS |
				syscall.CLONE_NEWUTS |
				syscall.CLONE_NEWIPC)
		case 'u':
			lflags |= syscall.CLONE_NEWUSER
		case 'p':
			lflags |= syscall.CLONE_NEWPID
		case 'n':
			lflags |= syscall.CLONE_NEWNET
		case 'm':
			lflags |= syscall.CLONE_NEWNS
		case 's':
			lflags |= syscall.CLONE_NEWUTS
		case 'i':
			lflags |= syscall.CLONE_NEWIPC
		default:
			return 0, fmt.Errorf("Wrong rfork flag: %c", flags[i])
		}
	}

	if lflags == 0 {
		return 0, fmt.Errorf("Rfork requires some flag")
	}

	return lflags, nil
}
