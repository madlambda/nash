// Package scanner is the lexical parser.
package scanner

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/NeowayLabs/nash/token"
)

type (
	Token struct {
		typ token.Token
		pos token.Pos // start position of this token
		val string
	}

	stateFn func(*Lexer) stateFn

	// Lexer holds the state of the scanner
	Lexer struct {
		name   string     // used only for error reports
		input  string     // the string being scanned
		start  int        // start position of this token
		pos    int        // current position in the input
		width  int        // width of last rune read
		Tokens chan Token // channel of scanned tokens

		lastNode Token
	}
)

const (
	eof        = -1
	spaceChars = " \t\r\n"

	RforkFlags = "cnsmifup"
)

func (i Token) Type() token.Token { return i.typ }
func (i Token) Value() string     { return i.val }
func (i Token) Pos() token.Pos    { return i.pos }

func (i Token) String() string {
	switch i.typ {
	case token.Illegal:
		return "ERROR: " + i.val
	case token.EOF:
		return "EOF"
	}

	if len(i.val) > 10 {
		return fmt.Sprintf("(%v) - pos: %d, val: %.10q...", i.typ, i.pos, i.val)
	}

	return fmt.Sprintf("(%v) - pos: %d, val: %q", i.typ, i.pos, i.val)
}

// run lexes the input by executing state functions until the state is nil
func (l *Lexer) run() {
	for state := lexStart; state != nil; {
		state = state(l)
	}

	l.emit(token.EOF)
	close(l.Tokens) // No more tokens will be delivered
}

func (l *Lexer) emitVal(t token.Token, val string) {
	l.Tokens <- Token{
		typ: t,
		val: val,
		pos: token.Pos(l.start),
	}

	l.start = l.pos
}

func (l *Lexer) emit(t token.Token) {
	l.Tokens <- Token{
		typ: t,
		val: l.input[l.start:l.pos],
		pos: token.Pos(l.start),
	}

	l.start = l.pos
}

func (l *Lexer) next() rune {
	var r rune

	if l.pos >= len(l.input) {
		l.width = 0
		return eof
	}

	r, l.width = utf8.DecodeRuneInString(l.input[l.pos:])

	l.pos += l.width
	return r
}

// ignore skips over the pending input before this point
func (l *Lexer) ignore() {
	l.start = l.pos
}

// backup steps back one rune
func (l *Lexer) backup() {
	l.pos -= l.width
}

// peek returns but does not consume the next rune
func (l *Lexer) peek() rune {
	rune := l.next()
	l.backup()
	return rune
}

// accept consumes the next rune if it's from the valid set
func (l *Lexer) accept(valid string) bool {
	if strings.IndexRune(valid, l.next()) >= 0 {
		return true
	}

	l.backup()
	return false
}

// acceptRun consumes a run of runes from the valid setup
func (l *Lexer) acceptRun(valid string) {
	for strings.IndexRune(valid, l.next()) >= 0 {

	}

	l.backup()
}

// errorf returns an error token
func (l *Lexer) errorf(format string, args ...interface{}) stateFn {
	l.Tokens <- Token{
		typ: token.Illegal,
		val: fmt.Sprintf(format, args...),
		pos: token.Pos(l.start),
	}

	l.start = len(l.input)
	l.pos = l.start

	return nil // finish the state machine
}

func (l *Lexer) String() string {
	return fmt.Sprintf("Lexer:\n\tPos: %d\n\tStart: %d\n",
		l.pos, l.start)
}

func Lex(name, input string) *Lexer {
	l := &Lexer{
		name:   name,
		input:  input,
		Tokens: make(chan Token),
	}

	go l.run() // concurrently run state machine

	return l
}

func lexStart(l *Lexer) stateFn {
	r := l.next()

	switch {
	case r == eof:
		return nil

	case isSpace(r):
		return lexSpace

	case isEndOfLine(r):
		l.ignore()
		return lexStart

	case r == '#':
		return lexComment

	case isIdentifier(r) || r == '$':
		return lexIdentifier

	case isSafePath(r):
		return lexInsideCommandName

	case r == '{':
		l.ignore()
		return l.errorf("Unexpected open block \"%#U\"", r)

	case r == '}':
		l.emit(token.RBrace)
		return lexStart
	case r == '(':
		l.emit(token.LParen)
		return lexInsideFnInv
	case r == ')':
		l.emit(token.RParen)
		return lexStart
	}

	return l.errorf("Unrecognized character in action: %#U at pos %d", r, l.pos)
}

