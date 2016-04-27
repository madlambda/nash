package nash

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"

	"golang.org/x/exp/ebnf"
)

func TestSpecificationIsSane(t *testing.T) {
	filename := os.Getenv("GOPATH") + "/src/github.com/NeowayLabs/nash/spec.ebnf"
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
