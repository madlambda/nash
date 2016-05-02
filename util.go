package nash

import (
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
