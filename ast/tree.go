package ast

type (
	// Tree is the AST
	Tree struct {
		Name string
		Root *BlockNode // top-level root of the tree.
	}
)

// NewTree creates a new AST tree
func NewTree(name string) *Tree {
	return &Tree{
		Name: name,
	}
}

func (t *Tree) IsEqual(other *Tree) bool {
	if t == other {
		return true
	}

	return t.Root.IsEqual(other.Root)
}

func (tree *Tree) String() string {
	if tree.Root == nil {
		return ""
	}

	if len(tree.Root.Nodes) == 0 {
		return ""
	}

	return tree.Root.String()
}