func absorbIdentifier(l *Lexer) {
	for {
		r := l.next()

		if isIdentifier(r) {
			continue // absorb
		}

		break
	}

	l.backup() // pos is now ahead of the alphanum
}

func lexIdentifier(l *Lexer) stateFn {
	absorbIdentifier(l)

	word := l.input[l.start:l.pos]

	r := l.peek()

	if r == '(' {
		l.emit(token.FnInv)
		l.next()
		l.emit(token.LParen)
		return lexInsideFnInv
	}

	if word[0] == '$' {
		return l.errorf("Unexpected '$' at pos %d. Variable can only start a statement if it's a function invocation",
			l.pos)
	}

	if len(word) > 0 && r == '-' {
		for r == '-' {
			l.next()

			absorbIdentifier(l)

			r = l.peek()
		}

		goto commandName
	}

	// name=val
	if isSpace(r) || r == '=' {
		// lookahead by hand, to avoid more complex Lexer API
		for i := l.pos; i < len(l.input); i++ {
			r, _ := utf8.DecodeRuneInString(l.input[i:])

			if !isSpace(r) {
				if r == '=' {
					l.emit(token.Ident)

					ignoreSpaces(l)
					l.next()
					l.emit(token.Assign)
					return lexInsideAssignment
				}

				break
			}
		}
	}

	// name <= cmd
	if isSpace(r) || r == '<' {
		// lookahead by hand, to avoid more complex Lexer API
		for i := l.pos; i < len(l.input); i++ {
			r, _ := utf8.DecodeRuneInString(l.input[i:])

			if !isSpace(r) {
				if r == '<' {
					r, _ := utf8.DecodeRuneInString(l.input[i+1:])

					if r != '=' {
						return l.errorf("Unexpected token '%v'. Expected '='", r)
					}

					l.emit(token.Ident)

					ignoreSpaces(l)

					l.next()
					l.next()

					l.emit(token.AssignCmd)

					ignoreSpaces(l)

					absorbIdentifier(l)

					word := l.input[l.start:l.pos]

					if len(word) == 0 {
						r = l.peek()

						if r != '-' {
							return l.errorf("Expected identifier")
						}

						l.next()
						absorbIdentifier(l)

						word = l.input[l.start:l.pos]
					}

					r = l.peek()

					if r == '(' {
						l.emit(token.FnInv)
						l.next()
						l.emit(token.LParen)
						return lexInsideFnInv
					}

					l.emit(token.Command)
					return lexInsideCommand
				}

				break
			}
		}
	}

	switch word {
	case "builtin":
		l.emit(token.Builtin)
		return lexStart
	case "import":
		l.emit(token.Import)
		return lexInsideImport
	case "rfork":
		l.emit(token.Rfork)
		return lexInsideRforkArgs
	case "cd":
		l.emit(token.Cd)
		return lexInsideCd
	case "setenv":
		l.emit(token.SetEnv)
		return lexInsideSetenv
	case "if":
		l.emit(token.If)
		return lexInsideIf
	case "fn":
		l.emit(token.FnDecl)
		return lexInsideFnDecl
	case "else":
		l.emit(token.Else)
		return lexInsideElse
	case "bindfn":
		l.emit(token.BindFn)

		return lexInsideBindFn
	case "dump":
		l.emit(token.Dump)
		return lexInsideDump
	case "return":
		l.emit(token.Return)
		return lexInsideReturn
	case "for":
		l.emit(token.For)
		return lexInsideFor
	}

commandName:
	l.emit(token.Command)
	return lexInsideCommand
}

func lexInsideDump(l *Lexer) stateFn {
	ignoreSpaces(l)

	r := l.peek()

	if r == '"' {
		l.next()
		l.ignore()
		return func(l *Lexer) stateFn {
			return lexQuote(l, lexInsideDump, lexStart)
		}
	}

	if isIdentifier(r) || isSafePath(r) {
		l.next()
		// parse as normal argument
		return func(l *Lexer) stateFn {
			return lexArg(l, lexInsideDump, lexStart)
		}
	}

	if r == '$' {
		return lexInsideCommonVariable(l, lexInsideDump, lexStart)
	}

	return lexStart
}

