package ast

import (
	"strings"

	"github.com/NeowayLabs/nash/token"
)

func stringify(s string) string {
	return strings.Replace(strings.Replace(s, "\n", "\\n", -1),
		"\t", "\\t", -1)
}

func NewSimpleArg(pos token.Pos, n string, quoted bool) Expr {
	return NewStringExpr(pos, n, quoted)
}
