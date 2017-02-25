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
		result string
		fail   bool
	}

	tests := map[string]exitDesc{
		"success": {
			script: "./testdata/exit.sh",
			status: "0",
			result: "0",
		},
		"failure": {
			script: "./testdata/exit.sh",
			status: "1",
			result: "1",
		},
		"maxStatus": {
			script: "./testdata/exit.sh",
			status: "255",
			result: "255",
		},
		"statusIsUnsigned": {
			script: "./testdata/exit.sh",
			status: "-1",
			result: "255",
		},
		"statusOverflow": {
			script: "./testdata/exit.sh",
			status: "666",
			result: "154", // Why ? For the glory of satan of course :-)
		},
	}

	//WHY: Not sure this is a great idea, but we need to exec with the
	//code under test nash, not the one installed on the system.
	//Can't circumvent the need for Exec here.
	//Other tests can just run nash inside their own process.
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
				if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
					expectedStatus, err := strconv.Atoi(desc.result)
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
