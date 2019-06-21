package sh_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	// FIXME: depending on other sh package on the internal sh tests seems very odd
	shtypes "github.com/NeowayLabs/nash/sh"

	"github.com/NeowayLabs/nash/internal/sh"
	"github.com/NeowayLabs/nash/internal/sh/internal/fixture"
	"github.com/NeowayLabs/nash/tests"
)

type (
	execTestCase struct {
		desc              string
		code              string
		expectedStdout    string
		expectedStderr    string
		expectedErr       string
		expectedPrefixErr string
	}

	testFixture struct {
		shell     *sh.Shell
		shellOut  *bytes.Buffer
		dir       string
		envDirs   fixture.NashDirs
		nashdPath string
	}
)

func TestInitEnv(t *testing.T) {
	os.Setenv("TEST", "abc=123=")

	f, teardown := setup(t)
	defer teardown()

	testEnv, ok := f.shell.Getenv("TEST")
	if !ok {
		t.Fatal("environment TEST not found")
	}
	expectedTestEnv := "abc=123="

	if testEnv.String() != expectedTestEnv {
		t.Fatalf("Expected TEST Env differs: '%s' != '%s'", testEnv, expectedTestEnv)
	}
}

func TestExecuteFile(t *testing.T) {
	type fileTests struct {
		path       string
		expected   string
		execBefore string
	}
	f, teardown := setup(t)
	defer teardown()

	for _, ftest := range []fileTests{
		{path: "/ex1.sh", expected: "hello world\n"},

		{path: "/sieve.sh", expected: "\n", execBefore: `var ARGS = ("" "0")`},
		{path: "/sieve.sh", expected: "\n", execBefore: `var ARGS = ("" "1")`},
		{path: "/sieve.sh", expected: "2 \n", execBefore: `var ARGS = ("" "2")`},
		{path: "/sieve.sh", expected: "2 3 \n", execBefore: `var ARGS = ("" "3")`},
		{path: "/sieve.sh", expected: "2 3 \n", execBefore: `var ARGS = ("" "4")`},
		{path: "/sieve.sh", expected: "2 3 5 \n", execBefore: `var ARGS = ("" "5")`},
		{path: "/sieve.sh", expected: "2 3 5 7 \n", execBefore: `var ARGS = ("" "10")`},

		{path: "/fibonacci.sh", expected: "1 \n", execBefore: `var ARGS = ("" "1")`},
		{path: "/fibonacci.sh", expected: "1 2 \n", execBefore: `var ARGS = ("" "2")`},
		{path: "/fibonacci.sh", expected: "1 2 3 \n", execBefore: `var ARGS = ("" "3")`},
		{path: "/fibonacci.sh", expected: "1 2 3 5 8 \n", execBefore: `var ARGS = ("" "5")`},
	} {
		testExecuteFile(t, f.dir+ftest.path, ftest.expected, ftest.execBefore)
	}
}

func TestExecuteCommand(t *testing.T) {
	echopath, err := exec.LookPath("echo")
	if err != nil {
		t.Fatal(err)
	}

	echodir := filepath.Dir(echopath)

	for _, test := range []execTestCase{
		{
			desc:              "command failed",
			code:              `non-existing-program`,
			expectedStdout:    "",
			expectedStderr:    "",
			expectedPrefixErr: `exec: "non-existing-program": executable file not found in `,
		},
		{
			desc:           "err ignored",
			code:           `-non-existing-program`,
			expectedStdout: "",
			expectedStderr: "",
			expectedErr:    "",
		},
		{
			desc:           "hello world",
			code:           "echo -n hello world",
			expectedStdout: "hello world",
			expectedStderr: "",
			expectedErr:    "",
		},
		{
			desc:           "cmd with concat",
			code:           `echo -n "hello " + "world"`,
			expectedStdout: "hello world",
			expectedStderr: "",
			expectedErr:    "",
		},
		{
			desc: "local command",
			code: fmt.Sprintf(`var echodir = "%s"
chdir($echodir)
./echo -n hello
`, strings.Replace(echodir, "\\", "\\\\", -1)),
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
			desc:           "wrong assignment",
			code:           `var name=i4k`,
			expectedStdout: "",
			expectedStderr: "",
			expectedErr:    "wrong assignment:1:9: Unexpected token IDENT. Expecting VARIABLE, STRING or (",
		},
		{
			desc: "assignment",
			code: `var name="i4k"
                         echo $name`,
			expectedStdout: "i4k\n",
			expectedStderr: "",
			expectedErr:    "",
		},
		{
			desc: "list assignment",
			code: `var name=(honda civic)
                         echo -n $name`,
			expectedStdout: "honda civic",
			expectedStderr: "",
			expectedErr:    "",
		},
		{
			desc: "list of lists",
			code: `var l = (
		(name Archlinux)
		(arch amd64)
		(kernel 4.7.1)
	)

	echo $l[0]
	echo $l[1]
	echo -n $l[2]`,
			expectedStdout: `name Archlinux
arch amd64
kernel 4.7.1`,
			expectedStderr: "",
			expectedErr:    "",
		},
		{
			desc: "list assignment",
			code: `var l = (0 1 2 3)
                         l[0] = "666"
                         echo -n $l`,
			expectedStdout: `666 1 2 3`,
			expectedStderr: "",
			expectedErr:    "",
		},
		{
			desc: "list assignment",
			code: `var l = (0 1 2 3)
                         var a = "2"
                         l[$a] = "666"
                         echo -n $l`,
			expectedStdout: `0 1 666 3`,
			expectedStderr: "",
			expectedErr:    "",
		},
	} {
		testExec(t, test)
	}
}

