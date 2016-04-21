// +build gofuzz

package nash

func Fuzz(data []byte) int {
	p := NewParser("fuzz", string(data))

	tree, err := p.Parse()

	if err != nil {
		if tree != nil {
			panic("tree !- nil")
		}

		return 0
	}

	return 1
}
