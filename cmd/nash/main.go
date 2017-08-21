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
	"os/user"

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
	flag.BoolVar(&noInit, "noinit", false, "do not load init/init.sh file")
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
		fmt.Printf("build tag: %s\n", VersionString)
		return
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

	usr, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("unable to infer nash directory. "+
			"You must set NASHPATH environment variable explicity: ",
			err.Error())
	}
	if usr.HomeDir == "" {
		return "", fmt.Errorf("User %s doesn't have a home directory. "+
			"Set NASHPATH environment variable to desired location.", usr.Name)
	}
	if usr.HomeDir == "" {
		return "", fmt.Errorf("Unable to automatically infer NASHPATH, requires setting it explicity.")
	}

	return fmt.Sprintf("%s%c%s", usr.HomeDir, os.PathSeparator, ".nash"), nil
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