func TestExecuteMultipleAssignment(t *testing.T) {
	for _, test := range []execTestCase{
		{
			desc: "multiple assignment",
			code: `var _1, _2 = "1", "2"
				echo -n $_1 $_2`,
			expectedStdout: "1 2",
			expectedStderr: "",
			expectedErr:    "",
		},
		{
			desc: "multiple assignment",
			code: `var _1, _2, _3 = "1", "2", "3"
				echo -n $_1 $_2 $_3`,
			expectedStdout: "1 2 3",
			expectedStderr: "",
			expectedErr:    "",
		},
		{
			desc: "multiple assignment",
			code: `var _1, _2 = (), ()
				echo -n $_1 $_2`,
			expectedStdout: "",
			expectedStderr: "",
			expectedErr:    "",
		},
		{
			desc: "multiple assignment",
			code: `var _1, _2 = (1 2 3 4 5), (6 7 8 9 10)
				echo -n $_1 $_2`,
			expectedStdout: "1 2 3 4 5 6 7 8 9 10",
			expectedStderr: "",
			expectedErr:    "",
		},
		{
			desc: "multiple assignment",
			code: `var _1, _2, _3, _4, _5, _6, _7, _8, _9, _10 = "1", "2", "3", "4", "5", "6", "7", "8", "9", "10"
				echo -n $_1 $_2 $_3 $_4 $_5 $_6 $_7 $_8 $_9 $_10`,
			expectedStdout: "1 2 3 4 5 6 7 8 9 10",
			expectedStderr: "",
			expectedErr:    "",
		},
		{
			desc: "multiple assignment",
			code: `var _1, _2 = (a b c), "d"
				echo -n $_1 $_2`,
			expectedStdout: "a b c d",
			expectedStderr: "",
			expectedErr:    "",
		},
		{
			desc: "multiple assignment",
			code: `fn a() { echo -n "a" }
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
			desc: "cmd assignment",
			code: `var name <= echo -n i4k
                         echo -n $name`,
			expectedStdout: "i4k",
			expectedStderr: "",
			expectedErr:    "",
		},
		{
			desc: "list cmd assignment",
			code: `var name <= echo "honda civic"
                         echo -n $name`,
			expectedStdout: "honda civic",
			expectedStderr: "",
			expectedErr:    "",
		},
		{
			desc:           "wrong cmd assignment",
			code:           `var name <= ""`,
			expectedStdout: "",
			expectedStderr: "",
			expectedErr:    "wrong cmd assignment:1:13: Invalid token STRING. Expected command or function invocation",
		},
		{
			desc: "fn must return value",
			code: `fn e() {}
                         var v <= e()`,
			expectedStdout: "",
			expectedStderr: "",
			expectedErr:    "<interactive>:2:29: Functions returns 0 objects, but statement expects 1",
		},
		{
			desc: "list assignment",
			code: `var l = (0 1 2 3)
                         l[0] <= echo -n 666
                         echo -n $l`,
			expectedStdout: `666 1 2 3`,
			expectedStderr: "",
			expectedErr:    "",
		},
		{
			desc: "list assignment",
			code: `var l = (0 1 2 3)
                         var a = "2"
                         l[$a] <= echo -n "666"
                         echo -n $l`,
			expectedStdout: `0 1 666 3`,
			expectedStderr: "",
			expectedErr:    "",
		},
	} {
		testExec(t, test)
	}
}

func TestExecuteCmdMultipleAssignment(t *testing.T) {
	for _, test := range []execTestCase{
		{
			desc: "cmd assignment",
			code: `var name, err <= echo -n i4k
                         if $err == "0" {
                             echo -n $name
                         }`,
			expectedStdout: "i4k",
			expectedStderr: "",
			expectedErr:    "",
		},
		{
			desc: "list cmd assignment",
			code: `var name, err2 <= echo "honda civic"
                         if $err2 == "0" {
                             echo -n $name
                         }`,
			expectedStdout: "honda civic",
			expectedStderr: "",
			expectedErr:    "",
		},
		{
			desc:           "wrong cmd assignment",
			code:           `var name, err <= ""`,
			expectedStdout: "",
			expectedStderr: "",
			expectedErr:    "wrong cmd assignment:1:18: Invalid token STRING. Expected command or function invocation",
		},
		{
			desc: "fn must return value",
			code: `fn e() {}
                         var v, err <= e()`,
			expectedStdout: "",
			expectedStderr: "",
			expectedErr:    "<interactive>:2:29: Functions returns 0 objects, but statement expects 2",
		},
		{
			desc: "list assignment",
			code: `var l = (0 1 2 3)
                         var l[0], err <= echo -n 666
                         if $err == "0" {
                             echo -n $l
                         }`,
			expectedStdout: `666 1 2 3`,
			expectedStderr: "",
			expectedErr:    "",
		},
		{
			desc: "list assignment",
			code: `var l = (0 1 2 3)
                         var a = "2"
                         var l[$a], err <= echo -n "666"
                         if $err == "0" {
                             echo -n $l
                         }`,
			expectedStdout: `0 1 666 3`,
			expectedStderr: "",
			expectedErr:    "",
		},
		{
			desc:           "cmd assignment works with 1 or 2 variables",
			code:           "var out, err, status <= echo something",
			expectedStdout: "",
			expectedStderr: "",
			expectedErr:    "",
		},
		{
			desc: "ignore error",
			code: `var out, _ <= cat /file-not-found/test >[2=]
					echo -n $out`,
			expectedStdout: "",
			expectedStderr: "",
			expectedErr:    "",
		},
		{
			desc: "exec without '-' and getting status still fails",
			code: `var out <= cat /file-not-found/test >[2=]
					echo $out`,
			expectedStdout: "",
			expectedStderr: "",
			expectedErr:    "exit status 1",
		},
		{
			desc: "check status",
			code: `var out, status <= cat /file-not-found/test >[2=]
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
			code: `fn fun() {
					return "1", "2"
				}

				var a, b <= fun()
				echo -n $a $b`,
			expectedStdout: "1 2",
			expectedStderr: "",
			expectedErr:    "",
		},
	} {
		testExec(t, test)
	}
}

