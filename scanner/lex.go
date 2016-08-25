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
		typ    token.Token
		pos    token.Pos // start position of this token
		line   int       // line of token
		column int       // column of token in the line
		val    string
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

		linenum    int
		column     int
		prevColumn int
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
func (i Token) Line() int         { return i.line }
func (i Token) Column() int       { return i.column }

func (i Token) String() string {
	switch i.typ {
	case token.Illegal:
		return "ERROR: " + i.val
	case token.EOF:
		return "EOF"
	}

	if len(i.typ.String()) > 10 {
		return fmt.Sprintf("%s...", i.typ.String()[0:10])
	}

	return fmt.Sprintf("%s", i.typ)
}

// run lexes the input by executing state functions until the state is nil
func (l *Lexer) run() {
	l.linenum = 1
	l.column = 0

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
		typ:    t,
		val:    l.input[l.start:l.pos],
		pos:    token.Pos(l.start),
		line:   l.linenum,
		column: l.column,
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
	l.prevColumn = l.column

	if r == '\n' {
		l.linenum++
		l.column = 0
	} else {
		l.column++
	}

	return r
}

// ignore skips over the pending input before this point
func (l *Lexer) ignore() {
	l.start = l.pos
}

// backup steps back one rune
func (l *Lexer) backup() {
	l.pos -= l.width

	r, _ := utf8.DecodeRuneInString(l.input[l.pos:])

	l.column = l.prevColumn

	if r == '\n' {
		l.linenum--
	}
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
	fname := l.name

	if fname == "" {
		fname = "<none>"
	}

	errMsg := fmt.Sprintf(format, args...)

	arguments := make([]interface{}, 0, len(args)+2)
	arguments = append(arguments, fname, l.linenum, l.column, errMsg)

	l.Tokens <- Token{
		typ:    token.Illegal,
		val:    fmt.Sprintf("%s:%d:%d: %s", arguments...),
		pos:    token.Pos(l.start),
		line:   l.linenum,
		column: l.column,
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

	case 'a' <= r && r <= 'z':
		// nash literals are lowecase
		lit := scanIdentifier(l)

		if len(lit) > 1 {
			l.emit(token.Lookup(lit))
		}

		l.emit(token.Ident)

		return lexStart
	case 'A' <= r && r <= 'Z':
		absorbIdentifier(l)
		l.emit(token.Ident)

		return lexStart
	case '0' <= r && r <= '9':
		return lexNumber
	case isSpace(r):
		return lexSpace

	case isEndOfLine(r):
		l.ignore()

		return lexStart
	case r == '"':
		l.ignore()

		return lexQuote
	case r == '#':
		return lexComment
	case r == '+':
		l.emit(token.Plus)
		return lexStart
	case r == '-':
		l.emit(token.Minus)
		return lexStart
	case r == '$':
		r = l.next()

		if !isIdentifier(r) {
			return l.errorf("Expected identifier, but found %q", r)
		}

		absorbIdentifier(l)
		l.emit(token.Variable)
		return lexStart
	case isSafePath(r):
		absorbPath(l)
		l.emit(token.Path)
		return lexStart
	case r == '{':
		l.emit(token.LBrace)
		return lexStart
	case r == '}':
		l.emit(token.RBrace)
		return lexStart
	case r == '(':
		l.emit(token.LParen)
		return lexStart
	case r == ')':
		l.emit(token.RParen)
		return lexStart
	case r == ',':
		l.emit(token.Comma)
		return lexStart
	}

	return l.errorf("Unrecognized character in action: %#U", r, l.pos)
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

func lexNumber(l *Lexer) stateFn {
	digits := "0123456789"

	if !l.accept(digits) {
		return l.errorf("Expected number or variable on variable indexing. Found %q", l.peek())
	}

	l.acceptRun(digits)

	l.emit(token.Number)
	return lexStart
}

func absorbPath(l *Lexer) {
	for {
		r := l.next()

		if isSafePath(r) {
			continue // absorb
		}

		break
	}

	l.backup()
}

func scanIdentifier(l *Lexer) string {
	absorbIdentifier(l)

	return l.input[l.start:l.pos]
}

func lexQuote(l *Lexer) stateFn {
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
