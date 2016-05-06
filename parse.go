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

func (p *Parser) parseCommand() (Node, error) {
	it := p.next()

	// paranoia check
	if it.typ != itemCommand {
		return nil, fmt.Errorf("Invalid command: %v", it)
	}

	n := NewCommandNode(it.pos, it.val)

	for {
		it = p.next()

		switch it.typ {
		case itemArg, itemString:
			var arg Arg

			if it.typ == itemString {
				arg = NewArg(it.pos, it.val, true)
			} else {
				arg = NewArg(it.pos, it.val, false)
			}

			n.AddArg(arg)
		case itemRedirRight:
			redir, err := p.parseRedirection(it)

			if err != nil {
				return nil, err
			}

			n.AddRedirect(redir)
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

	if it.typ != itemRedirLBracket && it.typ != itemRedirFile &&
		it.typ != itemRedirNetAddr {
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

	if it.typ != itemRedirFile && it.typ != itemRedirNetAddr {
		if rval != redirMapNoValue || lval != redirMapNoValue {
			return redir, nil
		}

		return nil, fmt.Errorf("Unexpected token '%v'", it)
	}

	redir.SetLocation(it.val)

	p.next()

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
		n.SetPath(NewArg(it.pos, it.val, true))
	} else {
		n.SetPath(NewArg(it.pos, it.val, false))
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

	it = p.next()

	if it.typ != itemArg && it.typ != itemString && it.typ != itemVariable {
		p.backup(it)
		return n, nil
	}

	if it.typ == itemString {
		n.SetDir(NewArg(it.pos, it.val, true))
	} else if it.typ == itemArg || it.typ == itemVariable {
		n.SetDir(NewArg(it.pos, it.val, false))
	}

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

func (p *Parser) parseAssignment() (Node, error) {
	it := p.next()

	if it.typ != itemVarName {
		return nil, fmt.Errorf("Invalid item: %v")
	}

	n := NewAssignmentNode(it.pos)
	n.SetVarName(it.val)

	it = p.next()

	if it.typ == itemVariable || it.typ == itemString {
		elems := make([]ElemNode, 0, 1)
		elem := ElemNode{
			elem:    it.val,
			concats: make([]string, 0, 16),
		}

		elems = append(elems, elem)

		n.SetValueList(elems)

		firstConcat := false

	hasConcat:
		it = p.peek()

		if it.typ == itemConcat {
			p.ignore()

			if !firstConcat {
				firstConcat = true
				elem.concats = append(elem.concats, elem.elem)
				elem.elem = ""
			}

			it = p.next()

			if it.typ == itemString || it.typ == itemVariable {
				elem.concats = append(elem.concats, it.val)
				elems[0] = elem
				n.SetValueList(elems)
				goto hasConcat
			} else {
				return nil, fmt.Errorf("Unexpected token %v", it)
			}
		}

	} else if it.typ == itemListOpen {
		values := make([]ElemNode, 0, 128)

		for it = p.next(); it.typ == itemListElem; it = p.next() {
			values = append(values, ElemNode{
				elem: it.val,
			})
		}

		if it.typ != itemListClose {
			return nil, fmt.Errorf("list variable assignment wrong. Expected ')' but found '%v'", it)
		}

		n.SetValueList(values)
	} else {
		return nil, fmt.Errorf("Unexpected token '%v'", it)
	}

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

	n.SetFlags(NewArg(it.pos, it.val, false))

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
		n.SetLvalue(NewArg(it.pos, it.val, true))
	} else {
		n.SetLvalue(NewArg(it.pos, it.val, false))
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
		n.SetRvalue(NewArg(it.pos, it.val, true))
	} else {
		n.SetRvalue(NewArg(it.pos, it.val, false))
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

			return nil, errors.New("Parser error: Blocks are only allowed inside rfork")
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
		return nil, fmt.Errorf("Open '{' not closed")
	}

	return ln, nil
}

// NewTree creates a new AST tree
func NewTree(name string) *Tree {
	return &Tree{
		Name: name,
	}
}