func TestExecuteRedirection(t *testing.T) {
	f, teardown := setup(t)
	defer teardown()

	shell := f.shell

	pathobj, err := ioutil.TempFile("", "nash-redir")
	if err != nil {
		t.Fatal(err)
	}
	path := strings.Replace(pathobj.Name(), "\\", "\\\\", -1)
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
	f, teardown := setup(t)
	defer teardown()

	shell := f.shell

	tmpfile, err := ioutil.TempFile("", "nash-redir-map")
	if err != nil {
		t.Fatal(err)
	}

	//path := strings.Replace(tmpfile.Name(), "\\", "\\\\", -1)
	defer os.Remove(tmpfile.Name())

	err = shell.Exec("redirect map", fmt.Sprintf(`
        echo -n "hello world" > %s
        `, tmpfile.Name()))
	if err != nil {
		t.Error(err)
		return
	}

	content, err := ioutil.ReadFile(tmpfile.Name())
	if err != nil {
		t.Fatal(err)
	}

	if string(content) != "hello world" {
		t.Fatalf("File differ: '%s' != '%s'", string(content), "hello world")
	}
}

func TestExecuteSetenv(t *testing.T) {
	f, teardown := setup(t)
	defer teardown()

	for _, test := range []execTestCase{
		{
			desc: "test setenv basic",
			code: `var setenvtest = "hello"
						 setenv setenvtest
                         ` + f.nashdPath + ` -c "echo $setenvtest"`,
			expectedStdout: "hello\n",
			expectedStderr: "",
			expectedErr:    "",
		},
		{
			desc: "test setenv assignment",
			code: `setenv setenvtest = "hello"
                         ` + f.nashdPath + ` -c "echo $setenvtest"`,
			expectedStdout: "hello\n",
			expectedStderr: "",
			expectedErr:    "",
		},
		{
			desc: "test setenv exec cmd",
			code: `setenv setenvtest <= echo -n "hello"
                         ` + f.nashdPath + ` -c "echo $setenvtest"`,
			expectedStdout: "hello\n",
			expectedStderr: "",
			expectedErr:    "",
		},
		{
			desc:           "test setenv semicolon",
			code:           `setenv a setenv b`,
			expectedStdout: "",
			expectedStderr: "",
			expectedErr:    "test setenv semicolon:1:9: Unexpected token setenv, expected semicolon (;) or EOL",
		},
	} {
		testExec(t, test)
	}
}

