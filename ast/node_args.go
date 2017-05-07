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
	if !s.equal(s, other) {
		return false
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

func NewIntExpr(info token.FileInfo, val int) *IntExpr {
	return &IntExpr{
		NodeType: NodeIntExpr,
		FileInfo: info,

		val: val,
	}
}

func (i *IntExpr) Value() int { return i.val }

func (i *IntExpr) IsEqual(other Node) bool {
	if !i.equal(i, other) {
		return false
	}

	o, ok := other.(*IntExpr)

	if !ok {
		return false
	}

	return i.val == o.val
}

func NewListExpr(info token.FileInfo, values []Expr) *ListExpr {
	return NewListVariadicExpr(info, values, false)
}

func NewListVariadicExpr(info token.FileInfo, values []Expr, variadic bool) *ListExpr {
	return &ListExpr{
		NodeType: NodeListExpr,
		FileInfo: info,

		List:       values,
		IsVariadic: variadic,
	}
}

// PushExpr push an expression to end of the list
func (l *ListExpr) PushExpr(a Expr) {
	l.List = append(l.List, a)
}

func (l *ListExpr) IsEqual(other Node) bool {
	if !l.equal(l, other) {
		return false
	}

	o, ok := other.(*ListExpr)

	if !ok {
		return false
	}

	if len(l.List) != len(o.List) {
		return false
	}

	for i := 0; i < len(l.List); i++ {
		if !l.List[i].IsEqual(o.List[i]) {
			debug("%v(%s) != %v(%s)", l.List[i], l.List[i].Type(),
				o.List[i], o.List[i].Type())
			return false
		}
	}

	return true
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
	if !c.equal(c, other) {
		return false
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

func NewVarExpr(info token.FileInfo, name string) *VarExpr {
	return NewVarVariadicExpr(info, name, false)
}

func NewVarVariadicExpr(info token.FileInfo, name string, isVariadic bool) *VarExpr {
	return &VarExpr{
		NodeType:   NodeVarExpr,
		FileInfo:   info,
		Name:       name,
		IsVariadic: isVariadic,
	}
}

func (v *VarExpr) IsEqual(other Node) bool {
	if !v.equal(v, other) {
		return false
	}

	o, ok := other.(*VarExpr)
	if !ok {
		return false
	}

	return v.Name == o.Name &&
		v.IsVariadic == o.IsVariadic
}

func NewIndexExpr(info token.FileInfo, va *VarExpr, idx Expr) *IndexExpr {
	return NewIndexVariadicExpr(info, va, idx, false)
}

func NewIndexVariadicExpr(info token.FileInfo, va *VarExpr, idx Expr, variadic bool) *IndexExpr {
	return &IndexExpr{
		NodeType: NodeIndexExpr,
		FileInfo: info,

		Var:        va,
		Index:      idx,
		IsVariadic: variadic,
	}
}

func (i *IndexExpr) IsEqual(other Node) bool {
	if !i.equal(i, other) {
		return false
	}

	o, ok := other.(*IndexExpr)
	if !ok {
		return false
	}

	return i.Var.IsEqual(o.Var) &&
		i.Index.IsEqual(o.Index) &&
		i.IsVariadic == o.IsVariadic
}
