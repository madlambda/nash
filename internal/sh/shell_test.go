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

type execTest struct {
	desc           string
	execStr        string
	expectedStdout string
	expectedStderr string
	expectedErr    string
}

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

func testExecuteFile(t *testing.T, path, expected string) {
	var out bytes.Buffer

	shell, err := NewShell()

	if err != nil {
		t.Error(err)
		return
	}

	shell.SetNashdPath(nashdPath)
	shell.SetStdout(&out)

	err = shell.ExecFile(path)

	if err != nil {
		t.Error(err)
		return
	}

	if string(out.Bytes()) != expected {
		t.Errorf("Wrong command output: '%s' != '%s'",
			string(out.Bytes()), expected)
		return
	}
}

func testShellExec(t *testing.T, shell *Shell, desc, execStr, expectedStdout, expectedStderr, expectedErr string) {

	var bout bytes.Buffer
	var berr bytes.Buffer
	shell.SetStderr(&berr)
	shell.SetStdout(&bout)

	err := shell.Exec(desc, execStr)

	if err != nil {
		if err.Error() != expectedErr {
			t.Errorf("Error differs: Expected '%s' but got '%s'",
				expectedErr, err.Error())
		}
	}

	if expectedStdout != string(bout.Bytes()) {
		t.Errorf("Stdout differs: '%s' != '%s'", expectedStdout,
			string(bout.Bytes()))
		return
	}

	if expectedStderr != string(berr.Bytes()) {
		t.Errorf("Stderr differs: '%s' != '%s'", expectedStderr,
			string(berr.Bytes()))
		return
	}
}

func testExec(t *testing.T, desc, execStr, expectedStdout, expectedStderr, expectedErr string) {
	shell, err := NewShell()

	if err != nil {
		t.Error(err)
		return
	}

	shell.SetNashdPath(nashdPath)

	testShellExec(t, shell, desc, execStr, expectedStdout, expectedStderr, expectedErr)
}

func testInteractiveExec(t *testing.T, desc, execStr, expectedStdout, expectedStderr, expectedErr string) {
	shell, err := NewShell()

	if err != nil {
		t.Error(err)
		return
	}

	shell.SetNashdPath(nashdPath)
	shell.SetInteractive(true)

	testShellExec(t, shell, desc, execStr, expectedStdout, expectedStderr, expectedErr)
}

func TestInitEnv(t *testing.T) {

	os.Setenv("TEST", "abc=123=")

	shell, err := NewShell()

	if err != nil {
		t.Error(err)
		return
	}

	testEnv, _ := shell.Getenv("TEST")
	expectedTestEnv := "abc=123="

	if testEnv.String() != expectedTestEnv {
		t.Errorf("Expected TEST Env differs: '%s' != '%s'", testEnv, expectedTestEnv)
		return
	}
}

func TestExecuteFile(t *testing.T) {
	type fileTests struct {
		path     string
		expected string
	}

	for _, ftest := range []fileTests{
		{path: "/ex1.sh", expected: "hello world\n"},
	} {
		testExecuteFile(t, testDir+ftest.path, ftest.expected)
	}
}

func TestExecuteCommand(t *testing.T) {
	for _, test := range []execTest{
		{
			"command failed",
			`non-existing-program`,
			"", "",
			`exec: "non-existing-program": executable file not found in $PATH`,
		},
		{
			"err ignored",
			`-non-existing-program`,
			"", "", "",
		},
		{
			"hello world",
			"echo -n hello world",
			"hello world", "", "",
		},
		{
			"cmd with concat",
			`echo -n "hello " + "world"`,
			"hello world", "", "",
		},
	} {
		testExec(t,
			test.desc,
			test.execStr,
			test.expectedStdout,
			test.expectedStderr,
			test.expectedErr,
		)
	}
}

func TestExecuteAssignment(t *testing.T) {
	for _, test := range []execTest{
		{ // wrong assignment
			"wrong assignment",
			`name=i4k`,
			"", "",
			"wrong assignment:1:5: Unexpected token IDENT. Expecting VARIABLE or STRING or (",
		},
		{
			"assignment",
			`name="i4k"
                         echo $name`,
			"i4k\n", "",
			"",
		},
		{
			"list assignment",
			`name=(honda civic)
                         echo -n $name`,
			"honda civic", "",
			"",
		},
		{
			"list of lists",
			`l = (
		(name Archlinux)
		(arch amd64)
		(kernel 4.7.1)
	)

	echo $l[0]
	echo $l[1]
	echo -n $l[2]`,
			`name Archlinux
arch amd64
kernel 4.7.1`,
			"",
			"",
		},
	} {
		testExec(t,
			test.desc,
			test.execStr,
			test.expectedStdout,
			test.expectedStderr,
			test.expectedErr,
		)
	}
}

