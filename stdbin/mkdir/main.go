package main 

import (
	"os"
	"fmt"
)

func mkdirs(dirnames []string) error {
	for _, d := range dirnames {
		if err := os.MkdirAll(d, 0644); err != nil {
			return err
		}
	}

	return nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: %s <dir1> <dir2> ...\n", os.Args[0])
		os.Exit(1)
	}
	err := mkdirs(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err.Error())
		os.Exit(1)
	}
}