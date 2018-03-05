package main

import (
	"io"
	"os"
)

func specialFile(path string) (io.WriteCloser, bool) {
	if path == "/dev/stdout" {
		return os.Stdout, true
	} else if path == "/dev/stderr" {
		return os.Stderr, true
	}
	return nil, false
}
