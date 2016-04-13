// Package main has two sides:
// - User mode: shell
// - tool mode: unix socket server for handling namespace operations
// When started, the program choses their side based on the argv[0].
// The name "rc" indicates a user shell and the name -nrc- indidcates
// the namespace server tool.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/tiago4orion/cnt"
)

var (
	debug bool
	file  string
	addr  string
)

func init() {
	flag.BoolVar(&debug, "debug", false, "enable debug")

	if os.Args[0] == "-rcd-" || os.Args[1] == "-rcd" {
		flag.Bool("rcd", false, "force enable rcd mode")
		flag.StringVar(&addr, "addr", "", "rcd unix file")
	}
}

func main() {
	var err error

	flag.Parse()

	if addr != "" {
		startRcd(addr, debug)
	} else if file == "" {
		err = cli(debug)
	} else {
		err = cnt.Execute(file, debug)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err.Error())
		os.Exit(1)
	}
}
