package nash

import (
	"errors"
	"fmt"
	"strconv"
)

type (
	// Parser parses an nash file
	Parser struct {
		name       string // filename or name of the buffer
		content    string
		l          *lexer
		tok        *item // token saved for lookahead
		openblocks int

		insidePipe bool
	}
)

// NewParser creates a new parser
func NewParser(name, content string) *Parser {
	return &Parser{
		name:    name,
		content: content,
		l:       lex(name, content),
	}
}

// Parse starts the parsing.
func (p *Parser) Parse() (*Tree, error) {
	root, err := p.parseBlock()

	if err != nil {
		return nil, err
	}

	tr := NewTree(p.name)
	tr.Root = root

	return tr, nil
}

// next returns the next item from lookahead buffer if not empty or
// from the lexer
func (p *Parser) next() item {
	if p.tok != nil {
		t := p.tok
		p.tok = nil
		return *t
	}

	return <-p.l.items
}

// backup puts the item into the lookahead buffer
func (p *Parser) backup(it item) error {
	if p.tok != nil {
		return errors.New("only one slot for backup/lookahead")
	}

	p.tok = &it

	return nil
}

// ignores the next item
func (p *Parser) ignore() {
	if p.tok != nil {
		p.tok = nil
	} else {
		<-p.l.items
	}
}

// peek gets but do not discards the next item (lookahead)
func (p *Parser) peek() item {
	i := p.next()
	p.tok = &i
	return i
}

func (p *Parser) parsePipe(first *CommandNode) (Node, error) {
	it := p.next()

	n := NewPipeNode(it.pos)

	n.AddCmd(first)

	for it = p.peek(); it.typ == itemCommand; it = p.peek() {
		cmd, err := p.parseCommand()

		if err != nil {
			return nil, err
		}

		n.AddCmd(cmd.(*CommandNode))
	}

	return n, nil
}

func (p *Parser) parseCommand() (Node, error) {
	it := p.next()

	// paranoia check
	if it.typ != itemCommand {
		return nil, fmt.Errorf("Invalid command: %v", it)
	}

	n := NewCommandNode(it.pos, it.val)

	for {
		it = p.peek()

		switch it.typ {
		case itemArg, itemString, itemVariable:
			arg, err := p.getArgument(true)

			if err != nil {
				return nil, err
			}

			n.AddArg(arg)
		case itemConcat:
			return nil, fmt.Errorf("Unexpected '+' at pos %d\n", it.pos)
		case itemRedirRight:
			p.next()
			redir, err := p.parseRedirection(it)

			if err != nil {
				return nil, err
			}

			n.AddRedirect(redir)
		case itemPipe:
			if p.insidePipe {
				p.next()
				return n, nil
			}

			p.insidePipe = true
			return p.parsePipe(n)
		case itemEOF:
			return n, nil
		case itemError:
			return nil, fmt.Errorf("Syntax error: %s", it.val)
		default:
			p.backup(it)
			return n, nil
		}
	}

	return nil, errors.New("unreachable")
}

func (p *Parser) parseRedirection(it item) (*RedirectNode, error) {
	var (
		lval, rval int = redirMapNoValue, redirMapNoValue
		err        error
	)

	redir := NewRedirectNode(it.pos)

	it = p.peek()

	if it.typ != itemRedirLBracket && it.typ != itemString && it.typ != itemArg && it.typ != itemVariable {
		return nil, fmt.Errorf("Unexpected token: %v", it)
	}

	// [
	if it.typ == itemRedirLBracket {
		p.next()
		it = p.peek()

		if it.typ != itemRedirMapLSide {
			return nil, fmt.Errorf("Expected lefthand side of redirection map, but found '%s'",
				it.val)
		}

		lval, err = strconv.Atoi(it.val)

		if err != nil {
			return nil, fmt.Errorf("Redirection map expects integers. Found: %s", it.val)
		}

		p.next()
		it = p.peek()

		if it.typ != itemRedirMapEqual && it.typ != itemRedirRBracket {
			return nil, fmt.Errorf("Unexpected token '%v'", it)
		}

		// [xxx=
		if it.typ == itemRedirMapEqual {
			p.next()
			it = p.peek()

			if it.typ != itemRedirMapRSide && it.typ != itemRedirRBracket {
				return nil, fmt.Errorf("Unexpected token '%v'", it)
			}

			if it.typ == itemRedirMapRSide {
				rval, err = strconv.Atoi(it.val)

				if err != nil {
					return nil, fmt.Errorf("Redirection map expects integers. Found: %s", it.val)
				}

				p.next()
				it = p.peek()
			} else {
				rval = redirMapSupress
			}
		}

		if it.typ != itemRedirRBracket {
			return nil, fmt.Errorf("Unexpected token '%v'", it)
		}

		// [xxx=yyy]

		redir.SetMap(lval, rval)

		p.next()
		it = p.peek()
	}

	if it.typ != itemString && it.typ != itemArg && it.typ != itemVariable {
		if rval != redirMapNoValue || lval != redirMapNoValue {
			return redir, nil
		}

		return nil, fmt.Errorf("Unexpected token '%v'", it)
	}

	arg, err := p.getArgument(true)

	if err != nil {
		return nil, err
	}

	redir.SetLocation(arg)

	return redir, nil
}

