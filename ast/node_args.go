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
		return NewStringExpr(val.Pos(), val.Value(), false), nil
	case token.String:
		return NewStringExpr(val.Pos(), val.Value(), true), nil
	case token.Variable:
		return NewVarExpr(val.Pos(), val.Value()), nil
	}

	return nil, fmt.Errorf("argFromToken doesn't support type %v", val)
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

func (s *StringExpr) IsEqual(other Node) bool {
	if s == other {
		return true
	}

	value, ok := other.(*StringExpr)

	if !ok {
		return false
	}

	if s.quoted != value.quoted {
		return false
	}

	return s.str == value.str
}

func NewIntExpr(pos token.Pos, val int) *IntExpr {
	return &IntExpr{
		NodeType: NodeIntExpr,
		Pos:      pos,

		val: val,
	}
}

func (i *IntExpr) Value() int     { return i.val }
func (i *IntExpr) String() string { return string(i.val) }
func (i *IntExpr) IsEqual(other Node) bool {
	if i == other {
		return true
	}

	o, ok := other.(*IntExpr)

	if !ok {
		return false
	}

	return i.val == o.val
}

func NewListExpr(pos token.Pos, values []Expr) *ListExpr {
	return &ListExpr{
		NodeType: NodeListExpr,
		Pos:      pos,

		list: values,
	}
}

// PushExpr push an expression to end of the list
func (l *ListExpr) PushExpr(a Expr) {
	l.list = append(l.list, a)
}

func (l *ListExpr) IsEqual(other Node) bool {
	if other == l {
		return true
	}

	o, ok := other.(*ListExpr)

	if !ok {
		return false
	}

	if len(l.list) != len(o.list) {
		return false
	}

	for i := 0; i < len(l.list); i++ {
		if !l.list[i].IsEqual(o.list[i]) {
			return false
		}
	}

	return true
}

func (l *ListExpr) String() string {
	elems := make([]string, len(l.list))
	columnCount := 0

	for i := 0; i < len(l.list); i++ {
		elems[i] = l.list[i].String()
		columnCount += len(elems[i])
	}

	if columnCount+len(elems) > 50 {
		return "(\n\t" + strings.Join(elems, "\n\t") + "\n)"
	}

	return "(" + strings.Join(elems, " ") + ")"
}

func NewConcatExpr(pos token.Pos, parts []Expr) *ConcatExpr {
	return &ConcatExpr{
		NodeType: NodeConcatExpr,
		Pos:      pos,

		concat: parts,
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

func (c *ConcatExpr) IsEqual(other Node) bool {
	if c == other {
		return true
	}

	o, ok := other.(*ConcatExpr)

	if !ok {
		return false
	}

	if len(c.concat) != len(o.concat) {
		return false
	}

	for i := 0; i < len(c.concat); i++ {
		if !c.concat[i].IsEqual(o.concat[i]) {
			return false
		}
	}

	return true
}

func (c *ConcatExpr) String() string {
	ret := ""

	for i := 0; i < len(c.concat); i++ {
		ret += c.concat[i].String()

		if i < (len(c.concat) - 1) {
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

func (v *VarExpr) Name() string { return v.name }

func (v *VarExpr) IsEqual(other Node) bool {
	if v == other {
		return true
	}

	o, ok := other.(*VarExpr)

	if ok {
		return true
	}

	return v.name == o.name
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
func (i *IndexExpr) IsEqual(other Node) bool {
	if i == other {
		return true
	}

	o, ok := other.(*IndexExpr)

	if !ok {
		return false
	}

	return i.variable.IsEqual(o.variable) && i.index.IsEqual(o.index)
}
