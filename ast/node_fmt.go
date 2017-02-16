package ast

import (
	"strconv"
	"strings"
)

func (s *StringExpr) String() string {
	if s.quoted {
		return `"` + stringify(s.str) + `"`
	}

	return s.str
}

func (i *IntExpr) String() string {
	return strconv.Itoa(i.val)
}

func (l *ListExpr) string() (string, bool) {
	elems := make([]string, len(l.list))
	columnCount := 0
	forceMulti := false

	for i := 0; i < len(l.list); i++ {
		if l.list[i].Type() == NodeListExpr {
			forceMulti = true
		}

		elems[i] = l.list[i].String()
		columnCount += len(elems[i])
	}

	if columnCount+len(elems) > 50 || forceMulti {
		forceMulti = true
		return "(\n\t" + strings.Join(elems, "\n\t") + "\n)", forceMulti
	}

	return "(" + strings.Join(elems, " ") + ")", false
}

func (l *ListExpr) String() string {
	str, _ := l.string()
	return str
}

func (c *ConcatExpr) String() string {
	ret := ""

	for i := 0; i < len(c.concat); i++ {
		ret += c.concat[i].String()

		if i < (len(c.concat) - 1) {
			ret += "+"
		}
	}

	return ret
}

func (v *VarExpr) String() string {
	return v.name
}

func (i *IndexExpr) String() string {
	return i.variable.String() + "[" + i.index.String() + "]"
}

func (l *BlockNode) adjustGroupAssign(node assignable, nodes []Node) {
	var (
		eqSpace int = node.getEqSpace()
		i       int
	)

	nodeName := node.Name().String()

	eqSpace = len(nodeName) + 1

	for i = 0; i < len(nodes); i++ {
		assign, ok := nodes[i].(assignable)

		if !ok {
			break
		}

		if len(assign.Name().String())+1 > eqSpace {
			eqSpace = len(assign.Name().String()) + 1
		}
	}

	for j := 0; j < i; j++ {
		knode := nodes[j].(assignable)
		knode.setEqSpace(eqSpace)
	}

	node.setEqSpace(eqSpace)
}

func (l *BlockNode) String() string {
	nodes := l.Nodes
	content := make([]string, 0, 8192)

	last := (len(nodes) - 1)

	for i := 0; i < len(nodes); i++ {
		addEOL := false
		node := nodes[i]

		nodebytes := node.String()

		if i == 0 && node.Type() == NodeComment &&
			strings.HasPrefix(node.String(), "#!") {
			addEOL = true
		} else if (node.Type() == NodeComment) && i < last {
			nextNode := nodes[i+1]

			if nextNode.Line() > node.Line()+1 {
				addEOL = true
			}
		} else if i < last {
			nextNode := nodes[i+1]

			if node.Type() != nextNode.Type() {
				addEOL = true
			} else if node.Type() == NodeFnDecl {
				addEOL = true
			} else if node.Type() == NodeAssignment || node.Type() == NodeExecAssign {
				nodeAssign := node.(assignable)

				if nodeAssign.getEqSpace() == -1 {
					// lookahead to decide about best '=' distance
					l.adjustGroupAssign(nodeAssign, nodes[i+1:])
				}

				nodebytes, addEOL = nodeAssign.string()
			}
		}

		if addEOL {
			nodebytes += "\n"
		}

		content = append(content, nodebytes)
	}

	return strings.Join(content, "\n")
}

// String returns the string representation of the import
func (n *ImportNode) String() string {
	return `import ` + n.path.String()
}

// String returns the string representation of assignment
func (n *SetenvNode) String() string {
	if n.assign == nil {
		return "setenv " + n.varName
	}

	return "setenv " + n.assign.String()
}

func (n *NameNode) String() string {
	if n.index != nil {
		return n.name + "[" + n.index.String() + "]"
	}

	return n.name
}

