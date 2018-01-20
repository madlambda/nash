package strings_test

import (
	"bytes"
	"testing"

	"github.com/NeowayLabs/nash/stdbin/strings"
)

func TestBinaryStartingWithText(t *testing.T) {
	expected := "textOnBeggining"
	bin := newBinary(512)
	input := append([]byte(expected), bin...)

	scanner := strings.Do(bytes.NewBuffer(input))
	assertTrue(t, scanner.Scan(), "expected to have data on Scan, none found")
	assertStrings(t, expected, scanner.Text())
	assertFalse(t, scanner.Scan(), "expected to have no data on Scan, found some")
}

func TestBinaryEndingWithText(t *testing.T) {
}

func TestBinaryWithTextOnMiddle(t *testing.T) {
}

func TestMinTextSizeIsAdjustable(t *testing.T) {
}

func TestEachTextOccurenceIsANewLine(t *testing.T) {
}

func TestJustText(t *testing.T) {
}

func TestJustBinary(t *testing.T) {
}

func newBinary(size uint) []byte {
	// TODO: Not the most awesome random binary =/
	bin := make([]byte, size)
	for i := 0; i < int(size); i++ {
		bin[i] = 0xFF
	}
	return bin
}

// TODO: Start an assert package on nash or use our other project ?
func assertTrue(t *testing.T, b bool, msg string) {
	t.Helper()
	if !b {
		t.Fatalf("want true, got false: %s", msg)
	}
}

func assertFalse(t *testing.T, b bool, msg string) {
	t.Helper()
	if b {
		t.Fatalf("want false, got true: %s", msg)
	}
}

func assertStrings(t *testing.T, want string, got string) {
	t.Helper()
	if want != got {
		t.Fatalf("want[%s] != got[%s]", want, got)
	}
}
