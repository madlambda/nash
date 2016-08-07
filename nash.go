package nash

import (
	"github.com/NeowayLabs/nash/internal/sh"
)

type (
	Shell struct {
		*sh.Shell
	}
)

func New() (*Shell, error) {
	sh, err := sh.NewShell()

	if err != nil {
		return nil, err
	}

	nash := Shell{
		Shell: sh,
	}

	return &nash, nil
}
