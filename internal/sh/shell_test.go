package sh

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/NeowayLabs/nash/sh"
)

type execTestCase struct {
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

func testShellExec(t *testing.T, shell *Shell, testcase execTestCase) {
	var bout bytes.Buffer
	var berr bytes.Buffer
	shell.SetStderr(&berr)
	shell.SetStdout(&bout)

	err := shell.Exec(testcase.desc, testcase.execStr)

	if err != nil {
		if err.Error() != testcase.expectedErr {
			t.Errorf("[%s] Error differs: Expected '%s' but got '%s'",
				testcase.desc,
				testcase.expectedErr,
				err.Error())
		}
	}

	if testcase.expectedStdout != string(bout.Bytes()) {
		t.Errorf("[%s] Stdout differs: '%s' != '%s'",
			testcase.desc,
			testcase.expectedStdout,
			string(bout.Bytes()))
		return
	}

	if testcase.expectedStderr != string(berr.Bytes()) {
		t.Errorf("[%s] Stderr differs: '%s' != '%s'",
			testcase.desc,
			testcase.expectedStderr,
			string(berr.Bytes()))
		return
	}
	bout.Reset()
	berr.Reset()
}

func testExec(t *testing.T, testcase execTestCase) {
	shell, err := NewShell()

	if err != nil {
		t.Error(err)
		return
	}

	shell.SetNashdPath(nashdPath)

	testShellExec(t, shell, testcase)
}

func testInteractiveExec(t *testing.T, testcase execTestCase) {
	shell, err := NewShell()

	if err != nil {
		t.Error(err)
		return
	}

	shell.SetNashdPath(nashdPath)
	shell.SetInteractive(true)

	testShellExec(t, shell, testcase)
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
	for _, test := range []execTestCase{
		{
			desc:           "command failed",
			execStr:        `non-existing-program`,
			expectedStdout: "",
			expectedStderr: "",
			expectedErr:    `exec: "non-existing-program": executable file not found in $PATH`,
		},
		{
			desc:           "err ignored",
			execStr:        `-non-existing-program`,
			expectedStdout: "",
			expectedStderr: "",
			expectedErr:    "",
		},
		{
			desc:           "hello world",
			execStr:        "echo -n hello world",
			expectedStdout: "hello world",
			expectedStderr: "",
			expectedErr:    "",
		},
		{
			desc:           "cmd with concat",
			execStr:        `echo -n "hello " + "world"`,
			expectedStdout: "hello world",
			expectedStderr: "",
			expectedErr:    "",
		},
		{
			desc: "local command",
			execStr: `echopath <= which echo
path <= dirname $echopath
chdir($path)
./echo -n hello`,
			expectedStdout: "hello",
			expectedStderr: "",
			expectedErr:    "",
		},
	} {
		testExec(t, test)
	}
}

func TestExecuteAssignment(t *testing.T) {
	for _, test := range []execTestCase{
		{ // wrong assignment
			"wrong assignment",
			`var name=i4k`,
			"", "",
			"wrong assignment:1:9: Unexpected token IDENT. Expecting VARIABLE, STRING or (",
		},
		{
			"assignment",
			`var name="i4k"
                         echo $name`,
			"i4k\n", "",
			"",
		},
		{
			"list assignment",
			`var name=(honda civic)
                         echo -n $name`,
			"honda civic", "",
			"",
		},
		{
			"list of lists",
			`var l = (
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
		{
			"list assignment",
			`var l = (0 1 2 3)
                         l[0] = "666"
                         echo -n $l`,
			`666 1 2 3`,
			"",
			"",
		},
		{
			"list assignment",
			`var l = (0 1 2 3)
                         var a = "2"
                         l[$a] = "666"
                         echo -n $l`,
			`0 1 666 3`,
			"",
			"",
		},
	} {
		testExec(t, test)
	}
}

func TestExecuteMultipleAssignment(t *testing.T) {
	for _, test := range []execTestCase{
		{
			desc: "multiple assignment",
			execStr: `var _1, _2 = "1", "2"
				echo -n $_1 $_2`,
			expectedStdout: "1 2",
			expectedStderr: "",
			expectedErr:    "",
		},
		{
			desc: "multiple assignment",
			execStr: `var _1, _2, _3 = "1", "2", "3"
				echo -n $_1 $_2 $_3`,
			expectedStdout: "1 2 3",
			expectedStderr: "",
			expectedErr:    "",
		},
		{
			desc: "multiple assignment",
			execStr: `var _1, _2 = (), ()
				echo -n $_1 $_2`,
			expectedStdout: "",
			expectedStderr: "",
			expectedErr:    "",
		},
		{
			desc: "multiple assignment",
			execStr: `var _1, _2 = (1 2 3 4 5), (6 7 8 9 10)
				echo -n $_1 $_2`,
			expectedStdout: "1 2 3 4 5 6 7 8 9 10",
			expectedStderr: "",
			expectedErr:    "",
		},
		{
			desc: "multiple assignment",
			execStr: `var _1, _2, _3, _4, _5, _6, _7, _8, _9, _10 = "1", "2", "3", "4", "5", "6", "7", "8", "9", "10"
				echo -n $_1 $_2 $_3 $_4 $_5 $_6 $_7 $_8 $_9 $_10`,
			expectedStdout: "1 2 3 4 5 6 7 8 9 10",
			expectedStderr: "",
			expectedErr:    "",
		},
		{
			desc: "multiple assignment",
			execStr: `var _1, _2 = (a b c), "d"
				echo -n $_1 $_2`,
			expectedStdout: "a b c d",
			expectedStderr: "",
			expectedErr:    "",
		},
		{
			desc: "multiple assignment",
			execStr: `fn a() { echo -n "a" }
				  fn b() { echo -n "b" }
				  var _a, _b = $a, $b
				  $_a(); $_b()`,
			expectedStdout: "ab",
			expectedStderr: "",
			expectedErr:    "",
		},
	} {
		testExec(t, test)
	}
}

func TestExecuteCmdAssignment(t *testing.T) {
	for _, test := range []execTestCase{
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
			"<interactive>:2:25: Functions returns 0 objects, but statement expects 1",
		},
		{
			"list assignment",
			`var l = (0 1 2 3)
                         l[0] <= echo -n 666
                         echo -n $l`,
			`666 1 2 3`,
			"",
			"",
		},
		{
			"list assignment",
			`var l = (0 1 2 3)
                         var a = "2"
                         l[$a] <= echo -n "666"
                         echo -n $l`,
			`0 1 666 3`,
			"",
			"",
		},
	} {
		testExec(t, test)
	}
}

func TestExecuteCmdMultipleAssignment(t *testing.T) {
	for _, test := range []execTestCase{
		{
			"cmd assignment",
			`name, err <= echo -n i4k
                         if $err == "0" {
                             echo -n $name
                         }`,
			"i4k", "",
			"",
		},
		{
			"list cmd assignment",
			`name, err2 <= echo "honda civic"
                         if $err2 == "0" {
                             echo -n $name
                         }`,
			"honda civic", "", "",
		},
		{
			"wrong cmd assignment",
			`name, err <= ""`,
			"", "", "wrong cmd assignment:1:14: Invalid token STRING. Expected command or function invocation",
		},
		{
			"fn must return value",
			`fn e() {}
                         v, err <= e()`,
			"",
			"",
			"<interactive>:2:25: Functions returns 0 objects, but statement expects 2",
		},
		{
			"list assignment",
			`var l = (0 1 2 3)
                         l[0], err <= echo -n 666
                         if $err == "0" {
                             echo -n $l
                         }`,
			`666 1 2 3`,
			"",
			"",
		},
		{
			desc: "list assignment",
			execStr: `var l = (0 1 2 3)
                         var a = "2"
                         l[$a], err <= echo -n "666"
                         if $err == "0" {
                             echo -n $l
                         }`,
			expectedStdout: `0 1 666 3`,
			expectedStderr: "",
			expectedErr:    "",
		},
		{
			desc:           "cmd assignment works with 1 or 2 variables",
			execStr:        "out, err1, err2 <= echo something",
			expectedStdout: "",
			expectedStderr: "",
			expectedErr:    "<interactive>:1:0: multiple assignment of commands requires two variable names, but got 3",
		},
		{
			desc: "ignore error",
			execStr: `out, _ <= cat /file-not-found/test >[2=]
					echo -n $out`,
			expectedStdout: "",
			expectedStderr: "",
			expectedErr:    "",
		},
		{
			desc: "exec without '-' and getting status still fails",
			execStr: `out <= cat /file-not-found/test >[2=]
					echo $out`,
			expectedStdout: "",
			expectedStderr: "",
			expectedErr:    "exit status 1",
		},
		{
			desc: "check status",
			execStr: `out, status <= cat /file-not-found/test >[2=]
					if $status == "0" {
						echo -n "must fail.. sniff"
					} else if $status == "1" {
						echo -n "it works"
					} else {
						echo -n "unexpected status:" $status
					}
				`,
			expectedStdout: "it works",
			expectedStderr: "",
			expectedErr:    "",
		},
		{
			desc: "multiple return in functions",
			execStr: `fn fun() {
					return "1", "2"
				}

				a, b <= fun()
				echo -n $a $b`,
			expectedStdout: "1 2",
			expectedStderr: "",
			expectedErr:    "",
		},
	} {
		testExec(t, test)
	}
}

// IFS *DO NOT* exists anymore.
// This tests only assure things works as expected (IFS has no power)
func TestExecuteCmdAssignmentIFSDontWork(t *testing.T) {
	for _, test := range []execTestCase{
		{
			"ifs",
			`var IFS = (" ")
range <= echo 1 2 3 4 5 6 7 8 9 10

for i in $range {
    echo "i = " + $i
}`,
			"", "",
			"<interactive>:4:9: Invalid variable type in for range: StringType",
		},
		{
			"ifs",
			`var IFS = (";")
range <= echo "1;2;3;4;5;6;7;8;9;10"

for i in $range {
    echo "i = " + $i
}`,
			"", "",
			"<interactive>:4:9: Invalid variable type in for range: StringType",
		},
		{
			"ifs",
			`var IFS = (" " ";")
range <= echo "1;2;3;4;5;6 7;8;9;10"

for i in $range {
    echo "i = " + $i
}`,
			"", "",
			"<interactive>:4:9: Invalid variable type in for range: StringType",
		},
		{
			"ifs",
			`var IFS = (" " "-")
range <= echo "1;2;3;4;5;6;7-8;9;10"

for i in $range {
    echo "i = " + $i
}`,
			"", "",
			"<interactive>:4:9: Invalid variable type in for range: StringType",
		},
	} {
		testExec(t, test)
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
	defer os.Remove(path)

	err = shell.Exec("redirect", fmt.Sprintf(`
        echo -n "hello world" > %s
        `, path))
	if err != nil {
		t.Fatal(err)
	}

	content, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if string(content) != "hello world" {
		t.Fatalf("File differ: '%s' != '%s'", string(content), "hello world")
	}

	// Test redirection truncate the file
	err = shell.Exec("redirect", fmt.Sprintf(`
        echo -n "a" > %s
        `, path))
	if err != nil {
		t.Fatal(err)
	}

	content, err = ioutil.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if string(content) != "a" {
		t.Fatalf("File differ: '%s' != '%s'", string(content), "a")
	}

	// Test redirection to variable
	err = shell.Exec("redirect", `
	var location = "`+path+`"
        echo -n "hello world" > $location
        `)

	if err != nil {
		t.Fatal(err)
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

	// Test redirection to concat
	err = shell.Exec("redirect", fmt.Sprintf(`
	location = "%s"
var a = ".2"
        echo -n "hello world" > $location+$a
        `, path))
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(path + ".2")
	content, err = ioutil.ReadFile(path + ".2")
	if err != nil {
		t.Fatal(err)
	}

	if string(content) != "hello world" {
		t.Fatalf("File differ: '%s' != '%s'", string(content), "hello world")
	}
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
	for _, test := range []execTestCase{
		{
			"test setenv basic",
			`var test = "hello"
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
		testExec(t, test)
	}
}

func TestExecuteCd(t *testing.T) {
	for _, test := range []execTestCase{
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
			"test cd into $val",
			`
        var val="/"
        cd $val
        pwd`,
			"/\n",
			"",
			"",
		},
		{
			"test error",
			`var val=("val1" "val2" "val3")
        cd $val
        pwd`,
			"", "",
			"<interactive>:2:12: lvalue is not comparable: (val1 val2 val3) -> ListType.",
		},
	} {
		testInteractiveExec(t, test)
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

	err = ioutil.WriteFile("/tmp/test.sh", []byte(`var TESTE="teste"`), 0644)
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
var OUTSIDE = "some value"

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
        var INSIDE = "camshaft"
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
	err = shell.Exec("test shadow", `var _path="AAA"
fn test(_path) {
echo -n $_path
}
        test("BBB")
`)

	if string(out.Bytes()) != "BBB" {
		t.Errorf("String differs: '%s' != '%s'", string(out.Bytes()), "BBB")
		return
	}

	out.Reset()

	err = shell.Exec("test shadow", `
fn test(_path) {
echo -n $_path
}

_path="AAA"
        test("BBB")
`)

	if string(out.Bytes()) != "BBB" {
		t.Errorf("String differs: '%s' != '%s'", string(out.Bytes()), "BBB")
		return
	}

	out.Reset()
	err = shell.Exec("test fn list arg", `
	var ids_luns = ()
	var id = "1"
	var lun = "lunar"
	ids_luns <= append($ids_luns, ($id $lun))
	print(len($ids_luns))`)
	if err != nil {
		t.Error(err)
		return
	}

	got := string(out.Bytes())
	expected := "1"
	if got != expected {
		t.Fatalf("String differs: '%s' != '%s'", got, expected)
	}

}

func TestFnComposition(t *testing.T) {
	for _, test := range []execTestCase{
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
		testExec(t, test)
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

	testShellExec(t, shell, execTestCase{
		"test bindfn interactive",
		`
        fn greeting() {
                echo "Hello"
        }

        bindfn greeting hello`,
		"", "", "",
	})

	shell.SetInteractive(false)
	shell.filename = "<non-interactive>"

	expectedErr := "<non-interactive>:1:0: " +
		"'hello' is a bind to 'greeting'." +
		" No binds allowed in non-interactive mode."

	testShellExec(t, shell, execTestCase{
		"test 'binded' function non-interactive",
		`hello`, "", "", expectedErr,
	})

	expectedErr = "<non-interactive>:6:8: 'bindfn' is not allowed in" +
		" non-interactive mode."

	testShellExec(t, shell,
		execTestCase{
			"test bindfn non-interactive",
			`
        fn goodbye() {
                echo "Ciao"
        }

        bindfn goodbye ciao`,
			"",
			"",
			expectedErr,
		})
}

func TestExecuteBindFn(t *testing.T) {
	for _, test := range []execTestCase{
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
		testInteractiveExec(t, test)
	}
}

func TestExecutePipe(t *testing.T) {
	var stderr bytes.Buffer
	var stdout bytes.Buffer

	// Case 1
	cmd := exec.Command(nashdPath, "-c", `echo hello | tr -d "[:space:]"`)

	cmd.Stderr = &stderr
	cmd.Stdout = &stdout

	err := cmd.Run()

	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}

	expectedOutput := "hello"
	actualOutput := string(stdout.Bytes())

	if actualOutput != expectedOutput {
		t.Errorf("'%s' != '%s'", actualOutput, expectedOutput)
		return
	}
	stdout.Reset()
	stderr.Reset()

	// Case 2
	cmd = exec.Command(nashdPath, "-c", `echo hello | wc -l | tr -d "[:space:]"`)

	cmd.Stderr = &stderr
	cmd.Stdout = &stdout

	err = cmd.Run()

	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}

	expectedOutput = "1"
	actualOutput = string(stdout.Bytes())

	if actualOutput != expectedOutput {
		t.Errorf("'%s' != '%s'", actualOutput, expectedOutput)
		return
	}
}

func TestExecuteRedirectionPipe(t *testing.T) {
	var stderr bytes.Buffer

	cmd := exec.Command(nashdPath, "-c", `cat stuff >[2=] | grep file`)

	cmd.Stderr = &stderr

	err := cmd.Run()

	expectedError := "exit status 1"

	if err != nil {
		if err.Error() != expectedError {
			t.Errorf("Error differs: Expected '%s' but got '%s'",
				expectedError,
				err.Error())
			return
		}
	}

	expectedStdErr := "<interactive>:1:16: exit status 1|success"
	strErr := strings.TrimSpace(string(stderr.Bytes()))

	if strErr != expectedStdErr {
		t.Errorf("Expected stderr to be '%s' but got '%s'",
			expectedStdErr,
			strErr)
		return
	}
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
	for _, test := range []execTestCase{
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
	var values = (0 1 2 3 4 5 6 7 8 9)

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
		testExec(t, test)
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

	err = shell.Exec("", `var TEST = "some value"`)

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

	err = shell.Exec("", `var dumpFile = "`+dumpFile+`"`)

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

	err = shell.Exec("", `var a = "A"
var b = "B"
var c = $a + $b + "C"
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

	err = shell.Exec("concat indexed var", `var tag = (Name some)
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

	err = shell.Exec("simple loop", `var files = (/etc/passwd /etc/shells)
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

	err = shell.Exec("indexing", `var list = ("1" "2" "3")
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

	err = shell.Exec("indexing", `var i = "0"
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

	err = shell.Exec("indexing", `var a = ("0")
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

	err = shell.Exec("interrupting loop", `var seq = (1 2 3 4 5)
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

	scode, ok := shell.Getvar("status", false)
	if !ok || scode.Type() != sh.StringType || scode.String() != strconv.Itoa(ENotFound) {
		t.Errorf("Invalid status code %s", scode.String())
		return
	}

	err = shell.Exec("-input-", `echo works >[1=]`)

	if err != nil {
		t.Error(err)
		return
	}

	scode, ok = shell.Getvar("status", false)
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

	scode, ok = shell.Getvar("status", false)

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

func TestExecuteVariadicFn(t *testing.T) {
	for _, test := range []execTestCase{
		{
			desc: "println",
			execStr: `fn println(fmt, arg...) {
	print($fmt+"\n", $arg...)
}
println("%s %s", "test", "test")`,
			expectedStdout: "test test\n",
			expectedStderr: "",
			expectedErr:    "",
		},
		{
			desc: "lots of args",
			execStr: `fn println(fmt, arg...) {
	print($fmt+"\n", $arg...)
}
println("%s%s%s%s%s%s%s%s%s%s", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10")`,
			expectedStdout: "12345678910\n",
			expectedStderr: "",
			expectedErr:    "",
		},
		{
			desc: "passing list to var arg fn",
			execStr: `fn puts(arg...) { for a in $arg { echo $a } }
				var a = ("1" "2" "3" "4" "5")
				puts($a...)`,
			expectedErr:    "",
			expectedStdout: "1\n2\n3\n4\n5\n",
			expectedStderr: "",
		},
		{
			desc: "passing empty list to var arg fn",
			execStr: `fn puts(arg...) { for a in $arg { echo $a } }
				var a = ()
				puts($a...)`,
			expectedErr:    "",
			expectedStdout: "",
			expectedStderr: "",
		},
		{
			desc: "... expansion",
			execStr: `var args = ("plan9" "from" "outer" "space")
print("%s %s %s %s", $args...)`,
			expectedStdout: "plan9 from outer space",
		},
		{
			desc:           "literal ... expansion",
			execStr:        `print("%s:%s:%s", ("a" "b" "c")...)`,
			expectedStdout: "a:b:c",
		},
		{
			desc:        "varargs only as last argument",
			execStr:     `fn println(arg..., fmt) {}`,
			expectedErr: "<interactive>:1:11: Vararg 'arg...' isn't the last argument",
		},
		{
			desc: "variadic argument are optional",
			execStr: `fn println(b...) {
	for v in $b {
		print($v)
	}
	print("\n")
}
println()`,
			expectedStdout: "\n",
		},
		{
			desc: "the first argument isn't optional",
			execStr: `fn a(b, c...) {
    print($b, $c...)
}
a("test")`,
			expectedStdout: "test",
		},
		{
			desc: "the first argument isn't optional",
			execStr: `fn a(b, c...) {
    print($b, $c...)
}
a()`,
			expectedErr: "<interactive>:4:0: Wrong number of arguments for function a. Expected at least 1 arguments but found 0",
		},
	} {
		testExec(t, test)
	}
}
