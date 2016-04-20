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

	"github.com/tiago4orion/nash"
)

var (
	debug bool
	file  string
	addr  string
)

func init() {
	flag.BoolVar(&debug, "debug", false, "enable debug")
	flag.StringVar(&file, "file", "", "script file")

	if os.Args[0] == "-rcd-" || (len(os.Args) > 1 && os.Args[1] == "-rcd") {
		flag.Bool("rcd", false, "force enable rcd mode")
		flag.StringVar(&addr, "addr", "", "rcd unix file")
	}
}

func main() {
	var err error

	flag.Parse()

	shell := nash.NewShell(debug)

	if addr != "" {
		startRcd(shell, addr)
	} else if file == "" {
		err = cli(shell)
	} else {
		err = shell.Execute(file)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}
}
