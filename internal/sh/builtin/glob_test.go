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

func TestGlobNoResult(t *testing.T) {
	dir, teardown := setup(t)
	defer teardown()

	filenotfound := dir + "/*.la"

	out := execSuccess(t, fmt.Sprintf(`
		res <= glob("%s")
		print($res)
	`, filenotfound))

	if out != "" {
		t.Fatalf("expected no results, got: %q", out)
	}
}

func TestGlobOneResult(t *testing.T) {
}

func TestGlobMultipleResults(t *testing.T) {
}

func TestGlobNoParamError(t *testing.T) {
}

func TestGlobInvalidPatternError(t *testing.T) {
}
