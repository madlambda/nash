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

func TestFailsToLoadLibIfHomeIsUnset(t *testing.T) {
	// TODO: Explode or just fails for not founding the lib ?
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

func TestFailsToLoadStdlibIfGOPATHIsUnset(t *testing.T) {
	// TODO: Explode or just fails for not founding the lib ?
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

func nashStdlibGoPath(gopath string) string {
	return gopath + "/src/github.com/NeowayLabs/nash/stdlib"
}
