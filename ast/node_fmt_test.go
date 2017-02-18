package ast

import (
	"testing"

	"github.com/NeowayLabs/nash/token"
)

func testPrinter(t *testing.T, node Node, expected string) {
	if node.String() != expected {
		t.Errorf("Values differ: '%s' != '%s'", node, expected)
	}
}

func TestAstPrinterStringExpr(t *testing.T) {
	for _, testcase := range []struct {
		expected string
		node     Node
	}{
		// quote
		{
			expected: `"\""`,
			node:     NewStringExpr(token.NewFileInfo(1, 0), "\"", true),
		},

		// escape
		{
			expected: `"\\this is a test\n"`,
			node:     NewStringExpr(token.NewFileInfo(1, 0), "\\this is a test\n", true),
		},

		// tab
		{
			expected: `"this is a test\t"`,
			node:     NewStringExpr(token.NewFileInfo(1, 0), "this is a test\t", true),
		},

		// linefeed
		{
			expected: `"this is a test\n"`,
			node:     NewStringExpr(token.NewFileInfo(1, 0), "this is a test\n", true),
		},
		{
			expected: `"\nthis is a test"`,
			node:     NewStringExpr(token.NewFileInfo(1, 0), "\nthis is a test", true),
		},
		{
			expected: `"\n\n\n"`,
			node:     NewStringExpr(token.NewFileInfo(1, 0), "\n\n\n", true),
		},

		// carriege return
		{
			expected: `"this is a test\r"`,
			node:     NewStringExpr(token.NewFileInfo(1, 0), "this is a test\r", true),
		},
		{
			expected: `"\rthis is a test"`,
			node:     NewStringExpr(token.NewFileInfo(1, 0), "\rthis is a test", true),
		},
		{
			expected: `"\r\r\r"`,
			node:     NewStringExpr(token.NewFileInfo(1, 0), "\r\r\r", true),
		},
	} {
		testPrinter(t, testcase.node, testcase.expected)
	}
}
