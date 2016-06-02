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
	command string
	addr    string
	noInit  bool
)

func init() {
	flag.BoolVar(&version, "version", false, "Show version")
	flag.BoolVar(&debug, "debug", false, "enable debug")
	flag.StringVar(&file, "file", "", "script file")
	flag.BoolVar(&noInit, "noinit", false, "do not load init file")
	flag.StringVar(&command, "c", "", "command to execute")

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

	shell, err := nash.NewShell(debug)

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}

	nashpath := os.Getenv("NASHPATH")

	if nashpath == "" {
		home := os.Getenv("HOME")

		if home == "" {
			user := os.Getenv("USER")

			if user != "" {
				home = "/home/" + user
			} else {
				fmt.Fprintf(os.Stderr, "Environment variable NASHPATH or $USER must be set")
				os.Exit(1)
			}
		}

		nashpath = home + "/.nash"
	}

	shell.SetDotDir(nashpath)

	os.Mkdir(nashpath, 0755)

	initFile := nashpath + "/init"

	if d, err := os.Stat(initFile); err == nil && !noInit {
		if m := d.Mode(); !m.IsDir() {
			err = shell.ExecuteString("init", `import "`+initFile+`"`)

			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to evaluate '%s': %s\n", initFile, err.Error())
				os.Exit(1)
			}
		}
	}

	if addr != "" {
		startNashd(shell, addr)
	} else if command != "" {
		err = shell.ExecuteString("<argument -c>", command)
	} else if file != "" {
		err = shell.ExecuteFile(file)
	} else {
		err = cli(shell)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}
}
