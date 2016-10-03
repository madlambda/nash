package sh

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"syscall"
	"time"

	"github.com/NeowayLabs/nash/sh"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func init() {
	rand.Seed(time.Now().UnixNano())
}

func randRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func buildenv(e Env) []string {
	env := make([]string, 0, len(e))

	for k, v := range e {
		if v == nil {
			continue
		}

		if v.Type() != sh.ListType &&
			v.Type() != sh.StringType {
			continue
		}

		if v.Type() == sh.ListType {
			vlist := v.(*sh.ListObj)
			env = append(env, k+"=("+vlist.String()+")")
		} else {
			vstr := v.(*sh.StrObj)
			env = append(env, k+"="+vstr.String())
		}
	}

	return env
}

func printVar(out io.Writer, name string, val sh.Obj) {
	if val.Type() == sh.StringType {
		valstr := val.(*sh.StrObj)
		fmt.Fprintf(out, "%s = \"%s\"\n", name, valstr.Str())
	} else if val.Type() == sh.ListType {
		vallist := val.(*sh.ListObj)
		fmt.Fprintf(out, "%s = (%s)\n", name, vallist.String())
	}
}

func printEnv(out io.Writer, name string) {
	fmt.Fprintf(out, "setenv %s\n", name)
}

func getErrStatus(err error, def string) string {
	status := def

	if exiterr, ok := err.(*exec.ExitError); ok {
		if statusObj, ok := exiterr.Sys().(syscall.WaitStatus); ok {
			status = strconv.Itoa(statusObj.ExitStatus())
		}
	}

	return status
}

func nashdAutoDiscover() string {
	path, err := os.Readlink("/proc/self/exe")

	if err != nil {
		path = os.Args[0]

		if _, err := os.Stat(path); err != nil {
			return ""
		}
	}

	return path
}
