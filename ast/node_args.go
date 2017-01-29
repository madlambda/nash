package ast

import (
	"fmt"

	"github.com/NeowayLabs/nash/scanner"
	"github.com/NeowayLabs/nash/token"
)

// ArgFromToken is a helper to get an argument based on the lexer token
func ExprFromToken(val scanner.Token) (Expr, error) {
	switch val.Type() {
	case token.Arg:
		return NewStringExpr(token.NewFileInfo(val.Line(), val.Column()), val.Value(), false), nil
	case token.String:
		return NewStringExpr(token.NewFileInfo(val.Line(), val.Column()), val.Value(), true), nil
	case token.Variable:
		return NewVarExpr(token.NewFileInfo(val.Line(), val.Column()), val.Value()), nil
	}

	return nil, fmt.Errorf("argFromToken doesn't support type %v", val)
}

// NewArgString creates a new string argument
func NewStringExpr(info token.FileInfo, value string, quoted bool) *StringExpr {
	return &StringExpr{
		NodeType: NodeStringExpr,
		FileInfo: info,

		str:    value,
		quoted: quoted,
	}
}

// Value returns the argument string value
func (s *StringExpr) Value() string {
	return s.str
}

func (s *StringExpr) SetValue(a string) {
	s.str = a
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

	if !cmpInfo(s, other) {
		return false
	}

	return s.str == value.str
}

func NewIntExpr(info token.FileInfo, val int) *IntExpr {
	return &IntExpr{
		NodeType: NodeIntExpr,
		FileInfo: info,

		val: val,
	}
}

func (i *IntExpr) Value() int { return i.val }

func (i *IntExpr) IsEqual(other Node) bool {
	if i == other {
		return true
	}

	o, ok := other.(*IntExpr)

	if !ok {
		return false
	}

	if !cmpInfo(i, other) {
		return false
	}

	return i.val == o.val
}

func NewListExpr(info token.FileInfo, values []Expr) *ListExpr {
	return &ListExpr{
		NodeType: NodeListExpr,
		FileInfo: info,

		list: values,
	}
}

func (l *ListExpr) List() []Expr { return l.list }

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
			debug("%v(%s) != %v(%s)", l.list[i], l.list[i].Type(),
				o.list[i], o.list[i].Type())
			return false
		}
	}

	return cmpInfo(l, other)
}

func NewConcatExpr(info token.FileInfo, parts []Expr) *ConcatExpr {
	return &ConcatExpr{
		NodeType: NodeConcatExpr,
		FileInfo: info,

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

func (c *ConcatExpr) List() []Expr { return c.concat }

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

	return cmpInfo(c, other)
}

func NewVarExpr(info token.FileInfo, name string) *VarExpr {
	return &VarExpr{
		NodeType: NodeVarExpr,
		FileInfo: info,
		name:     name,
	}
}

func (v *VarExpr) Name() string { return v.name }

func (v *VarExpr) IsEqual(other Node) bool {
	if v == other {
		return true
	}

	o, ok := other.(*VarExpr)

	if !ok {
		return false
	}

	if !cmpInfo(v, other) {
		return false
	}

	return v.name == o.name
}

func NewIndexExpr(info token.FileInfo, variable *VarExpr, index Expr) *IndexExpr {
	return &IndexExpr{
		NodeType: NodeIndexExpr,
		FileInfo: info,

		variable: variable,
		index:    index,
	}
}

func (i *IndexExpr) Var() *VarExpr { return i.variable }
func (i *IndexExpr) Index() Expr   { return i.index }

func (i *IndexExpr) IsEqual(other Node) bool {
	if i == other {
		return true
	}

	o, ok := other.(*IndexExpr)

	if !ok {
		return false
	}

	if !cmpInfo(i, other) {
		return false
	}

	return i.variable.IsEqual(o.variable) && i.index.IsEqual(o.index)
}
