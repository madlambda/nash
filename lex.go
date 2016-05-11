package nash

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Itemtype identifies the type of lex items
type (
	itemType int

	item struct {
		typ itemType
		pos Pos // start position of this item
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

		lastNode itemType
	}
)

//go:generate stringer -type=itemType

const (
	eof = -1

	itemError itemType = iota + 1 // error ocurred
	itemEOF
	itemImport
	itemComment
	itemSetEnv
	itemShowEnv
	itemVarName
	itemConcat
	itemVariable
	itemListOpen
	itemListClose
	itemListElem
	itemCommand // alphanumeric identifier that's not a keyword
	itemArg
	itemLeftBlock     // {
	itemRightBlock    // }
	itemLeftParen     // (
	itemRightParen    // )
	itemString        // "<string>"
	itemRedirRight    // >
	itemRedirRBracket // [ eg.: cmd >[1] file.out
	itemRedirLBracket // ]
	itemRedirFile
	itemRedirNetAddr
	itemRedirMapEqual // = eg.: cmd >[2=1]
	itemRedirMapLSide
	itemRedirMapRSide

	itemIf // if <condition> { <block> }
	itemElse
	itemComparison
	//	itemFor
	itemRfork
	itemRforkFlags
	itemCd

	itemFnDecl // fn <name>(<arg>) { <block> }
	itemFnInv  // <identifier>(<args>)
)

const (
	spaceChars = " \t\r\n"

	rforkFlags = "cnsmifup"
)

func (i item) String() string {
	switch i.typ {
	case itemError:
		return "Error: " + i.val
	case itemEOF:
		return "EOF"
	}

	if len(i.val) > 10 {
		return fmt.Sprintf("(%s) - pos: %d, val: %.10q...", i.typ, i.pos, i.val)
	}

	return fmt.Sprintf("(%s) - pos: %d, val: %q", i.typ, i.pos, i.val)
}

// run lexes the input by executing state functions until the state is nil
func (l *lexer) run() {
	for state := lexStart; state != nil; {
		state = state(l)
	}

	l.emit(itemEOF)
	close(l.items) // No more tokens will be delivered
}

func (l *lexer) emit(t itemType) {
	l.items <- item{
		typ: t,
		val: l.input[l.start:l.pos],
		pos: Pos(l.start),
	}

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
	l.items <- item{
		typ: itemError,
		val: fmt.Sprintf(format, args...),
		pos: Pos(l.start),
	}

	return nil // finish the state machine
}

func (l *lexer) String() string {
	return fmt.Sprintf("Lexer:\n\tPos: %d\n\tStart: %d\n",
		l.pos, l.start)
}

func lex(name, input string) *lexer {
	l := &lexer{
		name:  name,
		input: input,
		items: make(chan item),
	}

	go l.run() // concurrently run state machine

	return l
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

	case isIdentifier(r):
		return lexIdentifier

	case isSafePath(r):
		return lexInsideCommandName

	case r == '{':
		l.ignore()
		return l.errorf("Unexpected open block \"%#U\"", r)

	case r == '}':
		l.emit(itemRightBlock)
		return lexStart
	case r == '(':
		l.emit(itemLeftParen)
		return lexInsideFnInv
	case r == ')':
		l.emit(itemRightParen)
		return lexStart
	}

	return l.errorf("Unrecognized character in action: %#U", r)
}

func lexIdentifier(l *lexer) stateFn {
	for {
		r := l.next()

		if isIdentifier(r) {
			continue // absorb
		}

		break
	}

	l.backup() // pos is now ahead of the alphanum

	word := l.input[l.start:l.pos]

	r := l.peek()

	if r == '(' {
		l.emit(itemFnInv)
		l.next()
		l.emit(itemLeftParen)
		return lexInsideFnInv
	}

	if isSpace(r) || r == '=' {
		// lookahead by hand, to avoid more complex lexer API
		for i := l.pos; i < len(l.input); i++ {
			r, _ := utf8.DecodeRuneInString(l.input[i:])

			if !isSpace(r) {
				if r == '=' {
					l.emit(itemVarName)

					ignoreSpaces(l)
					l.next()
					l.ignore()
					return lexInsideAssignment
				}

				break
			}
		}
	}

	switch word {
	case "import":
		l.emit(itemImport)
		return lexInsideImport
	case "rfork":
		l.emit(itemRfork)
		return lexInsideRforkArgs
	case "cd":
		l.emit(itemCd)
		return lexInsideCd
	case "setenv":
		l.emit(itemSetEnv)
		return lexInsideSetenv
	case "if":
		l.emit(itemIf)
		return lexInsideIf
	case "fn":
		l.emit(itemFnDecl)
		return lexInsideFnDecl
	case "else":
		l.emit(itemElse)
		return lexInsideElse
	case "showenv":
		l.emit(itemShowEnv)

		ignoreSpaces(l)

		r := l.next()

		if !isEndOfLine(r) && r != eof {
			pos := l.pos

			l.backup()
			return l.errorf("Unexpected character %q at pos %d. Showenv doesn't have arguments.",
				r, pos)
		}

		l.backup()
		return lexStart
	}

	l.emit(itemCommand)
	return lexInsideCommand
}

