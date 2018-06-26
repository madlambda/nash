package main

import (
	"io"
	"os"
)

func specialFile(path string) (io.WriteCloser, bool) {
	if path == "CON" { // holycrap!
		return os.Stdout, true
	}
	return nil, false
}