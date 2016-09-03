package ast

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/NeowayLabs/nash/token"
)

const (
	// RedirMapNoValue indicates the pipe has not redirection
	RedirMapNoValue = -1
	// RedirMapSupress indicates the rhs of map was suppressed
	RedirMapSupress = -2

	RforkFlags = "umnips"
)

type (
	// Node represents nodes in the grammar
	Node interface {
		Type() NodeType
		Position() token.Pos
		String() string
		IsEqual(Node) bool
	}

	// Expr is the interface of expression nodes.
	Expr Node

	// NodeType is the types of grammar
	NodeType int

	// ListNode is the block
	ListNode struct {
		NodeType
		token.Pos

		Nodes []Node
	}

	// An ImportNode represents the node for an "import" keyword.
	ImportNode struct {
		NodeType
		token.Pos

		path *StringExpr // Import path
	}

	// A SetenvNode represents the node for a "setenv" keyword.
	SetenvNode struct {
		NodeType
		token.Pos

		varName string
	}

	// AssignmentNode is a node for variable assignments
	AssignmentNode struct {
		NodeType
		token.Pos

		name string
		val  Expr
	}

	// An ExecAssignNode represents the node for execution assignment.
	ExecAssignNode struct {
		NodeType
		token.Pos

		name string
		cmd  Node
	}

	// A CommandNode is a node for commands
	CommandNode struct {
		NodeType
		token.Pos

		name   string
		args   []Expr
		redirs []*RedirectNode
	}

	// PipeNode represents the node for a command pipeline.
	PipeNode struct {
		NodeType
		token.Pos

		cmds []*CommandNode
	}

	// StringExpr is a string argument
	StringExpr struct {
		NodeType
		token.Pos

		str    string
		quoted bool
	}

	// IntExpr is a integer used at indexing
	IntExpr struct {
		NodeType
		token.Pos

		val int
	}

	// ListExpr is a list argument
	ListExpr struct {
		NodeType
		token.Pos

		list []Expr
	}

	// ConcatExpr is a concatenation of arguments
	ConcatExpr struct {
		NodeType
		token.Pos

		concat []Expr
	}

	// VarExpr is a variable argument
	VarExpr struct {
		NodeType
		token.Pos

		name string
	}

	// IndexExpr is a indexed variable
	IndexExpr struct {
		NodeType
		token.Pos

		variable *VarExpr
		index    Expr
	}

	// RedirectNode represents the output redirection part of a command
	RedirectNode struct {
		NodeType
		token.Pos
		rmap     RedirMap
		location Expr
	}

	// RforkNode is a builtin node for rfork
	RforkNode struct {
		NodeType
		token.Pos
		arg  *StringExpr
		tree *Tree
	}

	// CdNode is a builtin node for change directories
	CdNode struct {
		NodeType
		token.Pos
		dir Expr
	}

	// CommentNode is the node for comments
	CommentNode struct {
		NodeType
		token.Pos
		val string
	}

	// RedirMap is the map of file descriptors of the redirection
	RedirMap struct {
		lfd int
		rfd int
	}

	// IfNode represents the node for the "if" keyword.
	IfNode struct {
		NodeType
		token.Pos
		lvalue Expr
		rvalue Expr
		op     string
		elseIf bool

		ifTree   *Tree
		elseTree *Tree
	}

	// A FnDeclNode represents a function declaration.
	FnDeclNode struct {
		NodeType
		token.Pos
		name string
		args []string
		tree *Tree
	}

	// A FnInvNode represents a function invocation statement.
	FnInvNode struct {
		NodeType
		token.Pos
		name string
		args []Expr
	}

	// A ReturnNode represents the "return" keyword.
	ReturnNode struct {
		NodeType
		token.Pos
		arg Expr
	}

	// A BindFnNode represents the "bindfn" keyword.
	BindFnNode struct {
		NodeType
		token.Pos
		name    string
		cmdname string
	}

	// A DumpNode represents the "dump" keyword.
	DumpNode struct {
		NodeType
		token.Pos
		filename Expr
	}

	// A ForNode represents the "for" keyword.
	ForNode struct {
		NodeType
		token.Pos
		identifier string
		inVar      string
		tree       *Tree
	}
)

