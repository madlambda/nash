package cnt

import (
	"fmt"
	"io/ioutil"
)

func Execute(path string) error {
	content, err := ioutil.ReadFile(path)

	if err != nil {
		return err
	}

	_, items := lex("cnt", string(content))

	for item := range items {
		fmt.Printf("Token: %v\n", item)
	}

	return nil
}