func lexInsideImport(l *Lexer) stateFn {
	ignoreSpaces(l)

	r := l.next()

	if r == '"' {
		l.ignore()
		return func(l *Lexer) stateFn {
			return lexQuote(l, lexInsideImport, lexStart)
		}
	}

	if isIdentifier(r) || isSafePath(r) {
		// parse as normal argument
		return func(l *Lexer) stateFn {
			return lexArg(l, lexInsideImport, lexStart)
		}
	}

	l.backup()
	return lexStart
}

func lexInsideSetenv(l *Lexer) stateFn {
	ignoreSpaces(l)

	for {
		r := l.next()

		if isIdentifier(r) {
			continue // absorb
		}

		break
	}

	l.backup() // pos is now ahead of the alphanum

	word := l.input[l.start:l.pos]

	if len(word) == 0 {
		// sanity check
		return l.errorf("internal error")
	}

	l.emit(token.Ident)
	return lexStart
}

func lexInsideAssignment(l *Lexer) stateFn {
	ignoreSpaces(l)

	r := l.peek()

	switch {
	case r == '(':
		l.next()
		l.emit(token.LParen)

		return lexInsideListVariable
	case r == '"':
		l.next()
		l.ignore()

		return func(l *Lexer) stateFn {
			return lexQuote(l, lexInsideAssignment, func(l *Lexer) stateFn {
				ignoreSpaces(l)

				r := l.peek()

				if !isEndOfLine(r) && r != eof {
					return l.errorf("Invalid assignment. Expected '+' or EOL, but found '%c' at pos '%d'", r, l.pos)
				}

				return lexStart
			})
		}

	case r == '$':
		return lexInsideCommonVariable(l, lexInsideAssignment, lexStart)
	}

	return l.errorf("Unexpected variable value '%c'. Expected '\"' for quoted string or '$' for variable.", r)
}

func lexInsideListVariable(l *Lexer) stateFn {
	ignoreSpaces(l)

	r := l.peek()

	switch {
	case isEndOfLine(r):
		l.next()
		l.ignore()
		return lexInsideListVariable
	case isSafeArg(r):
		return lexArg(l, lexInsideListVariable, lexInsideListVariable)
	case r == '"':
		l.next()
		l.ignore()
		return lexQuote(l, lexInsideListVariable, lexInsideListVariable)
	case r == '$':
		return lexInsideCommonVariable(l, lexInsideListVariable, lexInsideListVariable)
	case r == ')':
		l.next()
		l.emit(token.RParen)
		return lexStart
	}

	return l.errorf("Unexpected '%q'. Expected elements or ')'", r)
}

func lexInsideCommonVariable(l *Lexer, nextConcatFn stateFn, nextFn stateFn) stateFn {
	var r rune

	r = l.next()

	if r != '$' {
		return l.errorf("Invalid variable. Unexpected '%c'", r)
	}

	for {
		r = l.next()

		if !isIdentifier(r) {
			break
		}
	}

	l.backup()

	if r == '"' {
		l.ignore()
		return l.errorf("Invalid quote inside variable name")
	}

	l.emit(token.Variable)

	r = l.peek()

	if r == '[' {
		l.next()
		l.emit(token.LBrack)

		r = l.peek()

		if r == '$' {
			for state := func(l *Lexer) stateFn {
				return lexInsideCommonVariable(l, nextConcatFn, nil)
			}; state != nil; {
				state = state(l)
			}
		} else {
			digits := "0123456789"

			if !l.accept(digits) {
				return l.errorf("Expected number or variable on variable indexing. Found %q", l.peek())
			}

			l.accept(digits)
			l.acceptRun(digits)
			l.emit(token.Number)
		}

		r = l.next()

		if r != ']' {
			return l.errorf("Unexpected %q. Expecting ']'", r)
		}

		l.emit(token.RBrack)
	}

	ignoreSpaces(l)

	r = l.peek()

	switch {
	case r == '+':
		l.next()
		l.emit(token.Concat)

		return nextConcatFn
	}

	return nextFn
}

