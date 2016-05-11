package nash

import "testing"

// Test API
func TestTreeCreation(t *testing.T) {
	tr := NewTree("example")

	if tr.Name != "example" {
		t.Errorf("Invalid name")
		return
	}
}

func TestTreeRawCreation(t *testing.T) {
	tr := NewTree("creating a tree by hand")

	ln := NewListNode()
	rfarg := newSimpleArg(6, "unp", false)

	r := NewRforkNode(0)
	r.SetFlags(rfarg)
	ln.Push(r)

	tr.Root = ln

	if tr.String() != "rfork unp" {
		t.Error("Failed to build AST by hand")
	}
}
