package nash

import "testing"

func TestExecuteIssue68(t *testing.T) {
	sh, err := NewShell()

	if err != nil {
		t.Error(err)
		return
	}

}
