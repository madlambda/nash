package sh

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestDefaultNashPath(t *testing.T) {
}

func TestLoadsStdlibFromNASHROOT(t *testing.T) {
}

func TestLoadsStdlibFromDefaultNASHROOT(t *testing.T) {
}

func TestLoadsStdlibFromGOPATHOnDefaultFailure(t *testing.T) {
}

func setupHome(t *testing.T) (string, func()) {

	home, err := ioutil.TempDir("", "shellenvtests")
	if err != nil {
		t.Fatal(err)
	}
	err = os.Setenv("HOME", home)
	if err != nil {
		t.Fatal(err)
	}

	return home, func() {
		err := os.Unsetenv("HOME")
		if err != nil {
			t.Fatal(err)
		}
		err = os.RemoveAll(home)
		if err != nil {
			t.Fatal(err)
		}
	}
}
