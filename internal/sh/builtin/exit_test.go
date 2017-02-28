package builtin_test

import (
	"os/exec"
	"testing"
)

// getCmdStatusCode must return the status code of the given
// err returned by exec.Exec or fail if unable to.
type getCmdStatusCode func(err error) int

// testExit tests builtin exit function, you need to provide
// a platform dependent way to get the status code of a exited command
// to run the test on a platform.
func testExit(t *testing.T, getstatus getCmdStatusCode) {
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
		"maxStatus": {
			script: "./testdata/exit.sh",
			status: "255",
			result: 255,
		},
		"statusIsUnsigned": {
			script: "./testdata/exit.sh",
			status: "-1",
			result: 255,
		},
		"statusOverflow": {
			script: "./testdata/exit.sh",
			status: "666",
			result: 154, // Why ? For the glory of satan of course :-)
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
			got := getstatus(err)
			if desc.result != got {
				t.Fatalf("expected[%d] got[%d]", desc.result, got)
			}
		})
	}
}
