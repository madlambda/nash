// Package main has two sides:
// - User mode: shell
// - tool mode: unix socket server for handling namespace operations
// When started, the program choses their side based on the argv[0].
// The name "rc" indicates a user shell and the name -nrc- indicates
// the namespace server tool.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/NeowayLabs/nash"
)

var (
	// version is set at build time
	VersionString = "No version provided"

	version bool
	debug   bool
	file    string
	addr    string
)

func init() {
	flag.BoolVar(&version, "version", false, "Show version")
	flag.BoolVar(&debug, "debug", false, "enable debug")
	flag.StringVar(&file, "file", "", "script file")

	if os.Args[0] == "-nashd-" || (len(os.Args) > 1 && os.Args[1] == "-daemon") {
		flag.Bool("daemon", false, "force enable nashd mode")
		flag.StringVar(&addr, "addr", "", "rcd unix file")
	}
}

func main() {
	flag.Parse()

	if version {
		fmt.Printf("%s\n", VersionString)
		os.Exit(0)
	}

	shell := nash.NewShell(debug)

	home := os.Getenv("HOME")

	if home == "" {
		user := os.Getenv("USER")

		if user != "" {
			home = "/home/" + user
		} else {
			home = "/tmp"
		}
	}

	nashDir := home + "/.nash"
	shell.SetDotDir(nashDir)

	os.Mkdir(nashDir, 0755)

	initFile := home + "/.nash/init"

	if d, err := os.Stat(initFile); err == nil {
		if m := d.Mode(); !m.IsDir() {
			err = shell.ExecuteFile(initFile)

			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to evaluate '%s': %s\n", initFile, err.Error())
				os.Exit(1)
			}
		}
	}

	var err error

	if addr != "" {
		startNashd(shell, addr)
	} else if file == "" {
		err = cli(shell)
	} else {
		err = shell.ExecuteFile(file)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}
}
