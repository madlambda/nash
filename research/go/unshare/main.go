package main

import (
	"log"
	"os"
	"os/exec"
	"syscall"
)

func main() {
	err := syscall.Unshare(syscall.CLONE_NEWPID)

	if err != nil {
		log.Fatal(err)
	}

	cmd := exec.Command("/bin/bash")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	err = cmd.Start()

	if err != nil {
		log.Fatal(err)
	}

	err = cmd.Wait()

	if err != nil {
		log.Fatal(err)
	}
}
