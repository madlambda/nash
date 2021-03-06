package nash

import (
	"bytes"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/madlambda/nash/tests"
	"golang.org/x/exp/ebnf"
)

func TestSpecificationIsSane(t *testing.T) {
	filename := filepath.Join(tests.Projectpath, "spec.ebnf")
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
