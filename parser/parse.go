package parser

import (
	"fmt"

	"strconv"

	"github.com/NeowayLabs/nash/ast"
	"github.com/NeowayLabs/nash/errors"
	"github.com/NeowayLabs/nash/scanner"
	"github.com/NeowayLabs/nash/token"
)

type (
	// Parser parses an nash file
	Parser struct {
		name       string // filename or name of the buffer
		content    string
		l          *scanner.Lexer
		tok        *scanner.Token // token saved for lookahead
		openblocks int

		insidePipe bool

		keywordParsers map[token.Token]parserFn
	}

	parserFn func() (ast.Node, error)
)

// NewParser creates a new parser
func NewParser(name, content string) *Parser {
	p := &Parser{
		name:    name,
		content: content,
		l:       scanner.Lex(name, content),
	}

	p.keywordParsers = map[token.Token]parserFn{
		token.Builtin: p.parseBuiltin,
		token.Cd:      p.parseCd,
		token.For:     p.parseFor,
		token.If:      p.parseIf,
		token.FnDecl:  p.parseFnDecl,
		token.FnInv:   p.parseFnInv,
		token.Return:  p.parseReturn,
		token.Import:  p.parseImport,
		token.SetEnv:  p.parseSet,
		token.Rfork:   p.parseRfork,
		token.BindFn:  p.parseBindFn,
		token.Dump:    p.parseDump,
		token.Assign:  p.parseAssignment,
		token.Ident:   p.parseAssignment,
		token.Command: p.parseCommand,
		token.Comment: p.parseComment,
		token.Illegal: p.parseError,
	}

	return p
}

// Parse starts the parsing.
func (p *Parser) Parse() (*ast.Tree, error) {
	root, err := p.parseBlock()

	if err != nil {
		return nil, err
	}

	tr := ast.NewTree(p.name)
	tr.Root = root

	return tr, nil
}

// next returns the next item from lookahead buffer if not empty or
// from the Lexer
func (p *Parser) next() scanner.Token {
	if p.tok != nil {
		t := p.tok
		p.tok = nil
		return *t
	}

	return <-p.l.Tokens
}

// backup puts the item into the lookahead buffer
func (p *Parser) backup(it scanner.Token) error {
	if p.tok != nil {
		return errors.NewError("only one slot for backup/lookahead")
	}

	p.tok = &it

	return nil
}

// ignores the next item
func (p *Parser) ignore() {
	if p.tok != nil {
		p.tok = nil
	} else {
		<-p.l.Tokens
	}
}

// peek gets but do not discards the next item (lookahead)
func (p *Parser) peek() scanner.Token {
	i := p.next()
	p.tok = &i
	return i
}

func (p *Parser) parseVariable() (ast.Expr, error) {
	it := p.next()

	if it.Type() != token.Variable {
		return nil, errors.NewError("Unexpected token %v. ", it)
	}

	variable := ast.NewVarExpr(it.Pos(), it.Value())

	it = p.peek()

	if it.Type() == token.LBrack {
		p.ignore()
		it = p.next()

		if it.Type() != token.Number && it.Type() != token.Variable {
			return nil, errors.NewError("Expected number or variable in index. Found %v", it)
		}

		var index ast.Expr

		if it.Type() == token.Number {
			// only supports base10
			intval, err := strconv.Atoi(it.Value())

			if err != nil {
				return nil, err
			}

			index = ast.NewIntExpr(it.Pos(), intval)
		} else {
			index = ast.NewVarExpr(it.Pos(), it.Value())
		}

		it = p.next()

		if it.Type() != token.RBrack {
			return nil, errors.NewError("Unexpected token %v. Expecting ']'", it)
		}

		return ast.NewIndexExpr(variable.Position(), variable, index), nil
	}

	return variable, nil
}

func (p *Parser) parsePipe(first *ast.CommandNode) (ast.Node, error) {
	it := p.next()

	n := ast.NewPipeNode(it.Pos())

	n.AddCmd(first)

	for it = p.peek(); it.Type() == token.Command; it = p.peek() {
		cmd, err := p.parseCommand()

		if err != nil {
			return nil, err
		}

		n.AddCmd(cmd.(*ast.CommandNode))

		if !p.insidePipe {
			break
		}
	}

	return n, nil
}