func lexInsideCd(l *Lexer) stateFn {
	// parse the cd directory
	ignoreSpaces(l)

	r := l.next()

	if r == '"' {
		l.ignore()
		return func(l *Lexer) stateFn {
			return lexQuote(l, lexInsideCd, lexInsideCd)
		}
	}

	if r == '$' {
		for {
			r = l.next()

			if !isIdentifier(r) {
				break
			}
		}

		l.backup()

		if r == '"' {
			l.ignore()
			return l.errorf("Invalid quote inside variable name")
		}

		l.emit(token.Variable)

		ignoreSpaces(l)

		r = l.peek()

		switch {
		case r == '+':
			l.next()
			l.emit(token.Concat)
			return lexInsideCd
		}

		if !isEndOfLine(r) && r != eof {
			return l.errorf("Expected end of line, but found %c at pos %d", r, l.pos)
		}

		return lexStart
	}

	if isIdentifier(r) || isSafePath(r) {
		// parse as normal argument
		return func(l *Lexer) stateFn {
			return lexArg(l, lexInsideCd, lexStart)
		}
	}

	l.backup()
	return lexStart
}

func lexIfLRValue(l *Lexer) stateFn {
	ignoreSpaces(l)

	r := l.next()

	switch {
	case r == '"':
		l.ignore()
		for state := func(l *Lexer) stateFn {
			return lexQuote(l, lexIfLRValue, nil)
		}; state != nil; {
			state = state(l)
		}

		return nil
	case r == '$':
		l.backup()

		for state := func(l *Lexer) stateFn {
			return lexInsideCommonVariable(l, lexIfLRValue, nil)
		}; state != nil; {
			state = state(l)
		}

		return nil
	}

	return l.errorf("Unexpected char %q at pos %d. IfDecl expects string or variable", r, l.pos)
}

func lexInsideIf(l *Lexer) stateFn {
	errState := lexIfLRValue(l)

	if errState != nil {
		return errState
	}

	ignoreSpaces(l)

	// get first char of operator. Eg.: '!'
	if !l.accept("=!") {
		l.errorf("Unexpected char %q inside if statement", l.peek())
		return nil
	}

	// get second char. Eg.: '='
	if !l.accept("=!") {
		l.errorf("Unexpected char %q inside if statement", l.peek())
		return nil
	}

	word := l.input[l.start:l.pos]

	if word != "==" && word != "!=" {
		return l.errorf("Invalid comparison operator '%s'", word)
	}

	if word == "==" {
		l.emit(token.Equal)
	} else {
		l.emit(token.NotEqual)
	}

	errState = lexIfLRValue(l)

	if errState != nil {
		return errState
	}

	ignoreSpaces(l)

	r := l.next()

	if r != '{' {
		return l.errorf("Unexpected %q at pos %d. Expected '{'", r, l.pos)
	}

	l.emit(token.LBrace)

	return lexStart
}

func lexForEnd(l *Lexer) stateFn {
	ignoreSpaces(l)

	r := l.next()

	if r != '{' {
		return l.errorf("Unexpected %q. Expected '{'", r)
	}

	l.emit(token.LBrace)
	return lexStart
}

func lexInsideForIn(l *Lexer) stateFn {
	ignoreSpaces(l)

	r := l.next()

	switch {
	case r == '$':
		l.backup()
		return lexInsideCommonVariable(l, lexInsideForIn, lexForEnd)
	}

	return l.errorf("Unexpected %q on for in clause", r)
}

func lexInsideFor(l *Lexer) stateFn {
	ignoreSpaces(l)

	for {
		r := l.peek()

		if !isIdentifier(r) {
			break
		}

		l.next()
	}

	word := l.input[l.start:l.pos]

	if len(word) > 0 {
		l.emit(token.Ident)

		ignoreSpaces(l)

		ri := l.next()
		rn := l.next()

		if ri != 'i' && rn != 'n' {
			return l.errorf("Unexpected %q. Expected 'in'", ri)
		}

		l.emit(token.ForIn)
		return lexInsideForIn
	}

	return lexForEnd
}

