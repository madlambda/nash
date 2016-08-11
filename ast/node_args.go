package ast

import (
	"fmt"
	"strings"

	"github.com/NeowayLabs/nash/scanner"
	"github.com/NeowayLabs/nash/token"
)

// ArgFromToken is a helper to get an argument based on the lexer token
func ExprFromToken(val scanner.Token) (Expr, error) {
	switch val.Type() {
	case token.Arg:
		return NewArgString(val.Pos, val.Value(), false), nil
	case token.String:
		return NewArgString(val.Pos, val.Value(), true), nil
	case token.Variable:
		return NewArgVariable(val.Pos, val.Value()), nil
	}

	return fmt.Errorf("argFromToken doesn't support type %v", val), nil
}

// NewArgString creates a new string argument
func NewStringExpr(pos token.Pos, value string, quoted bool) *StringExpr {
	return &StringExpr{
		NodeType: NodeStringExpr,
		Pos:      pos,

		str:    value,
		quoted: quoted,
	}
}

// Value returns the argument string value
func (s *StringExpr) Value() string {
	return s.str
}

func (s *StringExpr) String() string {
	if s.quoted {
		return `"` + stringify(s.str) + `"`
	}

	return s.str
}

func NewListExpr(pos token.Pos) *ListExpr {
	return &ListExpr{
		NodeType: NodeListExpr,
		Pos:      pos,
	}
}

// PushExpr push an expression to end of the list
func (l *ListExpr) PushExpr(a Expr) {
	l.list = append(l.list, a)
}

func (l *ListElem) SetList(a []Expr) {
	l.list = a
}

func (l *ListElem) String() string {
	elems := make([]string, len(l.list))
	linecount := 0

	for i := 0; i < len(l.list); i++ {
		elems[i] = l.list[i].String()
		linecount += len(elems[i])
	}

	if linecount+len(l) > 50 {
		return "(\n\t" + strings.Join(elems, "\n\t") + "\n)"
	}

	return "(" + strings.Join(elems, " ") + ")"
}

func NewConcatExpr(pos token.Pos) *ConcatExpr {
	return &ConcatExpr{
		NodeType: NodeConcatExpr,
		Pos:      pos,
	}
}

// PushExpr push an expression to end of the concat list
func (c *ConcatExpr) PushExpr(a Expr) {
	c.concat = append(c.concat, a)
}

// SetConcatList set the concatenation parts
func (c *ConcatExpr) SetConcat(v []Expr) {
	c.concat = v
}

func (c *ConcatExpr) ConcatList() []Expr { return c.concat }

func (c *ConcatExpr) String() string {
	ret := ""

	for i := 0; i < len(c.concat); i++ {
		ret += c.concat[i].String()

		if i < (len(n.concat) - 1) {
			ret += "+"
		}
	}

	return ret
}

func NewVarExpr(pos token.Pos, name string) *VarExpr {
	return &VarExpr{
		NodeType: NodeVarExpr,
		Pos:      pos,
		name:     name,
	}
}

func (v *VarExpr) String() string {
	return v.name
}

func NewIndexExpr(pos token.Pos, variable *VarExpr, index Expr) *IndexExpr {
	return &IndexExpr{
		NodeType: NodeIndexExpr,
		Pos:      pos,

		variable: variable,
		index:    index,
	}
}

func (i *IndexExpr) String() string {
	return i.variable.String() + "[" + i.index.String() + "]"
}
