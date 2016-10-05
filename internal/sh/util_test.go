package sh

import (
	"sort"
	"testing"

	"github.com/NeowayLabs/nash/sh"
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
		"PATH": sh.NewStrObj("/bin:/usr/bin"),
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
		"PATH": sh.NewListObj([]sh.Obj{
			sh.NewStrObj("/bin"),
			sh.NewStrObj("/usr/bin"),
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
		"PATH": sh.NewListObj([]sh.Obj{
			sh.NewStrObj("/bin"),
			sh.NewStrObj("/usr/bin"),
		}),
		"path": sh.NewStrObj("abracadabra"),
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
