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

func stringify(s string) string {
	return strings.Replace(strings.Replace(s, "\n", "\\n", -1),
		"\t", "\\t", -1)
}