//go:generate stringer -type=NodeType

const (
	// NodeSetenv the type for "setenv" builtin keyword
	NodeSetenv NodeType = iota + 1

	// NodeAssignment is the type for simple variable assignment
	NodeAssignment

	// NodeExecAssign is the type for command or function assignment
	NodeExecAssign

	// NodeImport is the type for "import" builtin keyword
	NodeImport

	execBegin

	// NodeCommand is the type for command execution
	NodeCommand

	// NodePipe is the type for pipeline execution
	NodePipe

	// NodeFnInv is the type for function invocation
	NodeFnInv

	execEnd

	expressionBegin

	// NodeStringExpr is the type of string expression (quoted or not).
	NodeStringExpr

	// NodeIntExpr is the type of integer expression (commonly list indexing)
	NodeIntExpr

	// NodeVarExpr is the type of variable expressions.
	NodeVarExpr

	// NodeListExpr is the type of list expression.
	NodeListExpr

	// NodeIndexExpr is the type of indexing expressions.
	NodeIndexExpr

	// NodeConcatExpr is the type of concatenation expressions.
	NodeConcatExpr

	expressionEnd

	// NodeString are nodes for argument strings
	NodeString

	// NodeRfork is the type for rfork statement
	NodeRfork

	// NodeCd is the type for builtin cd
	NodeCd

	// NodeRforkFlags are nodes for rfork flags
	NodeRforkFlags

	// NodeIf is the type for if statements
	NodeIf

	// NodeComment are nodes for comment
	NodeComment

	// NodeFnDecl is the type for function declaration
	NodeFnDecl

	// NodeReturn is the type for return statement
	NodeReturn

	// NodeBindFn is the type for bindfn statements
	NodeBindFn

	// NodeDump is the type for dump statements
	NodeDump

	// NodeFor is the type for "for" statements
	NodeFor
)

var (
	DebugCmp bool
)

func debug(format string, args ...interface{}) {
	if DebugCmp {
		fmt.Printf("[debug] "+format+"\n", args...)
	}
}

// Type returns the type of the node
func (t NodeType) Type() NodeType {
	return t
}

// IsExpr returns if the node is an expression.
func (t NodeType) IsExpr() bool {
	return t > expressionBegin && t < expressionEnd
}

// IsExecutable returns if the node is executable
func (t NodeType) IsExecutable() bool {
	return t > execBegin && t < execEnd
}

// NewListNode creates a new block
func NewListNode() *ListNode {
	return &ListNode{}
}

// Push adds a new node for a block of nodes
func (l *ListNode) Push(n Node) {
	l.Nodes = append(l.Nodes, n)
}

func (l *ListNode) String() string {
	return "unimplemented"
}

// IsEqual returns if it is equal to the other node.
func (l *ListNode) IsEqual(other Node) bool {
	if l == other {
		return true
	}

	o, ok := other.(*ListNode)

	if !ok {
		debug("Failed to cast other node to ListNode")
		return false
	}

	if len(l.Nodes) != len(o.Nodes) {
		debug("Nodes differs in length")
		return false
	}

	for i := 0; i < len(l.Nodes); i++ {
		if !l.Nodes[i].IsEqual(o.Nodes[i]) {
			debug("List entry %d differ... '%s' != '%s'", i, l.Nodes[i], o.Nodes[i])
			return false
		}
	}

	return true
}

// NewImportNode creates a new ImportNode object
func NewImportNode(pos token.Pos, path *StringExpr) *ImportNode {
	return &ImportNode{
		NodeType: NodeImport,
		Pos:      pos,

		path: path,
	}
}

// Path returns the path of import.
func (n *ImportNode) Path() *StringExpr { return n.path }

// IsEqual returns if it is equal to the other node.
func (n *ImportNode) IsEqual(other Node) bool {
	if n == other {
		return true
	}

	o, ok := other.(*ImportNode)

	if !ok {
		debug("Failed to cast to ImportNode")
		return false
	}

	if n.path == o.path {
		return true
	}

	if n.path != nil {
		return n.path.IsEqual(o.path)
	}

	return false
}