func (p *Parser) parseImport() (Node, error) {
	it := p.next()

	if it.typ != itemImport {
		return nil, fmt.Errorf("Invalid item: %v", it)
	}

	n := NewImportNode(it.pos)

	it = p.next()

	if it.typ != itemArg && it.typ != itemString {
		return nil, fmt.Errorf("Unexpected token %v", it)
	}

	if it.typ == itemString {
		arg := NewArg(it.pos, ArgQuoted)
		arg.SetString(it.val)
		n.SetPath(arg)
	} else if it.typ == itemArg {
		arg := NewArg(it.pos, ArgUnquoted)
		arg.SetString(it.val)
		n.SetPath(arg)
	} else {
		return nil, fmt.Errorf("Parser error: Invalid token '%v' for import path", it)
	}

	return n, nil
}

func (p *Parser) parseShowEnv() (Node, error) {
	it := p.next()

	return NewShowEnvNode(it.pos), nil
}

func (p *Parser) parseCd() (Node, error) {
	it := p.next()

	if it.typ != itemCd {
		return nil, fmt.Errorf("Invalid item: %v", it)
	}

	n := NewCdNode(it.pos)

	it = p.peek()

	if it.typ != itemArg && it.typ != itemString && it.typ != itemVariable && it.typ != itemConcat {
		p.backup(it)
		return n, nil
	}

	arg, err := p.getArgument(true)

	if err != nil {
		return nil, err
	}

	n.SetDir(arg)

	return n, nil
}

func (p *Parser) parseSet() (Node, error) {
	it := p.next()

	if it.typ != itemSetEnv {
		return nil, fmt.Errorf("Failed sanity check. Unexpected %v", it)
	}

	pos := it.pos

	it = p.next()

	if it.typ != itemVarName {
		return nil, fmt.Errorf("Unexpected token %v, expected variable", it)
	}

	n := NewSetAssignmentNode(pos, it.val)

	return n, nil
}

func (p *Parser) getArgument(allowArg bool) (*Arg, error) {
	it := p.next()

	if it.typ != itemString && it.typ != itemVariable && it.typ != itemArg {
		return nil, fmt.Errorf("Unexpected token %v. Expected itemString, itemVariable or itemArg", it)
	}

	firstToken := it

	it = p.peek()

	if it.typ == itemConcat {
		return p.getConcatArg(firstToken)
	}

	if firstToken.typ == itemArg && !allowArg {
		return nil, fmt.Errorf("Unquoted string not allowed at pos %d (%s)", it.pos, it.val)
	}

	arg := NewArg(firstToken.pos, 0)

	if firstToken.typ == itemVariable {
		arg.SetArgType(ArgVariable)
		arg.SetString(firstToken.val)
	} else if firstToken.typ == itemString {
		arg.SetArgType(ArgQuoted)
		arg.SetString(firstToken.val)
	} else {
		arg.SetArgType(ArgUnquoted)
		arg.SetString(firstToken.val)
	}

	return arg, nil
}

