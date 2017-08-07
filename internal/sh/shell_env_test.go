package sh

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"
)

func TestLoadsLibFromNASHPATH(t *testing.T) {
	home, teardown := setupEnvTests(t)
	defer teardown()

	nashpath := home + "/testnashpath"
	os.Setenv("NASHPATH", nashpath)

	nashlib := nashpath + "/lib"
	mkdirAll(t, nashlib)

	writeFile(t, nashlib+"/lib.sh", `
		fn test() {
			echo "hasnashpath"
		}
	`)

	newTestShell(t).Exec(t, `
		import lib
		test()
	`, "hasnashpath\n")
}

func TestLoadsLibFromHOMEIfNASHPATHIsUnset(t *testing.T) {
	home, teardown := setupEnvTests(t)
	defer teardown()

	nashlib := home + "/nash/lib"
	mkdirAll(t, nashlib)

	writeFile(t, nashlib+"/lib.sh", `
		fn test() {
			echo "defaultnashpath"
		}
	`)

	newTestShell(t).Exec(t, `
		import lib
		test()
	`, "defaultnashpath\n")
}

func TestLoadsLibIfNoEnvVarIsSet(t *testing.T) {
	home, teardown := setupEnvTests(t)
	defer teardown()

	unsetAllEnvVars(t)

	writeFile(t, home+"/lib.sh", `
		fn test() {
			echo "noenv"
		}
	`)

	oldcwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	chdir(t, home)
	defer chdir(t, oldcwd)

	newTestShell(t).Exec(t, `
		import lib
		test()
	`, "noenv\n")
}

func TestLoadsStdlibFromNASHROOT(t *testing.T) {
	home, teardown := setupEnvTests(t)
	defer teardown()

	nashroot := home + "/testnashroot"
	os.Setenv("NASHROOT", nashroot)

	nashstdlib := nashroot + "/stdlib"
	mkdirAll(t, nashstdlib)

	writeFile(t, nashstdlib+"/lib.sh", `
		fn test() {
			echo "hasnashroot"
		}
	`)

	newTestShell(t).Exec(t, `
		import lib
		test()
	`, "hasnashroot\n")
}

func TestLoadsStdlibFromHOMEIfNASHROOTIsUnset(t *testing.T) {
	home, teardown := setupEnvTests(t)
	defer teardown()

	nashroot := home + "/nashroot"
	nashstdlib := nashroot + "/stdlib"
	mkdirAll(t, nashstdlib)

	writeFile(t, nashstdlib+"/lib.sh", `
		fn test() {
			echo "defaultnashroot"
		}
	`)

	newTestShell(t).Exec(t, `
		import lib
		test()
	`, "defaultnashroot\n")
}

func TestLoadsStdlibFromGOPATHOnIfHOMEIsUnset(t *testing.T) {
	home, teardown := setupEnvTests(t)
	defer teardown()

	os.Unsetenv("HOME")
	os.Setenv("GOPATH", home)
	// Avoid failure of no NASHPATH/HOME
	os.Setenv("NASHPATH", "/whatever")

	nashstdlib := nashStdlibGoPath(home)
	mkdirAll(t, nashstdlib)

	writeFile(t, nashstdlib+"/lib.sh", `
		fn test() {
			echo "gopathnashroot"
		}
	`)

	newTestShell(t).Exec(t, `
		import lib
		test()
	`, "gopathnashroot\n")
}

func TestLoadsStdlibFromGOPATHIfStdlibNotOnHOME(t *testing.T) {
	home, teardown := setupEnvTests(t)
	defer teardown()

	os.Setenv("GOPATH", home)

	nashstdlib := nashStdlibGoPath(home)
	mkdirAll(t, nashstdlib)

	writeFile(t, nashstdlib+"/lib.sh", `
		fn test() {
			echo "nostdlibonhome"
		}
	`)

	newTestShell(t).Exec(t, `
		import lib
		test()
	`, "nostdlibonhome\n")
}

type testshell struct {
	shell  *Shell
	stdout *bytes.Buffer
}

func (s *testshell) Exec(t *testing.T, code string, expectedOutupt string) {
	err := s.shell.Exec("shellenvtest", code)
	if err != nil {
		t.Fatal(err)
	}

	output := s.stdout.String()
	s.stdout.Reset()

	if output != expectedOutupt {
		t.Fatalf(
			"expected output: [%s] got: [%s]",
			expectedOutupt,
			output,
		)
	}
}

func newTestShell(t *testing.T) *testshell {

	shell, err := NewShell()
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	shell.SetStdout(&out)

	return &testshell{shell: shell, stdout: &out}
}

func setupEnvTests(t *testing.T) (string, func()) {

	home, err := ioutil.TempDir("", "nashenvtests")
	if err != nil {
		t.Fatal(err)
	}

	curhome := os.Getenv("HOME")
	curgopath := os.Getenv("GOPATH")

	err = os.Setenv("HOME", home)
	if err != nil {
		t.Fatal(err)
	}

	mkdirAll(t, home)

	return home, func() {
		errs := []error{}
		errs = append(errs, os.Setenv("HOME", curhome))
		errs = append(errs, os.Setenv("GOPATH", curgopath))
		errs = append(errs, os.Unsetenv("NASHPATH"))
		errs = append(errs, os.Unsetenv("NASHROOT"))
		errs = append(errs, os.RemoveAll(home))

		for _, err := range errs {
			if err != nil {
				t.Errorf("error tearing down: %s", err)
			}
		}
	}
}

func unsetEnv(t *testing.T, name string) {
	err := os.Unsetenv(name)
	if err != nil {
		t.Fatalf("error[%s] unsetting [%s]", err, name)
	}
}

func unsetAllEnvVars(t *testing.T) {
	unsetEnv(t, "HOME")
	unsetEnv(t, "GOPATH")
	unsetEnv(t, "NASHPATH")
	unsetEnv(t, "NASHROOT")
}

func mkdirAll(t *testing.T, nashlib string) {
	err := os.MkdirAll(nashlib, os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
}

func writeFile(t *testing.T, filename string, data string) {
	err := ioutil.WriteFile(filename, []byte(data), os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
}

func chdir(t *testing.T, dir string) {
	err := os.Chdir(dir)
	if err != nil {
		t.Fatalf("error[%s] changing dir to[%s]", err, dir)
	}
}

func nashStdlibGoPath(gopath string) string {
	return gopath + "/src/github.com/NeowayLabs/nash/stdlib"
}
