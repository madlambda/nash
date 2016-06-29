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

		keywordParsers map[itemType]parserFn
	}

	parserFn func() (Node, error)
)

// NewParser creates a new parser
func NewParser(name, content string) *Parser {
	p := &Parser{
		name:    name,
		content: content,
		l:       lex(name, content),
	}

	p.keywordParsers = map[itemType]parserFn{
		itemBuiltin:    p.parseBuiltin,
		itemCd:         p.parseCd,
		itemFor:        p.parseFor,
		itemIf:         p.parseIf,
		itemFnDecl:     p.parseFnDecl,
		itemFnInv:      p.parseFnInv,
		itemReturn:     p.parseReturn,
		itemImport:     p.parseImport,
		itemShowEnv:    p.parseShowEnv,
		itemSetEnv:     p.parseSet,
		itemRfork:      p.parseRfork,
		itemBindFn:     p.parseBindFn,
		itemDump:       p.parseDump,
		itemAssign:     p.parseAssignment,
		itemIdentifier: p.parseAssignment,
		itemCommand:    p.parseCommand,
		itemComment:    p.parseComment,
		itemError:      p.parseError,
	}

	return p
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

func (p *Parser) parseVariable() (*Arg, error) {
	it := p.next()

	if it.typ != itemVariable {
		return nil, newError("Unexpected token %v. ", it)
	}

	arg := NewArg(it.pos, ArgVariable)
	arg.SetString(it.val)

	it = p.peek()

	if it.typ == itemBracketOpen {
		p.ignore()
		it = p.next()

		if it.typ != itemNumber && it.typ != itemVariable {
			return nil, newError("Expected number or variable in index. Found %v", it)
		}

		var index *Arg

		if it.typ == itemNumber {
			index = NewArg(it.pos, ArgNumber)
		} else {
			index = NewArg(it.pos, ArgVariable)
		}

		index.SetString(it.val)
		arg.SetIndex(index)

		it = p.next()

		if it.typ != itemBracketClose {
			return nil, newError("Unexpected token %v. Expecting ']'", it)
		}
	}

	return arg, nil
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

		if !p.insidePipe {
			break
		}
	}

	return n, nil
}

func (p *Parser) parseCommand() (Node, error) {
	it := p.next()

	n := NewCommandNode(it.pos, it.val)

cmdLoop:
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
			break cmdLoop
		}
	}

	if p.insidePipe {
		p.insidePipe = false
	}

	return n, nil
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

func (p *Parser) parseBuiltin() (Node, error) {
	it := p.next()

	node, err := p.parseStatement()

	if err != nil {
		return nil, err
	}

	if node.Type() != NodeCd {
		return nil, newError("'builtin' must be used only with 'cd' keyword")
	}

	return NewBuiltinNode(it.pos, node), nil
}

func (p *Parser) parseImport() (Node, error) {
	it := p.next()

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

	pos := it.pos

	it = p.next()

	if it.typ != itemIdentifier {
		return nil, fmt.Errorf("Unexpected token %v, expected variable", it)
	}

	n := NewSetAssignmentNode(pos, it.val)

	return n, nil
}

