package cnt

import (
	"errors"
	"fmt"
	"io/ioutil"
)

// Execute the cnt file at given path
func Execute(path string) error {
	content, err := ioutil.ReadFile(path)

	if err != nil {
		return err
	}

	parser := NewParser(path, string(content))

	tr, err := parser.Parse()

	if err != nil {
		return err
	}

	if tr.Root == nil {
		return errors.New("nothing parsed")
	}

	root := tr.Root

	for _, node := range root.Nodes {
		fmt.Printf("Node: Type: %d, %v\n", node.Type(), node)
	}

	return nil
}
