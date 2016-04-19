package cnt

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"syscall"
	"time"
)

func getProcAttrs(flags uintptr) *syscall.SysProcAttr {
	uid := os.Getuid()
	gid := os.Getgid()

	return &syscall.SysProcAttr{
		Cloneflags: flags,
		UidMappings: []syscall.SysProcIDMap{
			{
				ContainerID: 0,
				HostID:      uid,
				Size:        1,
			},
		},
		GidMappings: []syscall.SysProcIDMap{
			{
				ContainerID: 0,
				HostID:      gid,
				Size:        1,
			},
		},
	}
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
func (sh *Shell) executeRfork(rfork *RforkNode) error {
	var (
		tr        *Tree
		i         int
		cntClient net.Conn
	)

	if sh.cntdPath == "" {
		return fmt.Errorf("Cntd not set")
	}

	unixfile := "/tmp/cnt." + randRunes(4) + ".sock"

	cmd := exec.Cmd{
		Path: sh.cntdPath,
		Args: append([]string{"-rcd-"}, "-addr", unixfile),
	}

	forkFlags, err := getflags(rfork.arg.val)

	if err != nil {
		return err
	}

	cmd.SysProcAttr = getProcAttrs(forkFlags)
	cmd.Stdin = sh.stdin
	cmd.Stdout = sh.stdout
	cmd.Stderr = sh.stderr

	err = cmd.Start()

	if err != nil {
		return err
	}

	cntClient, err = dialRc(unixfile)

	defer cntClient.Close()

	tr = rfork.Tree()

	if tr == nil || tr.Root == nil {
		return fmt.Errorf("Rfork with no sub block")
	}

	for i = 0; i < len(tr.Root.Nodes); i++ {
		node := tr.Root.Nodes[i]
		data := []byte(node.String() + "\n")

		n, err := cntClient.Write(data)

		if err != nil || n != len(data) {
			return fmt.Errorf("RPC call failed: Err: %v, bytes written: %d", err, n)
		}

		// read response

		var response [1024]byte
		n, err = cntClient.Read(response[:])

		if err != nil {
			break
		}

		status, err := strconv.Atoi(string(response[0:n]))

		if err != nil {
			err = fmt.Errorf("Invalid status: %s", string(response[0:n]))
			break
		}

		if status != 0 {
			err = fmt.Errorf("rc: Exited with status %d", status)
			break
		}
	}

	// we're done with rfork daemon
	cntClient.Write([]byte("quit"))

	if err != nil {
		return err
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
