package cnt

type (
	Node interface {
		Type() NodeType
		Position() Pos
		Tree() *Tree
	}

	NodeType int

	ListNode struct {
		NodeType
		Pos
		Nodes []Node
	}

	Pos int

	// cnt node types
	CommandNode struct {
		NodeType
		Pos
		name string
		args []Arg
	}

	Arg struct {
		NodeType
		Pos
		val string
	}

	RforkNode struct {
		NodeType
		Pos
		arg Arg
		tree *Tree
	}

	CommentNode struct {
		NodeType
		Pos
		val string
	}
		
)

const (
	NodeCommand NodeType = iota + 1
	NodeArg
	NodeString
	NodeRfork
	NodeRforkFlags
	NodeComment
)

func (p Pos) Position() Pos {
	return p
}

func (t NodeType) Type() NodeType {
	return t
}

func NewListNode() *ListNode {
	return &ListNode{
		
	}
}

func (l *ListNode) Push(n Node) {
	if l.Nodes == nil {
		l.Nodes = make([]Node, 0, 1024)
	}

	l.Nodes = append(l.Nodes, n)
}

func (l *ListNode) Tree() *Tree { return nil }

func NewCommandNode(pos Pos, name string) *CommandNode {
	return &CommandNode{
		NodeType: NodeCommand,
		Pos: pos,
		name: name,
		args: make([]Arg, 0, 1024),
	}
}

func (n *CommandNode) Nodes() []Node {
	return make([]Node, 0, 0)
}

func (n *CommandNode) AddArg(a Arg) {
	n.args = append(n.args, a)
}

func (n *CommandNode) SetArgs(args []Arg) {
	n.args = args
}

func (n *CommandNode) Tree() *Tree { return nil }

func NewRforkNode(pos Pos) *RforkNode {
	return &RforkNode{
		NodeType: NodeRfork,
		Pos: pos,
	}
}

func (n *RforkNode) SetFlags(a Arg) {
	n.arg = a
}

func (n *RforkNode) Tree() *Tree {
	return n.tree
}

func NewArg(pos Pos, val string) Arg {
	return Arg{
		NodeType: NodeArg,
		Pos: pos,
		val: val,
	}
}

func (n Arg) Tree() *Tree { return nil }

func NewCommentNode(pos Pos, val string) *CommentNode {
	return &CommentNode{
		NodeType: NodeComment,
		Pos: pos,
		val: val,
	}
}

func (n *CommentNode) Tree() *Tree { return nil }
