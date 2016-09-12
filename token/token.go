package token

import "strconv"

type (
	Token int

	FileInfo struct {
		line, column int
	}
)

const (
	Illegal Token = iota + 1 // error ocurred
	EOF
	Comment

	literal_beg

	Ident
	String // "<string>"
	Number // [0-9]+
	Arg

	literal_end

	operator_beg

	Assign    // =
	AssignCmd // <=
	Equal     // ==
	NotEqual  // !=
	Plus      // +
	Minus     // -
	Gt        // >
	Lt        // <

	Colon     // ,
	Semicolon // ;

	operator_end

	LBrace // {
	RBrace // }
	LParen // (
	RParen // )
	LBrack // [
	RBrack // ]
	Pipe

	Comma

	Variable

	keyword_beg

	Import
	SetEnv
	ShowEnv
	BindFn // "bindfn <fn> <cmd>
	Dump   // "dump" [ file ]
	Return
	If
	Else
	For
	Rfork
	Fn

	keyword_end
)

var tokens = [...]string{
	Illegal: "ILLEGAL",
	EOF:     "EOF",
	Comment: "COMMENT",

	Ident:  "IDENT",
	String: "STRING",
	Number: "NUMBER",
	Arg:    "ARG",

	Assign:    "=",
	AssignCmd: "<=",
	Equal:     "==",
	NotEqual:  "!=",
	Plus:      "+",
	Minus:     "-",
	Gt:        ">",
	Lt:        "<",

	Colon:     ",",
	Semicolon: ";",

	LBrace: "{",
	RBrace: "}",
	LParen: "(",
	RParen: ")",
	LBrack: "[",
	RBrack: "]",
	Pipe:   "|",

	Comma: ",",

	Variable: "VARIABLE",

	Import:  "import",
	SetEnv:  "setenv",
	ShowEnv: "showenv",
	BindFn:  "bindfn",
	Dump:    "dump",
	Return:  "return",
	If:      "if",
	Else:    "else",
	For:     "for",
	Rfork:   "rfork",
	Fn:      "fn",
}

var keywords map[string]Token

func init() {
	keywords = make(map[string]Token)
	for i := keyword_beg + 1; i < keyword_end; i++ {
		keywords[tokens[i]] = i
	}
}

func Lookup(ident string) Token {
	if tok, isKeyword := keywords[ident]; isKeyword {
		return tok
	}

	return Ident
}

func IsKeyword(t Token) bool {
	if t > keyword_beg && t < keyword_end {
		return true
	}

	return false
}

func NewFileInfo(l, c int) FileInfo { return FileInfo{l, c} }
func (info FileInfo) Line() int     { return info.line }
func (info FileInfo) Column() int   { return info.column }

func (tok Token) String() string {
	s := ""

	if 0 < tok && tok < Token(len(tokens)) {
		s = tokens[tok]
	}
	if s == "" {
		s = "token(" + strconv.Itoa(int(tok)) + ")"
	}
	return s
}
