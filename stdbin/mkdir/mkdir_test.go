package main 

import (
	"testing"
	"path"
	"path/filepath"
	"io/ioutil"
	"os"
)

type testcase struct {
	dirs []string
}

func testMkdir(t *testing.T, tc testcase) {
	tmpDir, err := ioutil.TempDir("", "nash-mkdir")
	if err != nil {
		t.Fatal(err)
	}

	defer os.RemoveAll(tmpDir)
	var dirs []string
	for _, p := range tc.dirs {
		dirs = append(dirs, filepath.FromSlash(path.Join(tmpDir, p)))
	}

	err = mkdirs(dirs)
	if err != nil {
		t.Fatal(err)
	}

	for _, d := range dirs {
		if s, err := os.Stat(d); err != nil {
			t.Fatal(err)
		} else if s.Mode()&os.ModeDir != os.ModeDir {
			t.Fatalf("Invalid directory mode: %v", s.Mode())
		}
	}
}

func TestMkdir(t *testing.T) {
	for _, tc := range []testcase{
		{
			dirs: []string{},
		},
		{
			dirs: []string{
				"1", "2", "3", "4", "5",
				"some", "thing", "_random_",
				"_",
			},
		},
		{
			dirs: []string{"a", "a"}, // directory already exists, silently works
		},
	} {
		tc := tc 
		testMkdir(t, tc)
	}
}