func (p *Parser) getConcatArg(firstToken item) (*Arg, error) {
	var it item
	parts := make([]*Arg, 0, 4)

	firstArg := NewArg(firstToken.pos, 0)
	firstArg.SetItem(firstToken)

	parts = append(parts, firstArg)

hasConcat:
	it = p.peek()

	if it.typ == itemConcat {
		p.ignore()

		it = p.next()

		if it.typ == itemString || it.typ == itemVariable || it.typ == itemArg {
			carg := NewArg(it.pos, 0)
			carg.SetItem(it)
			parts = append(parts, carg)
			goto hasConcat
		} else {
			return nil, fmt.Errorf("Unexpected token %v", it)
		}
	}

	arg := NewArg(firstToken.pos, ArgConcat)
	arg.SetConcat(parts)

	return arg, nil
}

func (p *Parser) parseAssignment() (Node, error) {
	varIt := p.next()

	if varIt.typ != itemVarName {
		return nil, fmt.Errorf("Invalid item: %v")
	}

	it := p.next()

	if it.typ != itemAssign && it.typ != itemAssignCmd {
		return nil, newError("Unexpected token %v, expected '=' or '<='", it)
	}

	if it.typ == itemAssign {
		return p.parseAssignValue(varIt)
	}

	return p.parseAssignCmdOut(varIt)
}

func (p *Parser) parseAssignValue(name item) (Node, error) {
	n := NewAssignmentNode(name.pos)
	n.SetVarName(name.val)

	it := p.peek()

	if it.typ == itemVariable || it.typ == itemString {
		arg, err := p.getArgument(false)

		if err != nil {
			return nil, err
		}

		n.SetValueList(append(make([]*Arg, 0, 1), arg))
	} else if it.typ == itemListOpen {
		p.next()

		values := make([]*Arg, 0, 128)

		for it = p.next(); it.typ == itemArg || it.typ == itemString || it.typ == itemVariable; it = p.next() {
			arg := NewArg(it.pos, 0)
			arg.SetItem(it)
			values = append(values, arg)
		}

		if it.typ != itemListClose {
			return nil, newUnfinishedListError()
		}

		n.SetValueList(values)
	} else {
		return nil, fmt.Errorf("Unexpected token '%v'", it)
	}

	return n, nil
}

func (p *Parser) parseAssignCmdOut(name item) (Node, error) {
	n := NewCmdAssignmentNode(name.pos, name.val)

	it := p.peek()

	if it.typ != itemCommand && it.typ != itemFnInv {
		return nil, newError("Invalid token %v. Expected command or function invocation", it)
	}

	if it.typ == itemCommand {
		cmd, err := p.parseCommand()

		if err != nil {
			return nil, err
		}

		n.SetCommand(cmd)
		return n, nil
	}

	fn, err := p.parseFnInv()

	if err != nil {
		return nil, err
	}

	n.SetCommand(fn)
	return n, nil
}

func (p *Parser) parseRfork() (Node, error) {
	it := p.next()

	if it.typ != itemRfork {
		return nil, fmt.Errorf("Invalid command: %v", it)
	}

	n := NewRforkNode(it.pos)

	it = p.next()

	if it.typ != itemRforkFlags {
		return nil, fmt.Errorf("rfork requires one or more of the following flags: %s", rforkFlags)
	}

	arg := NewArg(it.pos, ArgUnquoted)
	arg.SetString(it.val)
	n.SetFlags(arg)

	it = p.peek()

	if it.typ == itemLeftBlock {
		p.ignore() // ignore lookaheaded symbol
		p.openblocks++

		n.tree = NewTree("rfork block")
		r, err := p.parseBlock()

		if err != nil {
			return nil, err
		}

		n.tree.Root = r
	}

	return n, nil
}

