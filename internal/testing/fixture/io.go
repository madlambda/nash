package fixture

import (
	"testing"
	"io/ioutil"
	"path/filepath"
	"math/rand"
	"os"
	"fmt"
)

// Tmpdir creates a temporary dir and returns a function that can be used
// to remove it after usage. Any error on any operation returns on a Fatal
// call on the given testing.T.
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

// MkdirAll will do the same thing as os.Mkdirall but calling Fatal on
// the given testing.T if something goes wrong.
func MkdirAll(t *testing.T, nashlib string) {
	t.Helper()
	
	err := os.MkdirAll(nashlib, os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
}

// CreateFiles will create all files and its dirs if
// necessary calling Fatal on the given testing if anything goes wrong.
//
// The files contents will be randomly generated strings (not a lot random,
// just for test purposes) and will be returned on the map that will map
// the filepath to its contents
func CreateFiles(t *testing.T, filepaths []string) map[string]string {
	t.Helper()
	
	createdFiles := map[string]string{}
	
	for _, f := range filepaths {
		contents := CreateFile(t, f)	
		createdFiles[f] = contents
	}
	
	return createdFiles
}

// CreateFile will create the file and its dirs if
// necessary calling Fatal on the given testing if anything goes wrong.
//
// The file content will be randomly generated strings (not a lot random,
// just for test purposes) and will be returned on the map that will map
// the filepath to its contents.
//
// Return the contents generated for the file (and that has been written on it).
func CreateFile(t *testing.T, f string) string {
	t.Helper()
	
	dir := filepath.Dir(f)
	MkdirAll(t, dir)
		
	contents := fmt.Sprintf("randomContents=%d", rand.Int())
		
	err := ioutil.WriteFile(f, []byte(contents), 0644)
	if err != nil {
		t.Fatalf("error[%s] writing file[%s]", err, f)
	}
		
	return contents
}