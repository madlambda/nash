package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

var banner = fmt.Sprintf("%s <file> <data>\n", os.Args[0])

func fatal(msg string) {
	fmt.Fprintf(os.Stderr, "%s", msg)
	os.Exit(1)
}

func main() {
	if len(os.Args) <= 1 ||
		len(os.Args) > 3 {
		fatal(banner)
	}

	var (
		fname = os.Args[1]
		in    io.Reader
	)

	if len(os.Args) == 2 {
		in = os.Stdin
	} else {
		in = bytes.NewBufferString(os.Args[2])
	}

	err := write(fname, in)
	if err != nil {
		fatal(err.Error())
	}
}
