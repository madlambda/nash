package nash

import (
	"fmt"
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
		String() string
	}

	// NodeType is the types of grammar
	NodeType int

	// ArgType is the types of arguments
	// (quoted string, unquoted, variable and concat)
	ArgType int

	// ListNode is the block
	ListNode struct {
		NodeType
		Pos
		Nodes []Node
	}

	// Pos is the position of a node in file
	Pos int

	ImportNode struct {
		NodeType
		Pos
		path *Arg
	}

	SetAssignmentNode struct {
		NodeType
		Pos
		varName string
	}

	ShowEnvNode struct {
		NodeType
		Pos
	}

	// AssignmentNode is a node for variable assignments
	AssignmentNode struct {
		NodeType
		Pos
		name string
		list []*Arg
	}

	CmdAssignmentNode struct {
		NodeType
		Pos
		name string
		cmd  Node
	}

	// CommandNode is a node for commands
	CommandNode struct {
		NodeType
		Pos
		name   string
		args   []*Arg
		redirs []*RedirectNode
	}

	PipeNode struct {
		NodeType
		Pos
		cmds []*CommandNode
	}

	// Arg is a command argument
	Arg struct {
		NodeType
		Pos

		argType ArgType
		val     string
		concat  []*Arg
	}

	// RedirectNode represents the output redirection part of a command
	RedirectNode struct {
		NodeType
		Pos
		rmap     RedirMap
		location *Arg
	}

	// RforkNode is a builtin node for rfork
	RforkNode struct {
		NodeType
		Pos
		arg  *Arg
		tree *Tree
	}

	// CdNode is a builtin node for change directories
	CdNode struct {
		NodeType
		Pos
		dir *Arg
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

	IfNode struct {
		NodeType
		Pos
		lvalue *Arg
		rvalue *Arg
		op     string
		elseIf bool

		ifTree   *Tree
		elseTree *Tree
	}

	FnDeclNode struct {
		NodeType
		Pos
		name string
		args []string
		tree *Tree
	}

	FnInvNode struct {
		NodeType
		Pos
		name string
		args []*Arg
	}

	BindFnNode struct {
		NodeType
		Pos
		name    string
		cmdname string
	}

	DumpNode struct {
		NodeType
		Pos
		filename *Arg
	}
)

//go:generate stringer -type=NodeType

const (
	// NodeSetAssignment is the type for "setenv" builtin keyword
	NodeSetAssignment NodeType = iota + 1

	// NodeShowEnv is the type for "showenv" builtin keyword
	NodeShowEnv

	// NodeAssignment are nodes for variable assignment
	NodeAssignment

	// NodeCmdAssignment
	NodeCmdAssignment

	// NodeImport is the type for "import" builtin keyword
	NodeImport

	// NodeCommand are command statements
	NodeCommand

	// NodePipe is the node type for pipes
	NodePipe

	// NodeArg are nodes for command arguments
	NodeArg

	// NodeString are nodes for argument strings
	NodeString

	// NodeRfork are nodes for rfork command
	NodeRfork

	// NodeCd are nodes of builtin cd
	NodeCd

	// NodeRforkFlags are nodes rfork flags
	NodeRforkFlags

	NodeIf

	// NodeComment are nodes for comment
	NodeComment

	// NodeFn are function nodes
	NodeFnDecl

	// NodeFnInv is a node for function invocation
	NodeFnInv

	NodeBindFn

	NodeDump
)

//go:generate stringer -type=ArgType

