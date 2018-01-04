package builtin_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
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
		var res, err <= glob("%s")
		print($res)
		print($err)
		print(len($res))
	`, pattern))

	if out != "0" {
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
		var res, err <= glob("%s")
		print($res)
		print($err)
	`, pattern))

	if out != filename {
		t.Fatalf("expected %q, got: %q", filename, out)
	}
}

func TestGlobMultipleResults(t *testing.T) {
	dir, teardown := setup(t)
	defer teardown()

	filename1 := dir + "/whatever.h"
	filename2 := dir + "/whatever2.h"

	createFile(t, filename1)
	createFile(t, filename2)

	pattern := dir + "/*.h"

	out := execSuccess(t, fmt.Sprintf(`
		var res, err <= glob("%s")
		print($res)
		print($err)
	`, pattern))

	res := strings.Split(out, " ")
	if len(res) != 2 {
		t.Fatalf("expected 2 results, got: %d", len(res))
	}

	found1 := false
	found2 := false

	for _, r := range res {
		if r == filename1 {
			found1 = true
		}
		if r == filename2 {
			found2 = true
		}
	}

	if !found1 || !found2 {
		t.Fatalf("unable to found all files, got: %q", out)
	}
}

func TestGlobInvalidPatternError(t *testing.T) {
	out := execSuccess(t, `
		var res, err <= glob("*[.go")
		print($res)
		if $err != "" {
			print("got error")
		}
	`)

	if out != "got error" {
		t.Fatalf("expected error message on glob, got nothing, output[%s]", out)
	}
}

func TestFatalFailure(t *testing.T) {
	type tcase struct {
		name string
		code string
	}
	cases := []tcase{
		tcase{
			name: "noParam",
			code: `
				var res <= glob()
				print($res)
			`,
		},
		tcase{
			name: "multipleParams",
			code: `
				var res <= glob("/a/*", "/b/*")
				print($res)
			`,
		},
		tcase{
			name: "wrongType",
			code: `
				var param = ("hi")
				var res <= glob($param)
				print($res)
			`,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			execFailure(t, c.code)
		})
	}
}
