// Package main has two sides:
// - User mode: shell
// - tool mode: unix socket server for handling namespace operations
// When started, the program choses their side based on the argv[0].
// The name "nash" indicates a user shell and the name -nashd- indicates
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

	version     bool
	debug       bool
	file        string
	command     string
	addr        string
	noInit      bool
	interactive bool
)

func init() {
	flag.BoolVar(&version, "version", false, "Show version")
	flag.BoolVar(&debug, "debug", false, "enable debug")
	flag.BoolVar(&noInit, "noinit", false, "do not load init file")
	flag.StringVar(&command, "c", "", "command to execute")
	flag.BoolVar(&interactive, "i", false, "Interactive mode (default if no args)")

	if os.Args[0] == "-nashd-" || (len(os.Args) > 1 && os.Args[1] == "-daemon") {
		flag.Bool("daemon", false, "force enable nashd mode")
		flag.StringVar(&addr, "addr", "", "rcd unix file")
	}
}

func main() {
	var args []string
	var shell *nash.Shell
	var err error

	flag.Parse()

	if version {
		fmt.Printf("%s\n", VersionString)
		os.Exit(0)
	}

	if len(flag.Args()) > 0 {
		args = flag.Args()
		file = args[0]
	}

	if shell, err = initShell(); err != nil {
		goto Error
	}

	shell.SetDebug(debug)

	if addr != "" {
		startNashd(shell, addr)
		return
	}

	if (file == "" && command == "") || interactive {
		if err = cli(shell); err != nil {
			goto Error
		}

		return
	}

	if file != "" {
		if err = shell.ExecFile(file, args...); err != nil {
			goto Error
		}
	}

	if command != "" {
		err = shell.ExecuteString("<argument -c>", command)
		if err != nil {
			goto Error
		}
	}

Error:
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}
}

func getnashpath() (string, error) {
	nashpath := os.Getenv("NASHPATH")
	if nashpath != "" {
		return nashpath, nil
	}

	home := os.Getenv("HOME")
	if home == "" {
		user := os.Getenv("USER")

		if user != "" {
			home = "/home/" + user
		} else {
			return "", fmt.Errorf("Environment variable NASHPATH or $USER must be set")
		}
	}

	return home + "/.nash", nil
}

func initShell() (*nash.Shell, error) {
	shell, err := nash.New()
	if err != nil {
		return nil, err
	}

	nashpath, err := getnashpath()
	if err != nil {
		return nil, err
	}

	os.Mkdir(nashpath, 0755)
	shell.SetDotDir(nashpath)

	return shell, nil
}
