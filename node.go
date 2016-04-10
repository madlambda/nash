package cnt

type (
	Node interface {
		Type() NodeType
		Position() Pos
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

func NewRforkNode(pos Pos) *RforkNode {
	return &RforkNode{
		NodeType: NodeRfork,
		Pos: pos,
	}
}

func (n *RforkNode) SetFlags(a Arg) {
	n.arg = a
}

func NewArg(pos Pos, val string) Arg {
	return Arg{
		NodeType: NodeArg,
		Pos: pos,
		val: val,
	}
}

func NewCommentNode(pos Pos, val string) *CommentNode {
	return &CommentNode{
		NodeType: NodeComment,
		Pos: pos,
		val: val,
	}
}