const (
	ArgQuoted ArgType = iota + 1
	ArgUnquoted
	ArgVariable
	ArgConcat
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

func NewImportNode(pos Pos) *ImportNode {
	return &ImportNode{
		NodeType: NodeImport,
		Pos:      pos,
	}
}

func (n *ImportNode) SetPath(arg *Arg) {
	n.path = arg
}

func (n *ImportNode) String() string {
	if n.path.IsQuoted() {
		return `import "` + n.path.val + `"`
	} else {
		return "import " + n.path.val
	}
}

func NewSetAssignmentNode(pos Pos, name string) *SetAssignmentNode {
	return &SetAssignmentNode{
		NodeType: NodeSetAssignment,
		Pos:      pos,
		varName:  name,
	}
}

func (n *SetAssignmentNode) String() string {
	return "setenv " + n.varName
}

func NewShowEnvNode(pos Pos) *ShowEnvNode {
	return &ShowEnvNode{
		NodeType: NodeShowEnv,
		Pos:      pos,
	}
}

func (n *ShowEnvNode) String() string { return "showenv" }

// NewAssignmentNode creates a new assignment
func NewAssignmentNode(pos Pos) *AssignmentNode {
	return &AssignmentNode{
		NodeType: NodeAssignment,
		Pos:      pos,
	}
}

// SetVarName sets the name of the variable
func (n *AssignmentNode) SetVarName(a string) {
	n.name = a
}

// SetValueList sets the value of the variable
func (n *AssignmentNode) SetValueList(alist []*Arg) {
	n.list = alist
}

func (n *AssignmentNode) String() string {
	ret := n.name + "="

	if len(n.list) == 1 {
		elem := n.list[0]

		if !elem.IsConcat() {
			if elem.IsVariable() {
				return n.name + `=` + elem.val
			}

			return n.name + `="` + elem.val + `"`
		}

		for i := 0; i < len(elem.concat); i++ {
			e := elem.concat[i]

			if e.IsVariable() {
				ret += e.val
			} else {
				ret += `"` + e.val + `"`
			}

			if i < (len(elem.concat) - 1) {
				ret += " + "
			}
		}

		return ret
	} else if len(n.list) == 0 {
		return n.name + `=""`
	}

	ret += "("

	for i := 0; i < len(n.list); i++ {
		ret += n.list[i].val

		if i < len(n.list)-1 {
			ret += " "
		}
	}

	ret += ")"
	return ret
}

func NewCmdAssignmentNode(pos Pos, name string) *CmdAssignmentNode {
	return &CmdAssignmentNode{
		NodeType: NodeCmdAssignment,
		Pos:      pos,
		name:     name,
	}
}

func (n *CmdAssignmentNode) Name() string {
	return n.name
}

func (n *CmdAssignmentNode) Command() Node {
	return n.cmd
}

func (n *CmdAssignmentNode) SetName(name string) {
	n.name = name
}

func (n *CmdAssignmentNode) SetCommand(c Node) {
	n.cmd = c
}

func (n *CmdAssignmentNode) String() string {
	return n.name + " <= " + n.cmd.String()
}

// NewCommandNode creates a new node for commands
func NewCommandNode(pos Pos, name string) *CommandNode {
	return &CommandNode{
		NodeType: NodeCommand,
		Pos:      pos,
		name:     name,
		args:     make([]*Arg, 0, 1024),
	}
}

// AddArg adds a new argument to the command
func (n *CommandNode) AddArg(a *Arg) {
	n.args = append(n.args, a)
}

// SetArgs sets an array of args to command
func (n *CommandNode) SetArgs(args []*Arg) {
	n.args = args
}

// AddRedirect adds a new redirect node to command
func (n *CommandNode) AddRedirect(redir *RedirectNode) {
	n.redirs = append(n.redirs, redir)
}

func (n *CommandNode) Name() string { return n.name }

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

func NewPipeNode(pos Pos) *PipeNode {
	return &PipeNode{
		NodeType: NodePipe,
		Pos:      pos,
		cmds:     make([]*CommandNode, 0, 16),
	}
}

func (n *PipeNode) AddCmd(c *CommandNode) {
	n.cmds = append(n.cmds, c)
}

func (n *PipeNode) Commands() []*CommandNode {
	return n.cmds
}

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
func NewRedirectNode(pos Pos) *RedirectNode {
	return &RedirectNode{
		rmap: RedirMap{
			lfd: -1,
			rfd: -1,
		},
		location: nil,
	}
}

// SetMap sets the redirection map. Eg.: [2=1]
func (r *RedirectNode) SetMap(lfd int, rfd int) {
	r.rmap.lfd = lfd
	r.rmap.rfd = rfd
}

// SetLocation of the output
func (r *RedirectNode) SetLocation(s *Arg) {
	r.location = s
}

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
	} else if r.rmap.rfd == redirMapNoValue {
		result = ">[" + strconv.Itoa(r.rmap.lfd) + "]"
	} else if r.rmap.rfd == redirMapSupress {
		result = ">[" + strconv.Itoa(r.rmap.lfd) + "=]"
	}

	if r.location != nil {
		result = result + " " + r.location.String()
	}

	return result
}

