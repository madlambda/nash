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
		token.FileInfo

		val string
	}

	stateFn func(*Lexer) stateFn

	// Lexer holds the state of the scanner
	Lexer struct {
		name  string // identify the source, used only for error reports
		input string // the string being scanned
		start int    // start position of current token

		width  int        // width of last rune read
		Tokens chan Token // channel of scanned tokens

		// file positions
		pos         int // file offset
		line        int // current line position
		lineStart   int // line of the symbol's start
		prevColumn  int // previous column value
		column      int // current column position
		columnStart int // column of the symbol's start

		openParens int

		addSemicolon bool
	}
)

const (
	eof = -1
)

func (i Token) Type() token.Token { return i.typ }
func (i Token) Value() string     { return i.val }

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
	l.line, l.lineStart, l.column, l.columnStart = 1, 1, 0, 0

	for state := lexStart; state != nil; {
		state = state(l)
	}

	l.emit(token.EOF)
	close(l.Tokens) // No more tokens will be delivered
}

func (l *Lexer) emitVal(t token.Token, val string, line, column int) {
	l.Tokens <- Token{
		FileInfo: token.NewFileInfo(line, column),

		typ: t,
		val: val,
	}

	l.start = l.pos
	l.lineStart = l.line
	l.columnStart = l.column
}

func (l *Lexer) emit(t token.Token) {
	l.Tokens <- Token{
		FileInfo: token.NewFileInfo(l.lineStart, l.columnStart),

		typ: t,
		val: l.input[l.start:l.pos],
	}

	l.start = l.pos
	l.lineStart = l.line
	l.columnStart = l.column
}

// peek returns but does not consume the next rune from input
func (l *Lexer) peek() rune {
	rune := l.next()
	l.backup()
	return rune
}

// next consumes the next rune from input
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
		l.line++
		l.column = 0
	} else {
		l.column++
	}

	return r
}

// ignore skips over the pending input before this point
func (l *Lexer) ignore() {
	l.start = l.pos
	l.lineStart = l.line
	l.columnStart = l.column
}

// backup steps back one rune
func (l *Lexer) backup() {
	l.pos -= l.width

	r, _ := utf8.DecodeRuneInString(l.input[l.pos:])

	l.column = l.prevColumn

	if r == '\n' {
		l.line--
	}
}

// acceptRun consumes a run of runes from the valid setup
func (l *Lexer) acceptRun(valid string) {
	for strings.IndexRune(valid, l.next()) >= 0 {

	}

	l.backup()
}

