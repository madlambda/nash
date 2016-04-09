package cnt

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

const eof = -1

// Itemtype identifies the type of lex items
type (
	itemType int

	item struct {
		typ itemType
		val string
	}

	stateFn func(*lexer) stateFn

	// lexer holds the state of the scanner
	lexer struct {
		name  string    // used only for error reports
		input string    // the string being scanned
		start int       // start position of this item
		pos   int       // current position in the input
		width int       // width of last rune read
		items chan item // channel of scanned items
	}
)

const (
	itemError itemType = iota // error ocurred
	itemEOF
	itemComment
	itemCommand // alphanumeric identifier that's not a keyword
	itemArg
	itemLeftBlock  // {
	itemRightBlock // }
	itemString

	itemKeyword // used only to delimit the keywords
	//	itemIf
	//	itemFor
	itemRfork
	itemRforkFlags
)

const (
	spaceChars = " \t\r\n"

	rforkName  = "rfork"
	rforkFlags = "fupnsmi"
)

var (
	key = map[string]itemType{
		//		"if": itemIf,
		//		"for": itemFor,
		"rfork": itemRfork,
	}
)

func (i item) String() string {
	switch i.typ {
	case itemError:
		return "Error: " + i.val
	case itemEOF:
		return "EOF"
	}

	if len(i.val) > 10 {
		return fmt.Sprintf("%.10q...", i.val)
	}

	return fmt.Sprintf("%q", i.val)
}

// run lexes the input by executing state functions until the state is nil
func (l *lexer) run() {
	for state := lexStart; state != nil; {
		state = state(l)
	}

	close(l.items) // No more tokens will be delivered
}

func (l *lexer) emit(t itemType) {
	l.items <- item{t, l.input[l.start:l.pos]}
	l.start = l.pos
}

func (l *lexer) next() rune {
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
func (l *lexer) ignore() {
	l.start = l.pos
}

// backup steps back one rune
func (l *lexer) backup() {
	l.pos -= l.width
}

// peek returns but does not consume the next rune
func (l *lexer) peek() rune {
	rune := l.next()
	l.backup()
	return rune
}

// accept consumes the next rune if it's from the valid set
func (l *lexer) accept(valid string) bool {
	if strings.IndexRune(valid, l.next()) >= 0 {
		return true
	}

	l.backup()
	return false
}

// acceptRun consumes a run of runes from the valid setup
func (l *lexer) acceptRun(valid string) {
	for strings.IndexRune(valid, l.next()) >= 0 {

	}

	l.backup()
}

// errorf returns an error token
func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	l.items <- item{itemError, fmt.Sprintf(format, args...)}
	return nil // finish the state machine
}

func (l *lexer) String() string {
	return fmt.Sprintf("Lexer:\n\tPos: %d\n\tStart: %d\n",
		l.pos, l.start)
}

func lex(name, input string) (*lexer, chan item) {
	l := &lexer{
		name:  name,
		input: input,
		items: make(chan item),
	}

	go l.run() // concurrently run state machine

	return l, l.items
}

func lexStart(l *lexer) stateFn {
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

	case isAlphaNumeric(r):
		return lexIdentifier

	case r == '}':
		l.emit(itemRightBlock)
		return lexStart

	default:
		return l.errorf("Unrecognized character in action: %#U", r)
	}

	panic("Unreachable code")
}

func lexIdentifier(l *lexer) stateFn {
	for {
		r := l.next()

		if isAlphaNumeric(r) {
			continue // absorb
		}

		break
	}

	l.backup() // pos is now ahead of the alphanum

	word := l.input[l.start:l.pos]

	if word == rforkName {
		l.emit(itemRfork)
		return lexInsideRforkArgs
	}

	l.emit(itemCommand)
	return lexInsideCommand
}

// Rfork flags:
// c = create new process -> clone(2)
// u = user namespace
// p = pid namespace
// n = network namespace
// s = uts namespace
// m = mount namespace
// i = ipc namespace
func lexInsideRforkArgs(l *lexer) stateFn {
	// parse the rfork parameters

	if l.accept(" \t") {
		ignoreSpaces(l)
	}

	if !l.accept(rforkFlags) {
		return l.errorf("invalid rfork argument")
	}

	l.acceptRun(rforkFlags)

	l.emit(itemRforkFlags)

	if l.accept(" \t") {
		ignoreSpaces(l)
	}

	if l.accept("{") {
		l.emit(itemLeftBlock)
	}

	return lexStart
}

func lexInsideCommand(l *lexer) stateFn {
	r := l.next()

	switch {
	case isSpace(r):
		l.ignore()
		return lexSpaceArg
	case isEndOfLine(r):
		l.ignore()
		return lexStart
	case r == '"':
		l.ignore()
		return lexQuoteArg
	}

	return lexArg
}

func lexQuoteArg(l *lexer) stateFn {
	for {
		r := l.next()

		if r != '"' && r != eof {
			continue
		}

		if r == eof {
			return l.errorf("Quoted string not finished: %s", l.input[l.start:])
		}

		l.backup()
		l.emit(itemString)
		l.next()
		l.ignore()
		break
	}

	return lexInsideCommand
}

func lexArg(l *lexer) stateFn {
	for {
		r := l.next()

		if r == eof {
			if l.pos > l.start {
				l.emit(itemArg)
			}

			return nil
		}

		if isAlphaNumeric(r) {
			continue
		}

		l.backup()
		l.emit(itemArg)
		break
	}

	return lexInsideCommand
}

func lexComment(l *lexer) stateFn {
	for {
		r := l.next()

		if isEndOfLine(r) {
			l.backup()
			l.emit(itemComment)
			return lexStart
		}

		if r == eof {
			return nil
		}
	}

	panic("not reached")
}

func lexSpaceArg(l *lexer) stateFn {
	ignoreSpaces(l)
	return lexInsideCommand
}

func lexSpace(l *lexer) stateFn {
	ignoreSpaces(l)
	return lexStart
}

// isSpace reports whether r is a space character.
func isSpace(r rune) bool {
	return r == ' ' || r == '\t'
}

// isAlphaNumeric reports whether r is an alphabetic, digit, or underscore.
func isAlphaNumeric(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}

// isEndOfLine reports whether r is an end-of-line character.
func isEndOfLine(r rune) bool {
	return r == '\r' || r == '\n'
}

func ignoreSpaces(l *lexer) {
	for {
		r := l.next()

		if !isSpace(r) {
			break
		}
	}

	l.backup()
	l.ignore()
}
