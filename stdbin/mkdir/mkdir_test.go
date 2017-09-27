package main 

import (
	"testing"
	"path"
	"path/filepath"
	"io/ioutil"
	"os"
	"fmt"
)

func TestMkdir(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "nash-mkdir")
	if err != nil {
		t.Fatal(err)
	}

	testDirs := []string{
		"1", "2", "3", "4", "5",
		"some", "thing", "_random_",
		"_",
	}

	defer os.RemoveAll(tmpDir)
	for _, p := range testDirs {
		testdir := filepath.FromSlash(path.Join(tmpDir, p))
		err = mkdirs([]string{testdir})
		if err != nil {
			t.Fatal(err)
		}

		if s, err := os.Stat(testdir); err != nil {
			t.Fatal(err)
		} else if s.Mode()&os.ModeDir != os.ModeDir {
			t.Errorf("Invalid directory mode: %v", s.Mode())
		}
	}
}