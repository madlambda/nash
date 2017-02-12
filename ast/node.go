package ast

import (
	"errors"
	"fmt"

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
		IsEqual(Node) bool

		// Line of node in the file
		Line() int
		// Column of the node in the file
		Column() int

		// String representation of the node.
		// Note that it could not match the correspondent node in
		// the source code.
		String() string
	}

	assignable interface {
		Name() *NameNode
		setEqSpace(int)
		getEqSpace() int
		string() (string, bool)
	}

	// Expr is the interface of expression nodes.
	Expr Node

	// NodeType is the types of grammar
	NodeType int

	// BlockNode is the block
	BlockNode struct {
		NodeType
		token.FileInfo

		Nodes []Node
	}

	// An ImportNode represents the node for an "import" keyword.
	ImportNode struct {
		NodeType
		token.FileInfo

		path *StringExpr // Import path
	}

	// A SetenvNode represents the node for a "setenv" keyword.
	SetenvNode struct {
		NodeType
		token.FileInfo

		varName string
		assign  Node
	}

	NameNode struct {
		NodeType
		token.FileInfo

		name  string
		index Expr
	}

	// AssignmentNode is a node for variable assignments
	AssignmentNode struct {
		NodeType
		token.FileInfo

		name    *NameNode
		val     Expr
		eqSpace int
	}

	// An ExecAssignNode represents the node for execution assignment.
	ExecAssignNode struct {
		NodeType
		token.FileInfo

		name    *NameNode
		cmd     Node
		eqSpace int
	}

	// A CommandNode is a node for commands
	CommandNode struct {
		NodeType
		token.FileInfo

		name   string
		args   []Expr
		redirs []*RedirectNode

		multi bool
	}

	// PipeNode represents the node for a command pipeline.
	PipeNode struct {
		NodeType
		token.FileInfo

		cmds  []*CommandNode
		multi bool
	}

	// StringExpr is a string argument
	StringExpr struct {
		NodeType
		token.FileInfo

		str    string
		quoted bool
	}

	// IntExpr is a integer used at indexing
	IntExpr struct {
		NodeType
		token.FileInfo

		val int
	}

	// ListExpr is a list argument
	ListExpr struct {
		NodeType
		token.FileInfo

		list []Expr
	}

	// ConcatExpr is a concatenation of arguments
	ConcatExpr struct {
		NodeType
		token.FileInfo

		concat []Expr
	}

	// VarExpr is a variable argument
	VarExpr struct {
		NodeType
		token.FileInfo

		name string
	}

	// IndexExpr is a indexed variable
	IndexExpr struct {
		NodeType
		token.FileInfo

		variable *VarExpr
		index    Expr
	}

	// RedirectNode represents the output redirection part of a command
	RedirectNode struct {
		NodeType
		token.FileInfo

		rmap     RedirMap
		location Expr
	}

	// RforkNode is a builtin node for rfork
	RforkNode struct {
		NodeType
		token.FileInfo

		arg  *StringExpr
		tree *Tree
	}

	// CommentNode is the node for comments
	CommentNode struct {
		NodeType
		token.FileInfo

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
		token.FileInfo

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
		token.FileInfo

		name string
		args []string
		tree *Tree
	}

	// A FnInvNode represents a function invocation statement.
	FnInvNode struct {
		NodeType
		token.FileInfo

		name string
		args []Expr
	}

	// A ReturnNode represents the "return" keyword.
	ReturnNode struct {
		NodeType
		token.FileInfo

		arg Expr
	}

	// A BindFnNode represents the "bindfn" keyword.
	BindFnNode struct {
		NodeType
		token.FileInfo

		name    string
		cmdname string
	}

	// A DumpNode represents the "dump" keyword.
	DumpNode struct {
		NodeType
		token.FileInfo

		filename Expr
	}

	// A ForNode represents the "for" keyword.
	ForNode struct {
		NodeType
		token.FileInfo

		identifier string
		inExpr     Expr
		tree       *Tree
	}
)

//go:generate stringer -type=NodeType

