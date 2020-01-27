package parser

import (
	"fmt"
	"runtime"

	"strconv"

	"github.com/madlambda/nash/ast"
	"github.com/madlambda/nash/errors"
	"github.com/madlambda/nash/scanner"
	"github.com/madlambda/nash/token"
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

	parserFn func(tok scanner.Token) (ast.Node, error)

	exprConfig struct {
		allowArg      bool
		allowVariadic bool
		allowFuncall  bool
		allowConcat   bool
	}
)

// NewParser creates a new parser
func NewParser(name, content string) *Parser {
	p := &Parser{
		name:    name,
		content: content,
		l:       scanner.Lex(name, content),
	}

	p.keywordParsers = map[token.Token]parserFn{
		token.For:     p.parseFor,
		token.If:      p.parseIf,
		token.Fn:      p.parseFnDecl,
		token.Var:     p.parseVar,
		token.Return:  p.parseReturn,
		token.Import:  p.parseImport,
		token.SetEnv:  p.parseSetenv,
		token.Rfork:   p.parseRfork,
		token.BindFn:  p.parseBindFn,
		token.Comment: p.parseComment,
		token.Illegal: p.parseError,
	}

	return p
}

// Parse starts the parsing.
func (p *Parser) Parse() (tr *ast.Tree, err error) {
	var root *ast.BlockNode

	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(runtime.Error); ok {
				panic(r)
			}

			err = r.(error)
		}
	}()

	root, err = p.parseBlock(1, 0)

	if err != nil {
		return nil, err
	}

	tr = ast.NewTree(p.name)
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

	tok := <-p.l.Tokens

	if tok.Type() == token.Illegal {
		panic(errors.NewError(tok.Value()))
	}

	return tok
}

