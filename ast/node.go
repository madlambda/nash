package ast

import (
	"strconv"
	"strings"

	"github.com/NeowayLabs/nash/token"
)

const (
	RedirMapNoValue = -1
	RedirMapSupress = -2
)

type (
	// Node represents nodes in the grammar
	Node interface {
		Type() NodeType
		Position() token.Pos
		String() string
		IsEqual(Node) bool
	}

	Expr Node

	// NodeType is the types of grammar
	NodeType int

	// ListNode is the block
	ListNode struct {
		NodeType
		token.Pos

		Nodes []Node
	}

	ImportNode struct {
		NodeType
		token.Pos

		path *StringExpr // Import path
	}

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

	CmdAssignmentNode struct {
		NodeType
		token.Pos

		name string
		cmd  Node
	}

	// CommandNode is a node for commands
	CommandNode struct {
		NodeType
		token.Pos

		name   string
		args   []Expr
		redirs []*RedirectNode
	}

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

	FnDeclNode struct {
		NodeType
		token.Pos
		name string
		args []string
		tree *Tree
	}

	FnInvNode struct {
		NodeType
		token.Pos
		name string
		args []Expr
	}

	ReturnNode struct {
		NodeType
		token.Pos
		arg Expr
	}

	BindFnNode struct {
		NodeType
		token.Pos
		name    string
		cmdname string
	}

	DumpNode struct {
		NodeType
		token.Pos
		filename Expr
	}

	ForNode struct {
		NodeType
		token.Pos
		identifier string
		inVar      string
		tree       *Tree
	}

	BuiltinNode struct {
		NodeType
		token.Pos
		stmt Node
	}
)

//go:generate stringer -type=NodeType

const (
	// NodeSetenv the type for "setenv" builtin keyword
	NodeSetenv NodeType = iota + 1

	// NodeAssignment is the type for simple variable assignment
	NodeAssignment

	// NodeCmdAssignment is the type for command or function assignment
	NodeCmdAssignment

	// NodeImport is the type for "import" builtin keyword
	NodeImport

	// NodeCommand is the type for command execution
	NodeCommand

	// NodePipe is the type for pipeline execution
	NodePipe

	expr_beg

	NodeStringExpr
	NodeIntExpr
	NodeVarExpr
	NodeListExpr
	NodeIndexExpr
	NodeConcatExpr

	expr_end

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

	// NodeFnInv is the type for function invocation
	NodeFnInv

	// NodeBindFn is the type for bindfn statements
	NodeBindFn

	// NodeDump is the type for dump statements
	NodeDump

	// NodeFor is the type for "for" statements
	NodeFor

	// NodeBuiltin
	NodeBuiltin
)

// Type returns the type of the node
func (t NodeType) Type() NodeType {
	return t
}

