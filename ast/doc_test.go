package ast_test

import (
	"fmt"

	"github.com/madlambda/nash/ast"
	"github.com/madlambda/nash/token"
)

func Example_AssignmentNode() {
	one := ast.NewNameNode(token.NewFileInfo(1, 0), "one", nil)
	two := ast.NewNameNode(token.NewFileInfo(1, 4), "two", nil)
	value1 := ast.NewStringExpr(token.NewFileInfo(1, 8), "1", true)
	value2 := ast.NewStringExpr(token.NewFileInfo(1, 10), "2", true)
	assign := ast.NewAssignNode(token.NewFileInfo(1, 0),
		[]*ast.NameNode{one, two},
		[]ast.Expr{value1, value2},
	)

	fmt.Printf("%s", assign)

	// Output: one, two = "1", "2"
}

func Example_AssignmentNode_Single() {
	operatingSystems := ast.NewNameNode(token.NewFileInfo(1, 0), "operatingSystems", nil)
	values := []ast.Expr{
		ast.NewStringExpr(token.NewFileInfo(1, 19), "plan9 from bell labs", true),
		ast.NewStringExpr(token.NewFileInfo(2, 19), "unix", true),
		ast.NewStringExpr(token.NewFileInfo(3, 19), "linux", true),
		ast.NewStringExpr(token.NewFileInfo(4, 19), "oberon", true),
		ast.NewStringExpr(token.NewFileInfo(5, 19), "windows", true),
	}

	list := ast.NewListExpr(token.NewFileInfo(0, 18), values)
	assign := ast.NewSingleAssignNode(token.NewFileInfo(1, 0),
		operatingSystems,
		list,
	)

	fmt.Printf("%s", assign)

	// Output: operatingSystems = (
	// 	"plan9 from bell labs"
	// 	"unix"
	// 	"linux"
	// 	"oberon"
	// 	"windows"
	// )
}
