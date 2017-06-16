package sh

import (
	"io"

	"github.com/NeowayLabs/nash/ast"
	"github.com/NeowayLabs/nash/errors"
	"github.com/NeowayLabs/nash/internal/sh/builtin"
	"github.com/NeowayLabs/nash/sh"
)

type (
	fnDef struct {
		name     string
		Parent   *Shell
		Body     *ast.Tree
		argNames []sh.FnArg

		stdin          io.Reader
		stdout, stderr io.Writer
		environ        []string
	}

	userFnDef struct {
		*fnDef
	}

	builtinFnDef struct {
		*fnDef
		constructor builtin.Constructor
	}
)

// newFnDef creates a new function definition
func newFnDef(name string, parent *Shell, args []*ast.FnArgNode, body *ast.Tree) (*fnDef, error) {
	fn := fnDef{
		name:   name,
		Parent: parent,
		Body:   body,
		stdin:  parent.stdin,
		stdout: parent.stdout,
		stderr: parent.stderr,
	}

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if i < len(args)-1 && arg.IsVariadic {
			return nil, errors.NewEvalError(parent.filename,
				arg, "Vararg '%s' isn't the last argument",
				arg.String())
		}

		fn.argNames = append(fn.argNames, sh.FnArg{arg.Name, arg.IsVariadic})
	}
	return &fn, nil
}

func (fnDef *fnDef) Name() string         { return fnDef.name }
func (fnDef *fnDef) ArgNames() []sh.FnArg { return fnDef.argNames }
func (fnDef *fnDef) Environ() []string    { return fnDef.environ }

func (fnDef *fnDef) SetEnviron(env []string) {
	fnDef.environ = env
}

func (fnDef *fnDef) SetStdin(r io.Reader) {
	fnDef.stdin = r
}

func (fnDef *fnDef) SetStderr(w io.Writer) {
	fnDef.stderr = w
}

func (fnDef *fnDef) SetStdout(w io.Writer) {
	fnDef.stdout = w
}

func (fnDef *fnDef) Stdin() io.Reader  { return fnDef.stdin }
func (fnDef *fnDef) Stdout() io.Writer { return fnDef.stdout }
func (fnDef *fnDef) Stderr() io.Writer { return fnDef.stderr }

func newUserFnDef(name string, parent *Shell, args []*ast.FnArgNode, body *ast.Tree) (*userFnDef, error) {
	fnDef, err := newFnDef(name, parent, args, body)
	if err != nil {
		return nil, err
	}
	ufndef := userFnDef{
		fnDef: fnDef,
	}
	return &ufndef, nil
}

func (ufnDef *userFnDef) Build() sh.Fn {
	userfn := NewUserFn(ufnDef.Name(), ufnDef.ArgNames(), ufnDef.Body, ufnDef.Parent)
	userfn.SetStdin(ufnDef.stdin)
	userfn.SetStdout(ufnDef.stdout)
	userfn.SetStderr(ufnDef.stderr)
	userfn.SetEnviron(ufnDef.environ)
	return userfn
}

func newBuiltinFnDef(name string, parent *Shell, constructor builtin.Constructor) *builtinFnDef {
	return &builtinFnDef{
		fnDef: &fnDef{
			name:   name,
			stdin:  parent.stdin,
			stdout: parent.stdout,
			stderr: parent.stderr,
		},
		constructor: constructor,
	}
}

func (bfnDef *builtinFnDef) Build() sh.Fn {
	return NewBuiltinFn(bfnDef.Name(),
		bfnDef.constructor(),
		bfnDef.stdin,
		bfnDef.stdout,
		bfnDef.stderr,
	)
}
