package ast

import "strings"

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

	nodes := tree.Root.Nodes

	content := make([]string, 0, 8192)

	for i := 0; i < len(nodes); i++ {
		addEOL := false
		node := nodes[i]

		nodebytes := node.String()

		if i == 0 && node.Type() == NodeComment && strings.HasPrefix(node.String(), "#!") {
			nodebytes += "\n"
		}

		if (node.Type() == NodeAssignment) && i < (len(nodes)-1) {
			nextNode := nodes[i+1]

			switch nextNode.Type() {
			case NodeComment, NodeFnDecl:
				addEOL = true
			}
		}

		if addEOL {
			nodebytes += "\n"
		}

		content = append(content, nodebytes)
	}

	return strings.Join(content, "\n")
}
