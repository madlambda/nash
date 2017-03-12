package builtin_test

import (
	"os/exec"
	"testing"
)

func TestExit(t *testing.T) {
	type exitDesc struct {
		script string
		status string
		result int
		fail   bool
	}

	// exitResult is a common interface implemented by
	// all platforms.
	type exitResult interface {
		ExitStatus() int
	}

	tests := map[string]exitDesc{
		"success": {
			script: "./testdata/exit.sh",
			status: "0",
			result: 0,
		},
		"failure": {
			script: "./testdata/exit.sh",
			status: "1",
			result: 1,
		},
	}

	// WHY: We need to run Exec because the script will call the exit syscall,
	// killing the process (the test process on this case).
	// When calling Exec we need to guarantee that we are using the nash
	// built directly from the project, not the one installed on the host.
	projectnash := "../../../cmd/nash/nash"

	for name, desc := range tests {
		t.Run(name, func(t *testing.T) {
			cmd := exec.Command(projectnash, desc.script, desc.status)
			err := cmd.Run()
			if err == nil {
				if desc.status == "0" {
					return
				}
				t.Fatalf("expected error for status: %s", desc.status)

			}
			if exiterr, ok := err.(*exec.ExitError); ok {
				if status, ok := exiterr.Sys().(exitResult); ok {
					got := status.ExitStatus()
					if desc.result != got {
						t.Fatalf("expected[%d] got[%d]", desc.result, got)
					}
				} else {
					t.Fatal("exit result does not have a  ExitStatus method")
				}
			} else {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
