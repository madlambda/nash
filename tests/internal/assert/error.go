package assert

import "testing"

func NoError(t *testing.T, err error, operation string) {
	if err != nil {
		t.Fatalf("error[%s] %s", err, operation)
	}
}