const (
	// NodeSetenv is the type for "setenv" builtin keyword
	NodeSetenv NodeType = iota + 1

	// NodeBlock represents a program scope.
	NodeBlock

	// NodeName represents an identifier
	NodeName

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

	// NodeRedirect is the type for redirection nodes
	NodeRedirect

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

// NewBlockNode creates a new block
func NewBlockNode(info token.FileInfo) *BlockNode {
	return &BlockNode{
		NodeType: NodeBlock,
		FileInfo: info,
	}
}

// Push adds a new node for a block of nodes
func (l *BlockNode) Push(n Node) {
	l.Nodes = append(l.Nodes, n)
}

// IsEqual returns if it is equal to the other node.
func (l *BlockNode) IsEqual(other Node) bool {
	if l == other {
		return true
	}

	o, ok := other.(*BlockNode)

	if !ok {
		debug("Failed to cast other node to BlockNode")
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

	return cmpInfo(l, other)
}

// NewImportNode creates a new ImportNode object
func NewImportNode(info token.FileInfo, path *StringExpr) *ImportNode {
	return &ImportNode{
		NodeType: NodeImport,
		FileInfo: info,

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

	if !cmpInfo(n, other) {
		return false
	}

	if n.path != nil {
		return n.path.IsEqual(o.path)
	} else if o.path == nil {
		return true
	}

	return false
}

// NewSetenvNode creates a new assignment node
func NewSetenvNode(info token.FileInfo, name string, assign Node) (*SetenvNode, error) {
	if assign != nil && assign.Type() != NodeAssignment &&
		assign.Type() != NodeExecAssign {
		return nil, errors.New("Invalid assignment in setenv")
	}

	return &SetenvNode{
		NodeType: NodeSetenv,
		FileInfo: info,

		varName: name,
		assign:  assign,
	}, nil
}

// Name returns the environment name.
func (n *SetenvNode) Name() string { return n.varName }

// Assignment returns the setenv assignment (if any)
func (n *SetenvNode) Assignment() Node { return n.assign }

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

	if !cmpInfo(n, other) {
		return false
	}

	if n.assign != o.assign {
		if !n.assign.IsEqual(o.assign) {
			return false
		}
	}

	return n.varName == o.varName
}

func NewNameNode(info token.FileInfo, ident string, index Expr) *NameNode {
	return &NameNode{
		NodeType: NodeName,
		FileInfo: info,
		name:     ident,
		index:    index,
	}
}

func (n *NameNode) Ident() string { return n.name }
func (n *NameNode) Index() Expr   { return n.index }

func (n *NameNode) IsEqual(other Node) bool {
	if n == other {
		return true
	}

	o, ok := other.(*NameNode)

	if !ok {
		debug("Failed to convert to NameNode")
		return false
	}

	if n.name != o.name {
		return false
	}

	if n.index == o.index {
		return true
	}

	if n.index != nil && o.index != nil {
		return n.index.IsEqual(o.index)
	}

	return false
}

// NewAssignmentNode creates a new assignment
func NewAssignmentNode(info token.FileInfo, ident *NameNode, value Expr) *AssignmentNode {
	return &AssignmentNode{
		NodeType: NodeAssignment,
		FileInfo: info,
		eqSpace:  -1,

		name: ident,
		val:  value,
	}
}

// SetIdentifier sets the name of the variable
func (n *AssignmentNode) SetIdentifier(a *NameNode) {
	n.name = a
}

// Name return the name of the variable.
func (n *AssignmentNode) Name() *NameNode { return n.name }

func (n *AssignmentNode) getEqSpace() int      { return n.eqSpace }
func (n *AssignmentNode) setEqSpace(value int) { n.eqSpace = value }

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

	if n.name != o.name && !n.name.IsEqual(o.name) {
		debug("Assignment identifier doesn't match: '%s' != '%s'", n.name, o.name)
		return false
	}

	if n.val != nil && o.val != nil && !n.val.IsEqual(o.val) {
		return false
	}

	if !cmpInfo(n, other) {
		return false
	}

	return true
}

// NewExecAssignNode creates a new node for executing something and store the
// result on a new variable. The assignment could be made using an operating system
// command, a pipe of commands or a function invocation.
// It returns a *ExecAssignNode ready to be executed or error when n is not a valid
// node for execution.
func NewExecAssignNode(info token.FileInfo, name *NameNode, n Node) (*ExecAssignNode, error) {
	if !n.Type().IsExecutable() {
		return nil, errors.New("NewExecAssignNode expects a CommandNode, PipeNode or FninvNode")
	}

	return &ExecAssignNode{
		NodeType: NodeExecAssign,
		FileInfo: info,

		name:    name,
		cmd:     n,
		eqSpace: -1,
	}, nil
}

// Name returns the identifier (l-value)
func (n *ExecAssignNode) Name() *NameNode {
	return n.name
}

func (n *ExecAssignNode) getEqSpace() int      { return n.eqSpace }
func (n *ExecAssignNode) setEqSpace(value int) { n.eqSpace = value }

// Command returns the command (or r-value). Command could be a CommandNode or FnNode
func (n *ExecAssignNode) Command() Node {
	return n.cmd
}

// SetName set the assignment identifier (l-value)
func (n *ExecAssignNode) SetName(name *NameNode) {
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

	if n.name != o.name && !n.name.IsEqual(o.name) {
		debug("Exec assignment name differs")
		return false
	}

	if !cmpInfo(n, other) {
		debug("cmpInfo differs")
		return false
	}

	if n.cmd != nil && o.cmd != nil && !n.cmd.IsEqual(o.cmd) {
		return false
	}

	return true
}

// NewCommandNode creates a new node for commands
func NewCommandNode(info token.FileInfo, name string, multiline bool) *CommandNode {
	return &CommandNode{
		NodeType: NodeCommand,
		FileInfo: info,

		name:  name,
		multi: multiline,
	}
}

func (n *CommandNode) IsMulti() bool   { return n.multi }
func (n *CommandNode) SetMulti(b bool) { n.multi = b }

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

	if n.multi != o.multi {
		debug("Command multiline differs.")
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

	if !cmpInfo(n, other) {
		return false
	}

	return n.name == o.name
}

// NewPipeNode creates a new command pipeline
func NewPipeNode(info token.FileInfo, multi bool) *PipeNode {
	return &PipeNode{
		NodeType: NodePipe,
		FileInfo: info,

		multi: multi,
	}
}

func (n *PipeNode) IsMulti() bool   { return n.multi }
func (n *PipeNode) SetMulti(b bool) { n.multi = b }

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

	return cmpInfo(n, other)
}

