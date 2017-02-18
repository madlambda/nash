package sh

import (
	"bytes"
	"testing"
)

func TestExecuteBuiltinLen(t *testing.T) {
	sh, err := NewShell()

	if err != nil {
		t.Error(err)
		return
	}

	var out bytes.Buffer

	sh.SetStdout(&out)

	err = sh.Exec("test len", `a = (1 2 3 4 5 6 7 8 9 0)
len_a <= len($a)
echo -n $len_a`)

	if err != nil {
		t.Error(err)
		return
	}

	if "10" != string(out.Bytes()) {
		t.Errorf("String differs: '%s' != '%s'", "10", string(out.Bytes()))
		return
	}

	out.Reset()

	err = sh.Exec("test len fail", `a = "test"
l <= len($a)
echo -n $l
`)

	if err != nil {
		t.Errorf("Must fail... Len only should work= with lists")
		return
	}

	if "4" != string(out.Bytes()) {
		t.Errorf("String differs: '%s' != '%s'", "4", string(out.Bytes()))
		return
	}
}

func TestExecuteBuiltinAppend(t *testing.T) {
	sh, err := NewShell()

	if err != nil {
		t.Error(err)
		return
	}

	var out bytes.Buffer

	sh.SetStdout(&out)

	err = sh.Exec("test append", `a = ()
a <= append($a, "hello")
a <= append($a, "world")
echo -n $a`)

	if err != nil {
		t.Error(err)
		return
	}

	if "hello world" != string(out.Bytes()) {
		t.Errorf("String differs: '%s' != '%s'", "hello world", string(out.Bytes()))
		return
	}

	err = sh.Exec("test append fail", `a = "something"
a <= append($a, "other")
echo -n $a`)

	if err == nil {
		t.Errorf("Must fail... Append only should works with lists")
		return
	}

}
