package main

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/chzyer/readline"
)

func TestCompleteAbsolutePaths(t *testing.T) {
	var rCompleters = []readline.PrefixCompleterInterface{}
	c := NewCompleter(rCompleters...)

	f, err := ioutil.TempFile("/tmp", "test-nash-completers")

	if err != nil {
		t.Error(err)
		return
	}

	f.Close()

	defer os.Remove(f.Name())

	completeStr := []rune("/tmp/test-nash-comp")
	newLines, offset := c.Do(completeStr, len(completeStr))

	if offset != len(completeStr) {
		t.Errorf("Invalid offset: %d. Expected %d", offset, len(completeStr))
		return
	}

	if len(newLines) != 1 {
		t.Errorf("Expected only 1 complete file")
		return
	}

	if len(f.Name()) <= len(completeStr) {
		t.Errorf("Expected '%s' but got '%s'", f.Name()[len(completeStr):], newLines[0])
		return
	}

	if string(newLines[0]) != f.Name()[len(completeStr):] {
		t.Errorf("Expected '%s' but got '%s'", f.Name()[len(completeStr):], newLines[0])
		return
	}
}
