// +build darwin dragonfly freebsd linux nacl netbsd openbsd solaris windows

package builtin_test

import (
	"os/exec"
	"syscall"
	"testing"
)

func TestPosixExit(t *testing.T) {
	testExit(t, func(err error) int {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				return status.ExitStatus()
			} else {
				t.Fatal("unable to extract status code from exec")
			}
		} else {
			t.Fatalf("unexpected error: %v", err)
		}
		return 0 //unrecheable
	})
}
