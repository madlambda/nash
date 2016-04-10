package cnt

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
	rfarg := NewArg(6, "unp")

	r := NewRforkNode(0)
	r.SetFlags(rfarg)
	ln.Push(r)

	tr.Root = ln
}
