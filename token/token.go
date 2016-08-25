package token

import "strconv"

type (
	Token int
	Pos   int
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
	Path

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

	Builtin
	Import
	SetEnv
	ShowEnv
	BindFn // "bindfn <fn> <cmd>
	Dump   // "dump" [ file ]
	Return
	If
	Else
	For
	ForIn
	Rfork
	Cd
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
	Path:   "PATH",

	Assign:    "=",
	AssignCmd: "<=",
	Equal:     "==",
	NotEqual:  "!=",
	Plus:      "+",
	Minus:     "-",
	Gt:        ">",
	Lt:        "<",

	LBrace: "{",
	RBrace: "}",
	LParen: "(",
	RParen: ")",
	LBrack: "[",
	RBrack: "]",
	Pipe:   "|",

	Comma: ",",

	Variable: "VARIABLE",

	Builtin: "BUILTIN",
	Import:  "IMPORT",
	SetEnv:  "SETENV",
	ShowEnv: "SHOWENV",
	BindFn:  "BINDFN",
	Dump:    "DUMP",
	Return:  "RETURN",
	If:      "IF",
	Else:    "ELSE",
	For:     "FOR",
	ForIn:   "FOR-IN",
	Rfork:   "RFORK",
	Cd:      "CD",
	Fn:      "FN",
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

// token.Position returns the position of the node in file
func (p Pos) Position() Pos {
	return p
}

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
