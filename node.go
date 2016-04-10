package cnt

type (
	Node interface {
		Type() NodeType
		Position() Pos
		tree() *Tree
	}

	NodeType int

	ListNode struct {
		NodeType
		Pos
		tr *Tree
		Nodes []Node
	}

	Pos int

	// cnt node types
	CommandNode struct {
		NodeType
		Pos
		tr *Tree
		name string
		args []Arg
	}

	Arg struct {
		Pos
		val string
	}
)

const (
	NodeCommand NodeType = iota
	NodeArg
	NodeString
	NodeRfork
	NodeRforkFlags
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
		Pos: pos,
		name: name,
		args: make([]Arg, 0, 1024),
	}
}

func (n *CommandNode) AddArg(a Arg) {
	n.args = append(n.args, a)
}

func (n *CommandNode) SetArgs(args []Arg) {
	n.args = args
}

func NewArg(pos Pos, val string) Arg {
	return Arg{
		Pos: pos,
		val: val,
	}
}

func (c *CommandNode) tree() *Tree {
	return c.tr
}
