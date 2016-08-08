package sh

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
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

		if v.Type() == StringType {
			env = append(env, k+"="+v.Str())
		} else if v.Type() == ListType {
			env = append(env, k+"=("+strings.Join(v.List(), " ")+")")
		}
	}

	return env
}

func printVar(out io.Writer, name string, val *Obj) {
	if val.Type() == StringType {
		fmt.Fprintf(out, "%s = \"%s\"\n", name, val.Str())
	} else if val.Type() == ListType {
		fmt.Fprintf(out, "%s = ", name)
		listStr := strings.Join(val.List(), ", ")
		fmt.Fprintf(out, "(\"%s\")\n", listStr)
	}
}

func printEnv(out io.Writer, name string) {
	fmt.Fprintf(out, "setenv %s\n", name)
}

func getErrStatus(err error, def int) string {
	status := def

	if exiterr, ok := err.(*exec.ExitError); ok {
		if statusObj, ok := exiterr.Sys().(syscall.WaitStatus); ok {
			status = statusObj.ExitStatus()
		}
	}

	return strconv.Itoa(status)
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