func (p *Parser) parseIf() (Node, error) {
	it := p.next()

	n := NewIfNode(it.pos)

	it = p.next()

	if it.typ != itemString && it.typ != itemVariable {
		return nil, fmt.Errorf("if requires an lvalue of type string or variable. Found %v", it)
	}

	if it.typ == itemString {
		arg := NewArg(it.pos, ArgQuoted)
		arg.SetString(it.val)
		n.SetLvalue(arg)
	} else if it.typ == itemVariable {
		arg := NewArg(it.pos, ArgVariable)
		arg.SetString(it.val)
		n.SetLvalue(arg)
	} else {
		return nil, newError("Unexpected token %v, expected itemString or itemVariable", it)
	}

	it = p.next()

	if it.typ != itemComparison {
		return nil, fmt.Errorf("Expected comparison. but found %v", it)
	}

	if it.val != "==" && it.val != "!=" {
		return nil, fmt.Errorf("Invalid if operator '%s'. Valid comparison operators are '==' and '!='", it.val)
	}

	n.SetOp(it.val)

	it = p.next()

	if it.typ != itemString && it.typ != itemVariable {
		return nil, fmt.Errorf("if requires an rvalue of type string or variable. Found %v", it)
	}

	if it.typ == itemString {
		arg := NewArg(it.pos, ArgQuoted)
		arg.SetString(it.val)
		n.SetRvalue(arg)
	} else {
		arg := NewArg(it.pos, ArgUnquoted)
		arg.SetString(it.val)
		n.SetRvalue(arg)
	}

	it = p.next()

	if it.typ != itemLeftBlock {
		return nil, fmt.Errorf("Expected '{' but found %v", it)
	}

	p.openblocks++

	r, err := p.parseBlock()

	if err != nil {
		return nil, err
	}

	ifTree := NewTree("if block")
	ifTree.Root = r
	n.SetIfTree(ifTree)

	it = p.peek()

	if it.typ == itemElse {
		p.next()

		elseBlock, elseIf, err := p.parseElse()

		if err != nil {
			return nil, err
		}

		elseTree := NewTree("else tree")
		elseTree.Root = elseBlock

		n.SetElseIf(elseIf)
		n.SetElseTree(elseTree)
	}

	return n, nil
}

func (p *Parser) parseFnArgs() ([]string, error) {
	args := make([]string, 0, 16)

	for {
		it := p.next()

		if it.typ == itemRightParen {
			break
		} else if it.typ == itemVarName {
			args = append(args, it.val)
		} else {
			return nil, fmt.Errorf("Unexpected token %v. Expected identifier or ')'", it)
		}

	}

	return args, nil
}

func (p *Parser) parseFnDecl() (Node, error) {
	it := p.next()

	n := NewFnDeclNode(it.pos, "")

	it = p.next()

	if it.typ == itemVarName {
		n.SetName(it.val)

		it = p.next()
	}

	if it.typ != itemLeftParen {
		return nil, newError("Unexpected token %v. Expected '('", it)
	}

	args, err := p.parseFnArgs()

	if err != nil {
		return nil, err
	}

	for _, arg := range args {
		n.AddArg(arg)
	}

	it = p.next()

	if it.typ != itemLeftBlock {
		return nil, newError("Unexpected token %v. Expected '{'", it)
	}

	p.openblocks++

	tree := NewTree(fmt.Sprintf("fn %s body", n.Name()))

	r, err := p.parseBlock()

	if err != nil {
		return nil, err
	}

	tree.Root = r

	n.SetTree(tree)

	return n, nil

}

func (p *Parser) parseFnInv() (Node, error) {
	it := p.next()

	n := NewFnInvNode(it.pos, it.val)

	it = p.next()

	if it.typ != itemLeftParen {
		return nil, newError("Invalid token %v. Expected '('", it)
	}

	for {
		it = p.next()

		if it.typ == itemString || it.typ == itemVariable {
			arg := NewArg(it.pos, 0)

			if it.typ == itemString {
				arg.SetArgType(ArgQuoted)
			} else {
				arg.SetArgType(ArgVariable)
			}

			arg.SetString(it.val)
			n.AddArg(arg)
		} else if it.typ == itemRightParen {
			break
		} else {
			return nil, newError("Unexpected token %v", it)
		}
	}

	return n, nil
}

func (p *Parser) parseElse() (*ListNode, bool, error) {
	it := p.next()

	if it.typ == itemLeftBlock {
		p.openblocks++

		elseBlock, err := p.parseBlock()

		if err != nil {
			return nil, false, err
		}

		return elseBlock, false, nil
	}

	if it.typ == itemIf {
		p.backup(it)

		ifNode, err := p.parseIf()

		if err != nil {
			return nil, false, err
		}

		block := NewListNode()
		block.Push(ifNode)

		return block, true, nil
	}

	return nil, false, fmt.Errorf("Unexpected token: %v", it)
}

