package builtin_test

import (
	"os"
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
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr // to know why scripts were failing
			cmd.Run()

			if cmd.ProcessState == nil {
				t.Fatalf("expected cmd[%v] to have a process state, can't validate status code", cmd)
			}
			got := cmd.ProcessState.ExitCode()
			if desc.result != got {
				t.Fatalf("expected[%d] got[%d]", desc.result, got)
			}
		})
	}
}
