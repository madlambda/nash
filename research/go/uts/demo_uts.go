// This program is the Go version of namespace example program
// demo_uts.c of Linux Namespace series of articles:
// https://lwn.net/Articles/531381/
//
// Go has serious problems with fork, exec, and clone syscalls because
// of the runtime threads. The idea to implement the same behaviour of the
// C version was using os.Exec'ing the same program passing a special argument
// in the end of argument list. This idea was stolen of R. Minnich
// u-root project.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"
)

// Converts a C string into a Go string
func getstr(cstr [65]int8) string {
	b := make([]byte, 0, 65)

	for _, i := range cstr {
		if i == 0 {
			break
		}

		b = append(b, byte(i))
	}

	return string(b)
}

func main() {
	var err error

	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s <name-of-uts>\n", os.Args[0])
		fmt.Printf("%+v\n", os.Args)
		os.Exit(1)
	}

	a := os.Args

	if a[len(a)-1][0] != '#' {
		// This is the parent code
		a = append(a, "#")

		c := exec.Command(a[0], a[1:]...)

		c.SysProcAttr = &syscall.SysProcAttr{
			Cloneflags: syscall.CLONE_NEWUTS,
		}

		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr

		if err := c.Run(); err != nil {
			fmt.Printf(err.Error())
		}

		fmt.Printf("Process cloned successfully!\n")

		time.Sleep(1 * time.Second)

		uts := syscall.Utsname{}

		err = syscall.Uname(&uts)

		if err != nil {
			fmt.Printf(err.Error())
			os.Exit(1)
		}

		fmt.Printf("Parent uts.nodename = %s\n", getstr(uts.Nodename))

		os.Exit(0)
	} else {
		// Child namespace'd code
		err = syscall.Sethostname([]byte(a[1]))

		if err != nil {
			fmt.Printf("Error: %s\n", err.Error())
			os.Exit(1)
		}

		uts := syscall.Utsname{}

		err = syscall.Uname(&uts)

		if err != nil {
			fmt.Printf("Error: %s\n", err.Error())
			os.Exit(1)
		}

		fmt.Printf("Child uts.nodename = %s\n", getstr(uts.Nodename))

		// this holds the namespace open and allow other process to
		// join this namespace with setns
		time.Sleep(60 * time.Second)
	}

	os.Exit(0)
}
