package sh_test

import (
	"os"
	"testing"
	"io/ioutil"
)

func writeFile(t *testing.T, filename string, data string) {
	err := ioutil.WriteFile(filename, []byte(data), os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
}

func chdir(t *testing.T, dir string) {
	t.Helper()
	
	err := os.Chdir(dir)
	if err != nil {
		t.Fatal(err)
	}
}

func getwd(t *testing.T) string {
	t.Helper()
	
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	
	return dir
}