func lexInsideImport(l *lexer) stateFn {
	ignoreSpaces(l)

	r := l.next()

	if r == '"' {
		l.ignore()
		return func(l *lexer) stateFn {
			return lexQuote(l, lexStart)
		}
	}

	if isIdentifier(r) || isSafePath(r) {
		// parse as normal argument
		return func(l *lexer) stateFn {
			return lexArg(l, lexStart)
		}
	}

	l.backup()
	return lexStart
}

func lexInsideSetenv(l *lexer) stateFn {
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

	l.emit(itemVarName)
	return lexStart
}

func lexInsideAssignment(l *lexer) stateFn {
	ignoreSpaces(l)

	r := l.peek()

	switch {
	case r == '(':
		return lexInsideListVariable
	case r == '"':
		l.next()
		l.ignore()

		return func(l *lexer) stateFn {
			lexQuote(l, nil)

			ignoreSpaces(l)

			r := l.peek()

			switch {
			case r == '+':
				l.next()
				l.emit(itemConcat)

				return lexInsideAssignment
			}

			if !isEndOfLine(r) && r != eof {
				return l.errorf("Invalid assignment. Expected '+' or EOL, but found %q at pos '%d'",
					r, l.pos)
			}

			return lexStart
		}

	case r == '$':
		return lexInsideCommonVariable(l, lexInsideAssignment)
	}

	return l.errorf("Unexpected variable value '%c'. Expected '\"' for quoted string or '$' for variable.", r)
}

func lexInsideListVariable(l *lexer) stateFn {
	r := l.next()

	if r != '(' {
		return l.errorf("Invalid list, expected '(' but found '%c'", r)
	}

	l.emit(itemListOpen)
nextelem:
	for {
		r = l.peek()

		if !isIdentifier(r) {
			break
		}

		l.next()
	}

	if l.start < l.pos {
		l.emit(itemListElem)
	}

	r = l.next()

	if isSpace(r) {
		l.ignore()
		goto nextelem
	} else if r != ')' {
		return l.errorf("Expected end of list ')' but found '%c'", r)
	}

	l.emit(itemListClose)
	return lexStart
}

func lexInsideCommonVariable(l *lexer, nextConcatFn stateFn) stateFn {
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

	l.emit(itemVariable)

	ignoreSpaces(l)

	r = l.peek()

	switch {
	case r == '+':
		l.next()
		l.emit(itemConcat)

		return nextConcatFn
	}

	if !isEndOfLine(r) && r != eof {
		return l.errorf("Invalid assignment. Expected '+' or EOL, but found %q at pos '%d'",
			r, l.pos)
	}

	return lexStart
}

