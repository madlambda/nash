package ast

import (
	"strings"

	"github.com/NeowayLabs/nash/token"
)

func stringify(s string) string {
	return strings.Replace(strings.Replace(s, "\n", "\\n", -1),
		"\t", "\\t", -1)
}

func NewSimpleArg(pos token.Pos, n string, typ ArgType) *Arg {
	arg := NewArg(pos, typ)
	arg.SetString(n)
	return arg
}