// NewRforkNode creates a new node for rfork
func NewRforkNode(pos Pos) *RforkNode {
	return &RforkNode{
		NodeType: NodeRfork,
		Pos:      pos,
	}
}

// SetFlags sets the rfork flags
func (n *RforkNode) SetFlags(a *Arg) {
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

// SetDir sets the cd directory to dir
func (n *CdNode) SetDir(dir *Arg) {
	n.dir = dir
}

// Dir returns the directory of cd node
func (n *CdNode) Dir() *Arg {
	return n.dir
}

func (n *CdNode) String() string {
	dir := n.dir

	if dir == nil {
		return "cd"
	}

	if dir.IsQuoted() {
		return `cd "` + dir.val + `"`
	}

	if dir.IsUnquoted() || dir.IsVariable() {
		return `cd ` + dir.val
	}

	if dir.IsConcat() {
		ret := "cd "
		for i := 0; i < len(dir.concat); i++ {
			a := dir.concat[i]

			ret += a.String()

			if i < (len(dir.concat) - 1) {
				ret += "+"
			}
		}

		return ret
	}

	panic("internal error")
}

// NewArg creates a new argument
func NewArg(pos Pos, argType ArgType) *Arg {
	return &Arg{
		NodeType: NodeArg,
		Pos:      pos,
		argType:  argType,
	}
}

func (n *Arg) SetArgType(t ArgType) {
	n.argType = t
}

func (n *Arg) SetString(name string) {
	n.val = name
}

func (n *Arg) Value() string {
	return n.val
}

func (n *Arg) SetConcat(v []*Arg) {
	n.concat = v
}

func (n *Arg) SetItem(val item) error {
	if val.typ == itemArg {
		n.SetArgType(ArgUnquoted)
		n.SetString(val.val)
	} else if val.typ == itemString {
		n.SetArgType(ArgQuoted)
		n.SetString(val.val)
	} else if val.typ == itemVariable {
		n.SetArgType(ArgVariable)
		n.SetString(val.val)
	} else {
		return fmt.Errorf("Arg doesn't support type %v", val)
	}

	return nil
}

func (n *Arg) IsQuoted() bool   { return n.argType == ArgQuoted }
func (n *Arg) IsUnquoted() bool { return n.argType == ArgUnquoted }
func (n *Arg) IsVariable() bool { return n.argType == ArgVariable }
func (n *Arg) IsConcat() bool   { return n.argType == ArgConcat }

func (n Arg) String() string {
	if n.IsQuoted() {
		return "\"" + n.val + "\""
	} else if n.IsConcat() {
		ret := ""

		for i := 0; i < len(n.concat); i++ {
			a := n.concat[i]

			if a.IsQuoted() {
				ret += `"` + a.val + `"`
			} else {
				ret += a.val
			}

			if i < (len(n.concat) - 1) {
				ret += "+"
			}
		}

		return ret
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

func (n *CommentNode) String() string {
	return n.val
}

func NewIfNode(pos Pos) *IfNode {
	return &IfNode{
		NodeType: NodeIf,
		Pos:      pos,
	}
}

func (n *IfNode) Lvalue() *Arg {
	return n.lvalue
}

func (n *IfNode) Rvalue() *Arg {
	return n.rvalue
}

func (n *IfNode) SetLvalue(arg *Arg) {
	n.lvalue = arg
}

func (n *IfNode) SetRvalue(arg *Arg) {
	n.rvalue = arg
}

func (n *IfNode) Op() string { return n.op }

func (n *IfNode) SetOp(op string) {
	n.op = op
}

func (n *IfNode) IsElseIf() bool {
	return n.elseIf
}

func (n *IfNode) SetElseIf(b bool) {
	n.elseIf = b
}

func (n *IfNode) SetIfTree(t *Tree) {
	n.ifTree = t
}

func (n *IfNode) SetElseTree(t *Tree) {
	n.elseTree = t
}

func (n *IfNode) IfTree() *Tree   { return n.ifTree }
func (n *IfNode) ElseTree() *Tree { return n.elseTree }

func (n *IfNode) String() string {
	var lstr, rstr string

	if n.lvalue.IsQuoted() {
		lstr = `"` + n.lvalue.val + `"`
	} else {
		lstr = n.lvalue.val // in case of variable
	}

	if n.rvalue.IsQuoted() {
		rstr = `"` + n.rvalue.val + `"`
	} else {
		rstr = n.rvalue.val
	}

	ifStr := "if " + lstr + " " + n.op + " " + rstr + " {\n"

	ifTree := n.IfTree()

	block := ifTree.String()
	stmts := strings.Split(block, "\n")

	for i := 0; i < len(stmts); i++ {
		stmts[i] = "\t" + stmts[i]
	}

	ifStr += strings.Join(stmts, "\n") + "\n}"

	elseTree := n.ElseTree()

	if elseTree != nil {
		ifStr += " else "

		elseBlock := elseTree.String()
		elsestmts := strings.Split(elseBlock, "\n")

		for i := 0; i < len(elsestmts); i++ {
			if n.IsElseIf() {
				elsestmts[i] = elsestmts[i]
			} else {
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

func NewFnDeclNode(pos Pos, name string) *FnDeclNode {
	return &FnDeclNode{
		NodeType: NodeFnDecl,
		Pos:      pos,
		name:     name,
		args:     make([]string, 0, 16),
	}
}

func (n *FnDeclNode) SetName(a string) {
	n.name = a
}

func (n *FnDeclNode) Name() string {
	return n.name
}

func (n *FnDeclNode) Args() []string {
	return n.args
}

func (n *FnDeclNode) AddArg(arg string) {
	n.args = append(n.args, arg)
}

func (n *FnDeclNode) Tree() *Tree {
	return n.tree
}

func (n *FnDeclNode) SetTree(t *Tree) {
	n.tree = t
}

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

func NewFnInvNode(pos Pos, name string) *FnInvNode {
	return &FnInvNode{
		NodeType: NodeFnInv,
		Pos:      pos,
		name:     name,
		args:     make([]*Arg, 0, 16),
	}
}

func (n *FnInvNode) SetName(a string) {
	n.name = a
}

func (n *FnInvNode) AddArg(arg *Arg) {
	n.args = append(n.args, arg)
}

func (n *FnInvNode) String() string {
	fnInvStr := n.name + "("

	for i := 0; i < len(n.args); i++ {
		fnInvStr += n.args[i].Value()

		if i < (len(n.args) - 1) {
			fnInvStr += ", "
		}
	}

	fnInvStr += ")"

	return fnInvStr
}

func NewBindFnNode(pos Pos, name, cmd string) *BindFnNode {
	return &BindFnNode{
		NodeType: NodeBindFn,
		Pos:      pos,
		name:     name,
		cmdname:  cmd,
	}
}

func (n *BindFnNode) Name() string    { return n.name }
func (n *BindFnNode) CmdName() string { return n.cmdname }

func (n *BindFnNode) String() string {
	return "bindfn " + n.name + " " + n.cmdname
}

func NewDumpNode(pos Pos) *DumpNode {
	return &DumpNode{
		NodeType: NodeDump,
		Pos:      pos,
	}
}

func (n *DumpNode) Filename() *Arg {
	return n.filename
}

func (n *DumpNode) SetFilename(a *Arg) {
	n.filename = a
}

func (n *DumpNode) String() string {
	if n.filename != nil {
		return "dump " + n.filename.String()
	}

	return "dump"
}
