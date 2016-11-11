package sh

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/NeowayLabs/nash/sh"
)

var (
	testDir, gopath, nashdPath string
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

	os.Setenv("NASHPATH", "/tmp/.nash")
}

func TestExecuteFile(t *testing.T) {
	testfile := testDir + "/ex1.sh"

	var out bytes.Buffer

	shell, err := NewShell()

	if err != nil {
		t.Error(err)
		return
	}

	shell.SetNashdPath(nashdPath)
	shell.SetStdout(&out)

	err = shell.ExecFile(testfile)

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
	shell, err := NewShell()

	if err != nil {
		t.Error(err)
		return
	}

	shell.SetNashdPath(nashdPath)
	shell.SetStderr(ioutil.Discard)

	err = shell.Exec("command failed", `
        non-existent-program
        `)

	if err == nil {
		t.Error("Must fail")
		return
	}

	// test ignore works
	err = shell.Exec("command failed", `
        -non-existent-program
        `)

	if err != nil {
		t.Errorf("Dash at beginning must ignore errors: ERROR: %s", err.Error())
		return
	}

	var out bytes.Buffer
	shell.SetStderr(os.Stderr)
	shell.SetStdout(&out)

	err = shell.Exec("command failed", `
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

	out.Reset()

	err = shell.Exec("cmd with concat", `echo -n "hello " + "world"`)

	if err != nil {
		t.Error(err)
		return
	}

	if string(out.Bytes()) != "hello world" {
		t.Errorf("Error: '%s' != '%s'", string(out.Bytes()), "hello world")
		return
	}

	out.Reset()

	err = shell.Exec("cmd with concat", `echo -n $notexits`)

	if err == nil {
		t.Errorf("Must fail")
		return
	}
}

func TestExecuteAssignment(t *testing.T) {
	var out bytes.Buffer

	shell, err := NewShell()

	if err != nil {
		t.Error(err)
		return
	}

	shell.SetNashdPath(nashdPath)
	shell.SetStdout(&out)

	err = shell.Exec("assignment", `
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

	err = shell.Exec("list assignment", `
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

	out.Reset()

	shell, err = NewShell()

	if err != nil {
		t.Error(err)
		return
	}

	shell.SetNashdPath(nashdPath)

	err = shell.Exec("assignment", `
        name=i4k
        `)

	if err == nil {
		t.Error("Must fail")
		return
	}

	out.Reset()

	shell.SetStdout(&out)

	err = shell.Exec("list of lists", `l = (
		(name Archlinux)
		(arch amd64)
		(kernel 4.7.1)
	)

	echo $l[0]
	echo $l[1]
	echo -n $l[2]`)

	if err != nil {
		t.Error(err)
		return
	}

	expected := `name Archlinux
arch amd64
kernel 4.7.1`

	if expected != string(out.Bytes()) {
		t.Errorf("expected '%s' but got '%s'", expected, string(out.Bytes()))
		return
	}
}

func TestExecuteCmdAssignment(t *testing.T) {
	var out bytes.Buffer

	shell, err := NewShell()

	if err != nil {
		t.Error(err)
		return
	}

	shell.SetNashdPath(nashdPath)
	shell.SetStdout(&out)

	err = shell.Exec("assignment", `
        name <= echo i4k
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

	err = shell.Exec("list assignment", `
        name <= echo "honda civic"
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

	shell, err = NewShell()

	if err != nil {
		t.Error(err)
		return
	}

	shell.SetNashdPath(nashdPath)

	err = shell.Exec("assignment", `
        name <= ""
        `)

	if err == nil {
		t.Error("Must fail")
		return
	}

	err = shell.Exec("fn must return value", `fn e() {}
v <= e()`)

	if err == nil {
		t.Errorf("Must fail")
		return
	}
}

func TestExecuteCmdAssignmentIFS(t *testing.T) {
	var out bytes.Buffer

	shell, err := NewShell()

	if err != nil {
		t.Error(err)
		return
	}

	shell.SetNashdPath(nashdPath)
	shell.SetStdout(&out)

	err = shell.Exec("assignment", `IFS = (" ")
range <= echo 1 2 3 4 5 6 7 8 9 10

for i in $range {
    echo "i = " + $i
}`)

	if err != nil {
		t.Error(err)
		return
	}

	expected := `i = 1
i = 2
i = 3
i = 4
i = 5
i = 6
i = 7
i = 8
i = 9
i = 10`

	if strings.TrimSpace(string(out.Bytes())) != expected {
		t.Error("assignment not work")
		fmt.Printf("'%s' != '%s'\n", strings.TrimSpace(string(out.Bytes())), expected)

		return
	}

	out.Reset()

	err = shell.Exec("assignment", `IFS = (";")
range <= echo "1;2;3;4;5;6;7;8;9;10"

for i in $range {
    echo "i = " + $i
}`)

	if strings.TrimSpace(string(out.Bytes())) != expected {
		t.Error("assignment not work")
		fmt.Printf("'%s' != '%s'\n", strings.TrimSpace(string(out.Bytes())), expected)

		return
	}

	out.Reset()

	err = shell.Exec("assignment", `IFS = (" " ";")
range <= echo "1;2;3;4;5;6;7;8;9;10"

for i in $range {
    echo "i = " + $i
}`)

	if strings.TrimSpace(string(out.Bytes())) != expected {
		t.Error("assignment not work")
		fmt.Printf("'%s' != '%s'\n", strings.TrimSpace(string(out.Bytes())), expected)

		return
	}

	out.Reset()

	err = shell.Exec("assignment", `IFS = (" " "-")
range <= echo "1;2;3;4;5;6;7;8;9;10"

for i in $range {
    echo "i = " + $i
}`)

	expected = "i = 1;2;3;4;5;6;7;8;9;10"

	if strings.TrimSpace(string(out.Bytes())) != expected {
		t.Error("assignment not work")
		fmt.Printf("'%s' != '%s'\n", strings.TrimSpace(string(out.Bytes())), expected)
		return
	}

}

func TestExecuteRedirection(t *testing.T) {
	shell, err := NewShell()

	if err != nil {
		t.Error(err)
		return
	}

	shell.SetNashdPath(nashdPath)

	path := "/tmp/nashell.test.txt"

	err = shell.Exec("redirect", `
        echo -n "hello world" > `+path+`
        `)

	if err != nil {
		t.Error(err)
		return
	}

	content, err := ioutil.ReadFile(path)

	if err != nil {
		t.Error(err)
		return
	}

	os.Remove(path)

	if string(content) != "hello world" {
		t.Errorf("File differ: '%s' != '%s'", string(content), "hello world")
		return
	}

	// Test redirection to variable
	err = shell.Exec("redirect", `
	location = "`+path+`"
        echo -n "hello world" > $location
        `)

	if err != nil {
		t.Error(err)
		return
	}

	content, err = ioutil.ReadFile(path)

	if err != nil {
		t.Error(err)
		return
	}

	if string(content) != "hello world" {
		t.Errorf("File differ: '%s' != '%s'", string(content), "hello world")
		return
	}

	os.Remove(path)

	// Test redirection to concat
	err = shell.Exec("redirect", `
	location = "`+path+`"
a = ".2"
        echo -n "hello world" > $location+$a
        `)

	if err != nil {
		t.Error(err)
		return
	}

	content, err = ioutil.ReadFile(path + ".2")

	if err != nil {
		t.Error(err)
		return
	}

	if string(content) != "hello world" {
		t.Errorf("File differ: '%s' != '%s'", string(content), "hello world")
		return
	}

	os.Remove(path + ".2")
}

func TestExecuteRedirectionMap(t *testing.T) {
	shell, err := NewShell()

	if err != nil {
		t.Error(err)
		return
	}

	shell.SetNashdPath(nashdPath)

	err = shell.Exec("redirect map", `
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

	shell, err := NewShell()

	if err != nil {
		t.Error(err)
		return
	}

	shell.SetNashdPath(nashdPath)
	shell.SetStdout(&out)

	err = shell.Exec("test cd", `
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

	err = shell.Exec("test cd", `
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

	err = shell.Exec("test cd", `
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

	err = shell.Exec("test error", `
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

	shell, err := NewShell()

	if err != nil {
		t.Error(err)
		return
	}

	shell.SetNashdPath(nashdPath)
	shell.SetStdout(&out)

	err = ioutil.WriteFile("/tmp/test.sh", []byte(`TESTE="teste"`), 0644)

	if err != nil {
		t.Error(err)
		return
	}

	err = shell.Exec("test import", `import /tmp/test.sh
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

func TestExecuteIfEqual(t *testing.T) {
	var out bytes.Buffer

	shell, err := NewShell()

	if err != nil {
		t.Error(err)
		return
	}

	shell.SetNashdPath(nashdPath)
	shell.SetStdout(&out)

	err = shell.Exec("test if equal", `
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

	err = shell.Exec("test if equal 2", `
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

	shell, err := NewShell()

	if err != nil {
		t.Error(err)
		return
	}

	shell.SetNashdPath(nashdPath)
	shell.SetStdout(&out)

	err = shell.Exec("test if else", `
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

	err = shell.Exec("test if equal 2", `
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

	shell, err := NewShell()

	if err != nil {
		t.Error(err)
		return
	}

	shell.SetNashdPath(nashdPath)
	shell.SetStdout(&out)

	err = shell.Exec("test if else", `
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

	err = shell.Exec("test if equal 2", `
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

func TestExecuteFnDecl(t *testing.T) {
	shell, err := NewShell()

	if err != nil {
		t.Error(err)
		return
	}

	shell.SetNashdPath(nashdPath)

	err = shell.Exec("test fnDecl", `
        fn build(image, debug) {
                ls
        }`)

	if err != nil {
		t.Error(err)
		return
	}
}

func TestExecuteFnInv(t *testing.T) {
	shell, err := NewShell()

	if err != nil {
		t.Error(err)
		return
	}

	shell.SetNashdPath(nashdPath)

	var out bytes.Buffer

	shell.SetStdout(&out)

	err = shell.Exec("test fn inv", `
fn getints() {
        return ("1" "2" "3" "4" "5" "6" "7" "8" "9" "0")
}

integers <= getints()
echo -n $integers
`)

	if err != nil {
		t.Error(err)
		return
	}

	if string(out.Bytes()) != "1 2 3 4 5 6 7 8 9 0" {
		t.Errorf("'%s' != '%s'", string(out.Bytes()), "1 2 3 4 5 6 7 8 9 0")
		return
	}

	out.Reset()

	// Test fn scope
	err = shell.Exec("test fn inv", `
OUTSIDE = "some value"

fn getOUTSIDE() {
        return $OUTSIDE
}

val <= getOUTSIDE()
echo -n $val
`)

	if err != nil {
		t.Error(err)
		return
	}

	if string(out.Bytes()) != "some value" {
		t.Errorf("'%s' != '%s'", string(out.Bytes()), "some value")
		return
	}

	err = shell.Exec("test fn inv", `
fn notset() {
        INSIDE = "camshaft"
}

notset()
echo -n $INSIDE
`)

	if err == nil {
		t.Error("Must fail")
		return
	}

	out.Reset()

	// test variables shadow the global ones
	err = shell.Exec("test shadow", `path="AAA"
fn test(path) {
echo -n $path
}
        test("BBB")
`)

	if string(out.Bytes()) != "BBB" {
		t.Errorf("String differs: '%s' != '%s'", string(out.Bytes()), "BBB")
		return
	}

	out.Reset()

	err = shell.Exec("test shadow", `
fn test(path) {
echo -n $path
}

path="AAA"
        test("BBB")
`)

	if string(out.Bytes()) != "BBB" {
		t.Errorf("String differs: '%s' != '%s'", string(out.Bytes()), "BBB")
		return
	}

}

func TestExecuteFnInvOthers(t *testing.T) {
	shell, err := NewShell()

	if err != nil {
		t.Error(err)
		return
	}

	shell.SetNashdPath(nashdPath)

	var out bytes.Buffer

	shell.SetStdout(&out)

	err = shell.Exec("test fn inv", `
fn _getints() {
        return ("1" "2" "3" "4" "5" "6" "7" "8" "9" "0")
}

fn getints() {
        values <= _getints()

        return $values
}

integers <= getints()
echo -n $integers
`)

	if err != nil {
		t.Error(err)
		return
	}

	if string(out.Bytes()) != "1 2 3 4 5 6 7 8 9 0" {
		t.Errorf("'%s' != '%s'", string(out.Bytes()), "1 2 3 4 5 6 7 8 9 0")
		return
	}
}

func TestExecuteBindFn(t *testing.T) {
	var out bytes.Buffer

	shell, err := NewShell()

	if err != nil {
		t.Error(err)
		return
	}

	shell.SetNashdPath(nashdPath)
	shell.SetStdout(&out)

	err = shell.Exec("test bindfn", `
        fn cd(path) {
                echo "override builtin cd"
        }

        bindfn cd cd
        cd`)

	if err != nil {
		t.Error(err)
		return
	}

	if strings.TrimSpace(string(out.Bytes())) != "override builtin cd" {
		t.Errorf("Error: '%s' != 'override builtin cd'", strings.TrimSpace(string(out.Bytes())))
		return
	}
}

func TestExecutePipe(t *testing.T) {
	var out bytes.Buffer

	path := os.Getenv("PATH")
	path = "/bin:/sbin:/usr/bin:/usr/sbin:/usr/local/bin:/usr/local/sbin:" + path

	os.Setenv("PATH", path)

	shell, err := NewShell()

	if err != nil {
		t.Error(err)
		return
	}

	shell.SetNashdPath(nashdPath)
	shell.SetStdout(&out)

	err = shell.Exec("test pipe", `echo hello | wc -l`)

	if err != nil {
		t.Error(err)
		return
	}

	strOut := strings.TrimSpace(string(out.Bytes()))

	if strOut != "1" {
		t.Errorf("Expected '1' but found '%s'", strOut)
		return
	}

	out.Reset()

	err = shell.Exec("test pipe 3", `echo hello | wc -l | grep 1`)

	if err != nil {
		t.Error(err)
		return
	}

	strOut = strings.TrimSpace(string(out.Bytes()))

	if strOut != "1" {
		t.Errorf("Expected '1' but found '%s'", strOut)
		return
	}

	out.Reset()
}

func testTCPRedirection(t *testing.T, port, command string) {
	message := "hello world"

	done := make(chan bool)

	l, err := net.Listen("tcp", port)

	if err != nil {
		t.Fatal(err)
	}

	go func() {
		shell, err := NewShell()

		if err != nil {
			t.Error(err)
			return
		}

		shell.SetNashdPath(nashdPath)

		<-done

		err = shell.Exec("test net redirection", command)

		if err != nil {
			t.Error(err)
			return
		}
	}()

	defer l.Close()

	for {
		done <- true
		conn, err := l.Accept()
		if err != nil {
			return
		}

		defer conn.Close()

		buf, err := ioutil.ReadAll(conn)

		if err != nil {
			t.Fatal(err)
		}

		fmt.Println(string(buf[:]))

		if msg := string(buf[:]); msg != message {
			t.Fatalf("Unexpected message:\nGot:\t\t%s\nExpected:\t%s\n", msg, message)
		}

		return // Done
	}
}

func TestTCPRedirection(t *testing.T) {
	testTCPRedirection(t, ":6666", `echo -n "hello world" >[1] "tcp://localhost:6666"`)
	testTCPRedirection(t, ":6667", `echo -n "hello world" > "tcp://localhost:6667"`)
}

func TestExecuteUnixRedirection(t *testing.T) {
	message := "hello world"

	sockDir, err := ioutil.TempDir("/tmp", "nash-tests")

	if err != nil {
		t.Error(err)
		return
	}

	sockFile := sockDir + "/listen.sock"

	defer func() {
		os.Remove(sockFile)
		os.RemoveAll(sockDir)
	}()

	done := make(chan bool)
	writeDone := make(chan bool)

	go func() {
		defer func() {
			writeDone <- true
		}()

		shell, err := NewShell()

		if err != nil {
			t.Error(err)
			return
		}

		shell.SetNashdPath(nashdPath)

		<-done

		err = shell.Exec("test net redirection", `echo -n "`+message+`" >[1] "unix://`+sockFile+`"`)

		if err != nil {
			t.Error(err)
			return
		}
	}()

	l, err := net.Listen("unix", sockFile)

	if err != nil {
		t.Error(err)
		return
	}

	defer l.Close()

	go func() {
		conn, err := l.Accept()
		if err != nil {
			return
		}

		defer conn.Close()

		buf, err := ioutil.ReadAll(conn)

		if err != nil {
			t.Fatal(err)
		}

		fmt.Println(string(buf[:]))

		if msg := string(buf[:]); msg != message {
			t.Fatalf("Unexpected message:\nGot:\t\t%s\nExpected:\t%s\n", msg, message)
		}

		return // Done
	}()

	done <- true
	<-writeDone
}

func TestExecuteUDPRedirection(t *testing.T) {
	message := "hello world"

	done := make(chan bool)
	writeDone := make(chan bool)

	go func() {
		defer func() {
			writeDone <- true
		}()

		shell, err := NewShell()

		if err != nil {
			t.Error(err)
			return
		}

		shell.SetNashdPath(nashdPath)

		<-done

		err = shell.Exec("test net redirection", `echo -n "`+message+`" >[1] "udp://localhost:6667"`)

		if err != nil {
			t.Error(err)
			return
		}
	}()

	serverAddr, err := net.ResolveUDPAddr("udp", ":6667")

	if err != nil {
		t.Error(err)
		return
	}

	l, err := net.ListenUDP("udp", serverAddr)

	if err != nil {
		t.Fatal(err)
	}

	go func() {
		defer l.Close()

		buf := make([]byte, 1024)

		nb, _, err := l.ReadFromUDP(buf)

		if err != nil {
			t.Error(err)
			return
		}

		received := string(buf[:nb])

		if received != message {
			t.Errorf("Unexpected message:\nGot:\t\t'%s'\nExpected:\t'%s'\n", received, message)
		}
	}()

	time.Sleep(time.Second * 1)

	done <- true
	<-writeDone
}

func TestExecuteReturn(t *testing.T) {
	shell, err := NewShell()

	if err != nil {
		t.Error(err)
		return
	}

	shell.SetNashdPath(nashdPath)

	err = shell.Exec("test return fail", "return")

	if err == nil {
		t.Errorf("Must fail. Return is only valid inside function")
		return
	}

	err = shell.Exec("test return", `fn test() { return }
test()`)

	if err != nil {
		t.Error(err)
		return
	}

	var out bytes.Buffer

	shell.SetStdout(&out)

	err = shell.Exec("test return", `fn test() {
	if "1" == "1" {
		return "1"
	}

	return "0"
}

res <= test()
echo -n $res`)

	if err != nil {
		t.Error(err)
		return
	}

	got := string(out.Bytes())

	if got != "1" {
		t.Errorf("Expected '1' but got '%s'", got)
		return
	}

	out.Reset()

	err = shell.Exec("ret from for", `fn test() {
	values = (0 1 2 3 4 5 6 7 8 9)

	for i in $values {
		if $i == "5" {
			return $i
		}
	}

	return "0"
}
a <= test()
echo -n $a`)

	if err != nil {
		t.Error(err)
		return
	}

	got = string(out.Bytes())

	if "5" != got {
		t.Errorf("Expected '5' but got '%s'", got)
		return
	}

	out.Reset()

	err = shell.Exec("inf loop ret", `fn test() {
	for {
		if "1" == "1" {
			return "1"
		}
	}

	# never happen
	return "bleh"
}
a <= test()
echo -n $a`)

	if err != nil {
		t.Error(err)
		return
	}

	got = string(out.Bytes())

	if got != "1" {
		t.Errorf("Expected '1' but got '%s'", got)
		return
	}
}

func TestExecuteFnAsFirstClass(t *testing.T) {
	shell, err := NewShell()

	if err != nil {
		t.Error(err)
		return
	}

	shell.SetNashdPath(nashdPath)

	var out bytes.Buffer

	shell.SetStdout(&out)

	err = shell.Exec("test fn by arg", `
        fn printer(val) {
                echo -n $val
        }

        fn success(print, val) {
                $print("[SUCCESS] " + $val)
        }

        success($printer, "Command executed!")
        `)

	if err != nil {
		t.Error(err)
		return
	}

	expected := `[SUCCESS] Command executed!`

	if expected != string(out.Bytes()) {
		t.Errorf("Differs: '%s' != '%s'", expected, string(out.Bytes()))
		return
	}
}

func TestExecuteDump(t *testing.T) {
	shell, err := NewShell()

	if err != nil {
		t.Error(err)
		return
	}

	shell.SetNashdPath(nashdPath)
	shell.Reset()

	var out bytes.Buffer

	shell.SetStdout(&out)

	err = shell.Exec("exec dump", "dump")

	if err != nil {
		t.Error(err)
		return
	}

	if string(out.Bytes()) != "" {
		t.Errorf("Must be empty. Shell was reset'ed, but returns '%s'", string(out.Bytes()))
		return
	}

	err = shell.Exec("", `TEST = "some value"`)

	if err != nil {
		t.Error(err)
		return
	}

	out.Reset()

	err = shell.Exec("", "dump")

	if err != nil {
		t.Error(err)
		return
	}

	expected := `TEST = "some value"`

	if strings.TrimSpace(string(out.Bytes())) != expected {
		t.Errorf("'%s' != '%s'", string(out.Bytes()), expected)
		return
	}

	tempDir, err := ioutil.TempDir("/tmp", "nash-test")

	if err != nil {
		t.Error(err)
		return
	}

	dumpFile := tempDir + "/dump.test"

	defer func() {
		os.Remove(dumpFile)
		os.RemoveAll(tempDir)
	}()

	out.Reset()

	//	shell.SetStdout(os.Stdout)

	err = shell.Exec("", "dump "+dumpFile)

	if err != nil {
		t.Error(err)
		return
	}

	if string(out.Bytes()) != "" {
		t.Error("Must be empty")
		return
	}

	content, err := ioutil.ReadFile(dumpFile)

	if err != nil {
		t.Error(err)
		return
	}

	if strings.TrimSpace(string(content)) != expected {
		t.Errorf("Must be equal. '%s' != '%s'", strings.TrimSpace(string(content)), expected)
		return
	}
}

func TestExecuteDumpVariable(t *testing.T) {
	shell, err := NewShell()

	if err != nil {
		t.Error(err)
		return
	}

	shell.SetNashdPath(nashdPath)
	shell.Reset()

	var out bytes.Buffer

	shell.SetStdout(&out)

	tempDir, err := ioutil.TempDir("/tmp", "nash-test")

	if err != nil {
		t.Error(err)
		return
	}

	dumpFile := tempDir + "/dump.test"

	defer func() {
		os.Remove(dumpFile)
		os.RemoveAll(tempDir)
	}()

	err = shell.Exec("", `dumpFile = "`+dumpFile+`"`)

	if err != nil {
		t.Error(err)
		return
	}

	out.Reset()

	//	shell.SetStdout(os.Stdout)

	err = shell.Exec("", `dump $dumpFile`)

	if err != nil {
		t.Error(err)
		return
	}

	if string(out.Bytes()) != "" {
		t.Error("Must be empty")
		return
	}

	content, err := ioutil.ReadFile(dumpFile)

	if err != nil {
		t.Error(err)
		return
	}

	expected := `dumpFile = "` + dumpFile + `"`

	if strings.TrimSpace(string(content)) != expected {
		t.Errorf("Must be equal. '%s' != '%s'", strings.TrimSpace(string(content)), expected)
		return
	}
}

func TestExecuteConcat(t *testing.T) {
	shell, err := NewShell()

	if err != nil {
		t.Error(err)
		return
	}

	shell.SetNashdPath(nashdPath)
	shell.Reset()

	var out bytes.Buffer

	shell.SetStdout(&out)

	err = shell.Exec("", `a = "A"
b = "B"
c = $a + $b + "C"
echo -n $c`)

	if err != nil {
		t.Error(err)
		return
	}

	if string(out.Bytes()) != "ABC" {
		t.Errorf("Must be equal. '%s' != '%s'", string(out.Bytes()), "ABC")
		return
	}

	out.Reset()

	err = shell.Exec("concat indexed var", `tag = (Name some)
	echo -n "Key="+$tag[0]+",Value="+$tag[1]`)

	if err != nil {
		t.Error(err)
		return
	}

	expected := "Key=Name,Value=some"

	if expected != string(out.Bytes()) {
		t.Errorf("String differs: '%s' != '%s'", expected, string(out.Bytes()))
		return
	}
}

func TestExecuteFor(t *testing.T) {
	shell, err := NewShell()

	if err != nil {
		t.Error(err)
		return
	}

	shell.SetNashdPath(nashdPath)

	var out bytes.Buffer

	shell.SetStdout(&out)

	err = shell.Exec("simple loop", `files = (/etc/passwd /etc/shells)
for f in $files {
        echo $f
        echo "loop"
}`)

	if err != nil {
		t.Error(err)
		return
	}

	expected := `/etc/passwd
loop
/etc/shells
loop`
	value := strings.TrimSpace(string(out.Bytes()))

	if value != expected {
		t.Errorf("String differs: '%s' != '%s'", expected, value)
		return
	}

}

func TestExecuteInfiniteLoop(t *testing.T) {
	shell, err := NewShell()

	if err != nil {
		t.Error(err)
		return
	}

	shell.SetNashdPath(nashdPath)

	doneCtrlc := make(chan bool)
	doneLoop := make(chan bool)

	go func() {
		fmt.Printf("Waiting 2 second to abort infinite loop")
		time.Sleep(2 * time.Second)

		shell.TriggerCTRLC()
		doneCtrlc <- true
	}()

	go func() {
		err = shell.Exec("simple loop", `for {
        echo "infinite loop" >[1=]
}`)
		doneLoop <- true

		if err == nil {
			t.Errorf("Must fail with interrupted error")
			return
		}

		type interrupted interface {
			Interrupted() bool
		}

		if errInterrupted, ok := err.(interrupted); !ok || !errInterrupted.Interrupted() {
			t.Errorf("Loop not interrupted properly")
			return
		}
	}()

	for i := 0; i < 2; i++ {
		select {
		case <-doneCtrlc:
			fmt.Printf("CTRL-C Sent to subshell\n")
		case <-doneLoop:
			fmt.Printf("Loop finished.\n")
		case <-time.After(5 * time.Second):
			t.Errorf("Failed to stop infinite loop")
			return
		}
	}
}

func TestExecuteVariableIndexing(t *testing.T) {
	shell, err := NewShell()

	if err != nil {
		t.Error(err)
		return
	}

	var out bytes.Buffer

	shell.SetNashdPath(nashdPath)
	shell.SetStdout(&out)

	err = shell.Exec("indexing", `list = ("1" "2" "3")
        echo -n $list[0]`)

	if err != nil {
		t.Error(err)
		return
	}

	result := strings.TrimSpace(string(out.Bytes()))
	expected := "1"

	if expected != result {
		t.Errorf("Fail: '%s' != '%s'", expected, result)
		return
	}

	out.Reset()

	err = shell.Exec("indexing", `i = "0"
echo -n $list[$i]`)

	if err != nil {
		t.Error(err)
		return
	}

	result = strings.TrimSpace(string(out.Bytes()))
	expected = "1"

	if expected != result {
		t.Errorf("Fail: '%s' != '%s'", expected, result)
		return
	}

	out.Reset()

	err = shell.Exec("indexing", `IFS = ("\n")
seq <= seq 0 2

for i in $seq {
    echo -n $list[$i]
}`)

	if err != nil {
		t.Error(err)
		return
	}

	result = strings.TrimSpace(string(out.Bytes()))
	expected = "123"

	if expected != result {
		t.Errorf("Fail: '%s' != '%s'", expected, result)
		return
	}

	out.Reset()

	err = shell.Exec("indexing", `echo -n $list[5]`)

	if err == nil {
		t.Error("Must fail. Out of bounds")
		return
	}

	out.Reset()

	err = shell.Exec("indexing", `a = ("0")
echo -n $list[$a[0]]`)

	if err != nil {
		t.Error(err)
		return
	}

	result = strings.TrimSpace(string(out.Bytes()))
	expected = "1"

	if expected != result {
		t.Errorf("Fail: '%s' != '%s'", expected, result)
		return
	}
}

func TestExecuteSubShellDoesNotOverwriteparentEnv(t *testing.T) {
	shell, err := NewShell()

	if err != nil {
		t.Error(err)
		return
	}

	err = shell.Exec("set env", `SHELL = "bleh"
setenv SHELL`)

	if err != nil {
		t.Error(err)
		return
	}

	var out bytes.Buffer
	shell.SetStdout(&out)

	err = shell.Exec("set env from fn", `fn test() {
}

test()

echo -n $SHELL`)

	if err != nil {
		t.Error(err)
		return
	}

	if string(out.Bytes()) != "bleh" {
		t.Errorf("Differ: '%s' != '%s'", "bleh", string(out.Bytes()))
		return
	}
}

func TestExecuteInterruptDoesNotCancelLoop(t *testing.T) {
	shell, err := NewShell()

	if err != nil {
		t.Error(err)
		return
	}

	shell.TriggerCTRLC()

	time.Sleep(time.Second * 1)

	err = shell.Exec("interrupting loop", `seq = (1 2 3 4 5)
for i in $seq {}`)

	if err != nil {
		t.Error(err)
		return
	}
}

func TestExecuteErrorSuppressionAll(t *testing.T) {
	shell, err := NewShell()

	if err != nil {
		t.Error(err)
		return
	}

	err = shell.Exec("-input-", `-command-not-exists`)

	if err != nil {
		t.Errorf("Expected to not fail...: %s", err.Error())
		return
	}

	scode, ok := shell.Getvar("status")

	if !ok || scode.Type() != sh.StringType || scode.String() != strconv.Itoa(ENotFound) {
		t.Errorf("Invalid status code %s", scode.String())
		return
	}

	err = shell.Exec("-input-", `echo works >[1=]`)

	if err != nil {
		t.Error(err)
		return
	}

	scode, ok = shell.Getvar("status")

	if !ok || scode.Type() != sh.StringType || scode.String() != "0" {
		t.Errorf("Invalid status code %s", scode)
		return
	}

	err = shell.Exec("-input-", `echo works | cmd-does-not-exists`)

	if err == nil {
		t.Errorf("Must fail")
		return
	}

	if err.Error() != "not started|exec: \"cmd-does-not-exists\": executable file not found in $PATH" {
		t.Errorf("Unexpected error: %s", err.Error())
		return
	}

	scode, ok = shell.Getvar("status")

	if !ok || scode.Type() != sh.StringType || scode.String() != "255|127" {
		t.Errorf("Invalid status code %s", scode)
		return
	}
}

func TestExecuteGracefullyError(t *testing.T) {
	shell, err := NewShell()

	if err != nil {
		t.Error(err)
		return
	}

	err = shell.Exec("someinput.sh", "(")

	if err == nil {
		t.Errorf("Must fail...")
		return
	}

	expectErr := "someinput.sh:1:1: Multi-line command not finished. Found EOF but expect ')'"

	if err.Error() != expectErr {
		t.Errorf("Expect error: %s, but got: %s", expectErr, err.Error())
		return
	}

	err = shell.Exec("input", "echo(")

	if err == nil {
		t.Errorf("Must fail...")
		return
	}

	if err.Error() != "input:1:5: Unexpected token EOF. Expecting STRING, VARIABLE or )" {
		t.Errorf("Unexpected error: %s", err.Error())
		return
	}

}

func TestExecuteMultilineCmd(t *testing.T) {
	shell, err := NewShell()

	if err != nil {
		t.Error(err)
		return
	}

	var out bytes.Buffer

	shell.SetStdout(&out)

	err = shell.Exec("test", `(echo -n
		hello
		world)`)

	if err != nil {
		t.Error(err)
		return
	}

	expected := "hello world"

	if expected != string(out.Bytes()) {
		t.Errorf("Expected '%s' but got '%s'", expected, string(out.Bytes()))
		return
	}

	out.Reset()

	err = shell.Exec("test", `(
                echo -n 1 2 3 4 5 6 7 8 9 10
                        11 12 13 14 15 16 17 18 19 20
                )`)

	if err != nil {
		t.Error(err)
		return
	}

	expected = "1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 19 20"

	if expected != string(out.Bytes()) {
		t.Errorf("Expected '%s' but got '%s'", expected, string(out.Bytes()))
		return
	}
}

func TestExecuteMultilineCmdAssign(t *testing.T) {
	shell, err := NewShell()

	if err != nil {
		t.Error(err)
		return
	}

	var out bytes.Buffer

	shell.SetStdout(&out)

	err = shell.Exec("test", `val <= (echo -n
		hello
		world)

	echo -n $val`)

	if err != nil {
		t.Error(err)
		return
	}

	expected := "hello world"

	if expected != string(out.Bytes()) {
		t.Errorf("Expected '%s' but got '%s'", expected, string(out.Bytes()))
		return
	}

	out.Reset()

	err = shell.Exec("test", `val <= (
                echo -n 1 2 3 4 5 6 7 8 9 10
                        11 12 13 14 15 16 17 18 19 20
                )
		echo -n $val`)

	if err != nil {
		t.Error(err)
		return
	}

	expected = "1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 19 20"

	if expected != string(out.Bytes()) {
		t.Errorf("Expected '%s' but got '%s'", expected, string(out.Bytes()))
		return
	}
}

func TestExecuteMuliReturnUnfinished(t *testing.T) {
	shell, err := NewShell()

	if err != nil {
		t.Error(err)
		return
	}

	err = shell.Exec("test", "(")

	if err == nil {
		t.Errorf("Must fail... Must return an unfinished paren error")
		return
	}

	type unfinished interface {
		Unfinished() bool
	}

	if e, ok := err.(unfinished); !ok || !e.Unfinished() {
		t.Errorf("Must fail with unfinished paren error. Got %s", err.Error())
		return
	}

	err = shell.Exec("test", `(
echo`)

	if err == nil {
		t.Errorf("Must fail... Must return an unfinished paren error")
		return
	}

	if e, ok := err.(unfinished); !ok || !e.Unfinished() {
		t.Errorf("Must fail with unfinished paren error. Got %s", err.Error())
		return
	}

	err = shell.Exec("test", `(
echo hello
world`)

	if err == nil {
		t.Errorf("Must fail... Must return an unfinished paren error")
		return
	}

	if e, ok := err.(unfinished); !ok || !e.Unfinished() {
		t.Errorf("Must fail with unfinished paren error. Got %s", err.Error())
		return
	}
}
