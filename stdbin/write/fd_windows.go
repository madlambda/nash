package main

import (
	"io"
	"os"
)

func specialFile(path string) (io.Writer, bool) {
	if fname == "CON" { // holycrap!
		return os.Stdout, true
	}
	return nil, false
}
