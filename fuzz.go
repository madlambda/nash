// +build gofuzz

package nash

import "github.com/madlambda/nash/parser"

func Fuzz(data []byte) int {
	p := parser.NewParser("fuzz", string(data))

	tree, err := p.Parse()

	if err != nil {
		if tree != nil {
			panic("tree != nil")
		}

		return 0
	}

	return 1
}
