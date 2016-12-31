// Package nash provides a library to embed the `nash` scripting language
// within your program or create your own nash cli.
package nash

import (
	"io"

	"github.com/NeowayLabs/nash/ast"
	shell "github.com/NeowayLabs/nash/internal/sh"
	"github.com/NeowayLabs/nash/sh"
)

type (
	// Shell is the execution engine of the scripting language.
	Shell struct {
		interp *shell.Shell
	}
)

// New creates a new `nash.Shell` instance.
func New() (*Shell, error) {
	interp, err := shell.NewShell()

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

// SetInteractive enables interactive (shell) mode.
func (nash *Shell) SetInteractive(b bool) {
	nash.interp.SetInteractive(b)
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
func (nash *Shell) Environ() shell.Env {
	return nash.interp.Environ()
}

// GetFn gets the function object.
func (nash *Shell) GetFn(name string) (sh.Fn, bool) { return nash.interp.GetFn(name) }

// Prompt returns the environment prompt or the default one
func (nash *Shell) Prompt() string {
	value, ok := nash.interp.Getenv("PROMPT")

	if ok {
		return value.String()
	}

	return "<no prompt> "
}

// SetNashdPath sets an alternativa path to nashd
func (nash *Shell) SetNashdPath(path string) {
	nash.interp.SetNashdPath(path)
}

// Exec executes the code specified by string content.
// By default, nash uses os.Stdin, os.Stdout and os.Stderr as input, output
// and error file descriptors. You can change it with SetStdin, SetStdout and Stderr,
// respectively.
// The path is only used for error line reporting. If content represents a file, then
// setting path to this filename should improve debugging (or no).
func (nash *Shell) Exec(path, content string) error {
	return nash.interp.Exec(path, content)
}

// ExecuteString executes the script content.
// Deprecated: Use Exec instead.
func (nash *Shell) ExecuteString(path, content string) error {
	return nash.interp.Exec(path, content)
}

// ExecFile executes the script content of the file specified by path.
// See Exec for more information.
func (nash *Shell) ExecFile(path string) error {
	return nash.interp.ExecFile(path)
}

// ExecuteFile executes the given file.
// Deprecated: Use ExecFile instead.
func (nash *Shell) ExecuteFile(path string) error {
	return nash.interp.ExecFile(path)
}

// ExecuteTree executes the given tree.
// Deprecated: Use ExecTree instead.
func (nash *Shell) ExecuteTree(tr *ast.Tree) (sh.Obj, error) {
	return nash.interp.ExecuteTree(tr)
}

// ExecTree evaluates the given abstract syntax tree.
// it returns the object result of eval or nil when not applied and error.
func (nash *Shell) ExecTree(tree *ast.Tree) (sh.Obj, error) {
	return nash.interp.ExecuteTree(tree)
}

// SetStdout set the stdout of the nash engine.
func (nash *Shell) SetStdout(out io.Writer) {
	nash.interp.SetStdout(out)
}

// SetStderr set the stderr of nash engine
func (nash *Shell) SetStderr(err io.Writer) {
	nash.interp.SetStderr(err)
}

// SetStdin set the stdin of the nash engine
func (nash *Shell) SetStdin(in io.Reader) {
	nash.interp.SetStdin(in)
}

func (nash *Shell) Stdin() io.Reader  { return nash.interp.Stdin() }
func (nash *Shell) Stdout() io.Writer { return nash.interp.Stdout() }
func (nash *Shell) Stderr() io.Writer { return nash.interp.Stderr() }

// Setvar sets or updates the variable in the nash session
func (nash *Shell) Setvar(name string, value sh.Obj) {
	nash.interp.Setvar(name, value)
}

// Getvar retrieves a variable from nash session
func (nash *Shell) Getvar(name string) (sh.Obj, bool) {
	return nash.interp.Getvar(name)
}
