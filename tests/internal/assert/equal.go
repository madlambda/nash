package assert

import "testing"

func EqualStrings(t *testing.T, want string, got string) {
	// TODO: could use t.Helper here, but only on Go 1.9
	if want != got {
		t.Fatalf("wanted[%s] but got[%s]", want, got)
	}
}
