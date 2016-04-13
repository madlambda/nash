package cnt

import (
	"fmt"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"syscall"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func executeRfork(rfork *RforkNode) error {
	var (
		tr *Tree
		i  int
	)

	uid := os.Getuid()
	gid := os.Getgid()

	unixfile := "/tmp/cnt." + randRunes(4) + ".sock"

	cmd := exec.Cmd{
		Path: os.Args[0],
		Args: append([]string{"-rcd-"}, "-addr", unixfile),
	}

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: getflags(rfork.arg.val),
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

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout

	err := cmd.Start()

	if err != nil {
		return err
	}

	retries := 0

retryRforkDial:
	rforkClient, err := net.Dial("unix", unixfile)

	if err != nil {
		if retries < 3 {
			retries++
			time.Sleep(time.Duration(retries) * time.Second)
			fmt.Printf("Retrying to dial cnt-rfork...\n")

			goto retryRforkDial
		}

		goto rforkErr
	}

	defer rforkClient.Close()

	tr = rfork.Tree()

	if tr == nil || tr.Root == nil {
		return fmt.Errorf("Rfork with no sub block")
	}

	time.Sleep(10 * time.Second)

	for i = 0; i < len(tr.Root.Nodes); i++ {
		node := tr.Root.Nodes[i]

		n, err := rforkClient.Write([]byte(node.String()))

		if err != nil {
			return fmt.Errorf("RPC call failed: %s", err.Error())
		}

		fmt.Printf("Written %d bytes\n", n)
	}

	time.Sleep(40 * time.Second)

	return nil

rforkErr:
	return err

}

func getflags(flags string) uintptr {
	return syscall.CLONE_NEWUSER | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS
}