func (p *Parser) parseBindFn() (Node, error) {
	bindIt := p.next()

	nameIt := p.next()

	if nameIt.typ != itemVarName {
		return nil, newError("Expected identifier, but found '%v'", nameIt)
	}

	cmdIt := p.next()

	if cmdIt.typ != itemVarName {
		return nil, newError("Expected identifier, but found '%v'", cmdIt)
	}

	n := NewBindFnNode(bindIt.pos, nameIt.val, cmdIt.val)
	return n, nil
}

func (p *Parser) parseDump() (Node, error) {
	dumpIt := p.next()

	dump := NewDumpNode(dumpIt.pos)

	fnameIt := p.peek()

	if fnameIt.typ != itemString && fnameIt.typ != itemVariable && fnameIt.typ != itemArg {
		return dump, nil
	}

	p.next()

	arg := NewArg(fnameIt.pos, 0)
	arg.SetItem(fnameIt)

	dump.SetFilename(arg)

	return dump, nil
}

func (p *Parser) parseReturn() (Node, error) {
	retIt := p.next()

	ret := NewReturnNode(retIt.pos)

	valueIt := p.peek()

	if valueIt.typ != itemString && valueIt.typ != itemVariable && valueIt.typ != itemListOpen {
		return ret, nil
	}

	p.next()

	if valueIt.typ == itemListOpen {
		values := make([]*Arg, 0, 128)

		for valueIt = p.next(); valueIt.typ == itemArg || valueIt.typ == itemString || valueIt.typ == itemVariable; valueIt = p.next() {
			arg := NewArg(valueIt.pos, 0)
			arg.SetItem(valueIt)
			values = append(values, arg)
		}

		if valueIt.typ != itemListClose {
			return nil, newUnfinishedListError()
		}

		ret.SetReturn(values)
		return ret, nil
	}

	values := make([]*Arg, 1)

	arg := NewArg(valueIt.pos, 0)
	arg.SetItem(valueIt)
	values[0] = arg

	ret.SetReturn(values)
	return ret, nil
}

func (p *Parser) parseComment() (Node, error) {
	it := p.next()

	if it.typ != itemComment {
		return nil, fmt.Errorf("Invalid comment: %v", it)
	}

	return NewCommentNode(it.pos, it.val), nil
}

func (p *Parser) parseStatement() (Node, error) {
	it := p.peek()

	switch it.typ {
	case itemError:
		return nil, fmt.Errorf("Syntax error: %s", it.val)
	case itemBuiltin:
		panic("not implemented")
	case itemImport:
		return p.parseImport()
	case itemShowEnv:
		return p.parseShowEnv()
	case itemSetEnv:
		return p.parseSet()
	case itemVarName:
		return p.parseAssignment()
	case itemCommand:
		return p.parseCommand()
	case itemRfork:
		return p.parseRfork()
	case itemCd:
		return p.parseCd()
	case itemComment:
		return p.parseComment()
	case itemIf:
		return p.parseIf()
	case itemFnDecl:
		return p.parseFnDecl()
	case itemFnInv:
		return p.parseFnInv()
	case itemBindFn:
		return p.parseBindFn()
	case itemDump:
		return p.parseDump()
	case itemReturn:
		return p.parseReturn()
	}

	return nil, fmt.Errorf("Unexpected token parsing statement '%+v'", it)
}

func (p *Parser) parseBlock() (*ListNode, error) {
	ln := NewListNode()

	for {
		it := p.peek()

		switch it.typ {
		case 0, itemEOF:
			goto finish
		case itemError:
			return nil, fmt.Errorf("Syntax error: %s", it.val)
		case itemLeftBlock:
			p.ignore()

			return nil, errors.New("Parser error: Unexpected '{'")
		case itemRightBlock:
			p.ignore()

			if p.openblocks <= 0 {
				return nil, errors.New("Parser error: No block open for close")
			}

			p.openblocks--
			return ln, nil
		default:
			n, err := p.parseStatement()

			if err != nil {
				return nil, err
			}

			ln.Push(n)
		}
	}

finish:
	if p.openblocks != 0 {
		return nil, newUnfinishedBlockError()
	}

	return ln, nil
}

// NewTree creates a new AST tree
func NewTree(name string) *Tree {
	return &Tree{
		Name: name,
	}
}