func TestExecuteCd(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "nash-cd")
	if err != nil {
		t.Fatal(err)
	}
	tmpdir, err = filepath.EvalSymlinks(tmpdir)
	if err != nil {
		t.Fatal(err)
	}

	tmpdirEscaped := strings.Replace(tmpdir, "\\", "\\\\", -1)
	homeEnvVar := "HOME"
	if runtime.GOOS == "windows" {
		homeEnvVar = "HOMEPATH"

		// hack to use nash's pwd instead of gnu on windows
		projectDir := filepath.FromSlash(tests.Projectpath)
		pwdDir := filepath.Join(projectDir, "stdbin", "pwd")
		path := os.Getenv("Path")
		defer os.Setenv("Path", path) // TODO(i4k): very unsafe
		os.Setenv("Path", pwdDir+";"+path)
	}

	for _, test := range []execTestCase{
		{
			desc: "test cd 1",
			code: fmt.Sprintf(`cd %s
        pwd`, tmpdir),
			expectedStdout: tmpdir + "\n",
			expectedStderr: "",
			expectedErr:    "",
		},
		{
			desc: "test cd 2",
			code: fmt.Sprintf(`%s = "%s"
        setenv %s
        cd
        pwd`, homeEnvVar, tmpdirEscaped, homeEnvVar),
			expectedStdout: tmpdir + "\n",
			expectedStderr: "",
			expectedErr:    "",
		},
		{
			desc: "test cd into $var",
			code: fmt.Sprintf(`
        var v = "%s"
        cd $v
        pwd`, tmpdirEscaped),
			expectedStdout: tmpdir + "\n",
			expectedStderr: "",
			expectedErr:    "",
		},
	} {
		t.Run(test.desc, func(t *testing.T) {
			test := test
			testInteractiveExec(t, test)
		})
	}
}

func TestExecuteImport(t *testing.T) {
	f, teardown := setup(t)
	defer teardown()

	shell := f.shell
	out := f.shellOut

	tmpfile, err := ioutil.TempFile("", "nash-import")
	if err != nil {
		t.Fatal(err)
	}

	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write([]byte(`var TESTE="teste"`))
	if err != nil {
		t.Fatal(err)
	}

	fnameEscaped := strings.Replace(tmpfile.Name(), "\\", "\\\\", -1)

	err = shell.Exec("test import", fmt.Sprintf(`import %s
        echo $TESTE
        `, fnameEscaped))
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
	for _, test := range []execTestCase{
		{
			desc: "if equal",
			code: `
        if "" == "" {
            echo "empty string works"
        }`,
			expectedStdout: "empty string works\n",
			expectedStderr: "",
			expectedErr:    "",
		},
		{
			desc: "if equal",
			code: `
        if "i4k" == "_i4k_" {
            echo "do not print"
        }`,
			expectedStdout: "",
			expectedStderr: "",
			expectedErr:    "",
		},
		{
			desc: "if lvalue concat",
			code: `
        if "i4"+"k" == "i4k" {
            echo -n "ok"
        }`,
			expectedStdout: "ok",
			expectedStderr: "",
			expectedErr:    "",
		},
		{
			desc: "if lvalue concat",
			code: `var name = "something"
        if $name+"k" == "somethingk" {
            echo -n "ok"
        }`,
			expectedStdout: "ok",
			expectedStderr: "",
			expectedErr:    "",
		},
		{
			desc: "if lvalue concat",
			code: `var name = "something"
        if $name+"k"+"k" == "somethingkk" {
            echo -n "ok"
        }`,
			expectedStdout: "ok",
			expectedStderr: "",
			expectedErr:    "",
		},
		{
			desc: "if rvalue concat",
			code: `
        if "i4k" == "i4"+"k" {
            echo -n "ok"
        }`,
			expectedStdout: "ok",
			expectedStderr: "",
			expectedErr:    "",
		},
		{
			desc: "if lvalue funcall",
			code: `var a = ()
        if len($a) == "0" {
            echo -n "ok"
        }`,
			expectedStdout: "ok",
			expectedStderr: "",
			expectedErr:    "",
		},
		{
			desc: "if rvalue funcall",
			code: `var a = ("1")
        if "1" == len($a) {
            echo -n "ok"
        }`,
			expectedStdout: "ok",
			expectedStderr: "",
			expectedErr:    "",
		},
		{
			desc: "if lvalue funcall with concat",
			code: `var a = ()
        if len($a)+"1" == "01" {
            echo -n "ok"
        }`,
			expectedStdout: "ok",
			expectedStderr: "",
			expectedErr:    "",
		},
	} {
		testExec(t, test)
	}
}

