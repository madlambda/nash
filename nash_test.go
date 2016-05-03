package nash

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
	testDir, gopath, nashdPath string
	enableUserNS               bool
)

func init() {
	gopath = os.Getenv("GOPATH")

	if gopath == "" {
		panic("Please, run tests from inside GOPATH")
	}

	testDir = gopath + "/src/github.com/NeowayLabs/nash/" + "testfiles"
	nashdPath = gopath + "/src/github.com/NeowayLabs/nash/cmd/nash/nash"

	if _, err := os.Stat(nashdPath); err != nil {
		panic("Please, run make build before running tests")
	}

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

func TestExecuteFile(t *testing.T) {
	testfile := testDir + "/ex1.sh"

	var out bytes.Buffer

	sh := NewShell(false)
	sh.SetNashdPath(nashdPath)
	sh.SetStdout(&out)

	err := sh.Execute(testfile)

	if err != nil {
		t.Error(err)
		return
	}

	if string(out.Bytes()) != "hello world\n" {
		t.Errorf("Wrong command output: '%s'", string(out.Bytes()))
		return
	}
}

func TestExecuteRforkUserNS(t *testing.T) {
	if !enableUserNS {
		t.Skip("User namespace not enabled")
		return
	}

	var out bytes.Buffer

	sh := NewShell(false)
	sh.SetNashdPath(nashdPath)
	sh.SetStdout(&out)

	err := sh.ExecuteString("rfork test", `
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

func TestExecuteRforkUserNSNested(t *testing.T) {
	if !enableUserNS {
		t.Skip("User namespace not enabled")
		return
	}

	var out bytes.Buffer

	sh := NewShell(false)
	sh.SetNashdPath(nashdPath)
	sh.SetStdout(&out)

	err := sh.ExecuteString("rfork userns nested", `
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

func TestExecuteAssignment(t *testing.T) {
	var out bytes.Buffer

	sh := NewShell(false)
	sh.SetNashdPath(nashdPath)
	sh.SetStdout(&out)

	err := sh.ExecuteString("assignment", `
        name="i4k"
        echo $name
        `)

	if err != nil {
		t.Error(err)
		return
	}

	if strings.TrimSpace(string(out.Bytes())) != "i4k" {
		t.Error("assignment not work")
		fmt.Printf("'%s' != '%s'\n", strings.TrimSpace(string(out.Bytes())), "i4k")

		return
	}

	out.Reset()

	err = sh.ExecuteString("list assignment", `
        name=(honda civic)
        echo $name
        `)

	if err != nil {
		t.Error(err)
		return
	}

	if strings.TrimSpace(string(out.Bytes())) != "honda civic" {
		t.Error("assignment not work")
		fmt.Printf("'%s' != '%s'\n", strings.TrimSpace(string(out.Bytes())), "honda civic")

		return
	}
}

func TestExecuteRedirection(t *testing.T) {
	sh := NewShell(false)
	sh.SetNashdPath(nashdPath)

	err := sh.ExecuteString("redirect", `
        echo -n "hello world" > /tmp/test1.txt
        `)

	if err != nil {
		t.Error(err)
		return
	}

	content, err := ioutil.ReadFile("/tmp/test1.txt")

	if err != nil {
		t.Error(err)
		return
	}

	if string(content) != "hello world" {
		t.Errorf("File differ: '%s' != '%s'", string(content), "hello world")
		return
	}
}

func TestExecuteRedirectionMap(t *testing.T) {
	sh := NewShell(false)
	sh.SetNashdPath(nashdPath)

	err := sh.ExecuteString("redirect map", `
        echo -n "hello world" > /tmp/test1.txt
        `)

	if err != nil {
		t.Error(err)
		return
	}

	content, err := ioutil.ReadFile("/tmp/test1.txt")

	if err != nil {
		t.Error(err)
		return
	}

	if string(content) != "hello world" {
		t.Errorf("File differ: '%s' != '%s'", string(content), "hello world")
		return
	}
}

func TestExecuteCd(t *testing.T) {
	var out bytes.Buffer

	sh := NewShell(false)
	sh.SetNashdPath(nashdPath)
	sh.SetStdout(&out)

	err := sh.ExecuteString("test cd", `
        cd /tmp
        pwd
        `)

	if err != nil {
		t.Error(err)
		return
	}

	if strings.TrimSpace(string(out.Bytes())) != "/tmp" {
		t.Errorf("Cd failed. '%s' != '%s'", string(out.Bytes()), "/tmp")
		return
	}

	out.Reset()

	var out2 bytes.Buffer
	sh.SetStdout(&out2)

	err = sh.ExecuteString("test cd", `
        HOME="/"
        setenv HOME
        cd
        pwd
        `)

	if err != nil {
		t.Error(err)
		return
	}

	if strings.TrimSpace(string(out2.Bytes())) != "/" {
		t.Errorf("Cd failed. '%s' != '%s'", string(out2.Bytes()), "/")
		return
	}
}