func lexInsideCd(l *lexer) stateFn {
	// parse the cd directory
	ignoreSpaces(l)

	r := l.next()

	if r == '"' {
		l.ignore()
		return func(l *lexer) stateFn {
			lexQuote(l, lexStart)
			ignoreSpaces(l)

			r = l.peek()

			switch {
			case r == '+':
				l.next()
				l.emit(itemConcat)
				return lexInsideCd
			}

			return lexStart
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

		l.emit(itemVariable)

		ignoreSpaces(l)

		r = l.peek()

		switch {
		case r == '+':
			l.next()
			l.emit(itemConcat)
			return lexInsideCd
		}

		return lexStart
	}

	if isIdentifier(r) || isSafePath(r) {
		// parse as normal argument
		return func(l *lexer) stateFn {
			return lexArg(l, lexStart)
		}
	}

	l.backup()
	return lexStart
}

func lexIfLRValue(l *lexer) bool {
	ignoreSpaces(l)

	r := l.next()

	switch {
	case r == '"':
		l.ignore()
		lexQuote(l, nil)
		return true
	case r == '$':
		for {
			r = l.next()

			if !isIdentifier(r) {
				break
			}
		}

		l.backup()

		if r == '"' {
			l.errorf("Invalid quote inside variable name")
			return false
		}

		l.emit(itemVariable)
		return true
	}

	l.errorf("Unexpected char %q at pos %d. IfDecl expects string or variable", r, l.pos)
	return false
}

func lexInsideIf(l *lexer) stateFn {
	ok := lexIfLRValue(l)

	if !ok {
		return nil
	}

	ignoreSpaces(l)

	// get first char of operator. Eg.: '!'
	if !l.accept("=!") {
		l.errorf("Unexpected char %q inside if statement", l.peek())
		l.backup()
		return nil
	}

	// get second char. Eg.: '='
	if !l.accept("=!") {
		l.errorf("Unexpected char %q inside if statement", l.peek())
		l.backup()
		return nil
	}

	word := l.input[l.start:l.pos]

	if word != "==" && word != "!=" {
		return l.errorf("Invalid comparison operator '%s'", word)
	}

	l.emit(itemComparison)

	ok = lexIfLRValue(l)

	if !ok {
		return nil
	}

	ignoreSpaces(l)

	r := l.next()

	if r != '{' {
		return l.errorf("Unexpected %q at pos %d. Expected '{'", r, l.pos)
	}

	l.emit(itemLeftBlock)

	return lexStart
}

func lexInsideFnDecl(l *lexer) stateFn {
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

	l.emit(itemVarName)

	r = l.next()

	if r != '(' {
		return l.errorf("Unexpected symbol %q. Expected '('", r)
	}

	l.emit(itemLeftParen)

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
		l.emit(itemVarName)

		r = l.peek()

		if r == ',' {
			l.next()
			goto getnextarg
		} else if r != ')' {
			return l.errorf("Unexpected symbol %q. Expected ',' or ')'", r)
		}
	}

	l.next()
	l.emit(itemRightParen)

	ignoreSpaces(l)

	r = l.next()

	if r != '{' {
		return l.errorf("Unexpected symbol %q. Expected '{'", r)
	}

	l.emit(itemLeftBlock)

	return lexStart
}