// NewRedirectNode creates a new redirection node for commands
func NewRedirectNode(info token.FileInfo) *RedirectNode {
	return &RedirectNode{
		NodeType: NodeRedirect,
		FileInfo: info,

		rmap: RedirMap{
			lfd: -1,
			rfd: -1,
		},
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

	if !cmpInfo(r, other) {
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

// NewRforkNode creates a new node for rfork
func NewRforkNode(info token.FileInfo) *RforkNode {
	return &RforkNode{
		NodeType: NodeRfork,
		FileInfo: info,
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

	if !cmpInfo(n, other) {
		return false
	}

	return n.tree.IsEqual(o.tree)
}

// NewCommentNode creates a new node for comments
func NewCommentNode(info token.FileInfo, val string) *CommentNode {
	return &CommentNode{
		NodeType: NodeComment,
		FileInfo: info,

		val: val,
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

	if !cmpInfo(n, other) {
		return false
	}

	return n.val == o.val
}

// NewIfNode creates a new if block statement
func NewIfNode(info token.FileInfo) *IfNode {
	return &IfNode{
		NodeType: NodeIf,
		FileInfo: info,
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

	if !cmpInfo(n, other) {
		return false
	}

	expectedTree = n.ElseTree()
	valueTree = o.ElseTree()

	return expectedTree.IsEqual(valueTree)
}

// NewFnDeclNode creates a new function declaration
func NewFnDeclNode(info token.FileInfo, name string) *FnDeclNode {
	return &FnDeclNode{
		NodeType: NodeFnDecl,
		FileInfo: info,

		name: name,
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

	return cmpInfo(n, other)
}

// NewFnInvNode creates a new function invocation
func NewFnInvNode(info token.FileInfo, name string) *FnInvNode {
	return &FnInvNode{
		NodeType: NodeFnInv,
		FileInfo: info,

		name: name,
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

	return cmpInfo(n, other)
}

// NewBindFnNode creates a new bindfn statement
func NewBindFnNode(info token.FileInfo, name, cmd string) *BindFnNode {
	return &BindFnNode{
		NodeType: NodeBindFn,
		FileInfo: info,

		name:    name,
		cmdname: cmd,
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

	if !cmpInfo(n, other) {
		return false
	}

	return n.name == o.name && n.cmdname == o.cmdname
}

// NewDumpNode creates a new dump statement
func NewDumpNode(info token.FileInfo) *DumpNode {
	return &DumpNode{
		NodeType: NodeDump,
		FileInfo: info,
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

	if !cmpInfo(n, other) {
		return false
	}

	if n.filename == o.filename {
		return true
	}

	if n.filename != nil {
		return n.filename.IsEqual(o.filename)
	}

	return false
}

// NewReturnNode create a return statement
func NewReturnNode(info token.FileInfo) *ReturnNode {
	return &ReturnNode{
		FileInfo: info,
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

	if !cmpInfo(n, other) {
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

// NewForNode create a new for statement
func NewForNode(info token.FileInfo) *ForNode {
	return &ForNode{
		NodeType: NodeFor,
		FileInfo: info,
	}
}

// SetIdentifier set the for indentifier
func (n *ForNode) SetIdentifier(a string) {
	n.identifier = a
}

// Identifier return the identifier part
func (n *ForNode) Identifier() string { return n.identifier }

// InVar return the "in" variable
func (n *ForNode) InExpr() Expr { return n.inExpr }

// SetInVar set "in" expression
func (n *ForNode) SetInExpr(a Expr) { n.inExpr = a }

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

	if !cmpInfo(n, other) {
		return false
	}

	if n.identifier != o.identifier {
		return false
	}

	if n.inExpr == o.inExpr {
		return true
	}

	if n.inExpr == nil || o.inExpr == nil {
		return false
	}

	return n.inExpr.IsEqual(o.inExpr)
}

func cmpInfo(n, other Node) bool {
	if n.Line() != other.Line() ||
		n.Column() != other.Column() {
		debug("file info mismatch on %v (%s): (%d, %d) != (%d, %d)", n, n.Type(), n.Line(), n.Column(),
			other.Line(), other.Column())
		return false
	}

	return true
}
