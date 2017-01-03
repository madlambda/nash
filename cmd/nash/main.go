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
	var args  []string
	var shell *nash.Shell
	var err   error

	cliMode := false

	flag.Parse()

	if version {
		fmt.Printf("%s\n", VersionString)
		os.Exit(0)
	}

	if len(flag.Args()) > 0 {
		args = flag.Args()
		file = args[0]
	}

	if (file == "" && command == "") || interactive {
		cliMode = true
		shell, err = initShell()
	} else {
		shell, err = nash.New()
	}

	if err != nil {
		goto Error
	}

	shell.SetDebug(debug)

	if addr != "" {
		startNashd(shell, addr)

		return
	}

	if file != "" {
		err = executeFilename(shell, file, args)

		if err != nil {
			goto Error
		}
	}

	if command != "" {
		err = shell.ExecuteString("<argument -c>", command)

		if err != nil {
			goto Error
		}
	}

	if cliMode {
		err = cli(shell)

		if err != nil {
			goto Error
		}

		return
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

	// initShell will run only if the nash command is ran
	// without arguments or interactive flag, hence interactive mode
	shell.SetInteractive(true)

	nashpath, err := getnashpath()

	if err != nil {
		return nil, err
	}

	shell.SetDotDir(nashpath)

	os.Mkdir(nashpath, 0755)

	initFile := nashpath + "/init"

	if d, err := os.Stat(initFile); err == nil && !noInit {
		if m := d.Mode(); !m.IsDir() {
			err = shell.ExecuteString("init", `import "`+initFile+`"`)

			if err != nil {
				return nil, fmt.Errorf("Failed to evaluate '%s': %s\n", initFile, err.Error())
			}
		}
	}

	return shell, nil
}

func executeFilename(shell *nash.Shell, file string, args []string) error {
	err := shell.ExecuteString("setting args", `ARGS = `+args2Nash(args))

	if err != nil {
		err = fmt.Errorf("Failed to set nash arguments: %s", err.Error())

		return err
	}

	err = shell.ExecuteFile(file)

	if err != nil {
		return err
	}

	return nil
}

func args2Nash(args []string) string {
	ret := "("

	for i := 0; i < len(args); i++ {
		ret += `"` + args[i] + `"`

		if i < (len(args) - 1) {
			ret += " "
		}
	}

	return ret + ")"
}