// String returns the string representation of the import
func (n *ImportNode) String() string {
	return `import ` + n.path.String()
}

// NewSetenvNode creates a new assignment node
func NewSetenvNode(pos token.Pos, name string) *SetenvNode {
	return &SetenvNode{
		NodeType: NodeSetenv,
		Pos:      pos,
		varName:  name,
	}
}

// Identifier returns the environment name.
func (n *SetenvNode) Identifier() string { return n.varName }

// IsEqual returns if it is equal to the other node.
func (n *SetenvNode) IsEqual(other Node) bool {
	if n == other {
		return true
	}

	o, ok := other.(*SetenvNode)

	if !ok {
		debug("Failed to convert to SetenvNode")
		return false
	}

	return n.varName == o.varName
}

// String returns the string representation of assignment
func (n *SetenvNode) String() string {
	return "setenv " + n.varName
}

// NewAssignmentNode creates a new assignment
func NewAssignmentNode(pos token.Pos, ident string, value Expr) *AssignmentNode {
	return &AssignmentNode{
		NodeType: NodeAssignment,
		Pos:      pos,

		name: ident,
		val:  value,
	}
}

// SetIdentifier sets the name of the variable
func (n *AssignmentNode) SetIdentifier(a string) {
	n.name = a
}

// Identifier return the name of the variable.
func (n *AssignmentNode) Identifier() string { return n.name }

// SetValue sets the value of the list
func (n *AssignmentNode) SetValue(val Expr) {
	n.val = val
}

// Value returns the assigned object
func (n *AssignmentNode) Value() Expr {
	return n.val
}

// IsEqual returns if it is equal to the other node.
func (n *AssignmentNode) IsEqual(other Node) bool {
	if n == other {
		return true
	}

	o, ok := other.(*AssignmentNode)

	if !ok {
		debug("Failed to convert to AssignmentNode")
		return false
	}

	if n.name != o.name {
		debug("Assignment identifier doesn't match: '%s' != '%s'", n.name, o.name)
		return false
	}

	if n.val == o.val {
		return true
	}

	if n.val != nil {
		return n.val.IsEqual(o.val)
	}

	return false
}

// String returns the string representation of assignment statement
func (n *AssignmentNode) String() string {
	obj := n.val

	if obj.Type().IsExpr() {
		return n.name + " = " + obj.String()
	}

	return "<unknown>"
}

// NewExecAssignNode creates a new node for executing something and store the
// result on a new variable. The assignment could be made using an operating system
// command, a pipe of commands or a function invocation.
// It returns a *ExecAssignNode ready to be executed or error when n is not a valid
// node for execution.
func NewExecAssignNode(pos token.Pos, name string, n Node) (*ExecAssignNode, error) {
	if !n.Type().IsExecutable() {
		return nil, errors.New("NewExecAssignNode expects a CommandNode, or PipeNode or FninvNode")
	}

	return &ExecAssignNode{
		NodeType: NodeExecAssign,
		Pos:      pos,
		name:     name,

		cmd: n,
	}, nil
}

// Name returns the identifier (l-value)
func (n *ExecAssignNode) Name() string {
	return n.name
}

// Command returns the command (or r-value). Command could be a CommandNode or FnNode
func (n *ExecAssignNode) Command() Node {
	return n.cmd
}

// SetName set the assignment identifier (l-value)
func (n *ExecAssignNode) SetName(name string) {
	n.name = name
}

// SetCommand set the command part (NodeCommand or NodeFnDecl)
func (n *ExecAssignNode) SetCommand(c Node) {
	n.cmd = c
}

func (n *ExecAssignNode) IsEqual(other Node) bool {
	if n == other {
		return true
	}

	o, ok := other.(*ExecAssignNode)

	if !ok {
		debug("Failed to convert to ExecAssignNode")
		return false
	}

	if n.name != o.name {
		debug("Exec assignment name differs")
		return false
	}

	if n.cmd == o.cmd {
		return true
	}

	if n.cmd != nil {
		return n.cmd.IsEqual(o.cmd)
	}

	return false
}

