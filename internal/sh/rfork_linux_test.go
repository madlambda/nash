// +build linux

package sh

import (
	"bytes"
	"strings"
	"syscall"
	"testing"
)

func getletters() string {
	var a bytes.Buffer

	for i := 'a'; i < 'z'; i++ {
		a.Write(append([]byte{}, byte(i)))
	}

	all := string(a.Bytes())
	allCap := strings.ToUpper(all)
	return all + allCap
}

func getvalid() string {
	return "cumnpsi"
}

func testTblFlagsOK(flagstr string, expected uintptr, t *testing.T) {
	flags, err := getflags(flagstr)

	if err != nil {
		t.Error(err)
		return
	}

	if flags != expected {
		t.Errorf("Flags differ: expected %08x but %08x", expected, flags)
		return
	}
}

func TestRforkFlags(t *testing.T) {
	_, err := getflags("")

	if err == nil {
		t.Error("Empty flags should return error")
		return
	}

	_, err = getflags("a")

	if err == nil {
		t.Error("Unknow flag a")
		return
	}

	allchars := getletters()

	_, err = getflags(allchars)

	if err == nil {
		t.Error("Should fail")
		return
	}

	testTblFlagsOK("u", syscall.CLONE_NEWUSER, t)
	testTblFlagsOK("m", syscall.CLONE_NEWNS, t)
	testTblFlagsOK("n", syscall.CLONE_NEWNET, t)
	testTblFlagsOK("i", syscall.CLONE_NEWIPC, t)
	testTblFlagsOK("s", syscall.CLONE_NEWUTS, t)
	testTblFlagsOK("p", syscall.CLONE_NEWPID, t)
	testTblFlagsOK("c", syscall.CLONE_NEWUSER|
		syscall.CLONE_NEWNS|syscall.CLONE_NEWNET|
		syscall.CLONE_NEWIPC|syscall.CLONE_NEWUTS|
		syscall.CLONE_NEWUSER|syscall.CLONE_NEWPID, t)
	testTblFlagsOK("um", syscall.CLONE_NEWUSER|syscall.CLONE_NEWNS, t)
	testTblFlagsOK("umn", syscall.CLONE_NEWUSER|
		syscall.CLONE_NEWNS|
		syscall.CLONE_NEWNET, t)
	testTblFlagsOK("umni", syscall.CLONE_NEWUSER|
		syscall.CLONE_NEWNS|
		syscall.CLONE_NEWNET|
		syscall.CLONE_NEWIPC, t)
	testTblFlagsOK("umnip", syscall.CLONE_NEWUSER|
		syscall.CLONE_NEWNS|
		syscall.CLONE_NEWNET|
		syscall.CLONE_NEWIPC|
		syscall.CLONE_NEWPID, t)
	testTblFlagsOK("umnips", syscall.CLONE_NEWUSER|
		syscall.CLONE_NEWNS|
		syscall.CLONE_NEWNET|
		syscall.CLONE_NEWIPC|
		syscall.CLONE_NEWPID|
		syscall.CLONE_NEWUTS, t)
}
