package main

import (
	"flag"
	"fmt"
	"io"
	"os"
)

var (
	help *bool
)

func init() {
	help = flag.Bool("h", false, "Show this help")
}

func usage() {
	flag.PrintDefaults()
}

func main() {
	flag.Parse()

	if *help {
		usage()
		os.Exit(0)
	}

	if err := do(flag.Args(), os.Stdout, os.Stderr); err != nil {
		fmt.Printf("error: %s\n", err.Error())
		os.Exit(1)
	}
}

func do(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		return idoc(stdout, stderr)
	}

	return doc(stdout, stderr, args)
}