func TestExecuteIfElse(t *testing.T) {
	f, teardown := setup(t)
	defer teardown()

	shell := f.shell
	out := f.shellOut

	err := shell.Exec("test if else", `
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
	f, teardown := setup(t)
	defer teardown()

	shell := f.shell
	out := f.shellOut

	err := shell.Exec("test if else", `
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
	f, teardown := setup(t)
	defer teardown()

	err := f.shell.Exec("test fnDecl", `
        fn build(image, debug) {
                ls
        }`)
	if err != nil {
		t.Error(err)
		return
	}
}

func TestExecuteFnInv(t *testing.T) {
	f, teardown := setup(t)
	defer teardown()

	shell := f.shell
	out := f.shellOut

	err := shell.Exec("test fn inv", `
fn getints() {
        return ("1" "2" "3" "4" "5" "6" "7" "8" "9" "0")
}

var integers <= getints()
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

var val <= getOUTSIDE()
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
	var ids_luns <= append($ids_luns, ($id $lun))
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
			desc: "composition",
			code: `
                fn a(b) { echo -n $b }
                fn b()  { return "hello" }
                a(b())
        `,
			expectedStdout: "hello",
			expectedStderr: "",
			expectedErr:    "",
		},
		{
			desc: "composition",
			code: `
                fn a(b, c) { echo -n $b $c  }
                fn b()     { return "hello" }
                fn c()     { return "world" }
                a(b(), c())
        `,
			expectedStdout: "hello world",
			expectedStderr: "",
			expectedErr:    "",
		},
	} {
		testExec(t, test)
	}
}

func TestExecuteFnInvOthers(t *testing.T) {
	f, teardown := setup(t)
	defer teardown()

	shell := f.shell
	out := f.shellOut

	err := shell.Exec("test fn inv", `
fn _getints() {
        return ("1" "2" "3" "4" "5" "6" "7" "8" "9" "0")
}

fn getints() {
        var values <= _getints()

        return $values
}

var integers <= getints()
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
	f, teardown := setup(t)
	defer teardown()

	shell := f.shell

	shell.SetInteractive(true)

	testShellExec(t, shell, execTestCase{
		desc: "test bindfn interactive",
		code: `
        fn greeting() {
                echo "Hello"
        }

        bindfn greeting hello`,
	})

	shell.SetInteractive(false)
	// FIXME: using private stuff on tests ?
	// shell.filename = "<non-interactive>"
	t.Skip("FIXME: TEST USES PRIVATE STUFF")

	expectedErr := "<non-interactive>:1:0: " +
		"'hello' is a bind to 'greeting'." +
		" No binds allowed in non-interactive mode."

	testShellExec(t, shell, execTestCase{
		desc:           "test 'binded' function non-interactive",
		code:           `hello`,
		expectedStdout: "",
		expectedStderr: "",
		expectedErr:    expectedErr,
	})

	expectedErr = "<non-interactive>:6:8: 'bindfn' is not allowed in" +
		" non-interactive mode."

	testShellExec(t, shell,
		execTestCase{
			desc: "test bindfn non-interactive",
			code: `
        fn goodbye() {
                echo "Ciao"
        }

        bindfn goodbye ciao`,
			expectedStdout: "",
			expectedStderr: "",
			expectedErr:    expectedErr,
		})
}

func TestExecuteBindFn(t *testing.T) {
	for _, test := range []execTestCase{
		{
			desc: "test bindfn",
			code: `
				fn cd() {
					echo "override builtin cd"
				}

				bindfn cd cd
				cd
			`,
			expectedStdout: "override builtin cd\n",
		},
		{
			desc: "test bindfn vargs",
			code: `
				fn echoargs(args...) {
					for a in $args {
						echo $a
					}
				}

				bindfn echoargs echoargs
				echoargs
				echoargs "a"
				echoargs "b" "c"
			`,
			expectedStdout: "a\nb\nc\n",
		},
		{
			desc: "test empty bindfn vargs len",
			code: `
				fn echoargs(args...) {
					var l <= len($args)
					echo $l
				}

				bindfn echoargs echoargs
				echoargs
			`,
			expectedStdout: "0\n",
		},
		{
			desc: "test bindfn args",
			code: `
				fn foo(line) {
					echo $line
				}

				bindfn foo bar
				bar test test
			`,
			expectedErr: "Wrong number of arguments for function foo. Expected 1 but found 2",
		},
	} {
		t.Run(test.desc, func(t *testing.T) {
			testInteractiveExec(t, test)
		})
	}
}

func TestExecutePipe(t *testing.T) {
	var stderr bytes.Buffer
	var stdout bytes.Buffer

	f, teardown := setup(t)
	defer teardown()

	// Case 1
	cmd := exec.Command(f.nashdPath, "-c", `echo hello | tr -d "[:space:]"`)

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
	cmd = exec.Command(f.nashdPath, "-c", `echo hello | wc -l | tr -d "[:space:]"`)

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
	f, teardown := setup(t)
	defer teardown()

	err := f.shell.Exec("test", `cat stuff >[2=] | grep file`)
	expectedErr := "<interactive>:1:16: exit status 1|success"

	if err == nil {
		t.Fatalf("expected err[%s]", expectedErr)
	}

	if err.Error() != expectedErr {
		t.Errorf("Expected stderr to be '%s' but got '%s'",
			expectedErr,
			err.Error())
		return
	}
}

func testTCPRedirection(t *testing.T, port, command string) {
	message := "hello world"
	done := make(chan error)

	l, err := net.Listen("tcp", port)
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	go func() {
		f, teardown := setup(t)
		defer teardown()

		err := <-done
		if err != nil {
			t.Fatal(err)
		}

		done <- f.shell.Exec("test net redirection", command)
	}()

	done <- nil // synchronize peers
	conn, err := l.Accept()
	if err != nil {
		done <- err
		t.Fatal(err)
	}

	defer conn.Close()
	err = <-done
	if err != nil {
		t.Fatal(err)
	}

	buf, err := ioutil.ReadAll(conn)
	if err != nil {
		t.Fatal(err)
	}

	if msg := string(buf[:]); msg != message {
		t.Fatalf("Unexpected message:\nGot:\t\t%s\nExpected:\t%s\n", msg, message)
	}
}

func TestTCPRedirection(t *testing.T) {
	testTCPRedirection(t, ":4666", `echo -n "hello world" >[1] "tcp://localhost:4666"`)
	testTCPRedirection(t, ":4667", `echo -n "hello world" > "tcp://localhost:4667"`)
}

func TestExecuteUnixRedirection(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("windows does not support unix socket")
		return
	}
	message := "hello world"

	sockDir, err := ioutil.TempDir("", "nash-tests")
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

		f, teardown := setup(t)
		defer teardown()

		defer func() {
			writeDone <- true
		}()

		<-done

		err = f.shell.Exec("test net redirection", `echo -n "`+message+`" >[1] "unix://`+sockFile+`"`)

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
		f, teardown := setup(t)
		defer teardown()

		defer func() {
			writeDone <- true
		}()

		<-done

		err := f.shell.Exec("test net redirection", `echo -n "`+message+`" >[1] "udp://localhost:6667"`)

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
			desc:           "return invalid",
			code:           `return`,
			expectedStdout: "",
			expectedStderr: "",
			expectedErr:    "<interactive>:1:0: Unexpected return outside of function declaration.",
		},
		{
			desc: "test simple return",
			code: `fn test() { return }
test()`,
			expectedStdout: "",
			expectedStderr: "",
			expectedErr:    "",
		},
		{
			desc: "return must finish func evaluation",
			code: `fn test() {
	if "1" == "1" {
		return "1"
	}

	return "0"
}

var res <= test()
echo -n $res`,
			expectedStdout: "1",
			expectedStderr: "",
			expectedErr:    "",
		},
		{
			desc: "ret from for",
			code: `fn test() {
	var values = (0 1 2 3 4 5 6 7 8 9)

	for i in $values {
		if $i == "5" {
			return $i
		}
	}

	return "0"
}
var a <= test()
echo -n $a`,
			expectedStdout: "5",
			expectedStderr: "",
			expectedErr:    "",
		},
		{
			desc: "inf loop ret",
			code: `fn test() {
	for {
		if "1" == "1" {
			return "1"
		}
	}

	# never happen
	return "bleh"
}
var a <= test()
echo -n $a`,
			expectedStdout: "1",
			expectedStderr: "",
			expectedErr:    "",
		},
		{
			desc: "test returning funcall",
			code: `fn a() { return "1" }
                         fn b() { return a() }
                         var c <= b()
                         echo -n $c`,
			expectedStdout: "1",
			expectedStderr: "",
			expectedErr:    "",
		},
	} {
		testExec(t, test)
	}
}

