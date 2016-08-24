package ast

import "strings"

func stringify(s string) string {
	return strings.Replace(strings.Replace(s, "\n", "\\n", -1),
		"\t", "\\t", -1)
}
