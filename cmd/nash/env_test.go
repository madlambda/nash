package main_test

import (
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"testing"
	
	"github.com/NeowayLabs/nash/cmd/nash"
)

// TODO: No idea on how to inject failures like empty HOME folders for now

func TestLoadNASHPATH(t *testing.T) {
	cases := []EnvTest{
		{
			name: "Exported",
			env: map[string]string {
				"NASHPATH": "/etc/nash/tests",
			},
			want: "/etc/nash/tests",
		},
		{
			name: "NotExported",
			want: filepath.Join(home(t), "nash"),
		},
	}
	
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			restore := clearenv(t)
			defer restore()
			
			export(t, c.env)
			got, err := main.NashPath()
			if err != nil {
				t.Fatal(err)
			}
			if got != c.want {
				t.Fatalf("got[%s] != want[%s]", got, c.want)
			}
		})
	}
}

func TestLoadNASHROOT(t * testing.T) {
}

type EnvTest struct {
	name string
	env map[string]string
	want string
}

func clearenv(t * testing.T) func() {
	env := os.Environ()
	os.Clearenv()
	
	return func() {
		for _, envvar := range env {
			parsed := strings.Split(envvar, "=")
			name := parsed[0]
			val := parsed[1]
			
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
	
	usr, err := user.Current()
	if err != nil {
		t.Fatal(err)
	}
	if usr.HomeDir == "" {
		t.Fatalf("current user[%v] has empty home", usr)
	}
	return usr.HomeDir
}