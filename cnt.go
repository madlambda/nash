package cnt

import (
	"fmt"
	"io/ioutil"
	"strings"
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
	itemText
	itemIf
	itemFor
	itemRfork
)

const (
	rforkMeta = "rfork"
)

func (i item) String() string {
	switch i.typ {
	case itemError:
		return i.val
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
	for state := lexText; state != nil; {
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

func lex(name, input string) (*lexer, chan item) {
	l := &lexer{
		name:  name,
		input: input,
		items: make(chan item),
	}

	go l.run() // concurrently run state machine

	return l, l.items
}

func lexRforkMeta(l *lexer) stateFn {
	l.emit(itemRfork)

	return nil
}

func lexText(l *lexer) stateFn {
	for {
		if strings.HasPrefix(l.input[l.pos:], rforkMeta) {
			if l.pos > l.start {
				l.emit(itemText)
			}

			return lexRforkMeta // next state
		}

		if l.next() == eof {
			break
		}
	}

	// correctly reached EOF
	if l.pos > l.start {
		l.emit(itemText)
	}

	l.emit(itemEOF)
	return nil
}

func Execute(path string) error {
	fmt.Printf("Executing %s...\n", path)

	content, err := ioutil.ReadFile(path)

	if err != nil {
		return err
	}

	_, items := lex("cnt", string(content))

	for item := range items {
		fmt.Printf("Token: %v\n", item)
	}

	return nil
}