func (t NodeType) IsExpr() bool {
	return t > expr_beg && t < expr_end
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

func (l *ListNode) IsEqual(other Node) bool {
	if l == other {
		return true
	}

	o, ok := other.(*ListNode)

	if !ok {
		return false
	}

	if len(l.Nodes) != len(o.Nodes) {
		return false
	}

	for i := 0; i < len(l.Nodes); i++ {
		if !l.Nodes[i].IsEqual(o.Nodes[i]) {
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

// Path returns the path of import
func (n *ImportNode) Path() *StringExpr { return n.path }

func (n *ImportNode) IsEqual(other Node) bool {
	if n == other {
		return true
	}

	o, ok := other.(*ImportNode)

	if !ok {
		return false
	}

	return n.path.IsEqual(o.path)
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

func (n *SetenvNode) Identifier() string { return n.varName }

func (n *SetenvNode) IsEqual(other Node) bool {
	if n == other {
		return true
	}

	o, ok := other.(*SetenvNode)

	if !ok {
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

// SetVarName sets the name of the variable
func (n *AssignmentNode) SetIdentifier(a string) {
	n.name = a
}

func (n *AssignmentNode) Identifier() string { return n.name }

// SetValueList sets the value of the list
func (n *AssignmentNode) SetValue(val Expr) {
	n.val = val
}

// Value returns the assigned object
func (n *AssignmentNode) Value() Expr {
	return n.val
}

func (n *AssignmentNode) IsEqual(other Node) bool {
	if n == other {
		return true
	}

	o, ok := other.(*AssignmentNode)

	if !ok {
		return false
	}

	if n.name != o.name {
		return false
	}

	return n.val.IsEqual(o.val)
}

// String returns the string representation of assignment statement
func (n *AssignmentNode) String() string {
	obj := n.val

	if obj.Type().IsExpr() {
		return n.name + " = " + obj.String()
	}

	return "<unknown>"
}

// NewCmdAssignmentNode creates a new command assignment
func NewCmdAssignmentNode(pos token.Pos, name string) *CmdAssignmentNode {
	return &CmdAssignmentNode{
		NodeType: NodeCmdAssignment,
		Pos:      pos,
		name:     name,
	}
}

// Name returns the identifier (l-value)
func (n *CmdAssignmentNode) Name() string {
	return n.name
}

// Command returns the command (or r-value). Command could be a CommandNode or FnNode
func (n *CmdAssignmentNode) Command() Node {
	return n.cmd
}

// SetName set the assignment identifier (l-value)
func (n *CmdAssignmentNode) SetName(name string) {
	n.name = name
}

// SetCommand set the command part (NodeCommand or NodeFnDecl)
func (n *CmdAssignmentNode) SetCommand(c Node) {
	n.cmd = c
}

func (n *CmdAssignmentNode) IsEqual(other Node) bool {
	if n == other {
		return true
	}

	o, ok := other.(*CmdAssignmentNode)

	if !ok {
		return false
	}

	if n.name != o.name {
		return false
	}

	return n.cmd.IsEqual(o.cmd)
}

// String returns the string representation of command assignment statement
func (n *CmdAssignmentNode) String() string {
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

func (n *CommandNode) Args() []Expr { return n.args }

// AddRedirect adds a new redirect node to command
func (n *CommandNode) AddRedirect(redir *RedirectNode) {
	n.redirs = append(n.redirs, redir)
}

func (n *CommandNode) Redirects() []*RedirectNode { return n.redirs }

// Name returns the program name
func (n *CommandNode) Name() string { return n.name }

func (n *CommandNode) IsEqual(other Node) bool {
	if n == other {
		return true
	}

	o, ok := other.(*CommandNode)

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

	if len(n.redirs) != len(o.redirs) {
		return false
	}

	for i := 0; i < len(n.redirs); i++ {
		if !n.redirs[i].IsEqual(o.redirs[i]) {
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

func (n *PipeNode) IsEqual(other Node) bool {
	if n == other {
		return true
	}

	o, ok := other.(*PipeNode)

	if !ok {
		return false
	}

	if len(n.cmds) != len(o.cmds) {
		return false
	}

	for i := 0; i < len(n.cmds); i++ {
		if n.cmds[i].IsEqual(o.cmds[i]) {
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

func (r *RedirectNode) LeftFD() int  { return r.rmap.lfd }
func (r *RedirectNode) RightFD() int { return r.rmap.rfd }

// SetLocation of the output
func (r *RedirectNode) SetLocation(s Expr) { r.location = s }
func (r *RedirectNode) Location() Expr     { return r.location }

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

	return r.location.IsEqual(o)
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

func (n *CdNode) IsEqual(other Node) bool {
	if n == other {
		return true
	}

	o, ok := other.(*CdNode)

	if !ok {
		return false
	}

	return n.dir.IsEqual(o.dir)
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
func (n *IfNode) SetElseIf(b bool) {
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

func (n *IfNode) IsEqual(other Node) bool {
	value, ok := other.(*IfNode)

	if !ok {
		return false
	}

	if ok = cmpCommon(n, value); !ok {
		return false
	}

	elvalue := n.Lvalue()
	ervalue := n.Rvalue()
	vlvalue := value.Lvalue()
	vrvalue := value.Rvalue()

	if !elvalue.IsEqual(vlvalue) {
		return false
	}

	if !ervalue.IsEqual(vrvalue) {
		return false
	}

	if n.Op() != value.Op() {
		return false
	}

	expectedTree := n.IfTree()
	valueTree := value.IfTree()

	if !expectedTree.IsEqual(valueTree) {
		return false
	}

	expectedTree = n.ElseTree()
	valueTree = value.ElseTree()

	return expectedTree.IsEqual(valueTree)
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

func (n *FnInvNode) Args() []Expr { return n.args }

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

// NewBuiltinNode creates a new "builtin" node
func NewBuiltinNode(pos token.Pos, n Node) *BuiltinNode {
	return &BuiltinNode{
		NodeType: NodeBuiltin,
		Pos:      pos,
		stmt:     n,
	}
}

func (n *BuiltinNode) Stmt() Node {
	return n.stmt
}

func (n *BuiltinNode) IsEqual(other Node) bool {
	if n == other {
		return true
	}

	o, ok := other.(*BuiltinNode)

	if !ok {
		return false
	}

	return n.stmt.IsEqual(o.stmt)
}

func (n *BuiltinNode) String() string {
	return "builtin " + n.stmt.String()
}