func (n *AssignmentNode) string() (string, bool) {
	var (
		objStr string
		multi  bool
	)

	obj := n.val
	lhs := n.name.String()

	if obj.Type().IsExpr() {
		if obj.Type() == NodeListExpr {
			lobj := obj.(*ListExpr)
			objStr, multi = lobj.string()
		} else {
			objStr = obj.String()
		}

		if n.eqSpace > len(lhs) && !multi {
			ret := lhs + strings.Repeat(" ", n.eqSpace-len(lhs)) + "= " + objStr
			return ret, multi
		}

		ret := lhs + " = " + objStr
		return ret, multi
	}

	return "<unknown>", false
}

// String returns the string representation of assignment statement
func (n *AssignmentNode) String() string {
	str, _ := n.string()
	return str
}

func (n *ExecAssignNode) string() (string, bool) {
	var (
		cmdStr string
		multi  bool
	)

	lhs := n.name.String()

	if n.cmd.Type() == NodeCommand {
		cmd := n.cmd.(*CommandNode)
		cmdStr, multi = cmd.string()
	} else if n.cmd.Type() == NodePipe {
		cmd := n.cmd.(*PipeNode)
		cmdStr, multi = cmd.string()
	} else {
		cmd := n.cmd.(*FnInvNode)
		cmdStr, multi = cmd.string()
	}

	if n.eqSpace > len(lhs) {
		ret := lhs + strings.Repeat(" ", n.eqSpace-len(lhs)) + "<= " + cmdStr
		return ret, multi
	}

	return lhs + " <= " + cmdStr, multi
}

// String returns the string representation of command assignment statement
func (n *ExecAssignNode) String() string {
	str, _ := n.string()
	return str
}

func (n *CommandNode) toStringParts() ([]string, int) {
	var (
		content  []string
		line     string
		last     = len(n.args) - 1
		totalLen = 0
	)

	for i := 0; i < len(n.args); i += 2 {
		var next string

		arg := n.args[i].String()

		if i < last {
			next = n.args[i+1].String()
		}

		if i == 0 {
			arg = n.name + " " + arg
		}

		if arg[0] == '-' {
			if line != "" {
				content = append(content, line)
				line = ""
			}

			if len(next) > 0 && next[0] != '-' {
				if line == "" {
					line += arg + " " + next
				} else {
					line += " " + arg + " " + next
				}
			} else {
				content = append(content, arg, next)
			}
		} else if next != "" {
			if line == "" {
				line += arg + " " + next
			} else {
				line += " " + arg + " " + next
			}
		} else {
			if line == "" {
				line += arg
			} else {
				line += " " + arg
			}
		}

		totalLen += len(arg) + len(next) + 1

	}

	if line != "" {
		content = append(content, line)
	}

	if len(content) == 0 {
		content = append(content, n.name)
	}

	for i := 0; i < len(n.redirs); i++ {
		rstr := n.redirs[i].String()
		totalLen += len(rstr) + 1
		content = append(content, rstr)
	}

	return content, totalLen
}

func (n *CommandNode) multiString() string {
	content, totalLen := n.toStringParts()

	if totalLen < 50 {
		return "(" + strings.Join(content, " ") + ")"
	}

	content[0] = "\t" + content[0]

	gentab := func(n int) string { return strings.Repeat("\t", n) }
	tabLen := (len(content[0]) + 7) / 8

	for i := 1; i < len(content); i++ {
		content[i] = gentab(tabLen) + content[i]
	}

	return "(\n" + strings.Join(content, "\n") + "\n)"
}

// String returns the string representation of command statement
func (n *CommandNode) string() (string, bool) {
	if n.multi {
		return n.multiString(), true
	}

	var content []string

	content = append(content, n.name)

	for i := 0; i < len(n.args); i++ {
		content = append(content, n.args[i].String())
	}

	for i := 0; i < len(n.redirs); i++ {
		content = append(content, n.redirs[i].String())
	}

	return strings.Join(content, " "), false
}