func (p *Parser) parseCommand() (ast.Node, error) {
	it := p.next()

	n := ast.NewCommandNode(it.Pos(), it.Value())

cmdLoop:
	for {
		it = p.peek()

		switch it.Type() {
		case token.Arg, token.String, token.Variable:
			arg, err := p.getArgument(true)

			if err != nil {
				return nil, err
			}

			n.AddArg(arg)
		case token.Concat:
			return nil, fmt.Errorf("Unexpected '+' at pos %d\n", it.Pos())
		case token.RedirRight:
			p.next()
			redir, err := p.parseRedirection(it)

			if err != nil {
				return nil, err
			}

			n.AddRedirect(redir)
		case token.Pipe:
			if p.insidePipe {
				p.next()
				return n, nil
			}

			p.insidePipe = true
			return p.parsePipe(n)
		case token.EOF:
			return n, nil
		case token.Illegal:
			return nil, fmt.Errorf("Syntax error: %s", it.Value())
		default:
			break cmdLoop
		}
	}

	if p.insidePipe {
		p.insidePipe = false
	}

	return n, nil
}

func (p *Parser) parseRedirection(it scanner.Token) (*ast.RedirectNode, error) {
	var (
		lval, rval int = ast.RedirMapNoValue, ast.RedirMapNoValue
		err        error
	)

	redir := ast.NewRedirectNode(it.Pos())

	it = p.peek()

	if it.Type() != token.LBrack && it.Type() != token.String && it.Type() != token.Arg && it.Type() != token.Variable {
		return nil, fmt.Errorf("Unexpected token: %v", it)
	}

	// [
	if it.Type() == token.LBrack {
		p.next()
		it = p.peek()

		if it.Type() != token.RedirMapLSide {
			return nil, fmt.Errorf("Expected lefthand side of redirection map, but found '%s'",
				it.Value())
		}

		lval, err = strconv.Atoi(it.Value())

		if err != nil {
			return nil, fmt.Errorf("Redirection map expects integers. Found: %s", it.Value())
		}

		p.next()
		it = p.peek()

		if it.Type() != token.Assign && it.Type() != token.RBrack {
			return nil, fmt.Errorf("Unexpected token '%v'", it)
		}

		// [xxx=
		if it.Type() == token.Assign {
			p.next()
			it = p.peek()

			if it.Type() != token.RedirMapRSide && it.Type() != token.RBrack {
				return nil, fmt.Errorf("Unexpected token '%v'", it)
			}

			if it.Type() == token.RedirMapRSide {
				rval, err = strconv.Atoi(it.Value())

				if err != nil {
					return nil, fmt.Errorf("Redirection map expects integers. Found: %s", it.Value())
				}

				p.next()
				it = p.peek()
			} else {
				rval = ast.RedirMapSupress
			}
		}

		if it.Type() != token.RBrack {
			return nil, fmt.Errorf("Unexpected token '%v'", it)
		}

		// [xxx=yyy]

		redir.SetMap(lval, rval)

		p.next()
		it = p.peek()
	}

	if it.Type() != token.String && it.Type() != token.Arg && it.Type() != token.Variable {
		if rval != ast.RedirMapNoValue || lval != ast.RedirMapNoValue {
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

func (p *Parser) parseBuiltin() (ast.Node, error) {
	it := p.next()

	node, err := p.parseStatement()

	if err != nil {
		return nil, err
	}

	if node.Type() != ast.NodeCd {
		return nil, errors.NewError("'builtin' must be used only with 'cd' keyword")
	}

	return ast.NewBuiltinNode(it.Pos(), node), nil
}

func (p *Parser) parseImport() (ast.Node, error) {
	it := p.next()

	n := ast.NewImportNode(it.Pos())

	it = p.next()

	if it.Type() != token.Arg && it.Type() != token.String {
		return nil, fmt.Errorf("Unexpected token %v", it)
	}

	if it.Type() == token.String {
		arg := ast.NewArg(it.Pos(), ast.ArgQuoted)
		arg.SetString(it.Value())
		n.SetPath(arg)
	} else if it.Type() == token.Arg {
		arg := ast.NewArg(it.Pos(), ast.ArgUnquoted)
		arg.SetString(it.Value())
		n.SetPath(arg)
	} else {
		return nil, fmt.Errorf("Parser error: Invalid token '%v' for import path", it)
	}

	return n, nil
}

func (p *Parser) parseCd() (ast.Node, error) {
	it := p.next()

	n := ast.NewCdNode(it.Pos())

	it = p.peek()

	if it.Type() != token.Arg && it.Type() != token.String && it.Type() != token.Variable && it.Type() != token.Concat {
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

func (p *Parser) parseSet() (ast.Node, error) {
	it := p.next()

	pos := it.Pos()

	it = p.next()

	if it.Type() != token.Ident {
		return nil, fmt.Errorf("Unexpected token %v, expected variable", it)
	}

	n := ast.NewSetAssignmentNode(pos, it.Value())

	return n, nil
}

func (p *Parser) getArgument(allowArg bool) (*ast.Arg, error) {
	var err error

	it := p.next()

	if it.Type() != token.String && it.Type() != token.Variable && it.Type() != token.Arg {
		return nil, fmt.Errorf("Unexpected token %v. Expected %s, %s or %s",
			it, token.String, token.Variable, token.Arg)
	}

	firstToken := it

	it = p.peek()

	if it.Type() == token.Concat {
		return p.getConcatArg(firstToken)
	}

	if firstToken.Type() == token.Arg && !allowArg {
		return nil, fmt.Errorf("Unquoted string not allowed at pos %d (%s)", it.Pos(), it.Value())
	}

	arg := ast.NewArg(firstToken.Pos(), 0)

	if firstToken.Type() == token.Variable {
		arg.SetArgType(ast.ArgVariable)
		arg.SetString(firstToken.Value())

		if it.Type() == token.LBrack {
			p.ignore()
			it = p.next()

			var indexArg *ast.Arg

			if it.Type() == token.Number {
				indexArg = ast.NewArg(it.Pos(), ast.ArgNumber)
				indexArg.SetString(it.Value())
			} else if it.Type() == token.Variable {
				p.backup(it)
				indexArg, err = p.getArgument(false)

				if err != nil {
					return nil, err
				}
			} else {
				return nil, errors.NewError("Invalid index type: %v", it)
			}

			arg.SetIndex(indexArg)

			it = p.next()

			if it.Type() != token.RBrack {
				return nil, errors.NewError("Unexpected token %v. Expected ']'", it)
			}
		}
	} else if firstToken.Type() == token.String {
		arg.SetArgType(ast.ArgQuoted)
		arg.SetString(firstToken.Value())
	} else {
		arg.SetArgType(ast.ArgUnquoted)
		arg.SetString(firstToken.Value())
	}

	return arg, nil
}

func (p *Parser) getConcatArg(firstToken scanner.Token) (*ast.Arg, error) {
	var it scanner.Token
	parts := make([]*ast.Arg, 0, 4)

	firstArg := ast.NewArg(firstToken.Pos(), 0)
	firstArg.SetItem(firstToken)

	parts = append(parts, firstArg)

hasConcat:
	it = p.peek()

	if it.Type() == token.Concat {
		p.ignore()

		it = p.next()

		if it.Type() == token.String || it.Type() == token.Variable || it.Type() == token.Arg {
			carg := ast.NewArg(it.Pos(), 0)
			carg.SetItem(it)
			parts = append(parts, carg)
			goto hasConcat
		} else {
			return nil, fmt.Errorf("Unexpected token %v", it)
		}
	}

	arg := ast.NewArg(firstToken.Pos(), ast.ArgConcat)
	arg.SetConcat(parts)

	return arg, nil
}

func (p *Parser) parseAssignment() (ast.Node, error) {
	varIt := p.next()

	it := p.next()

	if it.Type() != token.Assign && it.Type() != token.AssignCmd {
		return nil, errors.NewError("Unexpected token %v, expected '=' or '<='", it)
	}

	if it.Type() == token.Assign {
		return p.parseAssignValue(varIt)
	}

	return p.parseAssignCmdOut(varIt)
}

func (p *Parser) parseAssignValue(name scanner.Token) (ast.Node, error) {
	n := ast.NewAssignmentNode(name.Pos())
	n.SetIdentifier(name.Value())

	it := p.peek()

	if it.Type() == token.Variable || it.Type() == token.String {
		arg, err := p.getArgument(false)

		if err != nil {
			return nil, err
		}

		n.SetValue(arg)
	} else if it.Type() == token.LParen {
		lit := p.next()

		values := make([]*ast.Arg, 0, 128)

		for it = p.next(); it.Type() == token.Arg || it.Type() == token.String || it.Type() == token.Variable; it = p.next() {
			arg := ast.NewArg(it.Pos(), 0)
			arg.SetItem(it)
			values = append(values, arg)
		}

		if it.Type() != token.RParen {
			return nil, errors.NewUnfinishedListError()
		}

		listArg := ast.NewArg(lit.Pos(), ast.ArgList)
		listArg.SetList(values)

		n.SetValue(listArg)
	} else {
		return nil, fmt.Errorf("Unexpected token '%v'", it)
	}

	return n, nil
}

func (p *Parser) parseAssignCmdOut(name scanner.Token) (ast.Node, error) {
	n := ast.NewCmdAssignmentNode(name.Pos(), name.Value())

	it := p.peek()

	if it.Type() != token.Command && it.Type() != token.FnInv {
		return nil, errors.NewError("Invalid token %v. Expected command or function invocation", it)
	}

	if it.Type() == token.Command {
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

func (p *Parser) parseRfork() (ast.Node, error) {
	it := p.next()

	n := ast.NewRforkNode(it.Pos())

	it = p.next()

	if it.Type() != token.String {
		return nil, fmt.Errorf("rfork requires one or more of the following flags: %s", scanner.RforkFlags)
	}

	arg := ast.NewArg(it.Pos(), ast.ArgUnquoted)
	arg.SetString(it.Value())
	n.SetFlags(arg)

	it = p.peek()

	if it.Type() == token.LBrace {
		p.ignore() // ignore lookaheaded symbol
		p.openblocks++

		tree := ast.NewTree("rfork block")
		r, err := p.parseBlock()

		if err != nil {
			return nil, err
		}

		tree.Root = r

		n.SetTree(tree)
	}

	return n, nil
}

func (p *Parser) parseIf() (ast.Node, error) {
	it := p.next()

	n := ast.NewIfNode(it.Pos())

	it = p.peek()

	if it.Type() != token.String && it.Type() != token.Variable {
		return nil, fmt.Errorf("if requires an lvalue of type string or variable. Found %v", it)
	}

	if it.Type() == token.String {
		p.next()
		arg := ast.NewArg(it.Pos(), ast.ArgQuoted)
		arg.SetString(it.Value())
		n.SetLvalue(arg)
	} else if it.Type() == token.Variable {
		arg, err := p.parseVariable()

		if err != nil {
			return nil, err
		}

		n.SetLvalue(arg)
	} else {
		return nil, errors.NewError("Unexpected token %v, expected %v or %v",
			it, token.String, token.Variable)
	}

	it = p.next()

	if it.Type() != token.Equal && it.Type() != token.NotEqual {
		return nil, fmt.Errorf("Expected comparison, but found %v", it)
	}

	if it.Value() != "==" && it.Value() != "!=" {
		return nil, fmt.Errorf("Invalid if operator '%s'. Valid comparison operators are '==' and '!='",
			it.Value())
	}

	n.SetOp(it.Value())

	it = p.next()

	if it.Type() != token.String && it.Type() != token.Variable {
		return nil, fmt.Errorf("if requires an rvalue of type string or variable. Found %v", it)
	}

	if it.Type() == token.String {
		arg := ast.NewArg(it.Pos(), ast.ArgQuoted)
		arg.SetString(it.Value())
		n.SetRvalue(arg)
	} else {
		arg := ast.NewArg(it.Pos(), ast.ArgUnquoted)
		arg.SetString(it.Value())
		n.SetRvalue(arg)
	}

	it = p.next()

	if it.Type() != token.LBrace {
		return nil, fmt.Errorf("Expected '{' but found %v", it)
	}

	p.openblocks++

	r, err := p.parseBlock()

	if err != nil {
		return nil, err
	}

	ifTree := ast.NewTree("if block")
	ifTree.Root = r
	n.SetIfTree(ifTree)

	it = p.peek()

	if it.Type() == token.Else {
		p.next()

		elseBlock, elseIf, err := p.parseElse()

		if err != nil {
			return nil, err
		}

		elseTree := ast.NewTree("else tree")
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

		if it.Type() == token.RParen {
			break
		} else if it.Type() == token.Ident {
			args = append(args, it.Value())
		} else {
			return nil, fmt.Errorf("Unexpected token %v. Expected identifier or ')'", it)
		}

	}

	return args, nil
}

func (p *Parser) parseFnDecl() (ast.Node, error) {
	it := p.next()

	n := ast.NewFnDeclNode(it.Pos(), "")

	it = p.next()

	if it.Type() == token.Ident {
		n.SetName(it.Value())

		it = p.next()
	}

	if it.Type() != token.LParen {
		return nil, errors.NewError("Unexpected token %v. Expected '('", it)
	}

	args, err := p.parseFnArgs()

	if err != nil {
		return nil, err
	}

	for _, arg := range args {
		n.AddArg(arg)
	}

	it = p.next()

	if it.Type() != token.LBrace {
		return nil, errors.NewError("Unexpected token %v. Expected '{'", it)
	}

	p.openblocks++

	tree := ast.NewTree(fmt.Sprintf("fn %s body", n.Name()))

	r, err := p.parseBlock()

	if err != nil {
		return nil, err
	}

	tree.Root = r

	n.SetTree(tree)

	return n, nil

}

func (p *Parser) parseFnInv() (ast.Node, error) {
	it := p.next()

	n := ast.NewFnInvNode(it.Pos(), it.Value())

	it = p.next()

	if it.Type() != token.LParen {
		return nil, errors.NewError("Invalid token %v. Expected '('", it)
	}

	for {
		it = p.peek()

		if it.Type() == token.String || it.Type() == token.Variable {
			arg, err := p.getArgument(false)

			if err != nil {
				return nil, err
			}

			n.AddArg(arg)
		} else if it.Type() == token.RParen {
			p.next()
			break
		} else {
			return nil, errors.NewError("Unexpected token %v", it)
		}
	}

	return n, nil
}

func (p *Parser) parseElse() (*ast.ListNode, bool, error) {
	it := p.next()

	if it.Type() == token.LBrace {
		p.openblocks++

		elseBlock, err := p.parseBlock()

		if err != nil {
			return nil, false, err
		}

		return elseBlock, false, nil
	}

	if it.Type() == token.If {
		p.backup(it)

		ifNode, err := p.parseIf()

		if err != nil {
			return nil, false, err
		}

		block := ast.NewListNode()
		block.Push(ifNode)

		return block, true, nil
	}

	return nil, false, fmt.Errorf("Unexpected token: %v", it)
}

func (p *Parser) parseBindFn() (ast.Node, error) {
	bindIt := p.next()

	nameIt := p.next()

	if nameIt.Type() != token.Ident {
		return nil, errors.NewError("Expected identifier, but found '%v'", nameIt)
	}

	cmdIt := p.next()

	if cmdIt.Type() != token.Ident {
		return nil, errors.NewError("Expected identifier, but found '%v'", cmdIt)
	}

	n := ast.NewBindFnNode(bindIt.Pos(), nameIt.Value(), cmdIt.Value())
	return n, nil
}

func (p *Parser) parseDump() (ast.Node, error) {
	dumpIt := p.next()

	dump := ast.NewDumpNode(dumpIt.Pos())

	fnameIt := p.peek()

	if fnameIt.Type() != token.String && fnameIt.Type() != token.Variable && fnameIt.Type() != token.Arg {
		return dump, nil
	}

	p.next()

	arg := ast.NewArg(fnameIt.Pos(), 0)
	arg.SetItem(fnameIt)

	dump.SetFilename(arg)

	return dump, nil
}

func (p *Parser) parseReturn() (ast.Node, error) {
	retIt := p.next()

	ret := ast.NewReturnNode(retIt.Pos())

	valueIt := p.peek()

	if valueIt.Type() != token.String && valueIt.Type() != token.Variable && valueIt.Type() != token.LParen {
		return ret, nil
	}

	retIt = p.next()

	retPos := retIt.Pos()

	if valueIt.Type() == token.LParen {
		values := make([]*ast.Arg, 0, 128)

		for valueIt = p.next(); valueIt.Type() == token.Arg || valueIt.Type() == token.String || valueIt.Type() == token.Variable; valueIt = p.next() {
			arg := ast.NewArg(valueIt.Pos(), 0)
			arg.SetItem(valueIt)
			values = append(values, arg)
		}

		if valueIt.Type() != token.RParen {
			return nil, errors.NewUnfinishedListError()
		}

		listArg := ast.NewArg(retPos, ast.ArgList)
		listArg.SetList(values)

		ret.SetReturn(listArg)
		return ret, nil
	}

	arg := ast.NewArg(valueIt.Pos(), 0)
	arg.SetItem(valueIt)

	ret.SetReturn(arg)
	return ret, nil
}

func (p *Parser) parseFor() (ast.Node, error) {
	it := p.next()

	forStmt := ast.NewForNode(it.Pos())

	it = p.peek()

	if it.Type() != token.Ident {
		goto forBlockParse
	}

	p.next()

	forStmt.SetIdentifier(it.Value())

	it = p.next()

	if it.Type() != token.ForIn {
		return nil, errors.NewError("Expected 'in' but found %q", it)
	}

	it = p.next()

	if it.Type() != token.Variable {
		return nil, errors.NewError("Expected variable but found %q", it)
	}

	forStmt.SetInVar(it.Value())

forBlockParse:
	it = p.peek()

	if it.Type() != token.LBrace {
		return nil, errors.NewError("Expected '{' but found %q", it)
	}

	p.ignore() // ignore lookaheaded symbol
	p.openblocks++

	tree := ast.NewTree("for block")

	r, err := p.parseBlock()

	if err != nil {
		return nil, err
	}

	tree.Root = r
	forStmt.SetTree(tree)

	return forStmt, nil
}

func (p *Parser) parseComment() (ast.Node, error) {
	it := p.next()

	return ast.NewCommentNode(it.Pos(), it.Value()), nil
}

func (p *Parser) parseStatement() (ast.Node, error) {
	it := p.peek()

	if fn, ok := p.keywordParsers[it.Type()]; ok {
		return fn()
	}

	return nil, fmt.Errorf("Unexpected token parsing statement '%+v'", it)
}

func (p *Parser) parseError() (ast.Node, error) {
	it := p.next()

	return nil, errors.NewError(it.Value())
}

func (p *Parser) parseBlock() (*ast.ListNode, error) {
	ln := ast.NewListNode()

	for {
		it := p.peek()

		switch it.Type() {
		case token.EOF:
			goto finish
		case token.Illegal:
			return nil, fmt.Errorf("Syntax error: %s", it.Value())
		case token.LBrace:
			p.ignore()

			return nil, errors.NewError("Parser error: Unexpected '{'")
		case token.RBrace:
			p.ignore()

			if p.openblocks <= 0 {
				return nil, errors.NewError("Parser error: No block open for close")
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
		return nil, errors.NewUnfinishedBlockError()
	}

	return ln, nil
}