func lexInsideFnDecl(l *Lexer) stateFn {
	var (
		r       rune
		argName string
	)

	ignoreSpaces(l)

	for {
		r = l.next()

		if isIdentifier(r) {
			continue
		}

		break
	}

	l.backup()

	l.emit(token.Ident)

	r = l.next()

	if r != '(' {
		return l.errorf("Unexpected symbol %q. Expected '('", r)
	}

	l.emit(token.LParen)

getnextarg:
	ignoreSpaces(l)

	for {
		r = l.next()

		if isIdentifier(r) {
			continue
		}

		break
	}

	l.backup()

	argName = l.input[l.start:l.pos]

	if len(argName) > 0 {
		l.emit(token.Ident)

		r = l.peek()

		if r == ',' {
			l.next()
			goto getnextarg
		} else if r != ')' {
			return l.errorf("Unexpected symbol %q. Expected ',' or ')'", r)
		}
	}

	l.next()
	l.emit(token.RParen)

	ignoreSpaces(l)

	r = l.next()

	if r != '{' {
		return l.errorf("Unexpected symbol %q. Expected '{'", r)
	}

	l.emit(token.LBrace)

	return lexStart
}

func lexInsideReturn(l *Lexer) stateFn {
	ignoreSpaces(l)

	r := l.peek()

	switch {
	case r == '(':
		l.next()
		l.emit(token.LParen)
		return lexInsideListVariable
	case r == '"':
		l.next()
		l.ignore()

		return lexQuote(l, lexInsideReturn, lexStart)
	case r == '$':
		return lexInsideCommonVariable(l, lexInsideReturn, lexStart)
	}

	return lexStart
}

func lexMoreFnArgs(l *Lexer) stateFn {
	ignoreSpaces(l)

	r := l.peek()

	if r == ',' {
		l.next()
		l.ignore()
		return lexInsideFnInv
	}

	if r == ')' {
		return lexStart
	}

	return l.errorf("Unexpected symbol %q. Expecting ',' or ')'", r)
}

func lexInsideFnInv(l *Lexer) stateFn {
	ignoreSpaces(l)

	var r rune

	r = l.peek()

	if r == '"' {
		l.next()
		l.ignore()
		return lexQuote(l, lexInsideFnInv, lexMoreFnArgs)
	} else if r == '$' {
		return lexInsideCommonVariable(l, lexInsideFnInv, lexMoreFnArgs)

	} else if r == ')' {
		l.next()
		l.emit(token.RParen)
		return lexStart
	}

	return l.errorf("Unexpected symbol %q. Expected quoted string or variable", r)
}

func lexInsideElse(l *Lexer) stateFn {
	ignoreSpaces(l)

	r := l.next()

	if r == '{' {
		l.emit(token.LBrace)
		return lexStart
	}

	for {
		r = l.next()

		if !isIdentifier(r) {
			break
		}
	}

	l.backup()

	word := l.input[l.start:l.pos]

	if word == "if" {
		l.emit(token.If)
		return lexInsideIf
	}

	return l.errorf("Unexpected word '%s' at pos %d", word, l.pos)
}

// Rfork flags:
// c = stands for container (c == upnsmi)
// u = user namespace
// p = pid namespace
// n = network namespace
// s = uts namespace
// m = mount namespace
// i = ipc namespace
func lexInsideRforkArgs(l *Lexer) stateFn {
	// parse the rfork parameters

	ignoreSpaces(l)

	if !l.accept(RforkFlags) {
		return l.errorf("invalid rfork argument: %s", string(l.peek()))
	}

	l.acceptRun(RforkFlags)

	l.emit(token.String)

	ignoreSpaces(l)

	if l.accept("{") {
		l.emit(token.LBrace)
	}

	return lexStart
}

func lexInsideCommandName(l *Lexer) stateFn {
	for {
		r := l.next()

		if isSafePath(r) {
			continue // absorb
		}

		break
	}

	l.backup() // pos is now ahead of the alphanum

	word := l.input[l.start:l.pos]

	if len(word) == 0 {
		// sanity check
		return l.errorf("internal error")
	}

	if len(word) == 1 && word == "-" {
		l.ignore()
		return l.errorf("- requires a command")
	}

	l.emit(token.Command)
	return lexInsideCommand
}

func lexInsideBindFn(l *Lexer) stateFn {
	var r rune

	ignoreSpaces(l)

	for {
		r = l.next()

		if isIdentifier(r) {
			continue
		}

		break
	}

	l.backup()

	word := l.input[l.start:l.pos]

	if len(word) == 0 {
		return l.errorf("Unexpected %q, expected identifier.", r)
	}

	l.emit(token.Ident)

	ignoreSpaces(l)

	for {
		r = l.next()

		if isIdentifier(r) || r == '-' {
			continue
		}

		break
	}

	l.backup()

	word = l.input[l.start:l.pos]

	if len(word) == 0 {
		return l.errorf("Unexpected %q, expected identifier.", r)
	}

	l.emit(token.Ident)

	return lexStart
}

