package nash

import "strings"

type (
	// Tree is the AST
	Tree struct {
		Name string
		Root *ListNode // top-level root of the tree.
	}
)

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
		node := nodes[i]
		nodebytes := node.String()

		content = append(content, nodebytes)
	}

	return strings.Join(content, "\n")
}
