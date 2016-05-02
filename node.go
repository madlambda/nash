package nash

import (
	"errors"
	"os"
	"strconv"
	"strings"
)

const (
	redirMapNoValue int = -1
	redirMapSupress     = -2
)

type (
	// Node represents nodes in the grammar
	Node interface {
		Type() NodeType
		Position() Pos
		Tree() *Tree
		String() string
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

	ElemNode struct {
		NodeType
		Pos
		concats []string
		elem    string
	}

	// AssignmentNode is a node for variable assignments
	AssignmentNode struct {
		NodeType
		Pos
		name string
		list []ElemNode
	}

	// CommandNode is a node for commands
	CommandNode struct {
		NodeType
		Pos
		name   string
		args   []Arg
		redirs []*RedirectNode
	}

	// Arg is a command argument
	Arg struct {
		NodeType
		Pos
		val    string
		quoted bool
	}

	// RedirectNode represents the output redirection part of a command
	RedirectNode struct {
		NodeType
		Pos
		rmap     RedirMap
		location string
	}

	// RforkNode is a builtin node for rfork
	RforkNode struct {
		NodeType
		Pos
		arg  Arg
		tree *Tree
	}

	// CdNode is a builtin node for change directories
	CdNode struct {
		NodeType
		Pos
		dir  Arg
		Home bool
	}

	// CommentNode is the node for comments
	CommentNode struct {
		NodeType
		Pos
		val string
	}

	// RedirMap is the map of file descriptors of the redirection
	RedirMap struct {
		lfd int
		rfd int
	}
)

//go:generate stringer -type=NodeType

const (
	// NodeAssignment are nodes for variable assignment
	NodeAssignment NodeType = iota + 1

	// NodeCommand are command statements
	NodeCommand

	// NodeArg are nodes for command arguments
	NodeArg

	// NodeElem represents one element of a list
	NodeElem

	// NodeString are nodes for argument strings
	NodeString

	// NodeRfork are nodes for rfork command
	NodeRfork

	// NodeCd are nodes of builtin cd
	NodeCd

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

// NewAssignmentNode creates a new assignment
func NewAssignmentNode(pos Pos) *AssignmentNode {
	return &AssignmentNode{
		NodeType: NodeAssignment,
		Pos:      pos,
	}
}

// Tree returns the tree for this node
func (n *AssignmentNode) Tree() *Tree { return nil }

// SetVarName sets the name of the variable
func (n *AssignmentNode) SetVarName(a string) {
	n.name = a
}

// SetValueList sets the value of the variable
func (n *AssignmentNode) SetValueList(alist []ElemNode) {
	n.list = alist
}

func (n *AssignmentNode) String() string {
	ret := n.name + "="

	if len(n.list) == 1 {
		elem := n.list[0]

		if len(elem.concats) == 0 {
			if len(elem.elem) > 0 && elem.elem[0] == '$' {
				return n.name + `=` + elem.elem
			}

			return n.name + `="` + elem.elem + `"`
		}

		for i := 0; i < len(elem.concats); i++ {
			e := elem.concats[i]

			if len(e) > 0 && e[0] == '$' {
				ret += e
			} else {
				ret += `"` + e + `"`
			}

			if i < (len(elem.concats) - 1) {
				ret += " + "
			}
		}

		return ret
	} else if len(n.list) == 0 {
		return n.name + `=""`
	}

	ret += "("

	for i := 0; i < len(n.list); i++ {
		ret += n.list[i].elem

		if i < len(n.list)-1 {
			ret += " "
		}
	}

	ret += ")"
	return ret
}

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

// AddRedirect adds a new redirect node to command
func (n *CommandNode) AddRedirect(redir *RedirectNode) {
	n.redirs = append(n.redirs, redir)
}

// Tree returns the child tree of node
func (n *CommandNode) Tree() *Tree { return nil }

func (n *CommandNode) String() string {
	content := make([]string, 0, 1024)
	args := make([]string, 0, len(n.args))
	redirs := make([]string, 0, len(n.redirs))

	for i := 0; i < len(n.args); i++ {
		args = append(args, n.args[i].String())
	}

	for i := 0; i < len(n.redirs); i++ {
		redirs = append(redirs, n.redirs[i].String())
	}

	content = append(content, n.name)
	content = append(content, args...)
	content = append(content, redirs...)

	return strings.Join(content, " ")
}

// NewRedirectNode creates a new redirection node for commands
func NewRedirectNode(pos Pos) *RedirectNode {
	return &RedirectNode{
		rmap: RedirMap{
			lfd: -1,
			rfd: -1,
		},
		location: "",
	}
}

// SetMap sets the redirection map. Eg.: [2=1]
func (r *RedirectNode) SetMap(lfd int, rfd int) {
	r.rmap.lfd = lfd
	r.rmap.rfd = rfd
}

// SetLocation of the output
func (r *RedirectNode) SetLocation(s string) {
	r.location = s
}

func (r *RedirectNode) String() string {
	var result string

	if r.rmap.lfd == r.rmap.rfd {
		if r.location != "" {
			return "> " + r.location
		}

		return ""
	}

	if r.rmap.rfd >= 0 {
		result = ">[" + strconv.Itoa(r.rmap.lfd) + "=" + strconv.Itoa(r.rmap.rfd) + "]"
	} else if r.rmap.rfd == redirMapNoValue {
		result = ">[" + strconv.Itoa(r.rmap.lfd) + "]"
	} else if r.rmap.rfd == redirMapSupress {
		result = ">[" + strconv.Itoa(r.rmap.lfd) + "=]"
	}

	if r.location != "" {
		result = result + " " + r.location
	}

	return result
}

// Tree returns the tree of the node
func (r *RedirectNode) Tree() *Tree { return nil }

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

// SetBlock sets the sub block of rfork
func (n *RforkNode) SetBlock(t *Tree) {
	n.tree = t
}

// Tree returns the child tree of node
func (n *RforkNode) Tree() *Tree {
	return n.tree
}

func (n *RforkNode) String() string {
	rforkstr := "rfork " + n.arg.val
	tree := n.Tree()

	if tree != nil {
		rforkstr += " {\n"
		block := tree.String()
		stmts := strings.Split(block, "\n")

		for i := 0; i < len(stmts); i++ {
			stmts[i] = "\t" + stmts[i]
		}

		rforkstr += strings.Join(stmts, "\n") + "\n}"
	}

	return rforkstr
}

// NewCdNode creates a new node for changing directory
func NewCdNode(pos Pos) *CdNode {
	return &CdNode{
		NodeType: NodeCd,
		Pos:      pos,
	}
}

// SetHome sets the directory as $home
func (n *CdNode) SetHome() {
	n.Home = true
}

// SetDir sets the cd directory to dir
func (n *CdNode) SetDir(dir Arg) {
	n.dir = dir
}

// Dir returns the directory of cd node
func (n *CdNode) Dir() (string, error) {
	if n.Home {
		homePath := os.Getenv("$home")

		if homePath == "" {
			homePath = os.Getenv("$HOME")

			if homePath == "" {
				return "", errors.New("No variable $home or $HOME set")
			}
		}

		return homePath, nil
	}

	return n.dir.val, nil
}

// Tree returns the child tree if any
func (n *CdNode) Tree() *Tree {
	return nil
}

func (n *CdNode) String() string {
	if n.Home {
		return "cd"
	}

	if n.dir.quoted {
		return `cd "` + n.dir.val + `"`
	}

	return "cd " + n.dir.val
}

// NewArg creates a new argument
func NewArg(pos Pos, val string, quoted bool) Arg {
	return Arg{
		NodeType: NodeArg,
		Pos:      pos,
		val:      val,
		quoted:   quoted,
	}
}

// Tree returns the child tree of node
func (n Arg) Tree() *Tree { return nil }

func (n Arg) String() string {
	if n.quoted {
		return "\"" + n.val + "\""
	}

	return n.val
}

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

func (n *CommentNode) String() string {
	return n.val
}
