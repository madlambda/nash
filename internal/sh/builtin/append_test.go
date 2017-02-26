package builtin_test

import (
	"bytes"
	"testing"

	"github.com/NeowayLabs/nash/internal/sh"
)

func TestAppend(t *testing.T) {
	sh, err := sh.NewShell()

	if err != nil {
		t.Error(err)
		return
	}

	var out bytes.Buffer

	sh.SetStdout(&out)

	err = sh.Exec(
		"test append",
		`a = ()
		 a <= append($a, "hello")
		 a <= append($a, "world")
		 echo -n $a`,
	)

	if err != nil {
		t.Error(err)
		return
	}

	if "hello world" != string(out.Bytes()) {
		t.Errorf("String differs: '%s' != '%s'", "hello world", string(out.Bytes()))
		return
	}

	err = sh.Exec(
		"test append fail",
		`a = "something"
		 a <= append($a, "other")
		 echo -n $a`,
	)

	if err == nil {
		t.Errorf("Must fail... Append only should works with lists")
		return
	}
}