// String returns the string representation of command assignment statement
func (n *ExecAssignNode) String() string {
	return n.name + " <= " + n.cmd.String()
}

// NewCommandNode creates a new node for commands
func NewCommandNode(pos token.Pos, name string) *CommandNode {
	return &CommandNode{
		NodeType: NodeCommand,
		Pos:      pos,
		name:     name,
	}
}

// AddArg adds a new argument to the command
func (n *CommandNode) AddArg(a Expr) {
	n.args = append(n.args, a)
}

// SetArgs sets an array of args to command
func (n *CommandNode) SetArgs(args []Expr) {
	n.args = args
}

// Args returns the list of arguments supplied to command.
func (n *CommandNode) Args() []Expr { return n.args }

// AddRedirect adds a new redirect node to command
func (n *CommandNode) AddRedirect(redir *RedirectNode) {
	n.redirs = append(n.redirs, redir)
}

// Redirects return the list of redirect maps of the command.
func (n *CommandNode) Redirects() []*RedirectNode { return n.redirs }

// Name returns the program name
func (n *CommandNode) Name() string { return n.name }

// IsEqual returns if it is equal to the other node.
func (n *CommandNode) IsEqual(other Node) bool {
	if n == other {
		return true
	}

	o, ok := other.(*CommandNode)

	if !ok {
		debug("Failed to convert to CommandNode")
		return false
	}

	if len(n.args) != len(o.args) {
		debug("Command argument length differs: %d (%+v) != %d (%+v)", len(n.args), n.args, len(o.args), o.args)
		return false
	}

	for i := 0; i < len(n.args); i++ {
		if !n.args[i].IsEqual(o.args[i]) {
			debug("Argument %d differs. '%s' != '%s'", i, n.args[i], o.args[i])
			return false
		}
	}

	if len(n.redirs) != len(o.redirs) {
		debug("Number of redirects differs. %d != %d", len(n.redirs), len(o.redirs))
		return false
	}

	for i := 0; i < len(n.redirs); i++ {
		if !n.redirs[i].IsEqual(o.redirs[i]) {
			debug("Redirect differs... %s != %s", n.redirs[i], o.redirs[i])
			return false
		}
	}

	return n.name == o.name
}

// String returns the string representation of command statement
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

// NewPipeNode creates a new command pipeline
func NewPipeNode(pos token.Pos) *PipeNode {
	return &PipeNode{
		NodeType: NodePipe,
		Pos:      pos,
	}
}

// AddCmd add another command to end of the pipeline
func (n *PipeNode) AddCmd(c *CommandNode) {
	n.cmds = append(n.cmds, c)
}

// Commands returns the list of pipeline commands
func (n *PipeNode) Commands() []*CommandNode {
	return n.cmds
}

// IsEqual returns if it is equal to the other node.
func (n *PipeNode) IsEqual(other Node) bool {
	if n == other {
		return true
	}

	o, ok := other.(*PipeNode)

	if !ok {
		debug("Failed to convert to PipeNode")
		return false
	}

	if len(n.cmds) != len(o.cmds) {
		debug("Number of pipe commands differ: %d != %d", len(n.cmds), len(o.cmds))
		return false
	}

	for i := 0; i < len(n.cmds); i++ {
		if !n.cmds[i].IsEqual(o.cmds[i]) {
			debug("Command differs. '%s' != '%s'", n.cmds[i], o.cmds[i])
			return false
		}
	}

	return true
}

// String returns the string representation of pipeline statement
func (n *PipeNode) String() string {
	ret := ""

	for i := 0; i < len(n.cmds); i++ {
		ret += n.cmds[i].String()

		if i < (len(n.cmds) - 1) {
			ret += " | "
		}
	}

	return ret
}

// NewRedirectNode creates a new redirection node for commands
func NewRedirectNode(pos token.Pos) *RedirectNode {
	return &RedirectNode{
		rmap: RedirMap{
			lfd: -1,
			rfd: -1,
		},
		Pos: pos,
	}
}