func lexInsideCommand(l *Lexer) stateFn {
	ignoreSpaces(l)

	r := l.next()

	switch {
	case r == eof:
		return nil
	case isEndOfLine(r):
		l.ignore()
		return lexStart
	case r == '#':
		return lexComment
	case r == '"':
		l.ignore()
		return func(l *Lexer) stateFn {
			return lexQuote(l, lexInsideCommand, lexInsideCommand)
		}
	case r == '}':
		l.emit(token.RBrace)
		return lexStart

	case r == '>':
		l.emit(token.RedirRight)
		return lexInsideRedirect
	case r == '|':
		l.emit(token.Pipe)
		return lexStart
	case r == '$':
		l.backup()
		return lexInsideCommonVariable(l, lexInsideCommand, lexInsideCommand)
	case isSafeArg(r):
		break
	default:
		return l.errorf("Invalid char %q at pos %d. String: %.10s", r, l.pos, l.input[l.pos:])
	}

	return func(l *Lexer) stateFn {
		return lexArg(l, lexInsideCommand, lexInsideCommand)
	}
}

func lexQuote(l *Lexer, concatFn, nextFn stateFn) stateFn {
	var data []rune

	data = make([]rune, 0, 256)

	for {
		r := l.next()

		if r != '"' && r != eof {
			if r == '\\' {
				r = l.next()

				switch r {
				case 'n':
					data = append(data, '\n')
				case 't':
					data = append(data, '\t')
				case '\\':
					data = append(data, '\\')
				case '"':
					data = append(data, '"')
				case 'x', 'u', 'U':
					return l.errorf("Escape types 'x', 'u' and 'U' aren't implemented yet")
				case '0', '1', '2', '3', '4', '5', '6', '7':
					x := r - '0'

					for i := 2; i > 0; i-- {
						r = l.next()

						if r >= '0' && r <= '7' {
							x = x*8 + r - '0'
							continue
						}

						return l.errorf("non-octal character in escape sequence: %c", r)
					}

					if x > 255 {
						return l.errorf("octal escape value > 255: %d", x)
					}

					data = append(data, x)
				}
			} else {
				data = append(data, r)
			}

			continue
		}

		if r == eof {
			return l.errorf("Quoted string not finished: %s", l.input[l.start:])
		}

		l.emitVal(token.String, string(data))

		l.ignore() // ignores last quote
		break
	}

	ignoreSpaces(l)

	r := l.peek()

	switch {
	case r == '+':
		if concatFn == nil {
			return l.errorf("Concatenation is not allowed at pos %d", l.pos)
		}

		l.next()
		l.emit(token.Concat)

		return concatFn
	}

	return nextFn
}

func lexArg(l *Lexer, concatFn, nextFn stateFn) stateFn {
	for {
		r := l.next()

		if r == eof {
			if l.pos > l.start {
				l.emit(token.Arg)
			}

			return nil
		}

		if isIdentifier(r) || isSafeArg(r) {
			continue
		}

		l.backup()
		l.emit(token.Arg)
		break
	}

	ignoreSpaces(l)

	r := l.peek()

	switch {
	case r == '+':
		l.next()
		l.emit(token.Concat)

		return concatFn
	}

	return nextFn
}

func lexInsideRedirect(l *Lexer) stateFn {
	ignoreSpaces(l)

	r := l.next()

	switch {
	case r == '[':
		l.emit(token.LBrack)
		return lexInsideRedirMapLeftSide
	case r == ']':
		return l.errorf("Unexpected ']' at pos %d", l.pos)
	case r == '"':
		l.ignore()

		for {
			r := l.next()

			if r != '"' && r != eof {
				continue
			}

			if r == eof {
				return l.errorf("Quoted string not finished: %s", l.input[l.start:])
			}

			l.backup()

			break
		}

		l.emit(token.String)

		l.next() // get last '"' again
		l.ignore()
	case r == '$':
		l.backup()
		return lexInsideCommonVariable(l, lexInsideRedirect, lexStart)
	case isSafePath(r):
		for {
			r = l.next()

			if !isSafePath(r) {
				l.backup()
				break
			}
		}

		l.emit(token.Arg)
	default:
		return l.errorf("Unexpected redirect identifier: %s", l.input[l.start:l.pos])
	}

	// verify if have more redirects

	ignoreSpaces(l)

	r = l.next()

	if r == '>' {
		l.emit(token.RedirRight)
		return lexInsideRedirect
	}

	if r == '|' {
		l.emit(token.Pipe)
		return lexStart
	}

	if !isEndOfLine(r) && r != eof {
		return l.errorf("Expected end of line or redirection, but found '%c'", r)
	}

	l.backup()
	return lexStart
}

