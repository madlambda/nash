package sh

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"path"
	"fmt"
)

func TestExecuteIssue68(t *testing.T) {
	sh, err := NewShell()
	if err != nil {
		t.Error(err)
		return
	}

	tmpDir, err := ioutil.TempDir("", "nash-tests")
	if err != nil {
		t.Fatal(err)
	}

	file := path.Join(tmpDir, "la")
	err = sh.Exec("-input-", fmt.Sprintf(`echo lalalala | grep la > %s`, file))
	if err != nil {
		t.Error(err)
		return
	}

	defer os.Remove(file)

	contents, err := ioutil.ReadFile(file)

	if err != nil {
		t.Fatal(err)
	}

	contentStr := strings.TrimSpace(string(contents))
	if contentStr != "lalalala" {
		t.Errorf("Strings differ: '%s' != '%s'", contentStr, "lalalala")
		return
	}
}

func TestExecuteErrorSuppression(t *testing.T) {
	sh, err := NewShell()

	if err != nil {
		t.Error(err)
	}

	err = sh.Exec("-input-", `-bllsdlfjlsd`)

	if err != nil {
		t.Errorf("Expected to not fail...: %s", err.Error())
		return
	}

	// issue #72
	err = sh.Exec("-input-", `echo lalala | -grep lelele`)

	if err != nil {
		t.Errorf("Expected to not fail...:(%s)", err.Error())
		return
	}
}
