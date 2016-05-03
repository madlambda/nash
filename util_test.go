package nash

import (
	"sort"
	"testing"
)

func TestBuildEnv(t *testing.T) {
	env := Env{
		"teste": []string{},
	}

	penv := buildenv(env)

	if len(penv) != 1 {
		t.Errorf("Invalid env length")
		return
	}

	if penv[0] != "teste=" {
		t.Errorf("Invalid env value: %s", penv[0])
		return
	}

	env = Env{
		"PATH": []string{
			"/bin:/usr/bin",
		},
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
		"PATH": []string{
			"/bin",
			"/usr/bin",
		},
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
		"PATH": []string{
			"/bin",
			"/usr/bin",
		},
		"path": []string{
			"abracadabra",
		},
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
