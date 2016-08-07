package nash

import (
	"github.com/NeowayLabs/nash/ast"
	"github.com/NeowayLabs/nash/internal/sh"
)

type (
	Shell struct {
		interp *sh.Shell
	}
)

func New() (*Shell, error) {
	interp, err := sh.NewShell()

	if err != nil {
		return nil, err
	}

	nash := Shell{
		interp: interp,
	}

	return &nash, nil
}

func (nash *Shell) SetDebug(b bool) {
	nash.interp.SetDebug(b)
}

func (nash *Shell) SetDotDir(path string) {
	obj := sh.NewStrObj(path)
	nash.interp.Setenv("NASHPATH", obj)
	nash.interp.Setvar("NASHPATH", obj)
}

func (nash *Shell) DotDir() string {
	if obj, ok := nash.interp.Getenv("NASHPATH"); ok {
		if obj.Type() != sh.StringType {
			return ""
		}

		return obj.String()
	}

	return ""
}

func (nash *Shell) Environ() sh.Env {
	return nash.interp.Environ()
}

// Prompt returns the environment prompt or the default one
func (nash *Shell) Prompt() string {
	value, ok := nash.interp.Getenv("PROMPT")

	if ok {
		return value.String()
	}

	return "<no prompt> "
}

// ExecuteString executes the commands specified by string content
func (nash *Shell) ExecuteString(path, content string) error {
	return nash.interp.ExecuteString(path, content)
}

// ExecuteFile execute the given path in the current shell environment
func (nash *Shell) ExecuteFile(path string) error {
	return nash.interp.ExecuteFile(path)
}

// ExecuteTree evaluates the given tree
func (nash *Shell) ExecuteTree(tr *ast.Tree) (*sh.Obj, error) {
	return nash.interp.ExecuteTree(tr)
}
