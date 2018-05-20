package fixture

import (
	"testing"
	"io/ioutil"
	"os"
)

func Tmpdir(t *testing.T) (string, func()) {
	t.Helper()
	
	dir, err := ioutil.TempDir("", "nash-tests")
	if err != nil {
		t.Fatal(err)
	}
	
	return dir, func() {
		err := os.RemoveAll(dir)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func MkdirAll(t *testing.T, nashlib string) {
	err := os.MkdirAll(nashlib, os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
}