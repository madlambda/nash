package cnt

import (
	"encoding/gob"
	"fmt"
	"math/rand"
	"net/rpc"
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
	uid := os.Getuid()
	gid := os.Getgid()

	unixfile := "/tmp/cnt." + randRunes(4) + ".sock"

	cmd := exec.Cmd{
		Path: os.Args[0],
		Args: append([]string{"cnt-rfork"}, "-rfork-sock", unixfile),
	}

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: getflags(rfork.args),
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

retryRforkDial:
	rforkClient, err := rpc.Dial("unix", unixfile)
	retries = 0

	if err != nil && retries < 3 {
		retries++
		time.Sleep(retries * time.Second)
		fmt.Printf("Retrying to dial cnt-rfork...\n")

		goto retryRforkDial
	}

	defer rforkClient.Close()

	tr := rfork.Tree()

	if tr == nil || tr.Root == nil {
		return fmt.Errorf("Rfork with no sub block")
	}

	enc := gob.NewEncoder(&rforkClient)

	for i := 0; i < tr.Root.Nodes; i++ {
		node := tr.Root.Nodes[i]

		var status error
		err = rforkClient.Call("RforkService.ExecuteNode", &node, &status)

		if err != nil {
			return fmt.Printf("RPC call failed: %s", err.Error())
		}

		if status != nil {
			return status
		}
	}

	return nil

}

func getflags(flags string) int {
	return syscall.CLONE_NEWUSER | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS
}
