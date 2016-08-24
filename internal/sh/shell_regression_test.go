package sh

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func TestExecuteIssue68(t *testing.T) {
	sh, err := NewShell()

	if err != nil {
		t.Error(err)
		return
	}

	err = sh.Exec("-input-", `echo lalalala | grep la > /tmp/la`)

	if err != nil {
		t.Error(err)
		return
	}

	defer os.Remove("/tmp/la")

	contents, err := ioutil.ReadFile("/tmp/la")

	if err != nil {
		t.Error(err)
		return
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
