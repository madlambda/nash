package assert

import (
	"strings"
	"testing"
)

func EqualStrings(t *testing.T, want string, got string) {
	// TODO: could use t.Helper here, but only on Go 1.9
	if want != got {
		t.Fatalf("wanted[%s] but got[%s]", want, got)
	}
}

func ContainsString(t *testing.T, str string, sub string) {
	// TODO: could use t.Helper here, but only on Go 1.9
	if !strings.Contains(str, sub) {
		t.Fatalf("[%s] is not a substring of [%s]", sub, str)
	}
}