func TestExecuteCmdAssignment(t *testing.T) {
	for _, test := range []execTest{
		{
			"cmd assignment",
			`name <= echo -n i4k
                         echo -n $name`,
			"i4k", "",
			"",
		},
		{
			"list cmd assignment",
			`name <= echo "honda civic"
                         echo -n $name`,
			"honda civic", "", "",
		},
		{
			"wrong cmd assignment",
			`name <= ""`,
			"", "", "wrong cmd assignment:1:9: Invalid token STRING. Expected command or function invocation",
		},
		{
			"fn must return value",
			`fn e() {}
                         v <= e()`,
			"",
			"",
			"<interactive>:2:25: Invalid assignment from function that does not return values: e()",
		},
	} {
		testExec(t,
			test.desc,
			test.execStr,
			test.expectedStdout,
			test.expectedStderr,
			test.expectedErr,
		)
	}
}

func TestExecuteCmdAssignmentIFS(t *testing.T) {
	for _, test := range []execTest{
		{
			"ifs",
			`IFS = (" ")
range <= echo 1 2 3 4 5 6 7 8 9 10

for i in $range {
    echo "i = " + $i
}`,
			"", "",
			"<interactive>:4:0: Invalid variable type in for range: StringType",
		},
		{
			"ifs",
			`IFS = (";")
range <= echo "1;2;3;4;5;6;7;8;9;10"

for i in $range {
    echo "i = " + $i
}`,
			"", "",
			"<interactive>:4:0: Invalid variable type in for range: StringType",
		},
		{
			"ifs",
			`IFS = (" " ";")
range <= echo "1;2;3;4;5;6 7;8;9;10"

for i in $range {
    echo "i = " + $i
}`,
			"", "",
			"<interactive>:4:0: Invalid variable type in for range: StringType",
		},
		{
			"ifs",
			`IFS = (" " "-")
range <= echo "1;2;3;4;5;6;7-8;9;10"

for i in $range {
    echo "i = " + $i
}`,
			"", "",
			"<interactive>:4:0: Invalid variable type in for range: StringType",
		},
	} {
		testExec(t,
			test.desc,
			test.execStr,
			test.expectedStdout,
			test.expectedStderr,
			test.expectedErr,
		)
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

func TestExecuteSetenv(t *testing.T) {
	for _, test := range []execTest{
		{
			"test setenv basic",
			`test = "hello"
                         setenv test
                         ` + nashdPath + ` -c "echo $test"`,
			"hello\n", "", "",
		},
		{
			"test setenv assignment",
			`setenv test = "hello"
                         ` + nashdPath + ` -c "echo $test"`,
			"hello\n", "", "",
		},
		{
			"test setenv exec cmd",
			`setenv test <= echo -n "hello"
                         ` + nashdPath + ` -c "echo $test"`,
			"hello\n", "", "",
		},
		{
			"test setenv semicolon",
			`setenv a setenv b`,
			"", "",
			"test setenv semicolon:1:9: Unexpected token setenv, expected semicolon (;) or EOL",
		},
	} {
		testExec(t,
			test.desc,
			test.execStr,
			test.expectedStdout,
			test.expectedStderr,
			test.expectedErr,
		)
	}
}

func TestExecuteCd(t *testing.T) {
	for _, test := range []execTest{
		{
			"test cd",
			`cd /
        pwd`,
			"/\n", "", "",
		},
		{
			"test cd",
			`HOME="/"
        setenv HOME
        cd
        pwd`,
			"/\n",
			"", "",
		},
		{
			"test cd into $var",
			`
        var="/"
        cd $var
        pwd`,
			"/\n",
			"",
			"",
		},
		{
			"test error",
			`var=("val1" "val2" "val3")
        cd $var
        pwd`,
			"", "",
			"<interactive>:2:12: lvalue is not comparable: (val1 val2 val3) -> ListType.",
		},
	} {
		testInteractiveExec(t,
			test.desc,
			test.execStr,
			test.expectedStdout,
			test.expectedStderr,
			test.expectedErr,
		)
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

func TestFnComposition(t *testing.T) {
	for _, test := range []execTest{
		{
			"composition",
			`
                fn a(b) { echo -n $b }
                fn b()  { return "hello" }
                a(b())
        `,
			"hello", "", "",
		},
		{
			"composition",
			`
                fn a(b, c) { echo -n $b $c  }
                fn b()     { return "hello" }
                fn c()     { return "world" }
                a(b(), c())
        `,
			"hello world", "", "",
		},
	} {
		testExec(t, test.desc,
			test.execStr,
			test.expectedStdout,
			test.expectedStderr,
			test.expectedErr,
		)
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

func TestNonInteractive(t *testing.T) {
	shell, err := NewShell()

	if err != nil {
		t.Error(err)
		return
	}

	shell.SetNashdPath(nashdPath)
	shell.SetInteractive(true)

	testShellExec(t, shell,
		"test bindfn interactive",
		`
        fn greeting() {
                echo "Hello"
        }

        bindfn greeting hello`,
		"", "", "")

	shell.SetInteractive(false)
	shell.filename = "<non-interactive>"

	expectedErr := "<non-interactive>:1:0: "+
		"'hello' is a bind to 'greeting'."+
		" No binds allowed in non-interactive mode."

	testShellExec(t, shell, "test 'binded' function non-interactive",
		`hello`, "", "", expectedErr)

	expectedErr = "<non-interactive>:6:8: 'bindfn' is not allowed in"+
		" non-interactive mode."

	testShellExec(t, shell, "test bindfn non-interactive",
	`
        fn goodbye() {
                echo "Ciao"
        }

        bindfn goodbye ciao`, "", "", expectedErr)
}

func TestExecuteBindFn(t *testing.T) {
	for _, test := range []execTest{
		{
			"test bindfn",
			`
        fn cd(path) {
                echo "override builtin cd"
        }

        bindfn cd cd
        cd`,
			"override builtin cd\n", "", "",
		},
		{
			"test bindfn args",
			`
        fn foo(line) {
                echo $line
        }

        bindfn foo bar
        bar test test`,
			"", "", "<interactive>:7:8: Too much arguments for function 'foo'. It expects 1 args, but given 2. Arguments: [\"test\" \"test\"]",
		},
	} {
		testInteractiveExec(t,
			test.desc,
			test.execStr,
			test.expectedStdout,
			test.expectedStderr,
			test.expectedErr,
		)
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
	for _, test := range []execTest{
		{
			"return invalid",
			`return`,
			"", "",
			"<interactive>:1:0: Unexpected return outside of function declaration.",
		},
		{
			"test simple return",
			`fn test() { return }
test()`,
			"", "", "",
		},
		{
			"return must finish func evaluation",
			`fn test() {
	if "1" == "1" {
		return "1"
	}

	return "0"
}

res <= test()
echo -n $res`,
			"1", "", "",
		},
		{
			"ret from for",
			`fn test() {
	values = (0 1 2 3 4 5 6 7 8 9)

	for i in $values {
		if $i == "5" {
			return $i
		}
	}

	return "0"
}
a <= test()
echo -n $a`,
			"5", "", "",
		},
		{
			"inf loop ret",
			`fn test() {
	for {
		if "1" == "1" {
			return "1"
		}
	}

	# never happen
	return "bleh"
}
a <= test()
echo -n $a`,
			"1", "", "",
		},
		{
			"test returning funcall",
			`fn a() { return "1" }
                         fn b() { return a() }
                         c <= b()
                         echo -n $c`,
			"1", "", "",
		},
	} {
		testExec(t,
			test.desc,
			test.execStr,
			test.expectedStdout,
			test.expectedStderr,
			test.expectedErr,
		)
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

	err = shell.Exec("indexing", `tmp <= seq 0 2
seq <= split($tmp, "\n")

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

	err = shell.Exec("set env", `setenv SHELL = "bleh"`)

	if err != nil {
		t.Error(err)
		return
	}

	var out bytes.Buffer
	shell.SetStdout(&out)

	err = shell.Exec("set env from fn", `fn test() {
        # test() should not call the setup func in Nash
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

	expectedError := `<interactive>:1:11: not started|exec: "cmd-does-not-exists": executable file not found in $PATH`

	if err.Error() != expectedError {
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

func TestExecuteMultiReturnUnfinished(t *testing.T) {
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
