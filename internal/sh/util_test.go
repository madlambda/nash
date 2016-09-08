package sh

import (
	"sort"
	"testing"
)

func TestBuildEnv(t *testing.T) {
	env := Env{
		"teste": nil,
	}

	penv := buildenv(env)

	if len(penv) != 0 {
		t.Errorf("Invalid env length")
		return
	}

	env = Env{
		"PATH": NewStrObj("/bin:/usr/bin"),
	}

	penv = buildenv(env)

	if len(penv) != 1 {
		t.Errorf("Invalid env length")
		return
	}

	if penv[0] != "PATH=/bin:/usr/bin" {
		t.Errorf("Invalid env value: %s", penv[0])
		return
	}

	env = Env{
		"PATH": NewListObj([]*Obj{
			NewStrObj("/bin"),
			NewStrObj("/usr/bin"),
		}),
	}

	penv = buildenv(env)

	if len(penv) != 1 {
		t.Errorf("Invalid env length")
		return
	}

	if penv[0] != "PATH=(/bin /usr/bin)" {
		t.Errorf("Invalid env value: %s", penv[0])
		return
	}

	env = Env{
		"PATH": NewListObj([]*Obj{
			NewStrObj("/bin"),
			NewStrObj("/usr/bin"),
		}),
		"path": NewStrObj("abracadabra"),
	}

	penv = buildenv(env)

	if len(penv) != 2 {
		t.Errorf("Invalid env length")
		return
	}

	sort.Strings(penv)

	if penv[0] != "PATH=(/bin /usr/bin)" {
		t.Errorf("Invalid env value: '%s'", penv[0])
		return
	}

	if penv[1] != "path=abracadabra" {
		t.Errorf("Invalid env value: '%s'", penv[1])
		return
	}
}
