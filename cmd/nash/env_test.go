package main_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/NeowayLabs/nash/cmd/nash"
)

// TODO: No idea on how to inject failures like empty HOME folders for now

func TestLoadNASHPATH(t *testing.T) {

	defaultNashPath := filepath.Join(home(t), "nash")

	runTests(t, main.NashPath, []EnvTest{
		{
			name: "Exported",
			env: map[string]string{
				"NASHPATH": filepath.Join("etc", "nash"),
			},
			want: filepath.Join("etc", "nash"),
		},
		{
			name: "IgnoresNASHROOT",
			env: map[string]string{
				"NASHROOT": "/etc/nashroot/tests",
				"HOME":     home(t),
			},
			want: defaultNashPath,
		},
		{
			name: "UseUserHomeWhenUnset",
			env: map[string]string{
				"NASHROOT": "/etc/nashroot/tests",
				"HOME":     home(t),
			},
			want: defaultNashPath,
		},
	})
}

func TestLoadNASHROOT(t *testing.T) {
	runTests(t, main.NashRoot, []EnvTest{
		{
			name: "Exported",
			env: map[string]string{
				"NASHROOT": filepath.Join("etc", "nashroot"),
			},
			want: filepath.Join("etc", "nashroot"),
		},
		{
			name: "IgnoresGOPATHIfSet",
			env: map[string]string{
				"GOPATH":   filepath.Join("go", "natel", "review"),
				"NASHROOT": filepath.Join("nashroot", "ignoredgopath"),
			},
			want: filepath.Join("nashroot", "ignoredgopath"),
		},
		{
			name: "UsesGOPATHIfUnset",
			env: map[string]string{
				"GOPATH": filepath.Join("go", "path"),
			},
			want: filepath.Join("go", "path", "src", "github.com", "NeowayLabs", "nash"),
		},
		{
			name: "UsesUserHomeWhenNASHROOTAndGOPATHAreUnset",
			env: map[string]string{
				"HOME": home(t),
			},
			want: filepath.Join(home(t), "nashroot"),
		},
	})
}

func runTests(t *testing.T, testfunc func() (string, error), cases []EnvTest) {

	t.Helper()

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			restore := clearenv(t)
			defer restore()

			export(t, c.env)
			got, err := testfunc()
			if err != nil {
				t.Fatal(err)
			}
			if got != c.want {
				t.Fatalf("got[%s] != want[%s]", got, c.want)
			}
		})
	}
}

type EnvTest struct {
	name string
	env  map[string]string
	want string
}

func clearenv(t *testing.T) func() {
	env := os.Environ()
	os.Clearenv()

	return func() {
		for _, envvar := range env {
			parsed := strings.Split(envvar, "=")
			name := parsed[0]
			val := strings.Join(parsed[1:], "=")

			err := os.Setenv(name, val)
			if err != nil {
				t.Fatalf("error[%s] restoring env var[%s]", err, envvar)
			}
		}
	}
}

func export(t *testing.T, env map[string]string) {
	t.Helper()

	for name, val := range env {
		err := os.Setenv(name, val)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func home(t *testing.T) string {
	t.Helper()

	homedir, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	return homedir
}
