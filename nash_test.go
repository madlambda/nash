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

func TestExecuteCommand(t *testing.T) {
	sh := NewShell(false)
	sh.SetNashdPath(nashdPath)

	err := sh.ExecuteString("command failed", `
        non-existent-program
        `)

	if err == nil {
		t.Error("Must fail")
		return
	}

	// test ignore works
	err = sh.ExecuteString("command failed", `
        -non-existent-program
        `)

	if err != nil {
		t.Error("Dash at beginning must ignore errors: ERROR: %s", err.Error())
		return
	}

	var out bytes.Buffer
	sh.SetStdout(&out)

	err = sh.ExecuteString("command failed", `
        echo -n "hello world"
        `)

	if err != nil {
		t.Error(err)
		return
	}

	if string(out.Bytes()) != "hello world" {
		t.Errorf("Invalid output: '%s'", string(out.Bytes()))
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

	if strings.TrimSpace(string(out.Bytes())) != "/" {
		t.Errorf("Cd failed. '%s' != '%s'", string(out.Bytes()), "/")
		return
	}

	// test cd into $var
	out.Reset()

	err = sh.ExecuteString("test cd", `
        var="/tmp"
        cd $var
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

	err = sh.ExecuteString("test error", `
        var=("val1" "val2" "val3")
        cd $var
        pwd
        `)

	if err == nil {
		t.Error("Must fail... Impossible to cd into variable containing a list")
		return
	}

	if strings.TrimSpace(string(out.Bytes())) != "" {
		t.Errorf("Cd failed. '%s' != '%s'", string(out.Bytes()), "")
		return
	}
}

func TestExecuteImport(t *testing.T) {
	var out bytes.Buffer

	sh := NewShell(false)
	sh.SetNashdPath(nashdPath)
	sh.SetStdout(&out)

	err := ioutil.WriteFile("/tmp/test.sh", []byte(`TESTE="teste"`), 0644)

	if err != nil {
		t.Error(err)
		return
	}

	err = sh.ExecuteString("test import", `import /tmp/test.sh
        echo $TESTE
        `)

	if err != nil {
		t.Error(err)
		return
	}

	if strings.TrimSpace(string(out.Bytes())) != "teste" {
		t.Error("Import does not work")
		return
	}
}

func TestExecuteShowEnv(t *testing.T) {
	var out bytes.Buffer

	sh := NewShell(false)
	sh.SetNashdPath(nashdPath)
	sh.SetStdout(&out)

	sh.SetEnv(make(Env)) // zero'ing the env

	err := sh.ExecuteString("test showenv", "showenv")

	if err != nil {
		t.Error(err)
		return
	}

	if string(out.Bytes()) != "" {
		t.Errorf("Must be empty. '%s' != ''", string(out.Bytes()))
		return
	}

	out.Reset()

	err = sh.ExecuteString("test showenv", `PATH="/bin"
        setenv PATH
        showenv
        `)

	if err != nil {
		t.Error(err)
		return
	}

	if strings.TrimSpace(string(out.Bytes())) != "PATH=/bin" {
		t.Errorf("Error: '%s' != 'PATH=/bin'", strings.TrimSpace(string(out.Bytes())))
		return
	}
}

func TestExecuteIfEqual(t *testing.T) {
	var out bytes.Buffer

	sh := NewShell(false)
	sh.SetNashdPath(nashdPath)
	sh.SetStdout(&out)

	err := sh.ExecuteString("test if equal", `
        if "" == "" {
            echo "empty string works"
        }`)

	if err != nil {
		t.Error(err)
		return
	}

	if strings.TrimSpace(string(out.Bytes())) != "empty string works" {
		t.Errorf("Must be empty. '%s' != 'empty string works'", string(out.Bytes()))
		return
	}

	out.Reset()

	err = sh.ExecuteString("test if equal 2", `
        if "i4k" == "_i4k_" {
            echo "do not print"
        }`)

	if err != nil {
		t.Error(err)
		return
	}

	if strings.TrimSpace(string(out.Bytes())) != "" {
		t.Errorf("Error: '%s' != ''", strings.TrimSpace(string(out.Bytes())))
		return
	}
}

func TestExecuteIfElse(t *testing.T) {
	var out bytes.Buffer

	sh := NewShell(false)
	sh.SetNashdPath(nashdPath)
	sh.SetStdout(&out)

	err := sh.ExecuteString("test if else", `
        if "" == "" {
            echo "if still works"
        } else {
            echo "nop"
        }`)

	if err != nil {
		t.Error(err)
		return
	}

	if strings.TrimSpace(string(out.Bytes())) != "if still works" {
		t.Errorf("'%s' != 'if still works'", strings.TrimSpace(string(out.Bytes())))
		return
	}

	out.Reset()

	err = sh.ExecuteString("test if equal 2", `
        if "i4k" == "_i4k_" {
            echo "do not print"
        } else {
            echo "print this"
        }`)

	if err != nil {
		t.Error(err)
		return
	}

	if strings.TrimSpace(string(out.Bytes())) != "print this" {
		t.Errorf("Error: '%s' != 'print this'", strings.TrimSpace(string(out.Bytes())))
		return
	}
}

func TestExecuteIfElseIf(t *testing.T) {
	var out bytes.Buffer

	sh := NewShell(false)
	sh.SetNashdPath(nashdPath)
	sh.SetStdout(&out)

	err := sh.ExecuteString("test if else", `
        if "" == "" {
            echo "if still works"
        } else if "bleh" == "bloh" {
            echo "nop"
        }`)

	if err != nil {
		t.Error(err)
		return
	}

	if strings.TrimSpace(string(out.Bytes())) != "if still works" {
		t.Errorf("'%s' != 'if still works'", strings.TrimSpace(string(out.Bytes())))
		return
	}

	out.Reset()

	err = sh.ExecuteString("test if equal 2", `
        if "i4k" == "_i4k_" {
            echo "do not print"
        } else if "a" != "b" {
            echo "print this"
        }`)

	if err != nil {
		t.Error(err)
		return
	}

	if strings.TrimSpace(string(out.Bytes())) != "print this" {
		t.Errorf("Error: '%s' != 'print this'", strings.TrimSpace(string(out.Bytes())))
		return
	}
}
