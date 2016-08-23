// +build linux

package sh

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
)

var (
	enableUserNS bool
)

func init() {
	// Travis build doesn't support /proc/config.gz but have userns enabled
	if os.Getenv("TRAVIS_BUILD") == "1" {
		enableUserNS = true

		return
	}

	usernsCmd := exec.Command("zgrep", "CONFIG_USER_NS", "/proc/config.gz")

	content, err := usernsCmd.CombinedOutput()

	if err != nil {
		fmt.Printf("ERROR: %s\n", err.Error())
		fmt.Printf("Warning: Impossible to know if kernel support USER namespace.\n")
		fmt.Printf("Warning: USER namespace tests will not run.\n")
		enableUserNS = false
	}

	switch strings.Trim(string(content), "\n \t") {
	case "CONFIG_USER_NS=y":
		enableUserNS = true
	default:
		enableUserNS = false
	}
}

func TestExecuteRforkUserNS(t *testing.T) {
	if !enableUserNS {
		t.Skip("User namespace not enabled")
		return
	}

	var out bytes.Buffer

	sh, err := NewShell()

	if err != nil {
		t.Error(err)
		return
	}

	sh.SetNashdPath(nashdPath)
	sh.SetStdout(&out)

	err = sh.Exec("rfork test", `
        rfork u {
            id -u
        }
        `)

	if err != nil {
		t.Error(err)
		return
	}

	if string(out.Bytes()) != "0\n" {
		t.Errorf("User namespace not supported in your kernel: %s", string(out.Bytes()))
		return
	}
}

func TestExecuteRforkEnvVars(t *testing.T) {
	if !enableUserNS {
		t.Skip("User namespace not enabled")
		return
	}

	sh, err := NewShell()

	if err != nil {
		t.Error(err)
		return
	}

	sh.SetNashdPath(nashdPath)

	err = sh.Exec("test env", `abra = "cadabra"
setenv abra
rfork up {
	echo $abra
}`)

	if err != nil {
		t.Error(err)
		return
	}
}

func TestExecuteRforkUserNSNested(t *testing.T) {
	if !enableUserNS {
		t.Skip("User namespace not enabled")
		return
	}

	var out bytes.Buffer

	sh, err := NewShell()

	if err != nil {
		t.Error(err)
		return
	}

	sh.SetNashdPath(nashdPath)
	sh.SetStdout(&out)

	err = sh.Exec("rfork userns nested", `
        rfork u {
            id -u
            rfork u {
                id -u
            }
        }
        `)

	if err != nil {
		t.Error(err)
		return
	}

	if string(out.Bytes()) != "0\n0\n" {
		t.Errorf("User namespace not supported in your kernel")
		return
	}
}
