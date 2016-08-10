// Package nash provides a library to embed the `nash` scripting language
// within your program or create your own nash cli.
package nash

import (
	"io"

	"github.com/NeowayLabs/nash/ast"
	"github.com/NeowayLabs/nash/internal/sh"
)

type (
	// Shell is the execution engine of the scripting language.
	Shell struct {
		interp *sh.Shell
	}
)

// New creates a new `nash.Shell` instance.
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

// SetDebug enable some logging for debug purposes.
func (nash *Shell) SetDebug(b bool) {
	nash.interp.SetDebug(b)
}

// SetDotDir sets the NASHPATH environment variable. The NASHPATH variable
// points to the location where nash will lookup for the init script and
// libraries installed.
func (nash *Shell) SetDotDir(path string) {
	obj := sh.NewStrObj(path)
	nash.interp.Setenv("NASHPATH", obj)
	nash.interp.Setvar("NASHPATH", obj)
}

// DotDir returns the value of the NASHPATH environment variable
func (nash *Shell) DotDir() string {
	if obj, ok := nash.interp.Getenv("NASHPATH"); ok {
		if obj.Type() != sh.StringType {
			return ""
		}

		return obj.String()
	}

	return ""
}

// Environ returns the set of environment variables in the shell
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

// SetNashdPath sets an alternativa path to nashd
func (sh *Shell) SetNashdPath(path string) {
	sh.interp.SetNashdPath(path)
}

// ExecuteString executes the code specified by string content.
// The `path` is only used for error line reporting.
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

func (nash *Shell) SetStdout(out io.Writer) {
	nash.interp.SetStdout(out)
}

func (nash *Shell) SetStderr(err io.Writer) {
	nash.interp.SetStderr(err)
}

func (nash *Shell) SetStdin(in io.Reader) {
	nash.interp.SetStdin(in)
}
