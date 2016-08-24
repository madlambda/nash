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

	literal_end

	operator_beg

	Assign    // =
	AssignCmd // <=
	Equal     // ==
	NotEqual  // !=
	Concat    // +

	operator_end

	LBrace // {
	RBrace // }
	LParen // (
	RParen // )
	LBrack // [
	RBrack // ]

	Variable
	ListElem
	Command // alphanumeric identifier that's not a keyword
	Pipe

	redirect_beg

	RedirRight // >
	RedirMapLSide
	RedirMapRSide

	redirect_end

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
	FnDecl
	FnInv // <identifier>(<args>)
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
	Concat:    "+",

	LBrace: "{",
	RBrace: "}",
	LParen: "(",
	RParen: ")",
	LBrack: "[",
	RBrack: "]",

	Variable: "VARIABLE",
	ListElem: "LIST-ELEM",
	Command:  "COMMAND",
	Pipe:     "|",

	RedirRight:    ">",
	RedirMapLSide: "MAP-LSIDE",
	RedirMapRSide: "MAP-RSIDE",

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
	FnDecl:  "FN",
	FnInv:   "FN-INV",
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