func (p *Parser) getArgument(allowArg bool) (*Arg, error) {
	var err error

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

		if it.typ == itemBracketOpen {
			p.ignore()
			it = p.next()

			var indexArg *Arg

			if it.typ == itemNumber {
				indexArg = NewArg(it.pos, ArgNumber)
				indexArg.SetString(it.val)
			} else if it.typ == itemVariable {
				p.backup(it)
				indexArg, err = p.getArgument(false)

				if err != nil {
					return nil, err
				}
			} else {
				return nil, newError("Invalid index type: %v", it)
			}

			arg.SetIndex(indexArg)

			it = p.next()

			if it.typ != itemBracketClose {
				return nil, newError("Unexpected token %v. Expected ']'", it)
			}
		}
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

		n.SetValue(arg)
	} else if it.typ == itemListOpen {
		lit := p.next()

		values := make([]*Arg, 0, 128)

		for it = p.next(); it.typ == itemArg || it.typ == itemString || it.typ == itemVariable; it = p.next() {
			arg := NewArg(it.pos, 0)
			arg.SetItem(it)
			values = append(values, arg)
		}

		if it.typ != itemListClose {
			return nil, newUnfinishedListError()
		}

		listArg := NewArg(lit.pos, ArgList)
		listArg.SetList(values)

		n.SetValue(listArg)
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

	n := NewRforkNode(it.pos)

	it = p.next()

	if it.typ != itemRforkFlags {
		return nil, fmt.Errorf("rfork requires one or more of the following flags: %s", rforkFlags)
	}

	arg := NewArg(it.pos, ArgUnquoted)
	arg.SetString(it.val)
	n.SetFlags(arg)

	it = p.peek()

	if it.typ == itemBracesOpen {
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

	it = p.peek()

	if it.typ != itemString && it.typ != itemVariable {
		return nil, fmt.Errorf("if requires an lvalue of type string or variable. Found %v", it)
	}

	if it.typ == itemString {
		p.next()
		arg := NewArg(it.pos, ArgQuoted)
		arg.SetString(it.val)
		n.SetLvalue(arg)
	} else if it.typ == itemVariable {
		arg, err := p.parseVariable()

		if err != nil {
			return nil, err
		}

		n.SetLvalue(arg)
	} else {
		return nil, newError("Unexpected token %v, expected itemString or itemVariable", it)
	}

	it = p.next()

	if it.typ != itemComparison {
		return nil, fmt.Errorf("Expected comparison, but found %v", it)
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

	if it.typ != itemBracesOpen {
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

		if it.typ == itemParenClose {
			break
		} else if it.typ == itemIdentifier {
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

	if it.typ == itemIdentifier {
		n.SetName(it.val)

		it = p.next()
	}

	if it.typ != itemParenOpen {
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

	if it.typ != itemBracesOpen {
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

	if it.typ != itemParenOpen {
		return nil, newError("Invalid token %v. Expected '('", it)
	}

	for {
		it = p.peek()

		if it.typ == itemString || it.typ == itemVariable {
			arg, err := p.getArgument(false)

			if err != nil {
				return nil, err
			}

			n.AddArg(arg)
		} else if it.typ == itemParenClose {
			p.next()
			break
		} else {
			return nil, newError("Unexpected token %v", it)
		}
	}

	return n, nil
}

func (p *Parser) parseElse() (*ListNode, bool, error) {
	it := p.next()

	if it.typ == itemBracesOpen {
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

	if nameIt.typ != itemIdentifier {
		return nil, newError("Expected identifier, but found '%v'", nameIt)
	}

	cmdIt := p.next()

	if cmdIt.typ != itemIdentifier {
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

	retIt = p.next()

	retPos := retIt.pos

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

		listArg := NewArg(retPos, ArgList)
		listArg.SetList(values)

		ret.SetReturn(listArg)
		return ret, nil
	}

	arg := NewArg(valueIt.pos, 0)
	arg.SetItem(valueIt)

	ret.SetReturn(arg)
	return ret, nil
}

func (p *Parser) parseFor() (Node, error) {
	it := p.next()

	forStmt := NewForNode(it.pos)

	it = p.peek()

	if it.typ != itemIdentifier {
		goto forBlockParse
	}

	p.next()

	forStmt.SetIdentifier(it.val)

	it = p.next()

	if it.typ != itemForIn {
		return nil, newError("Expected 'in' but found %q", it)
	}

	it = p.next()

	if it.typ != itemVariable {
		return nil, newError("Expected variable but found %q", it)
	}

	forStmt.SetInVar(it.val)

forBlockParse:
	it = p.peek()

	if it.typ != itemBracesOpen {
		return nil, newError("Expected '{' but found %q", it)
	}

	p.ignore() // ignore lookaheaded symbol
	p.openblocks++

	tree := NewTree("for block")

	r, err := p.parseBlock()

	if err != nil {
		return nil, err
	}

	tree.Root = r
	forStmt.SetTree(tree)

	return forStmt, nil
}

func (p *Parser) parseComment() (Node, error) {
	it := p.next()

	return NewCommentNode(it.pos, it.val), nil
}

func (p *Parser) parseStatement() (Node, error) {
	it := p.peek()

	if fn, ok := p.keywordParsers[it.typ]; ok {
		return fn()
	}

	return nil, fmt.Errorf("Unexpected token parsing statement '%+v'", it)
}

func (p *Parser) parseError() (Node, error) {
	it := p.next()

	return nil, newError(it.val)
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
		case itemBracesOpen:
			p.ignore()

			return nil, errors.New("Parser error: Unexpected '{'")
		case itemBracesClose:
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