func lexInsideRedirMapLeftSide(l *Lexer) stateFn {
	var r rune

	for {
		r = l.peek()

		if !unicode.IsDigit(r) {
			if len(l.input[l.start:l.pos]) == 0 {
				return l.errorf("Unexpected %c at pos %d", r, l.pos)
			}

			if r == ']' {
				// [xxx]
				l.emit(token.RedirMapLSide)
				l.next()
				l.emit(token.RBrack)

				ignoreSpaces(l)

				r = l.peek()

				if isSafePath(r) || r == '"' {
					return lexInsideRedirect
				}

				if r == '>' {
					l.next()
					l.emit(token.RedirRight)
					return lexInsideRedirect
				}

				if r == '|' {
					l.next()
					l.emit(token.Pipe)
				}

				return lexStart
			}

			if r != '=' {
				return l.errorf("Expected '=' but found '%c' at por %d", r, l.pos)
			}

			// [xxx=
			l.emit(token.RedirMapLSide)
			l.next()
			l.emit(token.Assign)

			return lexInsideRedirMapRightSide
		}

		r = l.next()
	}
}

func lexInsideRedirMapRightSide(l *Lexer) stateFn {
	var r rune

	// [xxx=yyy]
	for {
		r = l.peek()

		if !unicode.IsDigit(r) {
			if len(l.input[l.start:l.pos]) == 0 {
				if r == ']' {
					l.next()
					l.emit(token.RBrack)

					ignoreSpaces(l)

					r = l.peek()

					if isSafePath(r) || r == '"' {
						return lexInsideRedirect
					}

					if r == '>' {
						l.next()
						l.emit(token.RedirRight)
						return lexInsideRedirect
					}

					return lexStart
				}

				return l.errorf("Unexpected %c at pos %d", r, l.pos)
			}

			l.emit(token.RedirMapRSide)
			l.next()
			l.emit(token.RBrack)

			ignoreSpaces(l)

			r = l.peek()

			if isSafePath(r) || r == '"' {
				return lexInsideRedirect
			}

			if r == '>' {
				l.next()
				l.emit(token.RedirRight)
				return lexInsideRedirect
			}

			if r == '|' {
				l.next()
				l.emit(token.Pipe)
			}

			break
		}

		r = l.next()
	}

	return lexStart
}

func lexComment(l *Lexer) stateFn {
	for {
		r := l.next()

		if isEndOfLine(r) {
			l.backup()
			l.emit(token.Comment)

			break
		}

		if r == eof {
			l.backup()
			l.emit(token.Comment)
			break
		}
	}

	return lexStart
}

func lexSpaceArg(l *Lexer) stateFn {
	ignoreSpaces(l)
	return lexInsideCommand
}

func lexSpace(l *Lexer) stateFn {
	ignoreSpaces(l)
	return lexStart
}

// isSpace reports whether r is a space character.
func isSpace(r rune) bool {
	return r == ' ' || r == '\t'
}

func isSafePath(r rune) bool {
	isId := isIdentifier(r)
	return isId || r == '_' || r == '-' || r == '/' || r == '.'
}

func isSafeArg(r rune) bool {
	isPath := isSafePath(r)

	return isPath || r == '=' || r == ':'
}

// isIdentifier reports whether r is a valid identifier
func isIdentifier(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}

// isEndOfLine reports whether r is an end-of-line character.
func isEndOfLine(r rune) bool {
	return r == '\r' || r == '\n'
}

func ignoreSpaces(l *Lexer) {
	for {
		r := l.next()

		if !isSpace(r) {
			break
		}
	}

	l.backup()
	l.ignore()
}
