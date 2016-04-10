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
	rfargs := make([]Arg, 1)
	rfargs[0] = NewArg(6, "unp")
	
	ln.Push(NewCommandNode(0, "rfork", rfargs))

	tr.Root = ln
}
