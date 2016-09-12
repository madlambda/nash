package ast

import (
	"testing"

	"github.com/NeowayLabs/nash/token"
)

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

	ln := NewBlockNode(token.NewFileInfo(1, 0))
	rfarg := NewStringExpr(token.NewFileInfo(1, 0), "unp", false)

	r := NewRforkNode(token.NewFileInfo(1, 0))
	r.SetFlags(rfarg)
	ln.Push(r)

	tr.Root = ln

	if tr.String() != "rfork unp" {
		t.Error("Failed to build AST by hand")
	}
}
