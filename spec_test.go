package nash

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"
	"os/user"
	"path"
	"path/filepath"

	"golang.org/x/exp/ebnf"
)

func TestSpecificationIsSane(t *testing.T) {
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		usr, err := user.Current()
		if err != nil {
			t.Fatal(err)
		}
		if usr.HomeDir == "" {
			t.Fatal("Unable to discover GOPATH")	
		}
		gopath = path.Join(usr.HomeDir, "go")
	}
	filename := filepath.Join(gopath, filepath.FromSlash("/src/github.com/NeowayLabs/nash/spec.ebnf"))
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		t.Error(err)
		return
	}

	buf := bytes.NewBuffer(content)

	grammar, err := ebnf.Parse(filename, buf)

	if err != nil {
		t.Error(err)
		return
	}

	err = ebnf.Verify(grammar, "program")

	if err != nil {
		t.Error(err)
		return
	}
}
