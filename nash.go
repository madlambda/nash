// Package nash provides a library to embed the `nash` scripting language
// within your program or create your own nash cli.
package nash

import (
	"bytes"
	"fmt"
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

// ExecOutput executes the code specified by string content.
//
// It behaves like **Exec** with the exception that it will ignore any
// stdout parameter (and the default os.Stdout) and will return the
// whole stdout output in memory.
//
// This method has no side effects, it will preserve any previously
// setted stdout, it will only ignore the configured stdout to run
// the provided script content;
func (nash *Shell) ExecOutput(path, content string) ([]byte, error) {
	oldstdout := nash.Stdout()
	defer nash.SetStdout(oldstdout)

	var output bytes.Buffer
	nash.SetStdout(&output)

	err := nash.interp.Exec(path, content)
	return output.Bytes(), err
}

// ExecuteString executes the script content.
// Deprecated: Use Exec instead.
func (nash *Shell) ExecuteString(path, content string) error {
	return nash.interp.Exec(path, content)
}

// ExecFile executes the script content of the file specified by path
// and passes as arguments to the script the given args slice.
func (nash *Shell) ExecFile(path string, args ...string) error {
	if len(args) > 0 {
		err := nash.ExecuteString("setting args", `ARGS = `+args2Nash(args))
		if err != nil {
			return fmt.Errorf("Failed to set nash arguments: %s", err.Error())
		}
	}
	return nash.interp.ExecFile(path)
}

// ExecuteFile executes the given file.
// Deprecated: Use ExecFile instead.
func (nash *Shell) ExecuteFile(path string) error {
	return nash.interp.ExecFile(path)
}

// ExecuteTree executes the given tree.
// Deprecated: Use ExecTree instead.
func (nash *Shell) ExecuteTree(tr *ast.Tree) ([]sh.Obj, error) {
	return nash.interp.ExecuteTree(tr)
}

// ExecTree evaluates the given abstract syntax tree.
// it returns the object result of eval or nil when not applied and error.
func (nash *Shell) ExecTree(tree *ast.Tree) ([]sh.Obj, error) {
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

func args2Nash(args []string) string {
	ret := "("

	for i := 0; i < len(args); i++ {
		ret += `"` + args[i] + `"`

		if i < (len(args) - 1) {
			ret += " "
		}
	}

	return ret + ")"
}
