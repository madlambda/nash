// +build linux

package sh_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"testing"
)

var (
	enableUserNS bool
)

func init() {
	const usernsOk = "1"
	const kernelcfg = "CONFIG_USER_NS"

	logUsernsDetection := func(err error) {
		if enableUserNS {
			fmt.Printf("Linux user namespaces enabled!")
			return
		}

		fmt.Printf("Warning: Impossible to know if kernel support USER namespace.\n")
		fmt.Printf("Warning: USER namespace tests will not run.\n")
		if err != nil {
			fmt.Printf("ERROR: %s\n", err)
		}
	}

	usernsCfg := "/proc/sys/kernel/unprivileged_userns_clone"
	val, permerr := ioutil.ReadFile(usernsCfg)

	// Travis build doesn't support /proc/config.gz but kernel has userns
	if os.Getenv("TRAVIS_BUILD") == "1" {
		enableUserNS = permerr == nil && string(val) == usernsOk
		logUsernsDetection(permerr)
		return
	}

	if permerr == nil {
		enableUserNS = string(val) == usernsOk
		logUsernsDetection(permerr)
		return
	}

	// old kernels dont have sysctl configurations
	// than just checking the /proc/config suffices
	usernsCmd := exec.Command("zgrep", kernelcfg, "/proc/config.gz")

	content, err := usernsCmd.CombinedOutput()
	if err != nil {
		enableUserNS = false
		logUsernsDetection(fmt.Errorf("Failed to get kernel config: %s", err))
		return
	}

	cfgVal := strings.Trim(string(content), "\n\t ")
	enableUserNS = cfgVal == kernelcfg+"=y"
	logUsernsDetection(fmt.Errorf("%s not enabled in kernel config", kernelcfg))
}

func TestExecuteRforkUserNS(t *testing.T) {
	if !enableUserNS {
		t.Skip("User namespace not enabled")
		return
	}

	f, teardown := setup(t)
	defer teardown()

	err := f.shell.Exec("rfork test", `
        rfork u {
            id -u
        }
        `)

	if err != nil {
		t.Error(err)
		return
	}

	if string(f.shellOut.Bytes()) != "0\n" {
		t.Errorf("User namespace not supported in your kernel: %s", string(f.shellOut.Bytes()))
		return
	}
}

func TestExecuteRforkEnvVars(t *testing.T) {
	if !enableUserNS {
		t.Skip("User namespace not enabled")
		return
	}

	f, teardown := setup(t)
	defer teardown()

	sh := f.shell

	err := sh.Exec("test env", `var abra = "cadabra"
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
	f, teardown := setup(t)
	defer teardown()

	sh := f.shell

	sh.SetStdout(&out)

	err := sh.Exec("rfork userns nested", `
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
