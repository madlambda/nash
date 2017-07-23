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
			echo "hi"
		}
	`)

	newTestShell(t).Exec(t, `
		import lib

		test()
	`, "hi\n")
}

func TestLoadsLibFromHOMEIfNASHPATHIsUnset(t *testing.T) {
}

func TestFailsToLoadLibIfHomeIsUnset(t *testing.T) {
}

func TestLoadsStdlibFromNASHROOT(t *testing.T) {
}

func TestLoadsStdlibFromHOMEIfNASHROOTIsUnset(t *testing.T) {
}

func TestLoadsStdlibFromGOPATHOnIfStdlibNotOnHOME(t *testing.T) {
}

func TestLoadsStdlibFromGOPATHOnIfHOMEIsUnset(t *testing.T) {
}

func TestStdlibFailsIfStdlibNotOnGOPATH(t *testing.T) {
}

func TestStdlibFailsIfGOPATHIsUnset(t *testing.T) {
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

	err = os.Setenv("HOME", home)
	if err != nil {
		t.Fatal(err)
	}

	return home, func() {
		errs := []error{}
		errs = append(errs, os.Setenv("HOME", curhome))
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
