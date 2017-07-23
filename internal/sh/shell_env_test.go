package sh

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"
)

func TestLoadsLibFromNASHPATH(t *testing.T) {
	f, teardown := setupHome(t)
	defer teardown()

	writeFile(t, f.home+"/lib.sh", `
		fn test() {
			echo "hi"
		}
	`)

	f.Exec(t, `
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

type envfixture struct {
	home   string
	shell  *Shell
	stdout *bytes.Buffer
}

func (e *envfixture) Exec(t *testing.T, code string, expectedOutupt string) {
	err := e.shell.Exec("shellenvtest", code)
	if err != nil {
		t.Fatal(err)
	}

	output := e.stdout.String()
	e.stdout.Reset()

	if output != expectedOutupt {
		t.Fatalf(
			"expected output: [%s] got: [%s]",
			expectedOutupt,
			output,
		)
	}
}

func setupHome(t *testing.T) (envfixture, func()) {

	home, err := ioutil.TempDir("", "nashenvtests")
	if err != nil {
		t.Fatal(err)
	}

	curhome := os.Getenv("HOME")

	err = os.Setenv("HOME", home)
	if err != nil {
		t.Fatal(err)
	}

	shell, err := NewShell()
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	shell.SetStdout(&out)

	return envfixture{
			home:   home,
			shell:  shell,
			stdout: &out,
		}, func() {
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

func writeFile(t *testing.T, filename string, data string) {
	err := ioutil.WriteFile(filename, []byte(data), os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
}