func TestExecuteFnAsFirstClass(t *testing.T) {

	f, teardown := setup(t)
	defer teardown()

	shell := f.shell
	out := f.shellOut

	err := shell.Exec("test fn by arg", `
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

func TestExecuteConcat(t *testing.T) {
	f, teardown := setup(t)
	defer teardown()

	shell := f.shell
	out := f.shellOut

	err := shell.Exec("", `var a = "A"
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
	f, teardown := setup(t)
	defer teardown()

	shell := f.shell
	out := f.shellOut

	err := shell.Exec("simple loop", `var files = (/etc/passwd /etc/shells)
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
	f, teardown := setup(t)
	defer teardown()

	shell := f.shell

	doneCtrlc := make(chan bool)
	doneLoop := make(chan bool)

	go func() {
		fmt.Printf("Waiting 2 second to abort infinite loop")
		time.Sleep(2 * time.Second)

		err := shell.TriggerCTRLC()
		if err != nil {
			t.Fatal(err)
		}
		doneCtrlc <- true
	}()

	go func() {
		err := shell.Exec("simple loop", `for {
		echo "infinite loop" >[1=]
		sleep 1
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
	f, teardown := setup(t)
	defer teardown()

	shell := f.shell
	out := f.shellOut

	err := shell.Exec("indexing", `var list = ("1" "2" "3")
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

	err = shell.Exec("indexing", `var tmp <= seq 0 2
var seq <= split($tmp, "\n")

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
	f, teardown := setup(t)
	defer teardown()

	shell := f.shell
	out := f.shellOut

	err := shell.Exec("set env", `setenv SHELL = "bleh"`)

	if err != nil {
		t.Error(err)
		return
	}

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
	f, teardown := setup(t)
	defer teardown()

	shell := f.shell
	shell.TriggerCTRLC()

	time.Sleep(time.Second * 1)

	err := shell.Exec("interrupting loop", `var seq = (1 2 3 4 5)
for i in $seq {}`)

	if err != nil {
		t.Error(err)
		return
	}
}

func TestExecuteErrorSuppressionAll(t *testing.T) {
	f, teardown := setup(t)
	defer teardown()

	shell := f.shell

	err := shell.Exec("-input-", `var _, status <= command-not-exists`)
	if err != nil {
		t.Errorf("Expected to not fail...: %s", err.Error())
		return
	}

	// FIXME: depending on other sh package on the internal sh tests seems very odd
	scode, ok := shell.GetLocalvar("status")
	if !ok || scode.Type() != shtypes.StringType || scode.String() != strconv.Itoa(sh.ENotFound) {
		t.Errorf("Invalid status code %v", scode)
		return
	}

	err = shell.Exec("-input-", `var _, status <= echo works`)
	if err != nil {
		t.Error(err)
		return
	}

	// FIXME: depending on other sh package on the internal sh tests seems very odd
	scode, ok = shell.GetLocalvar("status")
	if !ok || scode.Type() != shtypes.StringType || scode.String() != "0" {
		t.Errorf("Invalid status code %v", scode)
		return
	}

	err = shell.Exec("-input-", `echo works | cmd-does-not-exists`)
	if err == nil {
		t.Errorf("Must fail")
		return
	}

	expectedError := `<interactive>:1:11: not started|exec: "cmd-does-not-exists": executable file not found in`

	if !strings.HasPrefix(err.Error(), expectedError) {
		t.Errorf("Unexpected error: %s", err.Error())
		return
	}
}

func TestExecuteGracefullyError(t *testing.T) {
	f, teardown := setup(t)
	defer teardown()

	shell := f.shell

	err := shell.Exec("someinput.sh", "(")
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
	f, teardown := setup(t)
	defer teardown()

	shell := f.shell
	out := f.shellOut

	err := shell.Exec("test", `(echo -n
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
	f, teardown := setup(t)
	defer teardown()

	shell := f.shell
	out := f.shellOut

	err := shell.Exec("test", `var val <= (echo -n
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
	f, teardown := setup(t)
	defer teardown()

	shell := f.shell

	err := shell.Exec("test", "(")

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
			code: `fn println(fmt, arg...) {
	print($fmt+"\n", $arg...)
}
println("%s %s", "test", "test")`,
			expectedStdout: "test test\n",
			expectedStderr: "",
			expectedErr:    "",
		},
		{
			desc: "lots of args",
			code: `fn println(fmt, arg...) {
	print($fmt+"\n", $arg...)
}
println("%s%s%s%s%s%s%s%s%s%s", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10")`,
			expectedStdout: "12345678910\n",
			expectedStderr: "",
			expectedErr:    "",
		},
		{
			desc: "passing list to var arg fn",
			code: `fn puts(arg...) { for a in $arg { echo $a } }
				var a = ("1" "2" "3" "4" "5")
				puts($a...)`,
			expectedErr:    "",
			expectedStdout: "1\n2\n3\n4\n5\n",
			expectedStderr: "",
		},
		{
			desc: "passing empty list to var arg fn",
			code: `fn puts(arg...) { for a in $arg { echo $a } }
				var a = ()
				puts($a...)`,
			expectedErr:    "",
			expectedStdout: "",
			expectedStderr: "",
		},
		{
			desc: "... expansion",
			code: `var args = ("plan9" "from" "outer" "space")
print("%s %s %s %s", $args...)`,
			expectedStdout: "plan9 from outer space",
		},
		{
			desc:           "literal ... expansion",
			code:           `print("%s:%s:%s", ("a" "b" "c")...)`,
			expectedStdout: "a:b:c",
		},
		{
			desc:        "varargs only as last argument",
			code:        `fn println(arg..., fmt) {}`,
			expectedErr: "<interactive>:1:11: Vararg 'arg...' isn't the last argument",
		},
		{
			desc: "variadic argument are optional",
			code: `fn println(b...) {
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
			code: `fn a(b, c...) {
    print($b, $c...)
}
a("test")`,
			expectedStdout: "test",
		},
		{
			desc: "the first argument isn't optional",
			code: `fn a(b, c...) {
    print($b, $c...)
}
a()`,
			expectedErr: "<interactive>:4:0: Wrong number of arguments for function a. Expected at least 1 arguments but found 0",
		},
	} {
		testExec(t, test)
	}
}

func setup(t *testing.T) (testFixture, func()) {
	dirs := fixture.SetupNashDirs(t)
	shell, err := sh.NewAbortShell(dirs.Path, dirs.Root)
	if err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	shell.SetStdout(&out)

	return testFixture{
		shell:     shell,
		shellOut:  &out,
		dir:       tests.Testdir,
		envDirs:   dirs,
		nashdPath: tests.Nashcmd,
	}, dirs.Cleanup
}

func testExecuteFile(t *testing.T, path, expected string, before string) {
	f, teardown := setup(t)
	defer teardown()

	if before != "" {
		f.shell.Exec("", before)
	}

	err := f.shell.ExecFile(path)

	if err != nil {
		t.Error(err)
		return
	}

	if string(f.shellOut.Bytes()) != expected {
		t.Errorf("Wrong command output: '%s' != '%s'",
			string(f.shellOut.Bytes()), expected)
		return
	}
}

func testShellExec(t *testing.T, shell *sh.Shell, testcase execTestCase) {
	t.Helper()

	var bout bytes.Buffer
	var berr bytes.Buffer
	shell.SetStderr(&berr)
	shell.SetStdout(&bout)

	err := shell.Exec(testcase.desc, testcase.code)
	if err != nil {
		if testcase.expectedPrefixErr != "" {
			if !strings.HasPrefix(err.Error(), testcase.expectedPrefixErr) {
				t.Errorf("[%s] Prefix of error differs: Expected prefix '%s' in '%s'",
					testcase.desc,
					testcase.expectedPrefixErr,
					err.Error())
			}
		} else if err.Error() != testcase.expectedErr {
			t.Errorf("[%s] Error differs: Expected '%s' but got '%s'",
				testcase.desc,
				testcase.expectedErr,
				err.Error())
		}
	} else if testcase.expectedErr != "" {
		t.Fatalf("Expected error[%s] but got nil", testcase.expectedErr)
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
	t.Helper()
	f, teardown := setup(t)
	defer teardown()

	testShellExec(t, f.shell, testcase)
}

func testInteractiveExec(t *testing.T, testcase execTestCase) {
	t.Helper()

	f, teardown := setup(t)
	defer teardown()

	f.shell.SetInteractive(true)
	testShellExec(t, f.shell, testcase)
}
