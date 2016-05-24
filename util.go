package nash

import (
	"fmt"
	"io"
	"math/rand"
	"strings"
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
		vlen := len(v)
		if vlen == 0 {
			env = append(env, k+"=")
		} else if len(v) == 1 {
			env = append(env, k+"="+v[0])
		} else {
			env = append(env, k+"=("+strings.Join(v, " ")+")")
		}
	}

	return env
}

func printVar(out io.Writer, name string, val []string) {
	if len(val) == 0 {
		return
	}

	if len(val) == 1 {
		fmt.Fprintf(out, "%s = \"%s\"\n", name, val[0])
		return
	}

	fmt.Fprintf(out, "%s = ", name)
	listStr := strings.Join(val, ", ")
	fmt.Fprintf(out, "(\"%s\")\n", listStr)
}

func printEnv(out io.Writer, name string) {
	fmt.Fprintf(out, "setenv %s\n", name)
}