func (n *CommandNode) String() string {
	str, _ := n.string()
	return str
}

func (n *PipeNode) multiString() string {
	totalLen := 0

	type cmdData struct {
		content  []string
		totalLen int
	}

	content := make([]cmdData, len(n.cmds))

	for i := 0; i < len(n.cmds); i++ {
		cmdContent, cmdLen := n.cmds[i].toStringParts()

		content[i] = cmdData{
			cmdContent,
			cmdLen,
		}

		totalLen += cmdLen
	}

	if totalLen+3 < 50 {
		result := "("

		for i := 0; i < len(content); i++ {
			result += strings.Join(content[i].content, " ")

			if i < len(content)-1 {
				result += " | "
			}
		}

		return result + ")"
	}

	gentab := func(n int) string { return strings.Repeat("\t", n) }

	result := "(\n"

	for i := 0; i < len(content); i++ {
		cmdContent := content[i].content

		cmdContent[0] = "\t" + cmdContent[0]
		tabLen := (len(cmdContent[0]) + 7) / 8

		for j := 1; j < len(cmdContent); j++ {
			cmdContent[j] = gentab(tabLen) + cmdContent[j]
		}

		result += strings.Join(cmdContent, "\n")

		if i < len(content)-1 {
			result += " |\n"
		}
	}

	return result + "\n)"
}

// String returns the string representation of pipeline statement
func (n *PipeNode) string() (string, bool) {
	if n.multi {
		return n.multiString(), true
	}

	ret := ""

	for i := 0; i < len(n.cmds); i++ {
		ret += n.cmds[i].String()

		if i < (len(n.cmds) - 1) {
			ret += " | "
		}
	}

	return ret, false
}

func (n *PipeNode) String() string {
	str, _ := n.string()
	return str
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

// String returns the string representation of comment
func (n *CommentNode) String() string {
	return n.val
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
		} else {
			fnStr += "\n"
		}
	}

	fnStr += "}"

	return fnStr
}

// String returns the string representation of function invocation
func (n *FnInvNode) string() (string, bool) {
	fnInvStr := n.name + "("

	for i := 0; i < len(n.args); i++ {
		fnInvStr += n.args[i].String()

		if i < (len(n.args) - 1) {
			fnInvStr += ", "
		}
	}

	fnInvStr += ")"

	return fnInvStr, false
}

func (n *FnInvNode) String() string {
	str, _ := n.string()
	return str
}

// String returns the string representation of bindfn
func (n *BindFnNode) String() string {
	return "bindfn " + n.name + " " + n.cmdname
}

// String returns the string representation of dump node
func (n *DumpNode) String() string {
	if n.filename != nil {
		return "dump " + n.filename.String()
	}

	return "dump"
}

// String returns the string representation of return statement
func (n *ReturnNode) String() string {
	if n.arg != nil {
		return "return " + n.arg.String()
	}

	return "return"
}

// String returns the string representation of for statement
func (n *ForNode) String() string {
	ret := "for"

	if n.identifier != "" {
		ret += " " + n.identifier + " in " + n.inExpr.String()
	}

	ret += " {\n"

	tree := n.Tree()

	stmts := strings.Split(tree.String(), "\n")

	for i := 0; i < len(stmts); i++ {
		if len(stmts[i]) > 0 {
			ret += "\t" + stmts[i] + "\n"
		} else {
			ret += "\n"
		}
	}

	ret += "}"

	return ret
}

func stringify(s string) string {
	buf := make([]byte, 0, len(s))

	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '"':
			buf = append(buf, '\\', '"')
		case '\t':
			buf = append(buf, '\\', 't')
		case '\n':
			buf = append(buf, '\\', 'n')
		case '\r':
			buf = append(buf, '\\', 'r')
		case '\\':
			buf = append(buf, '\\', '\\')
		default:
			buf = append(buf, s[i])
		}
	}

	return string(buf)
}