// SetMap sets the redirection map. Eg.: [2=1]
func (r *RedirectNode) SetMap(lfd int, rfd int) {
	r.rmap.lfd = lfd
	r.rmap.rfd = rfd
}

// LeftFD return the lhs of the redirection map.
func (r *RedirectNode) LeftFD() int { return r.rmap.lfd }

// RightFD return the rhs of the redirection map.
func (r *RedirectNode) RightFD() int { return r.rmap.rfd }

// SetLocation of the output
func (r *RedirectNode) SetLocation(s Expr) { r.location = s }

// Location return the location of the redirection.
func (r *RedirectNode) Location() Expr { return r.location }

// IsEqual return if it is equal to the other node.
func (r *RedirectNode) IsEqual(other Node) bool {
	if r == other {
		return true
	}

	o, ok := other.(*RedirectNode)

	if !ok {
		return false
	}

	if r.rmap.lfd != o.rmap.lfd ||
		r.rmap.rfd != o.rmap.rfd {
		return false
	}

	if r.location == o.location {
		return true
	}

	if r.location != nil {
		return r.location.IsEqual(o.location)
	}

	return false
}

// String returns the string representation of redirect
func (r *RedirectNode) String() string {
	var result string

	if r.rmap.lfd == r.rmap.rfd {
		if r.location != nil {
			return "> " + r.location.String()
		}

		return ""
	}

	if r.rmap.rfd >= 0 {
		result = ">[" + strconv.Itoa(r.rmap.lfd) + "=" + strconv.Itoa(r.rmap.rfd) + "]"
	} else if r.rmap.rfd == RedirMapNoValue {
		result = ">[" + strconv.Itoa(r.rmap.lfd) + "]"
	} else if r.rmap.rfd == RedirMapSupress {
		result = ">[" + strconv.Itoa(r.rmap.lfd) + "=]"
	}

	if r.location != nil {
		result = result + " " + r.location.String()
	}

	return result
}

// NewRforkNode creates a new node for rfork
func NewRforkNode(pos token.Pos) *RforkNode {
	return &RforkNode{
		NodeType: NodeRfork,
		Pos:      pos,
	}
}

// Arg return the string argument of the rfork.
func (n *RforkNode) Arg() *StringExpr {
	return n.arg
}

// SetFlags sets the rfork flags
func (n *RforkNode) SetFlags(a *StringExpr) {
	n.arg = a
}

// Tree returns the child tree of node
func (n *RforkNode) Tree() *Tree {
	return n.tree
}

// SetTree set the body of the rfork block.
func (n *RforkNode) SetTree(t *Tree) {
	n.tree = t
}

func (n *RforkNode) IsEqual(other Node) bool {
	if n == other {
		return true
	}

	o, ok := other.(*RforkNode)

	if !ok {
		return false
	}

	if !n.arg.IsEqual(o.arg) {
		return false
	}

	return n.tree.IsEqual(o.tree)
}