func lexMoreFnArgs(l *lexer) stateFn {
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

func lexInsideFnInv(l *lexer) stateFn {
	ignoreSpaces(l)

	var r rune

	r = l.peek()

	if r == '"' {
		l.next()
		l.ignore()
		lexQuote(l, nil)

		return lexMoreFnArgs
	} else if r == '$' {
		for {
			r = l.next()

			if isIdentifier(r) || r == '$' {
				continue
			}

			break
		}

		l.backup()

		word := l.input[l.start:l.pos]

		if len(word) > 0 {
			l.emit(itemVariable)

			return lexMoreFnArgs
		}
	} else if r == ')' {
		l.next()
		l.emit(itemRightParen)
		return lexStart
	}

	return l.errorf("Unexpected symbol %q. Expected quoted string or variable", r)
}

func lexInsideElse(l *lexer) stateFn {
	ignoreSpaces(l)

	r := l.next()

	if r == '{' {
		l.emit(itemLeftBlock)
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
		l.emit(itemIf)
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
func lexInsideRforkArgs(l *lexer) stateFn {
	// parse the rfork parameters

	ignoreSpaces(l)

	if !l.accept(rforkFlags) {
		return l.errorf("invalid rfork argument: %s", string(l.peek()))
	}

	l.acceptRun(rforkFlags)

	l.emit(itemRforkFlags)

	ignoreSpaces(l)

	if l.accept("{") {
		l.emit(itemLeftBlock)
	}

	return lexStart
}

func lexInsideCommandName(l *lexer) stateFn {
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

	l.emit(itemCommand)
	return lexInsideCommand
}

func lexInsideCommand(l *lexer) stateFn {
	r := l.next()

	switch {
	case r == eof:
		return nil
	case isSpace(r):
		l.ignore()
		return lexSpaceArg
	case isEndOfLine(r):
		l.ignore()
		return lexStart
	case r == '#':
		return lexComment
	case r == '"':
		l.ignore()
		return func(l *lexer) stateFn {
			return lexQuote(l, lexInsideCommand)
		}
	case r == '}':
		l.emit(itemRightBlock)
		return lexStart

	case r == '>':
		l.emit(itemRedirRight)
		return lexInsideRedirect
	case r == '$':
		l.backup()
		return lexInsideCommonVariable(l, lexInsideCommand)
	case isSafeArg(r):
		break
	default:
		return l.errorf("Invalid char %q at pos %d", r, l.pos)
	}

	return func(l *lexer) stateFn {
		return lexArg(l, lexInsideCommand)
	}
}

func lexQuote(l *lexer, nextFn stateFn) stateFn {
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

	return nextFn
}

func lexArg(l *lexer, nextFn stateFn) stateFn {
	for {
		r := l.next()

		if r == eof {
			if l.pos > l.start {
				l.emit(itemArg)
			}

			return nil
		}

		if isIdentifier(r) || isSafeArg(r) {
			continue
		}

		l.backup()
		l.emit(itemArg)
		break
	}

	return nextFn
}

func lexInsideRedirect(l *lexer) stateFn {
	var r rune

	for {
		r = l.next()

		if !isSpace(r) {
			break
		}

		l.ignore()
	}

	switch {
	case r == '[':
		l.emit(itemRedirLBracket)
		return lexInsideRedirMapLeftSide
	case r == ']':
		return l.errorf("Unexpected ']' at pos %d", l.pos)
	}

	if isSafePath(r) {
		for {
			r = l.next()

			if !isSafePath(r) {
				l.backup()
				break
			}
		}

		l.emit(itemRedirFile)
	} else if r == '"' {
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

		word := l.input[l.start:l.pos]

		if (len(word) > 6 && word[0:6] == "tcp://") ||
			(len(word) > 6 && word[0:6] == "udp://") ||
			(len(word) > 7 && word[0:7] == "unix://") {
			l.emit(itemRedirNetAddr)
		} else {
			l.emit(itemRedirFile)
		}

		l.next() // get last '"' again
		l.ignore()
	} else {
		return l.errorf("Unexpected redirect identifier: %s", l.input[l.start:l.pos])
	}

	// verify if have more redirects

	for {
		r = l.next()

		if !isSpace(r) {
			break
		}

		l.ignore()
	}

	if r == '>' {
		l.emit(itemRedirRight)
		return lexInsideRedirect
	}

	return lexStart
}

func lexInsideRedirMapLeftSide(l *lexer) stateFn {
	var r rune

	for {
		r = l.peek()

		if !unicode.IsDigit(r) {
			if len(l.input[l.start:l.pos]) == 0 {
				return l.errorf("Unexpected %c at pos %d", r, l.pos)
			}

			if r == ']' {
				// [xxx]
				l.emit(itemRedirMapLSide)
				l.next()
				l.emit(itemRedirRBracket)

				ignoreSpaces(l)

				r = l.next()

				if isSafePath(r) || r == '"' {
					return lexInsideRedirect
				}

				if r == '>' {
					l.emit(itemRedirRight)
					return lexInsideRedirect
				}

				l.backup()

				return lexStart
			}

			if r != '=' {
				return l.errorf("Expected '=' but found '%c' at por %d", r, l.pos)
			}

			// [xxx=
			l.emit(itemRedirMapLSide)
			l.next()
			l.emit(itemRedirMapEqual)

			return lexInsideRedirMapRightSide
		}

		r = l.next()
	}
}

func lexInsideRedirMapRightSide(l *lexer) stateFn {
	var r rune

	// [xxx=yyy]
	for {
		r = l.peek()

		if !unicode.IsDigit(r) {
			if len(l.input[l.start:l.pos]) == 0 {
				if r == ']' {
					l.next()
					l.emit(itemRedirRBracket)

					ignoreSpaces(l)

					r = l.next()

					if isSafePath(r) || r == '"' {
						return lexInsideRedirect
					}

					if r == '>' {
						l.emit(itemRedirRight)
						return lexInsideRedirect
					}

					l.backup()

					return lexStart
				}

				return l.errorf("Unexpected %c at pos %d", r, l.pos)
			}

			l.emit(itemRedirMapRSide)
			l.next()
			l.emit(itemRedirRBracket)

			ignoreSpaces(l)

			r = l.next()

			if isSafePath(r) || r == '"' {
				return lexInsideRedirect
			}

			if r == '>' {
				l.emit(itemRedirRight)
				return lexInsideRedirect
			}

			l.backup()
			break
		}

		r = l.next()
	}

	return lexStart
}

func lexComment(l *lexer) stateFn {
	for {
		r := l.next()

		if isEndOfLine(r) {
			l.backup()
			l.emit(itemComment)

			break
		}

		if r == eof {
			l.backup()
			l.emit(itemComment)
			break
		}
	}

	return lexStart
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