// errorf emit an error token
func (l *Lexer) errorf(format string, args ...interface{}) stateFn {
	fname := l.name

	if fname == "" {
		fname = "<none>"
	}

	errMsg := fmt.Sprintf(format, args...)

	arguments := make([]interface{}, 0, len(args)+2)
	arguments = append(arguments, fname, l.line, l.column, errMsg)

	l.Tokens <- Token{
		FileInfo: token.NewFileInfo(l.line, l.column),

		typ: token.Illegal,
		val: fmt.Sprintf("%s:%d:%d: %s", arguments...),
	}

	l.start = len(l.input)
	l.lineStart = l.line
	l.columnStart = l.column
	l.pos = l.start

	return nil // finish the state machine
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
		if l.addSemicolon {
			l.emitVal(token.Semicolon, ";", l.line, l.column)
		}

		l.addSemicolon = false

		return nil
	case '0' <= r && r <= '9':
		digits := "0123456789"

		l.acceptRun(digits)

		next := l.peek()

		// >[2=]
		// cmd[2]
		if next == '=' || next == ']' || (!isIdentifier(l.peek()) && !isArgument(l.peek())) {
			l.emit(token.Number)
		} else if isIdentifier(l.peek()) {
			absorbIdentifier(l)

			if isArgument(l.peek()) {
				absorbArgument(l)

				l.emit(token.Arg)
			} else {
				l.emit(token.Ident)
			}
		} else if isArgument(l.peek()) {
			absorbArgument(l)
			l.emit(token.Arg)
		}

		return lexStart
	case r == ';':
		l.emit(token.Semicolon)
		return lexStart
	case isSpace(r):
		return lexSpace

	case isEndOfLine(r):
		l.ignore()

		if l.addSemicolon && l.openParens == 0 {
			l.emitVal(token.Semicolon, ";", l.line, l.column)
		}

		l.addSemicolon = false

		return lexStart
	case r == '"':
		l.ignore()

		return lexQuote
	case r == '#':
		return lexComment
	case r == '+':
		l.emit(token.Plus)
		return lexStart
	case r == '>':
		l.emit(token.Gt)
		return lexStart
	case r == '|':
		l.emit(token.Pipe)
		return lexStart
	case r == '$':
		r = l.next()

		if !isIdentifier(r) {
			return l.errorf("Expected identifier, but found %q", r)
		}

		absorbIdentifier(l)

		next := l.peek()
		if next != eof && !isSpace(next) &&
			!isEndOfLine(next) && next != ';' &&
			next != ')' && next != ',' && next != '+' &&
			next != '[' && next != ']' && next != '(' {
			l.errorf("Unrecognized character in action: %#U", next)
			return nil
		}

		l.emit(token.Variable)
		return lexStart
	case r == '=':
		if l.peek() == '=' {
			l.next()
			l.emit(token.Equal)
		} else {
			l.emit(token.Assign)
		}

		return lexStart
	case r == '!':
		if l.peek() == '=' {
			l.next()
			l.emit(token.NotEqual)
		} else {
			l.emit(token.Arg)
		}

		return lexStart
	case r == '<':
		if l.peek() == '=' {
			l.next()
			l.emit(token.AssignCmd)
		} else {
			l.emit(token.Lt)
		}

		return lexStart
	case r == '{':
		l.addSemicolon = false
		l.emit(token.LBrace)
		return lexStart
	case r == '}':
		l.emit(token.RBrace)
		l.addSemicolon = false
		return lexStart
	case r == '[':
		l.emit(token.LBrack)
		return lexStart
	case r == ']':
		l.emit(token.RBrack)
		return lexStart
	case r == '(':
		l.openParens++

		l.emit(token.LParen)
		l.addSemicolon = false
		return lexStart
	case r == ')':
		l.openParens--

		l.emit(token.RParen)
		l.addSemicolon = true
		return lexStart
	case r == ',':
		l.emit(token.Comma)
		return lexStart
	case isIdentifier(r):
		// nash literals are lowercase
		absorbIdentifier(l)

		next := l.peek()

		if isEndOfLine(next) || isSpace(next) ||
			next == '=' || next == '(' ||
			next == ')' || next == ',' ||
			next == '[' || next == eof {
			lit := scanIdentifier(l)

			if len(lit) > 1 && r >= 'a' && r <= 'z' {
				l.emit(token.Lookup(lit))
			} else {
				l.emit(token.Ident)
			}
		} else {
			absorbArgument(l)
			l.emit(token.Arg)
		}

		if next == eof && l.openParens > 0 {
			l.addSemicolon = false
		} else {
			l.addSemicolon = true
		}

		return lexStart
	case isArgument(r):
		absorbArgument(l)
		l.emit(token.Arg)
		l.addSemicolon = true
		return lexStart
	}

	return l.errorf("Unrecognized character in action: %#U", r)
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

func absorbArgument(l *Lexer) {
	for {
		r := l.next()

		if isArgument(r) {
			continue // absorb
		}

		break
	}

	l.backup() // pos is now ahead of the alphanum
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

		l.emitVal(token.String, string(data), l.lineStart, l.columnStart)

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

// isSpace reports whether r is a space character.
func isSpace(r rune) bool {
	return r == ' ' || r == '\t'
}

func isArgument(r rune) bool {
	isId := isAlpha(r)

	return isId || (r != eof && !isEndOfLine(r) && !isSpace(r) &&
		r != '$' && r != '{' && r != '}' && r != '(' && r != ']' && r != '[' &&
		r != ')' && r != '>' && r != '"' && r != ',' && r != ';' && r != '|')
}

func isIdentifier(r rune) bool {
	return isAlpha(r) || r == '_'
}

// isIdentifier reports whether r is a valid identifier
func isAlpha(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r)
}

// isEndOfLine reports whether r is an end-of-line character.
func isEndOfLine(r rune) bool {
	return r == '\r' || r == '\n'
}
