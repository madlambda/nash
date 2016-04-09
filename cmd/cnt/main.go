package main

import (
	"fmt"
	"os"

	
	"github.com/tiago4orion/cnt"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <optional cnt file>\n", os.Args[0])
		os.Exit(1)
	}

	path := os.Args[1]
	err := cnt.Execute(path)

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err.Error())
		os.Exit(1)
	}
}
