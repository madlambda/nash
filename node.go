package cnt

type (
	// Node represents nodes in the grammar
	Node interface {
		Type() NodeType
		Position() Pos
		Tree() *Tree
	}

	// NodeType is the types of grammar
	NodeType int

	// ListNode is the block
	ListNode struct {
		NodeType
		Pos
		Nodes []Node
	}

	// Pos is the position of a node in file
	Pos int

	// CommandNode is a node for commands
	CommandNode struct {
		NodeType
		Pos
		name string
		args []Arg
	}

	// Arg is a command argument
	Arg struct {
		NodeType
		Pos
		val string
	}

	// RforkNode is a node for rfork
	RforkNode struct {
		NodeType
		Pos
		arg  Arg
		tree *Tree
	}

	// CommentNode is the node for comments
	CommentNode struct {
		NodeType
		Pos
		val string
	}
)

const (
	// NodeCommand are command statements
	NodeCommand NodeType = iota + 1

	// NodeArg are nodes for command arguments
	NodeArg

	// NodeString are nodes for argument strings
	NodeString

	// NodeRfork are nodes for rfork command
	NodeRfork

	// NodeRforkFlags are nodes rfork flags
	NodeRforkFlags

	// NodeComment are nodes for comment
	NodeComment
)

// Position returns the position of the node in file
func (p Pos) Position() Pos {
	return p
}

// Type returns the type of the node
func (t NodeType) Type() NodeType {
	return t
}

// NewListNode creates a new block
func NewListNode() *ListNode {
	return &ListNode{}
}

// Push adds a new node for a block of nodes
func (l *ListNode) Push(n Node) {
	if l.Nodes == nil {
		l.Nodes = make([]Node, 0, 1024)
	}

	l.Nodes = append(l.Nodes, n)
}

// Tree returns the tree for this node
func (l *ListNode) Tree() *Tree { return nil }

// NewCommandNode creates a new node for commands
func NewCommandNode(pos Pos, name string) *CommandNode {
	return &CommandNode{
		NodeType: NodeCommand,
		Pos:      pos,
		name:     name,
		args:     make([]Arg, 0, 1024),
	}
}

// AddArg adds a new argument to the command
func (n *CommandNode) AddArg(a Arg) {
	n.args = append(n.args, a)
}

// SetArgs sets an array of args to command
func (n *CommandNode) SetArgs(args []Arg) {
	n.args = args
}

// Tree returns the child tree of node
func (n *CommandNode) Tree() *Tree { return nil }

// NewRforkNode creates a new node for rfork
func NewRforkNode(pos Pos) *RforkNode {
	return &RforkNode{
		NodeType: NodeRfork,
		Pos:      pos,
	}
}

// SetFlags sets the rfork flags
func (n *RforkNode) SetFlags(a Arg) {
	n.arg = a
}

// Tree returns the child tree of node
func (n *RforkNode) Tree() *Tree {
	return n.tree
}

// NewArg creates a new argument
func NewArg(pos Pos, val string) Arg {
	return Arg{
		NodeType: NodeArg,
		Pos:      pos,
		val:      val,
	}
}

// Tree returns the child tree of node
func (n Arg) Tree() *Tree { return nil }

// NewCommentNode creates a new node for comments
func NewCommentNode(pos Pos, val string) *CommentNode {
	return &CommentNode{
		NodeType: NodeComment,
		Pos:      pos,
		val:      val,
	}
}

// Tree returns the child tree of node
func (n *CommentNode) Tree() *Tree { return nil }
