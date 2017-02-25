package builtin_test

import (
	"os/exec"
	"strconv"
	"syscall"
	"testing"
)

func TestLinuxExit(t *testing.T) {
	type exitDesc struct {
		script string
		status string
		fail   bool
	}

	tests := map[string]exitDesc{
		"success": {
			script: "./testdata/exit.sh",
			status: "0",
		},
		"failure1": {
			script: "./testdata/exit.sh",
			status: "1",
		},
		"failure-1": {
			script: "./testdata/exit.sh",
			status: "-1",
		},
	}

	for name, desc := range tests {
		t.Run(name, func(t *testing.T) {
			cmd := exec.Command(desc.script, desc.status)
			err := cmd.Run()
			if err == nil {
				if desc.status != "0" {
					t.Fatalf("expected error for status: %s", desc.status)
				}

			}
			if exiterr, ok := err.(*exec.ExitError); ok {
				if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
					expectedStatus, err := strconv.Atoi(desc.status)
					if err != nil {
						t.Fatalf("error[%s] converting[%s]", err, desc.status)
					}
					got := status.ExitStatus()
					if expectedStatus != got {
						t.Fatalf("expected[%d] got[%d]", expectedStatus, got)
					}
				}
			} else {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