// backup puts the item into the lookahead buffer
func (p *Parser) backup(it scanner.Token) error {
	if p.tok != nil {
		panic(errors.NewError("only one slot for backup/lookahead: %s", it))
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

func (p *Parser) parseBlock(lineStart, columnStart int) (*ast.BlockNode, error) {
	ln := ast.NewBlockNode(token.NewFileInfo(lineStart, columnStart))

	for {
		it := p.peek()

		switch it.Type() {
		case token.EOF:
			goto finish
		case token.LBrace:
			p.ignore()

			return nil, newParserError(it, p.name,
				"Unexpected '{'")
		case token.RBrace:
			p.ignore()

			if p.openblocks <= 0 {
				return nil, newParserError(it, p.name,
					"No block open for close")
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
		return nil, errors.NewUnfinishedBlockError(p.name, p.peek())
	}

	return ln, nil
}

func (p *Parser) parseStatement() (ast.Node, error) {
	it := p.next()
	next := p.peek()

	if fn, ok := p.keywordParsers[it.Type()]; ok {
		return fn(it)
	}

	// statement starting with ident:
	// - fn call
	// - variable assignment
	// - variable exec assignment
	// - Command

	if isFuncall(it.Type(), next.Type()) {
		return p.parseFnInv(it, true)
	}

	if it.Type() == token.Ident {
		if isAssignment(next.Type()) {
			return p.parseAssignment(it)
		}

		return p.parseCommand(it)
	} else if it.Type() == token.Arg {
		return p.parseCommand(it)
	}

	// statement starting with '('
	// -multiline command (echo hello)
	if it.Type() == token.LParen {
		return p.parseCommand(it)
	}

	return nil, newParserError(it, p.name, "Unexpected token parsing statement '%+v'", it)
}

func (p *Parser) parseIndexing() (ast.Expr, error) {
	it := p.next()

	if it.Type() != token.Number && it.Type() != token.Variable {
		return nil, newParserError(it, p.name,
			"Expected number or variable in index. Found %v", it)
	}

	var (
		index ast.Expr
		err   error
	)

	if it.Type() == token.Number {
		// only supports base10
		intval, err := strconv.Atoi(it.Value())

		if err != nil {
			return nil, err
		}

		index = ast.NewIntExpr(it.FileInfo, intval)
	} else {
		index, err = p.parseVariable(&it, false)

		if err != nil {
			return nil, err
		}
	}

	it = p.next()

	if it.Type() != token.RBrack {
		return nil, newParserError(it, p.name,
			"Unexpected token %v. Expecting ']'", it)
	}

	return index, nil
}

func (p *Parser) parseVariable(tok *scanner.Token, allowVararg bool) (ast.Expr, error) {
	var it scanner.Token

	if tok == nil {
		it = p.next()
	} else {
		it = *tok
	}

	if it.Type() != token.Variable {
		return nil, newParserError(it, p.name,
			"Unexpected token %v. Expected VARIABLE", it)
	}

	variadicErr := func(tok scanner.Token) (ast.Node, error) {
		return nil, newParserError(it, p.name,
			"Unexpected token '...'. Varargs allowed only in fn call and fn decl")
	}

	varTok := it
	it = p.peek()
	if it.Type() == token.LBrack {
		variable := ast.NewVarExpr(varTok.FileInfo, varTok.Value())
		p.ignore()
		index, err := p.parseIndexing()
		if err != nil {
			return nil, err
		}

		isVariadic := p.peek().Type() == token.Dotdotdot
		if isVariadic && !allowVararg {
			return variadicErr(p.peek())
		}
		indexedVar := ast.NewIndexVariadicExpr(variable.FileInfo, variable, index, isVariadic)
		if isVariadic {
			p.ignore()
		}
		return indexedVar, nil
	}

	isVariadic := p.peek().Type() == token.Dotdotdot
	if isVariadic {
		if !allowVararg {
			return variadicErr(p.peek())
		}
		p.ignore()
	}

	return ast.NewVarVariadicExpr(varTok.FileInfo, varTok.Value(), isVariadic), nil
}

func (p *Parser) parsePipe(first *ast.CommandNode) (ast.Node, error) {
	it := p.next()

	n := ast.NewPipeNode(it.FileInfo, first.IsMulti())
	first.SetMulti(false)

	n.AddCmd(first)

	for it = p.peek(); it.Type() == token.Ident || it.Type() == token.Arg; it = p.peek() {
		p.next()
		cmd, err := p.parseCommand(it)

		if err != nil {
			return nil, err
		}

		n.AddCmd(cmd.(*ast.CommandNode))

		if !p.insidePipe {
			break
		}
	}

	if n.IsMulti() {
		it = p.peek()
		if it.Type() != token.RParen {
			if it.Type() == token.EOF {
				return nil, errors.NewUnfinishedCmdError(p.name, it)
			}

			return nil, newParserError(it, p.name, "Unexpected symbol '%s'", it)
		}

		p.ignore()
	}

	it = p.peek()

	if it.Type() == token.RBrace {
		return n, nil
	}

	if it.Type() != token.Semicolon {
		return nil, newParserError(it, p.name, "Unexpected symbol %s", it)
	}

	p.ignore()

	return n, nil
}

func (p *Parser) parseCommand(it scanner.Token) (ast.Node, error) {
	isMulti := false

	if it.Type() == token.LParen {
		// multiline command
		isMulti = true

		it = p.next()
	}

	if it.Type() != token.Ident && it.Type() != token.Arg {
		if isMulti && it.Type() == token.EOF {
			return nil, errors.NewUnfinishedCmdError(p.name, it)
		}

		return nil, newParserError(it, p.name, "Unexpected token %v. Expecting IDENT or ARG", it)
	}

	n := ast.NewCommandNode(it.FileInfo, it.Value(), isMulti)

cmdLoop:
	for {
		it = p.peek()

		switch typ := it.Type(); {
		case typ == token.RBrace:
			if p.openblocks > 0 {
				if p.insidePipe {
					p.insidePipe = false
				}

				return n, nil
			}

			break cmdLoop
		case isValidArgument(it):
			arg, err := p.getArgument(nil, exprConfig{
				allowConcat:   true,
				allowArg:      true,
				allowVariadic: true,
				allowFuncall:  false,
			})

			if err != nil {
				return nil, err
			}

			n.AddArg(arg)
		case typ == token.Plus:
			return nil, newParserError(it, p.name,
				"Unexpected '+'")
		case typ == token.Gt:
			p.next()
			redir, err := p.parseRedirection(it)

			if err != nil {
				return nil, err
			}

			n.AddRedirect(redir)
		case typ == token.Pipe:
			if p.insidePipe {
				p.next()
				// TODO(i4k): test against pipes and multiline cmds
				return n, nil
			}

			p.insidePipe = true
			return p.parsePipe(n)
		case typ == token.EOF:
			break cmdLoop
		case typ == token.Illegal:
			return nil, errors.NewError(it.Value())
		default:
			break cmdLoop
		}
	}

	it = p.peek()

	if isMulti {
		if it.Type() != token.RParen {
			if it.Type() == token.EOF {
				return nil, errors.NewUnfinishedCmdError(p.name, it)
			}

			return nil, newParserError(it, p.name, "Unexpected symbol '%s'", it)
		}

		p.ignore()

		it = p.peek()
	}

	if p.insidePipe {
		p.insidePipe = false
		return n, nil
	}

	if it.Type() != token.Semicolon {
		return nil, newParserError(it, p.name, "Unexpected symbol '%s'", it)
	}

	p.ignore()

	return n, nil
}

func (p *Parser) parseRedirection(it scanner.Token) (*ast.RedirectNode, error) {
	var (
		lval, rval int = ast.RedirMapNoValue, ast.RedirMapNoValue
		err        error
	)

	redir := ast.NewRedirectNode(it.FileInfo)

	it = p.peek()

	if !isValidArgument(it) && it.Type() != token.LBrack {
		return nil, newParserError(it, p.name, "Unexpected token: %v", it)
	}

	// [
	if it.Type() == token.LBrack {
		p.next()
		it = p.peek()

		if it.Type() != token.Number {
			return nil, newParserError(it, p.name, "Expected lefthand side of redirection map, but found '%s'",
				it.Value())
		}

		lval, err = strconv.Atoi(it.Value())

		if err != nil {
			return nil, newParserError(it, p.name, "Redirection map expects integers. Found: %s",
				it.Value())
		}

		p.next()
		it = p.peek()

		if it.Type() != token.Assign && it.Type() != token.RBrack {
			return nil, newParserError(it, p.name, "Unexpected token %v. Expecting ASSIGN or ]",
				it)
		}

		// [xxx=
		if it.Type() == token.Assign {
			p.next()
			it = p.peek()

			if it.Type() != token.Number && it.Type() != token.RBrack {
				return nil, newParserError(it, p.name, "Unexpected token %v. Expecting REDIRMAPRSIDE or ]", it)
			}

			if it.Type() == token.Number {
				rval, err = strconv.Atoi(it.Value())

				if err != nil {
					return nil, newParserError(it, p.name, "Redirection map expects integers. Found: %s", it.Value())
				}

				p.next()
				it = p.peek()
			} else {
				rval = ast.RedirMapSupress
			}
		}

		if it.Type() != token.RBrack {
			return nil, newParserError(it, p.name, "Unexpected token %v. Expecting ]", it)
		}

		// [xxx=yyy]

		redir.SetMap(lval, rval)

		p.next()
		it = p.peek()
	}

	if !isValidArgument(it) {
		if rval != ast.RedirMapNoValue || lval != ast.RedirMapNoValue {
			return redir, nil
		}

		return nil, newParserError(it, p.name, "Unexpected token %v. Expecting STRING or ARG or VARIABLE", it)
	}

	arg, err := p.getArgument(nil, exprConfig{
		allowConcat:   true,
		allowArg:      true,
		allowVariadic: false,
		allowFuncall:  false,
	})
	if err != nil {
		return nil, err
	}

	redir.SetLocation(arg)

	return redir, nil
}

func (p *Parser) parseImport(importToken scanner.Token) (ast.Node, error) {
	it := p.next()

	if it.Type() != token.Arg && it.Type() != token.String && it.Type() != token.Ident {
		return nil, newParserError(it, p.name, "Unexpected token %v. Expecting ARG or STRING", it)
	}

	var arg *ast.StringExpr

	if it.Type() == token.String {
		arg = ast.NewStringExpr(it.FileInfo, it.Value(), true)
	} else if it.Type() == token.Arg || it.Type() == token.Ident {
		arg = ast.NewStringExpr(it.FileInfo, it.Value(), false)
	} else {
		return nil, newParserError(it, p.name, "Parser error: Invalid token '%v' for import path", it)
	}

	if p.peek().Type() == token.Semicolon {
		p.ignore()
	}

	return ast.NewImportNode(importToken.FileInfo, arg), nil
}

func (p *Parser) parseSetenv(it scanner.Token) (ast.Node, error) {
	var (
		setenv   *ast.SetenvNode
		assign   ast.Node
		err      error
		fileInfo = it.FileInfo
	)

	it = p.next()
	next := p.peek()

	if it.Type() != token.Ident {
		return nil, newParserError(it, p.name, "Unexpected token %v, expected identifier", it)
	}

	if next.Type() == token.Assign || next.Type() == token.AssignCmd {
		assign, err = p.parseAssignment(it)

		if err != nil {
			return nil, err
		}

		setenv, err = ast.NewSetenvNode(fileInfo, it.Value(), assign)
	} else {
		setenv, err = ast.NewSetenvNode(fileInfo, it.Value(), nil)

		if p.peek().Type() != token.Semicolon {
			return nil, newParserError(p.peek(),
				p.name,
				"Unexpected token %v, expected semicolon (;) or EOL",
				p.peek())
		}

		p.ignore()
	}

	if err != nil {
		return nil, err
	}

	return setenv, nil
}

func (p *Parser) getArgument(tok *scanner.Token, cfg exprConfig) (ast.Expr, error) {
	var (
		err       error
		it        scanner.Token
		isFuncall bool
	)

	if tok != nil {
		it = *tok
	} else {
		it = p.next()
	}
	if !isValidArgument(it) {
		return nil, newParserError(it, p.name, "Unexpected token %v. Expected %s, %s, %s or %s",
			it, token.Ident, token.String, token.Variable, token.Arg)
	}

	firstToken := it
	var arg ast.Expr

	if firstToken.Type() == token.Variable {
		next := p.peek()

		if cfg.allowFuncall && next.Type() == token.LParen {
			arg, err = p.parseFnInv(firstToken, false)
			isFuncall = true
		} else {
			// makes "echo $list" == "echo $list..."
			arg, err = p.parseVariable(&firstToken, cfg.allowVariadic)
		}
	} else if firstToken.Type() == token.String {
		arg = ast.NewStringExpr(firstToken.FileInfo, firstToken.Value(), true)
	} else {
		// Arg, Ident, Number, Dotdotdot, etc

		next := p.peek()

		if cfg.allowFuncall && next.Type() == token.LParen {
			arg, err = p.parseFnInv(firstToken, false)
			isFuncall = true
		} else {
			arg = ast.NewStringExpr(firstToken.FileInfo, firstToken.Value(), false)
		}
	}

	if err != nil {
		return nil, err
	}

	it = p.peek()
	if it.Type() == token.Plus && cfg.allowConcat {
		return p.getConcatArg(arg)
	}

	if (firstToken.Type() == token.Arg || firstToken.Type() == token.Ident) && (!cfg.allowArg && !isFuncall) {
		return nil, newParserError(it, p.name, "Unquoted string not allowed at pos %d (%s)", it.FileInfo, it.Value())
	}

	return arg, nil
}

func (p *Parser) getConcatArg(firstArg ast.Expr) (ast.Expr, error) {
	var (
		it    scanner.Token
		parts []ast.Expr
	)

	parts = append(parts, firstArg)

hasConcat:
	it = p.peek()

	if it.Type() == token.Plus {
		p.ignore()

		arg, err := p.getArgument(nil, exprConfig{
			allowArg:      false,
			allowConcat:   false,
			allowFuncall:  true,
			allowVariadic: false,
		})
		if err != nil {
			return nil, err
		}

		parts = append(parts, arg)
		goto hasConcat
	}

	return ast.NewConcatExpr(token.NewFileInfo(firstArg.Line(), firstArg.Column()), parts), nil
}

func (p *Parser) parseAssignment(ident scanner.Token) (ast.Node, error) {
	// we're here
	// |
	// V
	// ident = ...
	// ident <= ...
	// ident, ident2, ..., identN = ...
	// ident, ident2, ..., identN <= ...
	it := p.next()

	if !isAssignment(it.Type()) {
		return nil, newParserError(it, p.name,
			"Unexpected token %v, expected '=' ,'<=', ',' or '['", it)
	}

	var (
		index ast.Expr
		err   error
	)

	if it.Type() == token.LBrack {
		index, err = p.parseIndexing()

		if err != nil {
			return nil, err
		}

		it = p.next()
	}

	names := []*ast.NameNode{
		ast.NewNameNode(ident.FileInfo, ident.Value(), index),
	}

	if it.Type() != token.Comma {
		goto assignOp
	}

	for it = p.next(); it.Type() == token.Ident; it = p.next() {
		var index ast.Expr

		name := it
		it = p.next()

		if it.Type() == token.LBrack {
			index, err = p.parseIndexing()

			if err != nil {
				return nil, err
			}

			it = p.next()
		}

		names = append(names, ast.NewNameNode(name.FileInfo, name.Value(), index))

		if it.Type() != token.Comma {
			break
		}
	}

assignOp:
	if it.Type() != token.AssignCmd && it.Type() != token.Assign {
		return nil, newParserError(it, p.name, "Unexpected token %v, expected ',' '=' or '<='", it)
	}

	if it.Type() == token.AssignCmd {
		return p.parseAssignCmdOut(names)
	}

	return p.parseAssignValues(names)
}

func (p *Parser) parseList(tok *scanner.Token) (ast.Node, error) {
	var (
		arg ast.Expr
		err error
		lit scanner.Token
	)

	if tok != nil {
		lit = *tok
	} else {
		lit = p.next()
	}

	if lit.Type() != token.LParen {
		return nil, newParserError(lit, p.name, "Unexpected token %v. Expecting (", lit)
	}

	var values []ast.Expr

	it := p.peek()

	for isValidArgument(it) || it.Type() == token.LParen {
		if it.Type() == token.LParen {
			arg, err = p.parseList(nil)
		} else {
			arg, err = p.getArgument(nil, exprConfig{
				allowArg:      true,
				allowConcat:   true,
				allowFuncall:  false,
				allowVariadic: false,
			})
		}

		if err != nil {
			return nil, err
		}

		it = p.peek()

		values = append(values, arg)
	}

	if it.Type() != token.RParen {
		if it.Type() == token.EOF {
			return nil, errors.NewUnfinishedListError(p.name, it)
		}

		return nil, newParserError(it, p.name, "Expected ) but found %s", it)
	}

	p.ignore()

	var isVariadic bool
	if p.peek().Type() == token.Dotdotdot {
		isVariadic = true
		p.ignore()
	}
	return ast.NewListVariadicExpr(lit.FileInfo, values, isVariadic), nil
}

func (p *Parser) parseAssignValues(names []*ast.NameNode) (ast.Node, error) {
	var values []ast.Expr

	if len(names) == 0 {
		return nil, newParserError(p.peek(), p.name, "parser error: expect names non nil")
	}

	for it := p.peek(); isExpr(it.Type()); it = p.peek() {
		var (
			value ast.Expr
			err   error
		)

		if it.Type() == token.Variable || it.Type() == token.String {
			value, err = p.getArgument(nil, exprConfig{
				allowArg:      false,
				allowFuncall:  true,
				allowVariadic: false,
				allowConcat:   true,
			})
		} else if it.Type() == token.LParen { // list
			value, err = p.parseList(nil)
		} else {
			return nil, newParserError(it, p.name, "Unexpected token %v. Expecting VARIABLE or STRING or (", it)
		}

		if err != nil {
			return nil, err
		}

		values = append(values, value)

		if p.peek().Type() != token.Comma {
			break
		}

		p.ignore()
	}

	if len(values) == 0 {
		return nil, newParserError(p.peek(), p.name, "Unexpected token %v. Expecting VARIABLE, STRING or (", p.peek())
	} else if len(values) != len(names) {
		return nil, newParserError(p.peek(), p.name, "assignment count mismatch: %d = %d",
			len(names), len(values))
	}

	if p.peek().Type() == token.Semicolon {
		p.ignore()
	}

	return ast.NewAssignNode(names[0].FileInfo, names, values), nil
}

func (p *Parser) parseAssignCmdOut(identifiers []*ast.NameNode) (ast.Node, error) {
	var (
		exec ast.Node
		err  error
	)

	it := p.next()

	if it.Type() != token.Ident && it.Type() != token.Arg && it.Type() != token.Variable && it.Type() != token.LParen {
		return nil, newParserError(it, p.name,
			"Invalid token %v. Expected command or function invocation", it)
	}

	if it.Type() == token.LParen {
		// command invocation
		exec, err = p.parseCommand(it)
	} else {
		nextIt := p.peek()

		if nextIt.Type() != token.LParen {
			// it == (Ident || Arg)
			exec, err = p.parseCommand(it)
		} else {
			// <ident>()
			// <arg>()
			// <var>()
			exec, err = p.parseFnInv(it, true)
		}
	}

	if err != nil {
		return nil, err
	}

	if len(identifiers) == 0 {
		// should not happen... pray
		panic("internal error parsing assignment")
	}

	return ast.NewExecAssignNode(identifiers[0].FileInfo, identifiers, exec)
}

func (p *Parser) parseRfork(it scanner.Token) (ast.Node, error) {
	n := ast.NewRforkNode(it.FileInfo)

	it = p.next()

	if it.Type() != token.Ident {
		return nil, newParserError(it, p.name, "rfork requires one or more of the following flags: %s", ast.RforkFlags)
	}

	arg := ast.NewStringExpr(it.FileInfo, it.Value(), false)
	n.SetFlags(arg)

	it = p.peek()

	if it.Type() == token.LBrace {
		blockPos := it.FileInfo

		p.ignore() // ignore lookaheaded symbol
		p.openblocks++

		tree := ast.NewTree("rfork block")
		r, err := p.parseBlock(blockPos.Line(), blockPos.Column())

		if err != nil {
			return nil, err
		}

		tree.Root = r

		n.SetTree(tree)
	}

	if p.peek().Type() == token.Semicolon {
		p.ignore()
	}

	return n, nil
}

func (p *Parser) parseIfExpr() (ast.Node, error) {
	it := p.peek()
	if it.Type() != token.Ident && it.Type() != token.String &&
		it.Type() != token.Variable {
		return nil, newParserError(it, p.name, "if requires lhs/rhs of type string, variable or function invocation. Found %v", it)
	}

	return p.getArgument(nil, exprConfig{
		allowArg:      false,
		allowVariadic: false,
		allowFuncall:  true,
		allowConcat:   true,
	})
}

func (p *Parser) parseIf(it scanner.Token) (ast.Node, error) {
	n := ast.NewIfNode(it.FileInfo)

	lvalue, err := p.parseIfExpr()
	if err != nil {
		return nil, err
	}

	n.SetLvalue(lvalue)

	it = p.next()

	if it.Type() != token.Equal && it.Type() != token.NotEqual {
		return nil, newParserError(it, p.name, "Expected comparison, but found %v", it)
	}

	if it.Value() != "==" && it.Value() != "!=" {
		return nil, newParserError(it, p.name, "Invalid if operator '%s'. Valid comparison operators are '==' and '!='",
			it.Value())
	}

	n.SetOp(it.Value())

	rvalue, err := p.parseIfExpr()

	if err != nil {
		return nil, err
	}

	n.SetRvalue(rvalue)

	it = p.next()

	if it.Type() != token.LBrace {
		return nil, newParserError(it, p.name, "Expected '{' but found %v", it)
	}

	p.openblocks++

	r, err := p.parseBlock(it.Line(), it.Column())

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

		n.SetElseif(elseIf)
		n.SetElseTree(elseTree)
	}

	return n, nil
}

func (p *Parser) parseFnArgs() ([]*ast.FnArgNode, error) {
	var args []*ast.FnArgNode

	if p.peek().Type() == token.RParen {
		// no argument
		p.ignore()
		return args, nil
	}

	for {
		it := p.next()
		if it.Type() == token.Ident {
			argName := it.Value()
			isVariadic := false
			if p.peek().Type() == token.Dotdotdot {
				isVariadic = true
				p.ignore()
			}
			args = append(args, ast.NewFnArgNode(it.FileInfo,
				argName, isVariadic))
		} else {
			return nil, newParserError(it, p.name, "Unexpected token %v. Expected identifier or ')'", it)
		}

		it = p.peek()
		if it.Type() == token.Comma {
			p.ignore()
			it = p.peek()

			if it.Type() == token.RParen {
				break
			}
			continue
		}

		if it.Type() != token.RParen {
			return nil, newParserError(it, p.name, "Unexpected '%v'. Expected ')'", it)
		}

		p.ignore()
		break
	}

	return args, nil
}

func (p *Parser) parseVar(it scanner.Token) (ast.Node, error) {
	var varTok = it

	it = p.next()
	next := p.peek()

	if it.Type() != token.Ident {
		return nil, newParserError(it, p.name,
			"Unexpected token %v. Expected IDENT",
			next,
		)
	}

	if !isAssignment(next.Type()) {
		return nil, newParserError(next, p.name,
			"Unexpected token %v. Expected '=' or ','",
			next,
		)
	}

	assign, err := p.parseAssignment(it)
	if err != nil {
		return nil, err
	}

	switch assign.Type() {
	case ast.NodeAssign:
		return ast.NewVarAssignDecl(
			varTok.FileInfo,
			assign.(*ast.AssignNode),
		), nil
	case ast.NodeExecAssign:
		return ast.NewVarExecAssignDecl(
			varTok.FileInfo,
			assign.(*ast.ExecAssignNode),
		), nil
	}

	return nil, newParserError(next, p.name,
		"Unexpected token %v. Expected ASSIGN or EXECASSIGN",
		next,
	)
}

func (p *Parser) parseFnDecl(it scanner.Token) (ast.Node, error) {
	var n *ast.FnDeclNode

	it = p.next()
	if it.Type() == token.Ident {
		n = ast.NewFnDeclNode(it.FileInfo, it.Value())
		it = p.next()
	} else {
		n = ast.NewFnDeclNode(it.FileInfo, "")
	}

	if it.Type() != token.LParen {
		return nil, newParserError(it, p.name,
			"Unexpected token %v. Expected '('", it)
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
		return nil, newParserError(it, p.name,
			"Unexpected token %v. Expected '{'", it)
	}

	p.openblocks++

	tree := ast.NewTree(fmt.Sprintf("fn %s body", n.Name()))
	r, err := p.parseBlock(it.Line(), it.Column())

	if err != nil {
		return nil, err
	}

	tree.Root = r
	n.SetTree(tree)
	return n, nil
}

func (p *Parser) parseFnInv(ident scanner.Token, allowSemicolon bool) (ast.Node, error) {
	n := ast.NewFnInvNode(ident.FileInfo, ident.Value())

	it := p.next()
	if it.Type() != token.LParen {
		return nil, newParserError(it, p.name, "Invalid token %v. Expected '('", it)
	}

	for {
		it = p.next()
		next := p.peek()
		if isFuncall(it.Type(), next.Type()) ||
			isValidArgument(it) {
			arg, err := p.getArgument(&it, exprConfig{
				allowArg:      false,
				allowFuncall:  true,
				allowConcat:   true,
				allowVariadic: true,
			})
			if err != nil {
				return nil, err
			}

			n.AddArg(arg)
		} else if it.Type() == token.LParen {
			listArg, err := p.parseList(&it)
			if err != nil {
				return nil, err
			}
			n.AddArg(listArg)
		} else if it.Type() == token.RParen {
			//			p.next()
			break
		} else if it.Type() == token.EOF {
			goto parseError
		}

		it = p.peek()
		if it.Type() == token.Comma {
			p.ignore()

			continue
		}

		if it.Type() == token.RParen {
			p.next()
			break
		}

		goto parseError
	}

	// semicolon is optional here
	if allowSemicolon && p.peek().Type() == token.Semicolon {
		p.next()
	}

	return n, nil

parseError:
	return nil, newParserError(it, p.name,
		"Unexpected token %v. Expecting STRING, VARIABLE or )", it)
}

func (p *Parser) parseElse() (*ast.BlockNode, bool, error) {
	it := p.next()

	if it.Type() == token.LBrace {
		p.openblocks++

		elseBlock, err := p.parseBlock(it.Line(), it.Column())

		if err != nil {
			return nil, false, err
		}

		return elseBlock, false, nil
	}

	if it.Type() == token.If {
		ifNode, err := p.parseIf(it)

		if err != nil {
			return nil, false, err
		}

		block := ast.NewBlockNode(it.FileInfo)
		block.Push(ifNode)

		return block, true, nil
	}

	return nil, false, newParserError(it, p.name, "Unexpected token: %v", it)
}

func (p *Parser) parseBindFn(bindIt scanner.Token) (ast.Node, error) {
	nameIt := p.next()

	if nameIt.Type() != token.Ident {
		return nil, newParserError(nameIt, p.name,
			"Expected identifier, but found '%v'", nameIt)
	}

	cmdIt := p.next()

	if cmdIt.Type() != token.Ident {
		return nil, newParserError(cmdIt, p.name, "Expected identifier, but found '%v'", cmdIt)
	}

	if p.peek().Type() == token.Semicolon {
		p.ignore()
	}

	n := ast.NewBindFnNode(bindIt.FileInfo, nameIt.Value(), cmdIt.Value())
	return n, nil
}

func (p *Parser) parseReturn(retTok scanner.Token) (ast.Node, error) {
	ret := ast.NewReturnNode(retTok.FileInfo)

	tok := p.peek()

	// return;
	// return }
	// return $v
	// return "<some>"
	// return ( ... values ... )
	// return <fn name>()
	// return "val1", "val2", $val3, test()
	if tok.Type() != token.Semicolon &&
		tok.Type() != token.RBrace &&
		tok.Type() != token.Variable &&
		tok.Type() != token.String &&
		tok.Type() != token.LParen &&
		tok.Type() != token.Ident {
		return nil, newParserError(tok, p.name,
			"Expected ';', STRING, VARIABLE, FUNCALL or LPAREN, but found %v",
			tok)
	}

	var returnExprs []ast.Expr

	for {
		tok = p.peek()
		if tok.Type() == token.Semicolon {
			p.ignore()
			break
		}

		if tok.Type() == token.RBrace {
			break
		}

		if tok.Type() == token.LParen {
			listArg, err := p.parseList(nil)
			if err != nil {
				return nil, err
			}
			returnExprs = append(returnExprs, listArg)
		} else if tok.Type() == token.Ident {
			p.next()
			next := p.peek()

			if next.Type() != token.LParen {
				return nil, newParserError(tok, p.name,
					"Expected FUNCALL, STRING, VARIABLE or LPAREN, but found '%v' %v",
					tok.Value(), next)
			}

			arg, err := p.parseFnInv(tok, true)
			if err != nil {
				return nil, err
			}

			returnExprs = append(returnExprs, arg)
		} else {
			arg, err := p.getArgument(nil, exprConfig{
				allowArg:      false,
				allowConcat:   true,
				allowFuncall:  true,
				allowVariadic: false,
			})
			if err != nil {
				return nil, err
			}

			returnExprs = append(returnExprs, arg)
		}

		next := p.peek()

		if next.Type() == token.Comma {
			p.ignore()
			continue
		}

		if next.Type() == token.Semicolon {
			p.ignore()
		}

		break
	}

	ret.Returns = returnExprs

	return ret, nil
}

func (p *Parser) parseFor(it scanner.Token) (ast.Node, error) {
	var (
		inExpr ast.Expr
		err    error
		next   scanner.Token
	)

	forStmt := ast.NewForNode(it.FileInfo)

	it = p.peek()

	if it.Type() != token.Ident {
		goto forBlockParse
	}

	p.next()

	forStmt.SetIdentifier(it.Value())

	it = p.next()

	if it.Type() != token.Ident || it.Value() != "in" {
		return nil, newParserError(it, p.name,
			"Expected 'in' but found %q", it)
	}

	// ignores 'in' keyword
	// TODO: make 'in' a real keyword

	it = p.next()
	next = p.peek()

	if it.Type() != token.Variable &&
		(it.Type() != token.Ident || (it.Type() == token.Ident && next.Type() != token.LParen)) &&
		it.Type() != token.LParen {
		return nil, newParserError(it, p.name,
			"Expected (variable, list or fn invocation) but found %q", it)
	}

	if (it.Type() == token.Ident || it.Type() == token.Variable) && next.Type() == token.LParen {
		inExpr, err = p.parseFnInv(it, false)
	} else if it.Type() == token.Variable {
		inExpr, err = p.parseVariable(&it, false)
	} else if it.Type() == token.LParen {
		inExpr, err = p.parseList(&it)
	}

	if err != nil {
		return nil, err
	}

	forStmt.SetInExpr(inExpr)
forBlockParse:
	it = p.peek()

	if it.Type() != token.LBrace {
		return nil, newParserError(it, p.name,
			"Expected '{' but found %q", it)
	}

	blockPos := it.FileInfo

	p.ignore() // ignore lookaheaded symbol
	p.openblocks++

	tree := ast.NewTree("for block")

	r, err := p.parseBlock(blockPos.Line(), blockPos.Column())

	if err != nil {
		return nil, err
	}

	tree.Root = r
	forStmt.SetTree(tree)

	return forStmt, nil
}

func (p *Parser) parseComment(it scanner.Token) (ast.Node, error) {
	return ast.NewCommentNode(it.FileInfo, it.Value()), nil
}

func (p *Parser) parseError(it scanner.Token) (ast.Node, error) {
	return nil, errors.NewError(it.Value())
}

func newParserError(item scanner.Token, name, format string, args ...interface{}) error {
	if item.Type() == token.Illegal {
		// scanner error
		return errors.NewError(item.Value())
	}

	errstr := fmt.Sprintf(format, args...)

	return errors.NewError("%s:%d:%d: %s", name, item.Line(), item.Column(), errstr)
}

func isValidArgument(t scanner.Token) bool {
	if t.Type() == token.String ||
		t.Type() == token.Number ||
		t.Type() == token.Arg ||
		t.Type() == token.Dotdotdot ||
		t.Type() == token.Ident ||
		token.IsKeyword(t.Type()) ||
		t.Type() == token.Variable {
		return true
	}

	return false
}

func isFuncall(tok, next token.Token) bool {
	return (tok == token.Ident || tok == token.Variable) &&
		next == token.LParen
}

func isAssignment(tok token.Token) bool {
	return tok == token.Assign ||
		tok == token.AssignCmd ||
		tok == token.LBrack ||
		tok == token.Comma
}

func isExpr(tok token.Token) bool {
	return tok == token.Variable ||
		tok == token.String ||
		tok == token.LParen
}
