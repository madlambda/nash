package builtin_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

func setup(t *testing.T) (string, func()) {
	dir, err := ioutil.TempDir("", "globtest")
	if err != nil {
		t.Fatalf("error on setup: %s", err)
	}

	return dir, func() {
		os.RemoveAll(dir)
	}
}

func createFile(t *testing.T, path string) {
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	f.WriteString("hi")
	f.Close()
}

func TestGlobNoResult(t *testing.T) {
	dir, teardown := setup(t)
	defer teardown()

	pattern := dir + "/*.la"

	out := execSuccess(t, fmt.Sprintf(`
		res <= glob("%s")
		print($res)
	`, pattern))

	if out != "" {
		t.Fatalf("expected no results, got: %q", out)
	}
}

func TestGlobOneResult(t *testing.T) {
	dir, teardown := setup(t)
	defer teardown()

	filename := dir + "/whatever.go"
	createFile(t, filename)
	pattern := dir + "/*.go"

	out := execSuccess(t, fmt.Sprintf(`
		res <= glob("%s")
		print($res)
	`, pattern))

	if out != filename {
		t.Fatalf("expected %q, got: %q", filename, out)
	}
}

func TestGlobMultipleResults(t *testing.T) {
}

func TestGlobNoParamError(t *testing.T) {
}

func TestGlobInvalidPatternError(t *testing.T) {
}