// String returns the string representation of rfork statement
func (n *RforkNode) String() string {
	rforkstr := "rfork " + n.arg.String()
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
func NewCdNode(pos token.Pos, dir Expr) *CdNode {
	return &CdNode{
		NodeType: NodeCd,
		Pos:      pos,

		dir: dir,
	}
}

// Dir returns the directory of cd node
func (n *CdNode) Dir() Expr {
	return n.dir
}

// IsEqual returns if it is equal to other node.
func (n *CdNode) IsEqual(other Node) bool {
	if n == other {
		return true
	}

	o, ok := other.(*CdNode)

	if !ok {
		debug("other node is not CdNode")
		return false
	}

	if n.dir == o.dir {
		return true
	}

	if n.dir != nil {
		return n.dir.IsEqual(o.dir)
	}

	return false
}

// String returns the string representation of cd node
func (n *CdNode) String() string {
	dir := n.dir

	if dir == nil {
		return "cd"
	}

	if !dir.Type().IsExpr() {
		return "cd <invalid path>"
	}

	return "cd " + dir.String()
}

// NewCommentNode creates a new node for comments
func NewCommentNode(pos token.Pos, val string) *CommentNode {
	return &CommentNode{
		NodeType: NodeComment,
		Pos:      pos,
		val:      val,
	}
}

func (n *CommentNode) IsEqual(other Node) bool {
	if n == other {
		return true
	}

	if n.Type() != other.Type() {
		return false
	}

	o, ok := other.(*CommentNode)

	if !ok {
		return false
	}

	return n.val == o.val
}

// String returns the string representation of comment
func (n *CommentNode) String() string {
	return n.val
}

// NewIfNode creates a new if block statement
func NewIfNode(pos token.Pos) *IfNode {
	return &IfNode{
		NodeType: NodeIf,
		Pos:      pos,
	}
}

// Lvalue returns the lefthand part of condition
func (n *IfNode) Lvalue() Expr {
	return n.lvalue
}

// Rvalue returns the righthand side of condition
func (n *IfNode) Rvalue() Expr {
	return n.rvalue
}

// SetLvalue set the lefthand side of condition
func (n *IfNode) SetLvalue(arg Expr) {
	n.lvalue = arg
}

// SetRvalue set the righthand side of condition
func (n *IfNode) SetRvalue(arg Expr) {
	n.rvalue = arg
}

// Op returns the condition operation
func (n *IfNode) Op() string { return n.op }

// SetOp set the condition operation
func (n *IfNode) SetOp(op string) {
	n.op = op
}

// IsElseIf tells if the if is an else-if statement
func (n *IfNode) IsElseIf() bool {
	return n.elseIf
}

// SetElseif sets the else-if part
func (n *IfNode) SetElseif(b bool) {
	n.elseIf = b
}

// SetIfTree sets the block of statements of the if block
func (n *IfNode) SetIfTree(t *Tree) {
	n.ifTree = t
}

// SetElseTree sets the block of statements of the else block
func (n *IfNode) SetElseTree(t *Tree) {
	n.elseTree = t
}

// IfTree returns the if block
func (n *IfNode) IfTree() *Tree { return n.ifTree }

// ElseTree returns the else block
func (n *IfNode) ElseTree() *Tree { return n.elseTree }

// IsEqual returns if it is equal to the other node.
func (n *IfNode) IsEqual(other Node) bool {
	if n == other {
		return true
	}

	o, ok := other.(*IfNode)

	if !ok {
		debug("Failed to convert to ifNode")
		return false
	}

	elvalue := n.Lvalue()
	ervalue := n.Rvalue()
	vlvalue := o.Lvalue()
	vrvalue := o.Rvalue()

	if !elvalue.IsEqual(vlvalue) {
		debug("Lvalue differs: '%s' != '%s'", elvalue, vlvalue)
		return false
	}

	if !ervalue.IsEqual(vrvalue) {
		debug("Rvalue differs: '%s' != '%s'", ervalue, vrvalue)
		return false
	}

	if n.Op() != o.Op() {
		debug("Operation differs: %s != %s", n.Op(), o.Op())
		return false
	}

	expectedTree := n.IfTree()
	valueTree := o.IfTree()

	if !expectedTree.IsEqual(valueTree) {
		debug("If tree differs: '%s' != '%s'", expectedTree, valueTree)
		return false
	}

	expectedTree = n.ElseTree()
	valueTree = o.ElseTree()

	return expectedTree.IsEqual(valueTree)
}

// String returns the string representation of if statement
func (n *IfNode) String() string {
	var lstr, rstr string

	lstr = n.lvalue.String()
	rstr = n.rvalue.String()

	ifStr := "if " + lstr + " " + n.op + " " + rstr + " {\n"

	ifTree := n.IfTree()

	block := ifTree.String()
	stmts := strings.Split(block, "\n")

	if strings.TrimSpace(block) != "" {
		for i := 0; i < len(stmts); i++ {
			stmts[i] = "\t" + stmts[i]
		}
	}

	ifStr += strings.Join(stmts, "\n") + "\n}"

	elseTree := n.ElseTree()

	if elseTree != nil {
		ifStr += " else "

		elseBlock := elseTree.String()
		elsestmts := strings.Split(elseBlock, "\n")

		for i := 0; i < len(elsestmts); i++ {
			if !n.IsElseIf() {
				elsestmts[i] = "\t" + elsestmts[i]
			}
		}

		if !n.IsElseIf() {
			ifStr += "{\n"
		}

		ifStr += strings.Join(elsestmts, "\n")

		if !n.IsElseIf() {
			ifStr += "\n}"
		}
	}

	return ifStr
}

// NewFnDeclNode creates a new function declaration
func NewFnDeclNode(pos token.Pos, name string) *FnDeclNode {
	return &FnDeclNode{
		NodeType: NodeFnDecl,
		Pos:      pos,
		name:     name,
	}
}

// SetName set the function name
func (n *FnDeclNode) SetName(a string) {
	n.name = a
}

// Name return the function name
func (n *FnDeclNode) Name() string {
	return n.name
}

// Args returns function arguments
func (n *FnDeclNode) Args() []string {
	return n.args
}

// AddArg add a new argument to end of argument list
func (n *FnDeclNode) AddArg(arg string) {
	n.args = append(n.args, arg)
}

// Tree return the function block
func (n *FnDeclNode) Tree() *Tree {
	return n.tree
}

// SetTree set the function tree
func (n *FnDeclNode) SetTree(t *Tree) {
	n.tree = t
}

func (n *FnDeclNode) IsEqual(other Node) bool {
	if n == other {
		return true
	}

	o, ok := other.(*FnDeclNode)

	if !ok {
		return false
	}

	if n.name != o.name || len(n.args) != len(o.args) {
		return false
	}

	for i := 0; i < len(n.args); i++ {
		if n.args[i] != o.args[i] {
			return false
		}
	}

	return true
}

// String returns the string representation of function declaration
func (n *FnDeclNode) String() string {
	fnStr := "fn"

	if n.name != "" {
		fnStr += " " + n.name + "("
	}

	for i := 0; i < len(n.args); i++ {
		fnStr += n.args[i]

		if i < (len(n.args) - 1) {
			fnStr += ", "
		}
	}

	fnStr += ") {\n"

	tree := n.Tree()

	stmts := strings.Split(tree.String(), "\n")

	for i := 0; i < len(stmts); i++ {
		if len(stmts[i]) > 0 {
			fnStr += "\t" + stmts[i] + "\n"
		}
	}

	fnStr += "}\n"

	return fnStr
}

// NewFnInvNode creates a new function invocation
func NewFnInvNode(pos token.Pos, name string) *FnInvNode {
	return &FnInvNode{
		NodeType: NodeFnInv,
		Pos:      pos,
		name:     name,
	}
}

// SetName set the function name
func (n *FnInvNode) SetName(a string) {
	n.name = a
}

// Name return the function name
func (n *FnInvNode) Name() string {
	return n.name
}

// AddArg add another argument to end of argument list
func (n *FnInvNode) AddArg(arg Expr) {
	n.args = append(n.args, arg)
}

// Args return the invocation arguments.
func (n *FnInvNode) Args() []Expr { return n.args }

// IsEqual returns if it is equal to the other node.
func (n *FnInvNode) IsEqual(other Node) bool {
	if n == other {
		return true
	}

	o, ok := other.(*FnInvNode)

	if !ok {
		return false
	}

	if len(n.args) != len(o.args) {
		return false
	}

	for i := 0; i < len(n.args); i++ {
		if !n.args[i].IsEqual(o.args[i]) {
			return false
		}
	}

	return true
}

// String returns the string representation of function invocation
func (n *FnInvNode) String() string {
	fnInvStr := n.name + "("

	for i := 0; i < len(n.args); i++ {
		fnInvStr += n.args[i].String()

		if i < (len(n.args) - 1) {
			fnInvStr += ", "
		}
	}

	fnInvStr += ")"

	return fnInvStr
}

// NewBindFnNode creates a new bindfn statement
func NewBindFnNode(pos token.Pos, name, cmd string) *BindFnNode {
	return &BindFnNode{
		NodeType: NodeBindFn,
		Pos:      pos,
		name:     name,
		cmdname:  cmd,
	}
}

// Name return the function name
func (n *BindFnNode) Name() string { return n.name }

// CmdName return the command name
func (n *BindFnNode) CmdName() string { return n.cmdname }

func (n *BindFnNode) IsEqual(other Node) bool {
	if n == other {
		return true
	}

	o, ok := other.(*BindFnNode)

	if !ok {
		return false
	}

	return n.name == o.name && n.cmdname == o.cmdname
}

// String returns the string representation of bindfn
func (n *BindFnNode) String() string {
	return "bindfn " + n.name + " " + n.cmdname
}

// NewDumpNode creates a new dump statement
func NewDumpNode(pos token.Pos) *DumpNode {
	return &DumpNode{
		NodeType: NodeDump,
		Pos:      pos,
	}
}

// Filename return the dump filename argument
func (n *DumpNode) Filename() Expr {
	return n.filename
}

// SetFilename set the dump filename
func (n *DumpNode) SetFilename(a Expr) {
	n.filename = a
}

func (n *DumpNode) IsEqual(other Node) bool {
	if n == other {
		return true
	}

	o, ok := other.(*DumpNode)

	if !ok {
		debug("Failed to convert to DumpNode")
		return ok
	}

	if n.filename == o.filename {
		return true
	}

	if n.filename != nil {
		return n.filename.IsEqual(o.filename)
	}

	return false
}

// String returns the string representation of dump node
func (n *DumpNode) String() string {
	if n.filename != nil {
		return "dump " + n.filename.String()
	}

	return "dump"
}

// NewReturnNode create a return statement
func NewReturnNode(pos token.Pos) *ReturnNode {
	return &ReturnNode{
		Pos:      pos,
		NodeType: NodeReturn,
	}
}

// SetReturn set the arguments to return
func (n *ReturnNode) SetReturn(a Expr) {
	n.arg = a
}

// Return returns the argument being returned
func (n *ReturnNode) Return() Expr { return n.arg }

func (n *ReturnNode) IsEqual(other Node) bool {
	if n == other {
		return true
	}

	if n.Type() != other.Type() {
		return false
	}

	o, ok := other.(*ReturnNode)

	if !ok {
		return false
	}

	if n.arg == o.arg {
		return true
	}

	if n.arg != nil {
		return n.arg.IsEqual(o.arg)
	}

	return false
}

// String returns the string representation of return statement
func (n *ReturnNode) String() string {
	if n.arg != nil {
		return "return " + n.arg.String()
	}

	return "return"
}

// NewForNode create a new for statement
func NewForNode(pos token.Pos) *ForNode {
	return &ForNode{
		NodeType: NodeFor,
		Pos:      pos,
	}
}

// SetIdentifier set the for indentifier
func (n *ForNode) SetIdentifier(a string) {
	n.identifier = a
}

// Identifier return the identifier part
func (n *ForNode) Identifier() string { return n.identifier }

// InVar return the "in" variable
func (n *ForNode) InVar() string { return n.inVar }

// SetInVar set "in" variable
func (n *ForNode) SetInVar(a string) { n.inVar = a }

// SetTree set the for block of statements
func (n *ForNode) SetTree(a *Tree) {
	n.tree = a
}

// Tree return the for block
func (n *ForNode) Tree() *Tree { return n.tree }

func (n *ForNode) IsEqual(other Node) bool {
	if n == other {
		return true
	}

	if n.Type() != other.Type() {
		return false
	}

	o, ok := other.(*ForNode)

	if !ok {
		return false
	}

	return n.identifier == o.identifier &&
		n.inVar == o.inVar
}

// String returns the string representation of for statement
func (n *ForNode) String() string {
	ret := "for"

	if n.identifier != "" {
		ret += " " + n.identifier + " in " + n.inVar
	}

	ret += " {\n"

	if n.tree != nil {
		ret += n.tree.String() + "\n"
	}

	ret += "}"

	return ret
}
