package sh

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"github.com/madlambda/nash/ast"
	"github.com/madlambda/nash/errors"
	"github.com/madlambda/nash/internal/sh/builtin"
	"github.com/madlambda/nash/parser"
	"github.com/madlambda/nash/sh"
	"github.com/madlambda/nash/token"
)

const (
	logNS     = "nashell.Shell"
	defPrompt = "\033[31mλ>\033[0m "
)

type (
	// Env is the environment map of lists
	Env map[string]sh.Obj
	Var Env
	Fns map[string]sh.FnDef

	StatusCode uint8

	// Shell is the core data structure.
	Shell struct {
		name        string
		debug       bool
		interactive bool
		abortOnErr  bool
		logf        LogFn
		nashdPath   string
		isFn        bool
		filename    string // current file being executed or imported

		sigs        chan os.Signal
		interrupted bool
		looping     bool

		stdin  io.Reader
		stdout io.Writer
		stderr io.Writer

		env   Env
		vars  Var
		binds Fns

		root   *ast.Tree
		parent *Shell

		repr string // string representation

		nashpath string
		nashroot string

		*sync.Mutex
	}

	errIgnore struct {
		*errors.NashError
	}

	errInterrupted struct {
		*errors.NashError
	}

	errStopWalking struct {
		*errors.NashError
	}
)

const (
	ESuccess    StatusCode = 0
	ENotFound              = 127
	ENotStarted            = 255
)

func newErrIgnore(format string, arg ...interface{}) error {
	e := &errIgnore{
		NashError: errors.NewError(format, arg...),
	}

	return e
}

func (e *errIgnore) Ignore() bool { return true }

func newErrInterrupted(format string, arg ...interface{}) error {
	return &errInterrupted{
		NashError: errors.NewError(format, arg...),
	}
}

func (e *errInterrupted) Interrupted() bool { return true }

func newErrStopWalking() *errStopWalking {
	return &errStopWalking{
		NashError: errors.NewError("return"),
	}
}

func (e *errStopWalking) StopWalking() bool { return true }

func NewAbortShell(nashpath string, nashroot string) (*Shell, error) {
	return newShell(nashpath, nashroot, true)
}

// NewShell creates a new shell object
// nashpath will be used to search libraries and nashroot will be used to
// search for the standard library shipped with the language.
func NewShell(nashpath string, nashroot string) (*Shell, error) {
	return newShell(nashpath, nashroot, false)
}

func newShell(nashpath string, nashroot string, abort bool) (*Shell, error) {
	shell := &Shell{
		name:        "parent scope",
		interactive: false,
		abortOnErr:  abort,
		isFn:        false,
		logf:        NewLog(logNS, false),
		nashdPath:   nashdAutoDiscover(),
		stdout:      os.Stdout,
		stderr:      os.Stderr,
		stdin:       os.Stdin,
		env:         make(Env),
		vars:        make(Var),
		binds:       make(Fns),
		Mutex:       &sync.Mutex{},
		sigs:        make(chan os.Signal, 1),
		filename:    "<interactive>",
		nashpath:    nashpath,
		nashroot:    nashroot,
	}

	err := shell.setup()
	if err != nil {
		return nil, err
	}

	shell.setupSignals()
	err = validateDirs(nashpath, nashroot)
	if err != nil {
		if shell.abortOnErr {
			return nil, err
		}

		printerr := func(msg string) {
			shell.Stderr().Write([]byte(msg + "\n"))
		}
		printerr(err.Error())
		printerr("please check your NASHPATH and NASHROOT so they point to valid locations")
	}

	return shell, nil
}

// NewSubShell creates a nashell.Shell that inherits the parent shell stdin,
// stdout, stderr and mutex lock.
// Every variable and function lookup is done first in the subshell and then, if
// not found, in the parent shell recursively.
func NewSubShell(name string, parent *Shell) *Shell {
	return &Shell{
		name:      name,
		isFn:      true,
		parent:    parent,
		logf:      NewLog(logNS, false),
		nashdPath: nashdAutoDiscover(),
		stdout:    parent.Stdout(),
		stderr:    parent.Stderr(),
		stdin:     parent.Stdin(),
		env:       make(Env),
		vars:      make(Var),
		binds:     make(Fns),
		Mutex:     parent.Mutex,
		filename:  parent.filename,
	}
}

func (shell *Shell) NashPath() string {
	return shell.nashpath
}

// initEnv creates a new environment from old one
func (shell *Shell) initEnv(processEnv []string) error {
	largs := make([]sh.Obj, len(os.Args))

	for i := 0; i < len(os.Args); i++ {
		largs[i] = sh.NewStrObj(os.Args[i])
	}

	argv := sh.NewListObj(largs)

	shell.Setenv("argv", argv)
	shell.Newvar("argv", argv)

	for _, penv := range processEnv {
		var value sh.Obj
		p := strings.Split(penv, "=")

		if len(p) >= 2 {
			// TODO(i4k): handle lists correctly in the future
			// argv is not special, every list must be handled correctly
			if p[0] == "argv" {
				continue
			}

			value = sh.NewStrObj(strings.Join(p[1:], "="))

			shell.Setenv(p[0], value)
			shell.Newvar(p[0], value)
		}
	}

	pidVal := sh.NewStrObj(strconv.Itoa(os.Getpid()))

	shell.Setenv("PID", pidVal)
	shell.Newvar("PID", pidVal)

	if _, ok := shell.Getenv("SHELL"); !ok {
		shellVal := sh.NewStrObj(nashdAutoDiscover())
		shell.Setenv("SHELL", shellVal)
		shell.Newvar("SHELL", shellVal)
	}

	cwd, err := os.Getwd()

	if err != nil {
		return err
	}

	cwdObj := sh.NewStrObj(cwd)
	shell.Setenv("PWD", cwdObj)
	shell.Newvar("PWD", cwdObj)

	return nil
}

// Reset internal state
func (shell *Shell) Reset() {
	shell.vars = make(Var)
	shell.env = make(Env)
	shell.binds = make(Fns)
}

// SetDebug enable/disable debug in the shell
func (shell *Shell) SetDebug(d bool) {
	shell.debug = d
	shell.logf = NewLog(logNS, d)
}

func (shell *Shell) Log(format string, args ...interface{}) {
	shell.logf(format, args...)
	// WHY: not using fmt.Sprintf to avoid formatting operation if
	// logging is disabled, but we always want newlines on the log calls.
	shell.logf("\n")
}

// SetInteractive enable/disable shell interactive mode
func (shell *Shell) SetInteractive(i bool) {
	shell.interactive = i

	if i {
		_ = shell.setupDefaultBindings()
	}
}

func (shell *Shell) Interactive() bool {
	if shell.parent != nil {
		return shell.parent.Interactive()
	}

	return shell.interactive
}

func (shell *Shell) SetName(a string) {
	shell.name = a
}

func (shell *Shell) Name() string { return shell.name }

func (shell *Shell) SetParent(a *Shell) {
	shell.parent = a
}

func (shell *Shell) Environ() Env {
	if shell.parent != nil {
		return shell.parent.Environ()
	}

	return shell.env
}

func (shell *Shell) Getenv(name string) (sh.Obj, bool) {
	if shell.parent != nil {
		return shell.parent.Getenv(name)
	}

	value, ok := shell.env[name]
	return value, ok
}

func (shell *Shell) Setenv(name string, value sh.Obj) {
	if shell.parent != nil {
		shell.parent.Setenv(name, value)
		return
	}

	shell.Newvar(name, value)

	shell.env[name] = value
	os.Setenv(name, value.String())
}

func (shell *Shell) SetEnviron(processEnv []string) {
	shell.env = make(Env)

	for _, penv := range processEnv {
		var value sh.Obj
		p := strings.Split(penv, "=")

		if len(p) == 2 {
			value = sh.NewStrObj(p[1])

			shell.Setenv(p[0], value)
			shell.Newvar(p[0], value)
		}
	}
}

// GetLocalvar returns a local scoped variable.
func (shell *Shell) GetLocalvar(name string) (sh.Obj, bool) {
	v, ok := shell.vars[name]
	return v, ok
}

// Getvar returns the value of the variable name. It could look in their
// parent scopes if not found locally.
func (shell *Shell) Getvar(name string) (sh.Obj, bool) {
	if value, ok := shell.vars[name]; ok {
		return value, ok
	}

	if shell.parent != nil {
		return shell.parent.Getvar(name)
	}

	return nil, false
}

// GetFn returns the function name or error if not found.
func (shell *Shell) GetFn(name string) (*sh.FnObj, error) {
	shell.logf("Looking for function '%s' on shell '%s'\n", name, shell.name)
	if obj, ok := shell.vars[name]; ok {
		if obj.Type() == sh.FnType {
			fnObj := obj.(*sh.FnObj)
			return fnObj, nil
		}
		return nil, errors.NewError("Identifier '%s' is not a function", name)
	}

	if shell.parent != nil {
		return shell.parent.GetFn(name)
	}

	return nil, fmt.Errorf("function '%s' not found", name)
}

func (shell *Shell) Setbindfn(name string, value sh.FnDef) {
	shell.binds[name] = value
}

func (shell *Shell) Getbindfn(cmdName string) (sh.FnDef, bool) {
	if fn, ok := shell.binds[cmdName]; ok {
		return fn, true
	}

	if shell.parent != nil {
		return shell.parent.Getbindfn(cmdName)
	}

	return nil, false
}

// Newvar creates a new variable in the scope.
func (shell *Shell) Newvar(name string, value sh.Obj) {
	shell.vars[name] = value
}

// Setvar updates the value of an existing variable. If the variable
// doesn't exist in current scope it looks up in their parent scopes.
// It returns true if the variable was found and updated.
func (shell *Shell) Setvar(name string, value sh.Obj) bool {
	_, ok := shell.vars[name]
	if ok {
		shell.vars[name] = value
		return true
	}

	if shell.parent != nil {
		return shell.parent.Setvar(name, value)
	}

	return false
}

func (shell *Shell) IsFn() bool     { return shell.isFn }
func (shell *Shell) SetIsFn(b bool) { shell.isFn = b }

// SetNashdPath sets an alternativa path to nashd
func (shell *Shell) SetNashdPath(path string) {
	shell.nashdPath = path
}

// SetStdin sets the stdin for commands
func (shell *Shell) SetStdin(in io.Reader) {
	shell.stdin = in
}

// SetStdout sets stdout for commands
func (shell *Shell) SetStdout(out io.Writer) {
	shell.stdout = out
}

// SetStderr sets stderr for commands
func (shell *Shell) SetStderr(err io.Writer) {
	shell.stderr = err
}

func (shell *Shell) Stdout() io.Writer { return shell.stdout }
func (shell *Shell) Stderr() io.Writer { return shell.stderr }
func (shell *Shell) Stdin() io.Reader  { return shell.stdin }

// SetTree sets the internal tree of the interpreter. It's used for
// sub-shells like `fn`.
func (shell *Shell) SetTree(t *ast.Tree) {
	shell.root = t
}

// Tree returns the internal tree of the subshell.
func (shell *Shell) Tree() *ast.Tree { return shell.root }

// SetRepr set the string representation of the shell
func (shell *Shell) SetRepr(a string) {
	shell.repr = a
}

func (shell *Shell) setupBuiltin() {
	for name, constructor := range builtin.Constructors() {
		fnDef := newBuiltinFnDef(name, shell, constructor)
		shell.Newvar(name, sh.NewFnObj(fnDef))
	}
}

func (shell *Shell) setupDefaultBindings() error {
	// only one builtin fn... no need for advanced machinery yet
	homeEnvVar := "HOME"
	if runtime.GOOS == "windows" {
		homeEnvVar = "HOMEPATH"
	}
	err := shell.Exec(shell.name, fmt.Sprintf(`fn nash_builtin_cd(args...) {
	    var path = ""
	    var l <= len($args)
            if $l == "0" {
                    path = $%s
            } else {
                    path = $args[0]
	    }

            chdir($path)
        }

        bindfn nash_builtin_cd cd`, homeEnvVar))

	return err
}

func (shell *Shell) setup() error {
	err := shell.initEnv(os.Environ())
	if err != nil {
		return err
	}

	if shell.env["PROMPT"] == nil {
		pobj := sh.NewStrObj(defPrompt)
		shell.Setenv("PROMPT", pobj)
		shell.Newvar("PROMPT", pobj)
	}

	_, ok := shell.Getvar("_")
	if !ok {
		shell.Newvar("_", sh.NewStrObj(""))
	}

	shell.setupBuiltin()
	return err
}

func (shell *Shell) setupSignals() {
	signal.Notify(shell.sigs, syscall.SIGINT)

	go func() {
		for {
			sig := <-shell.sigs

			switch sig {
			case syscall.SIGINT:
				shell.Lock()

				// TODO(i4k): Review implementation when interrupted inside
				// function loops
				if shell.looping {
					shell.setIntr(true)
				}

				shell.Unlock()
			default:
				fmt.Printf("%s\n", sig)
			}
		}
	}()
}

// TriggerCTRLC mock the user pressing CTRL-C in the terminal
func (shell *Shell) TriggerCTRLC() error {
	p, err := os.FindProcess(os.Getpid())
	if err != nil {
		return err
	}

	return p.Signal(syscall.SIGINT)
}

// setIntr *do not lock*. You must do it yourself!
func (shell *Shell) setIntr(b bool) {
	if shell.parent != nil {
		shell.parent.setIntr(b)
		return
	}

	shell.interrupted = b
}

// getIntr returns true if nash was interrupted by CTRL-C
func (shell *Shell) getIntr() bool {
	if shell.parent != nil {
		return shell.parent.getIntr()
	}

	return shell.interrupted
}

// Exec executes the commands specified by string content
func (shell *Shell) Exec(path, content string) error {
	p := parser.NewParser(path, content)

	tr, err := p.Parse()

	if err != nil {
		return err
	}

	_, err = shell.ExecuteTree(tr)
	return err
}

// Execute the nash file at given path
func (shell *Shell) ExecFile(path string) error {
	bkCurFile := shell.filename

	content, err := ioutil.ReadFile(path)

	if err != nil {
		return err
	}

	shell.filename = path

	defer func() {
		shell.filename = bkCurFile
	}()

	return shell.Exec(path, string(content))
}

func (shell *Shell) newvar(name *ast.NameNode, value sh.Obj) error {
	if name.Index == nil {
		shell.Newvar(name.Ident, value)
		return nil
	}

	// handles ident[x] = v

	obj, ok := shell.Getvar(name.Ident)
	if !ok {
		return errors.NewEvalError(shell.filename,
			name, "Variable %s not found", name.Ident)
	}

	index, err := shell.evalIndex(name.Index)
	if err != nil {
		return err
	}

	col, err := sh.NewWriteableCollection(obj)
	if err != nil {
		return errors.NewEvalError(shell.filename, name, err.Error())
	}

	err = col.Set(index, value)
	if err != nil {
		return errors.NewEvalError(
			shell.filename,
			name,
			"error[%s] setting var",
			err,
		)
	}

	shell.Newvar(name.Ident, obj)
	return nil
}

func (shell *Shell) setvar(name *ast.NameNode, value sh.Obj) error {
	if name.Index == nil {
		if !shell.Setvar(name.Ident, value) {
			return errors.NewEvalError(shell.filename,
				name, "Variable '%s' is not initialized. Use 'var %s = <value>'",
				name, name)
		}
		return nil
	}

	obj, ok := shell.Getvar(name.Ident)
	if !ok {
		return errors.NewEvalError(shell.filename,
			name, "Variable %s not found", name.Ident)
	}

	index, err := shell.evalIndex(name.Index)
	if err != nil {
		return err
	}

	col, err := sh.NewWriteableCollection(obj)
	if err != nil {
		return errors.NewEvalError(shell.filename, name, err.Error())
	}

	err = col.Set(index, value)
	if err != nil {
		return errors.NewEvalError(
			shell.filename,
			name,
			"error[%s] setting var",
			err,
		)
	}

	if !shell.Setvar(name.Ident, obj) {
		return errors.NewEvalError(shell.filename,
			name, "Variable '%s' is not initialized. Use 'var %s = <value>'",
			name, name)
	}
	return nil
}

func (shell *Shell) setvars(names []*ast.NameNode, values []sh.Obj) error {
	for i := 0; i < len(names); i++ {
		err := shell.setvar(names[i], values[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func (shell *Shell) newvars(names []*ast.NameNode, values []sh.Obj) {
	for i := 0; i < len(names); i++ {
		shell.newvar(names[i], values[i])
	}
}

func (shell *Shell) setcmdvars(names []*ast.NameNode, stdout, stderr, status sh.Obj) error {
	if len(names) == 3 {
		err := shell.setvar(names[0], stdout)
		if err != nil {
			return err
		}
		err = shell.setvar(names[1], stderr)
		if err != nil {
			return err
		}
		return shell.setvar(names[2], status)
	} else if len(names) == 2 {
		err := shell.setvar(names[0], stdout)
		if err != nil {
			return err
		}
		return shell.setvar(names[1], status)
	} else if len(names) == 1 {
		return shell.setvar(names[0], stdout)
	}

	panic(fmt.Sprintf("internal error: expects 1 <= len(names) <= 3,"+
		" but got %d",
		len(names)))

	return nil
}

func (shell *Shell) newcmdvars(names []*ast.NameNode, stdout, stderr, status sh.Obj) {
	if len(names) == 3 {
		shell.newvar(names[0], stdout)
		shell.newvar(names[1], stderr)
		shell.newvar(names[2], status)
	} else if len(names) == 2 {
		shell.newvar(names[0], stdout)
		shell.newvar(names[1], status)
	} else if len(names) == 1 {
		shell.newvar(names[0], stdout)
	} else {
		panic(fmt.Sprintf("internal error: expects 1 <= len(names) <= 3,"+
			" but got %d",
			len(names)))
	}
}

// evalConcat reveives the AST representation of a concatenation of objects and
// returns the string representation, or error.
func (shell *Shell) evalConcat(path ast.Expr) (string, error) {
	var pathStr string

	if path.Type() != ast.NodeConcatExpr {
		return "", fmt.Errorf("Invalid node %+v", path)
	}

	concatExpr := path.(*ast.ConcatExpr)
	concat := concatExpr.List()

	for i := 0; i < len(concat); i++ {
		part := concat[i]

		switch part.Type() {
		case ast.NodeConcatExpr:
			return "", errors.NewEvalError(shell.filename, part,
				"Nested concat is not allowed: %s", part)
		case ast.NodeVarExpr, ast.NodeIndexExpr:
			partValue, err := shell.evalVariable(part)
			if err != nil {
				return "", err
			}

			if partValue.Type() == sh.ListType {
				return "", errors.NewEvalError(shell.filename,
					part, "Concat of list variables is not allowed: %v = %v",
					part, partValue)
			} else if partValue.Type() != sh.StringType {
				return "", errors.NewEvalError(shell.filename, part,
					"Invalid concat element: %v", partValue)
			}

			strval := partValue.(*sh.StrObj)
			pathStr += strval.Str()
		case ast.NodeStringExpr:
			str, ok := part.(*ast.StringExpr)
			if !ok {
				return "", errors.NewEvalError(shell.filename, part,
					"Failed to eval string: %s", part)
			}

			pathStr += str.Value()
		case ast.NodeFnInv:
			fnNode := part.(*ast.FnInvNode)
			result, err := shell.executeFnInv(fnNode)
			if err != nil {
				return "", err
			}

			if len(result) == 0 || len(result) > 1 {
				return "", errors.NewEvalError(shell.filename, part,
					"Function '%s' used in string concat but returns %d values.",
					fnNode.Name)
			}
			obj := result[0]
			if obj.Type() != sh.StringType {
				return "", errors.NewEvalError(shell.filename, part,
					"Function '%s' used in concat but returns a '%s'", obj.Type())
			}

			str := obj.(*sh.StrObj)
			pathStr += str.Str()
		case ast.NodeListExpr:
			return "", errors.NewEvalError(shell.filename, part,
				"Concat of lists is not allowed: %+v", part.String())
		default:
			return "", errors.NewEvalError(shell.filename, part,
				"Invalid argument: %+v", part)
		}
	}

	return pathStr, nil
}

func (shell *Shell) executeNode(node ast.Node) ([]sh.Obj, error) {
	var (
		objs []sh.Obj
		err  error
	)

	shell.logf("Executing node: %v\n", node)

	switch node.Type() {
	case ast.NodeImport:
		err = shell.executeImport(node.(*ast.ImportNode))
	case ast.NodeComment:
		// ignore
	case ast.NodeSetenv:
		err = shell.executeSetenv(node.(*ast.SetenvNode))
	case ast.NodeVarAssignDecl:
		err = shell.executeVarAssign(node.(*ast.VarAssignDeclNode))
	case ast.NodeVarExecAssignDecl:
		err = shell.executeVarExecAssign(node.(*ast.VarExecAssignDeclNode))
	case ast.NodeAssign:
		err = shell.executeAssignment(node.(*ast.AssignNode))
	case ast.NodeExecAssign:
		err = shell.executeExecAssign(node.(*ast.ExecAssignNode))
	case ast.NodeCommand:
		_, err = shell.executeCommand(node.(*ast.CommandNode))
	case ast.NodePipe:
		_, err = shell.executePipe(node.(*ast.PipeNode))
	case ast.NodeRfork:
		err = shell.executeRfork(node.(*ast.RforkNode))
	case ast.NodeIf:
		objs, err = shell.executeIf(node.(*ast.IfNode))
	case ast.NodeFnDecl:
		err = shell.executeFnDecl(node.(*ast.FnDeclNode))
	case ast.NodeFnInv:
		// invocation ignoring output
		_, err = shell.executeFnInv(node.(*ast.FnInvNode))
	case ast.NodeFor:
		objs, err = shell.executeFor(node.(*ast.ForNode))
	case ast.NodeBindFn:
		err = shell.executeBindFn(node.(*ast.BindFnNode))
	case ast.NodeReturn:
		if shell.IsFn() {
			objs, err = shell.executeReturn(node.(*ast.ReturnNode))
		} else {
			err = errors.NewEvalError(shell.filename,
				node,
				"Unexpected return outside of function declaration.")
		}
	default:
		// should never get here
		return nil, errors.NewEvalError(shell.filename, node,
			"invalid node: %v.", node.Type())
	}

	return objs, err
}

func (shell *Shell) ExecuteTree(tr *ast.Tree) ([]sh.Obj, error) {
	return shell.executeTree(tr, true)
}

// executeTree evaluates the given tree
func (shell *Shell) executeTree(tr *ast.Tree, stopable bool) ([]sh.Obj, error) {
	if tr == nil || tr.Root == nil {
		return nil, errors.NewError("empty abstract syntax tree to execute")
	}

	root := tr.Root

	for _, node := range root.Nodes {
		objs, err := shell.executeNode(node)
		if err != nil {
			type (
				IgnoreError interface {
					Ignore() bool
				}

				InterruptedError interface {
					Interrupted() bool
				}

				StopWalkingError interface {
					StopWalking() bool
				}
			)

			if errIgnore, ok := err.(IgnoreError); ok && errIgnore.Ignore() {
				continue
			}

			if errInterrupted, ok := err.(InterruptedError); ok && errInterrupted.Interrupted() {
				return nil, err
			}

			if errStopWalking, ok := err.(StopWalkingError); stopable && ok && errStopWalking.StopWalking() {
				return objs, nil
			}

			return objs, err
		}
	}

	return nil, nil
}

func (shell *Shell) executeReturn(n *ast.ReturnNode) ([]sh.Obj, error) {
	var returns []sh.Obj

	returnExprs := n.Returns

	for i := 0; i < len(returnExprs); i++ {
		retExpr := returnExprs[i]

		obj, err := shell.evalExpr(retExpr)
		if err != nil {
			return nil, err
		}

		returns = append(returns, obj)
	}

	return returns, newErrStopWalking()
}

func (shell *Shell) getNashRootFromGOPATH(preverr error) (string, error) {
	g, hasgopath := shell.Getenv("GOPATH")
	if !hasgopath {
		return "", errors.NewError("%s\nno GOPATH env var setted", preverr)
	}
	gopath := g.String()
	return filepath.Join(gopath, filepath.FromSlash("/src/github.com/madlambda/nash")), nil
}

func isValidNashRoot(nashroot string) bool {
	_, err := os.Stat(filepath.Join(nashroot, "stdlib"))
	return err == nil
}

func (shell *Shell) executeImport(node *ast.ImportNode) error {
	obj, err := shell.evalExpr(node.Path)
	if err != nil {
		return errors.NewEvalError(shell.filename,
			node, err.Error())
	}

	if obj.Type() != sh.StringType {
		return errors.NewEvalError(shell.filename,
			node.Path,
			"Invalid type on import argument: %s", obj.Type())
	}

	objstr := obj.(*sh.StrObj)
	fname := objstr.Str()

	shell.logf("Importing '%s'", fname)

	var (
		tries  []string
		hasExt bool
	)

	hasExt = filepath.Ext(fname) != ""
	if filepath.IsAbs(fname) {
		tries = append(tries, fname)

		if !hasExt {
			tries = append(tries, fname+".sh")
		}
	}

	if shell.filename != "" {
		localFile := filepath.Join(filepath.Dir(shell.filename), fname)
		tries = append(tries, localFile)

		if !hasExt {
			tries = append(tries, localFile+".sh")
		}
	}

	tries = append(tries, filepath.Join(shell.nashpath, "lib", fname))
	if !hasExt {
		tries = append(tries, filepath.Join(shell.nashpath, "lib", fname+".sh"))
	}

	tries = append(tries, filepath.Join(shell.nashroot, "stdlib", fname+".sh"))

	shell.logf("Trying %q\n", tries)

	for _, path := range tries {
		d, err := os.Stat(path)

		if err != nil {
			continue
		}

		if m := d.Mode(); !m.IsDir() {
			return shell.ExecFile(path)
		}
	}

	errmsg := fmt.Sprintf(
		"Failed to import path '%s'. The locations below have been tried:\n \"%s\"",
		fname,
		strings.Join(tries, `", "`),
	)

	return errors.NewEvalError(shell.filename, node, errmsg)
}

// executePipe executes a pipe of ast.Command's. Each command can be
// a path command in the operating system or a function bind to a
// command name.
// The error of each command can be suppressed prepending it with '-' (dash).
// The error returned will be a string representing the errors (or none) of
// each command separated by '|'. The $status of pipe execution will be
// the $status of each command separated by '|'.
func (shell *Shell) executePipe(pipe *ast.PipeNode) (sh.Obj, error) {
	var (
		closeFiles     []io.Closer
		closeAfterWait []io.Closer
		errIndex       int
		err            error
	)

	defer func() {
		for _, c := range closeAfterWait {
			c.Close()
		}
	}()

	nodeCommands := pipe.Commands()

	if len(nodeCommands) < 2 {
		return sh.NewStrObj(strconv.Itoa(ENotStarted)),
			errors.NewEvalError(shell.filename,
				pipe, "Pipe requires at least two commands.")
	}

	cmds := make([]sh.Runner, len(nodeCommands))
	errs := make([]string, len(nodeCommands))
	igns := make([]bool, len(nodeCommands)) // ignoreErrors
	cods := make([]string, len(nodeCommands))

	for i := 0; i < len(nodeCommands); i++ {
		errs[i] = "not started"
		cods[i] = strconv.Itoa(ENotStarted)
	}

	last := len(nodeCommands) - 1

	envVars := buildenv(shell.Environ())

	// Create all commands
	for i := 0; i < len(nodeCommands); i++ {
		var (
			cmd    sh.Runner
			ignore bool
			args   []sh.Obj
		)

		nodeCmd := nodeCommands[i]

		cmd, ignore, err = shell.getCommand(nodeCmd)

		igns[i] = ignore

		if err != nil {
			errIndex = i
			cods[i] = strconv.Itoa(ENotFound)
			goto pipeError
		}

		// SetEnviron must be called before SetArgs
		// otherwise the subshell will have the arguments
		// shadowed by parent env
		cmd.SetEnviron(envVars)
		args, err = shell.evalExprs(nodeCmd.Args())

		if err != nil {
			errIndex = i
			goto pipeError
		}

		err = cmd.SetArgs(args)
		if err != nil {
			errIndex = i
			goto pipeError
		}

		cmd.SetStdin(shell.stdin)

		if i < last {
			closeFiles, err = shell.setRedirects(cmd, nodeCmd.Redirects())
			closeAfterWait = append(closeAfterWait, closeFiles...)

			if err != nil {
				errIndex = i
				goto pipeError
			}
		}

		cmds[i] = cmd
	}

	// Shell does not support stdin redirection yet
	cmds[0].SetStdin(shell.stdin)

	// Setup the commands. Pointing the stdin of next command to stdout of previous.
	// Except the stdout of last one
	for i, cmd := range cmds[:last] {
		var (
			stdin io.ReadCloser
		)

		// StdoutPipe complains if Stdout is already set
		cmd.SetStdout(nil)
		stdin, err = cmd.StdoutPipe()

		if err != nil {
			errIndex = i
			goto pipeError
		}

		cmds[i+1].SetStdin(stdin)
	}

	cmds[last].SetStdout(shell.stdout)
	cmds[last].SetStderr(shell.stderr)

	closeFiles, err = shell.setRedirects(cmds[last], nodeCommands[last].Redirects())
	closeAfterWait = append(closeAfterWait, closeFiles...)

	if err != nil {
		errIndex = last
		goto pipeError
	}

	for i := 0; i < len(cmds); i++ {
		cmd := cmds[i]

		err = cmd.Start()

		if err != nil {
			errIndex = i
			goto pipeError
		}

		errs[i] = "success"
		cods[i] = "0"
	}

	for i, cmd := range cmds {
		err = cmd.Wait()

		if err != nil {
			errIndex = i
			goto pipeError
		}

		errs[i] = "success"
		cods[i] = "0"
	}

	return sh.NewStrObj("0"), nil

pipeError:
	if igns[errIndex] {
		errs[errIndex] = "none"
	} else {
		errs[errIndex] = err.Error()
	}

	cods[errIndex] = getErrStatus(err, cods[errIndex])

	err = errors.NewEvalError(shell.filename,
		pipe, strings.Join(errs, "|"))

	// verify if all status codes are the same
	uniqCodes := make(map[string]struct{})
	var uniqCode string

	for i := 0; i < len(cods); i++ {
		uniqCodes[cods[i]] = struct{}{}
		uniqCode = cods[i]
	}

	var status sh.Obj

	if len(uniqCodes) == 1 {
		// if all status are the same
		status = sh.NewStrObj(uniqCode)
	} else {
		status = sh.NewStrObj(strings.Join(cods, "|"))
	}

	if igns[errIndex] {
		return status, nil
	}

	return status, err
}

func (shell *Shell) openRedirectLocation(location ast.Expr) (io.WriteCloser, error) {
	var protocol string

	locationObj, err := shell.evalExpr(location)
	if err != nil {
		return nil, err
	}

	if locationObj.Type() != sh.StringType {
		return nil, errors.NewEvalError(shell.filename,
			location,
			"Redirection to invalid object type: %v (%s)", locationObj, locationObj.Type())
	}

	objstr := locationObj.(*sh.StrObj)
	locationStr := objstr.Str()

	if len(locationStr) > 6 {
		if locationStr[0:6] == "tcp://" {
			protocol = "tcp"
		} else if locationStr[0:6] == "udp://" {
			protocol = "udp"
		} else if len(locationStr) > 7 && locationStr[0:7] == "unix://" {
			protocol = "unix"
		}
	}

	if protocol == "" {
		return os.OpenFile(locationStr, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	}

	switch protocol {
	case "tcp", "udp":
		netParts := strings.Split(locationStr[6:], ":")

		if len(netParts) != 2 {
			return nil, errors.NewEvalError(shell.filename,
				location,
				"Invalid tcp/udp address: %s", locationStr)
		}

		url := netParts[0] + ":" + netParts[1]

		return net.Dial(protocol, url)
	case "unix":
		return net.Dial(protocol, locationStr[7:])
	}

	return nil, errors.NewEvalError(shell.filename, location,
		"Unexpected redirection value: %s", locationStr)
}

func (shell *Shell) setRedirects(cmd sh.Runner, redirDecls []*ast.RedirectNode) ([]io.Closer, error) {
	var closeAfterWait []io.Closer

	for _, r := range redirDecls {
		closeFiles, err := shell.buildRedirect(cmd, r)
		closeAfterWait = append(closeAfterWait, closeFiles...)

		if err != nil {
			return closeAfterWait, err
		}
	}

	return closeAfterWait, nil
}

func (shell *Shell) buildRedirect(cmd sh.Runner, redirDecl *ast.RedirectNode) ([]io.Closer, error) {
	var closeAfterWait []io.Closer

	if redirDecl.LeftFD() > 2 || redirDecl.LeftFD() < ast.RedirMapSupress {
		return closeAfterWait, errors.NewEvalError(shell.filename,
			redirDecl,
			"Invalid file descriptor redirection: fd=%d", redirDecl.LeftFD())
	}

	if redirDecl.RightFD() > 2 || redirDecl.RightFD() < ast.RedirMapSupress {
		return closeAfterWait, errors.NewEvalError(shell.filename,
			redirDecl,
			"Invalid file descriptor redirection: fd=%d", redirDecl.RightFD())
	}

	var err error

	// Note(i4k): We need to remove the repetitive code in some smarter way
	switch redirDecl.LeftFD() {
	case 0:
		return closeAfterWait, fmt.Errorf("Does not support stdin redirection yet")
	case 1:
		switch redirDecl.RightFD() {
		case 0:
			return closeAfterWait, errors.NewEvalError(shell.filename,
				redirDecl,
				"Invalid redirect mapping: %d -> %d", 1, 0)
		case 1: // do nothing
		case 2:
			cmd.SetStdout(cmd.Stderr())
		case ast.RedirMapNoValue:
			if redirDecl.Location() == nil {
				return closeAfterWait, errors.NewEvalError(shell.filename,
					redirDecl,
					"Missing file in redirection: >[%d] <??>", redirDecl.LeftFD())
			}

			file, err := shell.openRedirectLocation(redirDecl.Location())
			if err != nil {
				return closeAfterWait, err
			}

			cmd.SetStdout(file)
			closeAfterWait = append(closeAfterWait, file)
		case ast.RedirMapSupress:
			file := ioutil.Discard
			cmd.SetStdout(file)
		}
	case 2:
		switch redirDecl.RightFD() {
		case 0:
			return closeAfterWait, errors.NewEvalError(shell.filename,
				redirDecl, "Invalid redirect mapping: %d -> %d", 2, 1)
		case 1:
			cmd.SetStderr(cmd.Stdout())
		case 2: // do nothing
		case ast.RedirMapNoValue:
			if redirDecl.Location() == nil {
				return closeAfterWait, errors.NewEvalError(shell.filename,
					redirDecl,
					"Missing file in redirection: >[%d] <??>", redirDecl.LeftFD())
			}

			file, err := shell.openRedirectLocation(redirDecl.Location())
			if err != nil {
				return closeAfterWait, err
			}

			cmd.SetStderr(file)
			closeAfterWait = append(closeAfterWait, file)
		case ast.RedirMapSupress:
			cmd.SetStderr(ioutil.Discard)
		}
	case ast.RedirMapNoValue:
		if redirDecl.Location() == nil {
			return closeAfterWait, errors.NewEvalError(shell.filename,
				redirDecl, "Missing file in redirection: >[%d] <??>", redirDecl.LeftFD())
		}

		file, err := shell.openRedirectLocation(redirDecl.Location())
		if err != nil {
			return closeAfterWait, err
		}

		cmd.SetStdout(file)
		closeAfterWait = append(closeAfterWait, file)
	}

	return closeAfterWait, err
}

func (shell *Shell) newBindfnRunner(
	c *ast.CommandNode,
	cmdName string,
	fnDef sh.FnDef,
) (sh.Runner, error) {
	shell.logf("Executing bind %s", cmdName)
	shell.logf("%s bind to %s", cmdName, fnDef.Name())

	if !shell.Interactive() {
		err := errors.NewEvalError(shell.filename,
			c, "'%s' is a bind to '%s'. "+
				"No binds allowed in non-interactive mode.",
			cmdName,
			fnDef.Name())
		return nil, err
	}

	return fnDef.Build(), nil
}

func (shell *Shell) getCommand(c *ast.CommandNode) (sh.Runner, bool, error) {
	var (
		ignoreError bool
		cmd         sh.Runner
		err         error
	)

	cmdName := c.Name()

	shell.logf("Executing: %s\n", c.Name())

	if len(cmdName) > 1 && cmdName[0] == '-' {
		ignoreError = true
		cmdName = cmdName[1:]

		shell.logf("Ignoring error\n")
	}

	if cmdName == "" {
		return nil, false, errors.NewEvalError(shell.filename,
			c, "Empty command name...")
	}

	if fnDef, ok := shell.Getbindfn(cmdName); ok {
		runner, err := shell.newBindfnRunner(c, cmdName, fnDef)
		return runner, ignoreError, err
	}

	cmd, err = NewCmd(cmdName)

	if err != nil {
		type NotFound interface {
			NotFound() bool
		}

		shell.logf("Command fails: %s", err.Error())

		if errNotFound, ok := err.(NotFound); ok && errNotFound.NotFound() {
			return nil, ignoreError, err
		}

		return nil, ignoreError, err
	}

	cmd.SetStdin(shell.stdin)
	cmd.SetStdout(shell.stdout)
	cmd.SetStderr(shell.stderr)

	return cmd, ignoreError, nil
}

func (shell *Shell) executeCommand(c *ast.CommandNode) (sh.Obj, error) {
	var (
		ignoreError    bool
		status         = "127"
		envVars        []string
		closeAfterWait []io.Closer
		cmd            sh.Runner
		err            error
		args           []sh.Obj
	)

	defer func() {
		for _, c := range closeAfterWait {
			c.Close()
		}
	}()

	cmd, ignoreError, err = shell.getCommand(c)
	if err != nil {
		goto cmdError
	}

	// SetEnviron must be called before SetArgs
	// otherwise the subshell will have the arguments
	// shadowed by parent env
	envVars = buildenv(shell.Environ())
	cmd.SetEnviron(envVars)

	args, err = shell.evalExprs(c.Args())
	if err != nil {
		goto cmdError
	}

	err = cmd.SetArgs(args)
	if err != nil {
		goto cmdError
	}

	closeAfterWait, err = shell.setRedirects(cmd, c.Redirects())
	if err != nil {
		goto cmdError
	}

	err = cmd.Start()
	if err != nil {
		goto cmdError
	}

	err = cmd.Wait()
	if err != nil {
		goto cmdError
	}

	return sh.NewStrObj("0"), nil

cmdError:
	statusObj := sh.NewStrObj(getErrStatus(err, status))
	if ignoreError {
		return statusObj, newErrIgnore(err.Error())
	}

	return statusObj, err
}

func (shell *Shell) evalList(argList *ast.ListExpr) (sh.Obj, error) {
	values := make([]sh.Obj, 0, len(argList.List))

	for _, arg := range argList.List {
		obj, err := shell.evalExpr(arg)
		if err != nil {
			return nil, err
		}

		values = append(values, obj)
	}

	return sh.NewListObj(values), nil
}

func (shell *Shell) evalArgList(argList *ast.ListExpr) ([]sh.Obj, error) {
	values := make([]sh.Obj, 0, len(argList.List))

	for _, arg := range argList.List {
		obj, err := shell.evalExpr(arg)
		if err != nil {
			return nil, err
		}

		values = append(values, obj)
	}

	if argList.IsVariadic {
		return values, nil
	}

	return []sh.Obj{sh.NewListObj(values)}, nil
}

func (shell *Shell) evalIndex(index ast.Expr) (int, error) {
	if index.Type() != ast.NodeIntExpr && index.Type() != ast.NodeVarExpr && index.Type() != ast.NodeIndexExpr {
		return 0, errors.NewEvalError(shell.filename,
			index, "Invalid indexing type: %s", index.Type())
	}

	if index.Type() == ast.NodeIntExpr {
		idxArg := index.(*ast.IntExpr)
		return idxArg.Value(), nil
	}

	idxObj, err := shell.evalVariable(index)
	if err != nil {
		return 0, err
	}

	if idxObj.Type() != sh.StringType {
		return 0, errors.NewEvalError(shell.filename,
			index, "Invalid object type on index value: %s", idxObj.Type())
	}

	objstr := idxObj.(*sh.StrObj)
	indexNum, err := strconv.Atoi(objstr.Str())

	if err != nil {
		return 0, err
	}

	return indexNum, nil
}

func (shell *Shell) evalIndexedVar(indexVar *ast.IndexExpr) (sh.Obj, error) {
	v, err := shell.evalVariable(indexVar.Var)

	if err != nil {
		return nil, err
	}

	col, err := sh.NewCollection(v)
	if err != nil {
		return nil, errors.NewEvalError(shell.filename, indexVar.Var, err.Error())
	}

	indexNum, err := shell.evalIndex(indexVar.Index)
	if err != nil {
		return nil, err
	}

	val, err := col.Get(indexNum)
	if err != nil {
		return nil, errors.NewEvalError(shell.filename, indexVar.Var, err.Error())
	}
	return val, nil
}

func (shell *Shell) evalArgIndexedVar(indexVar *ast.IndexExpr) ([]sh.Obj, error) {
	v, err := shell.evalVariable(indexVar.Var)
	if err != nil {
		return nil, err
	}

	col, err := sh.NewCollection(v)
	if err != nil {
		return nil, errors.NewEvalError(shell.filename, indexVar.Var, err.Error())
	}

	indexNum, err := shell.evalIndex(indexVar.Index)
	if err != nil {
		return nil, err
	}

	retval, err := col.Get(indexNum)
	if err != nil {
		return nil, errors.NewEvalError(shell.filename, indexVar.Var, err.Error())
	}

	if indexVar.IsVariadic {
		if retval.Type() != sh.ListType {
			return nil, errors.NewEvalError(shell.filename,
				indexVar, "Use of '...' on a non-list variable")
		}
		retlist := retval.(*sh.ListObj)
		return retlist.List(), nil
	}
	return []sh.Obj{retval}, nil
}

func (shell *Shell) evalVariable(a ast.Expr) (sh.Obj, error) {
	var (
		value sh.Obj
		ok    bool
	)

	if a.Type() == ast.NodeIndexExpr {
		return shell.evalIndexedVar(a.(*ast.IndexExpr))
	}

	if a.Type() != ast.NodeVarExpr {
		return nil, errors.NewEvalError(shell.filename,
			a, "Invalid eval of non variable argument: %s", a)
	}

	vexpr := a.(*ast.VarExpr)
	varName := vexpr.Name

	if value, ok = shell.Getvar(varName[1:]); !ok {
		return nil, errors.NewEvalError(shell.filename,
			a, "Variable %s not set on shell %s", varName, shell.name)
	}
	return value, nil
}

func (shell *Shell) evalArgVariable(a ast.Expr) ([]sh.Obj, error) {
	var (
		value sh.Obj
		ok    bool
	)

	if a.Type() == ast.NodeIndexExpr {
		return shell.evalArgIndexedVar(a.(*ast.IndexExpr))
	}

	if a.Type() != ast.NodeVarExpr {
		return nil, errors.NewEvalError(shell.filename,
			a, "Invalid eval of non variable argument: %s", a)
	}

	vexpr := a.(*ast.VarExpr)
	if value, ok = shell.Getvar(vexpr.Name[1:]); !ok {
		return nil, errors.NewEvalError(shell.filename,
			a, "Variable %s not set on shell %s", vexpr.Name,
			shell.name)
	}

	if vexpr.IsVariadic {
		if value.Type() != sh.ListType {
			return nil, errors.NewEvalError(shell.filename,
				a, "Variable expansion (%s) on a non-list object",
				vexpr.String())
		}

		return value.(*sh.ListObj).List(), nil
	}

	return []sh.Obj{value}, nil
}

func (shell *Shell) evalExprs(exprs []ast.Expr) ([]sh.Obj, error) {
	objs := make([]sh.Obj, 0, len(exprs))

	for _, expr := range exprs {
		obj, err := shell.evalExpr(expr)
		if err != nil {
			return nil, err
		}

		objs = append(objs, obj)
	}

	return objs, nil
}

func (shell *Shell) evalArgExprs(exprs []ast.Expr) ([]sh.Obj, error) {
	ret := make([]sh.Obj, 0, len(exprs))

	for _, expr := range exprs {
		objs, err := shell.evalArgExpr(expr)
		if err != nil {
			return nil, err
		}

		ret = append(ret, objs...)
	}

	return ret, nil
}

func (shell *Shell) evalArgExpr(expr ast.Expr) ([]sh.Obj, error) {
	switch expr.Type() {
	case ast.NodeStringExpr:
		if str, ok := expr.(*ast.StringExpr); ok {
			return []sh.Obj{
				sh.NewStrObj(str.Value()),
			}, nil
		}
	case ast.NodeConcatExpr:
		if concat, ok := expr.(*ast.ConcatExpr); ok {
			argVal, err := shell.evalConcat(concat)
			if err != nil {
				return nil, err
			}

			return []sh.Obj{
				sh.NewStrObj(argVal),
			}, nil
		}
	case ast.NodeVarExpr:
		return shell.evalArgVariable(expr)
	case ast.NodeIndexExpr:
		if indexedVar, ok := expr.(*ast.IndexExpr); ok {
			return shell.evalArgIndexedVar(indexedVar)
		}
	case ast.NodeListExpr:
		if listExpr, ok := expr.(*ast.ListExpr); ok {
			return shell.evalArgList(listExpr)
		}
	case ast.NodeFnInv:
		if fnInv, ok := expr.(*ast.FnInvNode); ok {
			objs, err := shell.executeFnInv(fnInv)
			if err != nil {
				return nil, err
			}

			if len(objs) == 0 {
				return nil, errors.NewEvalError(shell.filename,
					expr,
					"Function used in"+
						" expression but do not return any value: %s",
					fnInv)
			} else if len(objs) != 1 {
				return nil, errors.NewEvalError(shell.filename,
					expr,
					"Function used in"+
						" expression but it returns %d values: %10q",
					len(objs), objs)
			}

			return []sh.Obj{objs[0]}, nil
		}
	}

	return nil, errors.NewEvalError(shell.filename,
		expr, "Failed to eval expression: %+v", expr)
}

func (shell *Shell) evalExpr(expr ast.Expr) (sh.Obj, error) {
	switch expr.Type() {
	case ast.NodeStringExpr:
		if str, ok := expr.(*ast.StringExpr); ok {
			return sh.NewStrObj(str.Value()), nil
		}
	case ast.NodeConcatExpr:
		if concat, ok := expr.(*ast.ConcatExpr); ok {
			argVal, err := shell.evalConcat(concat)
			if err != nil {
				return nil, err
			}

			return sh.NewStrObj(argVal), nil
		}
	case ast.NodeVarExpr:
		return shell.evalVariable(expr)
	case ast.NodeIndexExpr:
		if indexedVar, ok := expr.(*ast.IndexExpr); ok {
			return shell.evalIndexedVar(indexedVar)
		}
	case ast.NodeListExpr:
		if listExpr, ok := expr.(*ast.ListExpr); ok {
			return shell.evalList(listExpr)
		}
	case ast.NodeFnInv:
		if fnInv, ok := expr.(*ast.FnInvNode); ok {
			objs, err := shell.executeFnInv(fnInv)
			if err != nil {
				return nil, err
			}

			if len(objs) == 0 {
				return nil, errors.NewEvalError(shell.filename,
					expr,
					"Function used in"+
						" expression but do not return any value: %s",
					fnInv)
			} else if len(objs) != 1 {
				return nil, errors.NewEvalError(shell.filename,
					expr,
					"Function used in"+
						" expression but it returns %d values: %10q",
					len(objs), objs)
			}

			return objs[0], nil
		}
	}

	return nil, errors.NewEvalError(shell.filename,
		expr, "Failed to eval expression: %+v", expr)
}

func (shell *Shell) executeSetenvAssign(assign *ast.AssignNode) error {
	for i := 0; i < len(assign.Names); i++ {
		name := assign.Names[i]
		value := assign.Values[i]
		err := shell.initVar(name, value)
		if err != nil {
			return err
		}
		obj, ok := shell.GetLocalvar(name.Ident)
		if !ok {
			return errors.NewEvalError(shell.filename,
				assign,
				"internal error: Setenv not setting local variable '%s'",
				name.Ident,
			)
		}
		shell.Setenv(name.Ident, obj)
	}
	return nil
}

func (shell *Shell) executeSetenvExec(assign *ast.ExecAssignNode) error {
	err := shell.executeExecAssign(assign)
	if err != nil {
		return err
	}
	for i := 0; i < len(assign.Names); i++ {
		name := assign.Names[i]
		obj, ok := shell.GetLocalvar(name.Ident)
		if !ok {
			return errors.NewEvalError(shell.filename,
				assign,
				"internal error: Setenv not setting local variable '%s'",
				name.Ident,
			)
		}
		shell.Setenv(name.Ident, obj)
	}
	return nil
}

func (shell *Shell) executeSetenv(v *ast.SetenvNode) error {
	var (
		varValue sh.Obj
		ok       bool
		assign   = v.Assignment()
	)

	if assign != nil {
		switch assign.Type() {
		case ast.NodeAssign:
			return shell.executeSetenvAssign(assign.(*ast.AssignNode))
		case ast.NodeExecAssign:
			return shell.executeSetenvExec(assign.(*ast.ExecAssignNode))
		}
		return errors.NewEvalError(shell.filename,
			v, "Failed to eval setenv, invalid assignment type: %+v",
			assign)
	}

	varValue, ok = shell.Getvar(v.Name)
	if !ok {
		return errors.NewEvalError(shell.filename,
			v, "Variable '%s' not set on shell %s", v.Name,
			shell.name,
		)
	}
	shell.Setenv(v.Name, varValue)
	return nil
}

func (shell *Shell) concatElements(expr *ast.ConcatExpr) (string, error) {
	value := ""

	list := expr.List()

	for i := 0; i < len(list); i++ {
		ec := list[i]

		obj, err := shell.evalExpr(ec)

		if err != nil {
			return "", err
		}

		if obj.Type() != sh.StringType {
			return "", errors.NewEvalError(shell.filename,
				expr, "Impossible to concat elements of type %s", obj.Type())
		}

		value = value + obj.String()
	}

	return value, nil
}

func (shell *Shell) execCmdOutput(cmd ast.Node,
	getstderr, ignoreError bool) ([]byte, []byte, sh.Obj, error) {
	var (
		outBuf, errBuf bytes.Buffer
		err            error
		status         sh.Obj
	)
	if cmd.Type() != ast.NodeCommand &&
		cmd.Type() != ast.NodePipe {
		return nil, nil, nil, errors.NewEvalError(shell.filename,
			cmd, "Invalid node type (%v). Expected command or pipe",
			cmd)
	}

	bkStdout, bkStderr := shell.stdout, shell.stderr
	shell.SetStdout(&outBuf)
	if getstderr {
		shell.SetStderr(&errBuf)
	}
	defer func() {
		shell.SetStdout(bkStdout)
		shell.SetStderr(bkStderr)
	}()

	if cmd.Type() == ast.NodeCommand {
		status, err = shell.executeCommand(cmd.(*ast.CommandNode))
	} else {
		status, err = shell.executePipe(cmd.(*ast.PipeNode))
	}

	outb := outBuf.Bytes()
	errb := errBuf.Bytes()

	trimnl := func(data []byte) []byte {
		if len(data) > 0 && data[len(data)-1] == '\n' {
			// remove the trailing new line
			// Why? because it's what user wants in 99.99% of times...

			data = data[0 : len(data)-1]
		}
		return data[:]
	}

	if ignoreError {
		err = nil
	}

	return trimnl(outb), trimnl(errb), status, err
}

func (shell *Shell) executeExecAssignCmd(v ast.Node) (stdout, stderr, status sh.Obj, err error) {
	assign := v.(*ast.ExecAssignNode)
	cmd := assign.Command()

	mustIgnoreErr := len(assign.Names) > 1
	collectStderr := len(assign.Names) == 3

	outb, errb, status, err := shell.execCmdOutput(cmd, collectStderr, mustIgnoreErr)
	if err != nil {
		return nil, nil, nil, err
	}

	return sh.NewStrObj(string(outb)), sh.NewStrObj(string(errb)), status, nil
}

func (shell *Shell) executeExecAssignFn(assign *ast.ExecAssignNode) ([]sh.Obj, error) {
	var (
		err      error
		fnValues []sh.Obj
	)

	cmd := assign.Command()
	if cmd.Type() != ast.NodeFnInv {
		return nil, errors.NewEvalError(shell.filename,
			cmd, "Invalid node type (%v). Expected function call",
			cmd)
	}

	fnValues, err = shell.executeFnInv(cmd.(*ast.FnInvNode))
	if err != nil {
		return nil, err
	}

	if len(fnValues) != len(assign.Names) {
		return nil, errors.NewEvalError(shell.filename,
			assign, "Functions returns %d objects, but statement expects %d",
			len(fnValues), len(assign.Names))
	}

	return fnValues, nil
}

func (shell *Shell) executeExecAssign(v *ast.ExecAssignNode) (err error) {
	exec := v.Command()
	switch exec.Type() {
	case ast.NodeFnInv:
		var values []sh.Obj
		values, err = shell.executeExecAssignFn(v)
		if err != nil {
			return err
		}
		err = shell.setvars(v.Names, values)
	case ast.NodeCommand, ast.NodePipe:
		var stdout, stderr, status sh.Obj
		stdout, stderr, status, err = shell.executeExecAssignCmd(v)
		if err != nil {
			return err
		}

		err = shell.setcmdvars(v.Names, stdout, stderr, status)
	default:
		err = errors.NewEvalError(shell.filename,
			exec, "Invalid node type (%v). Expected function call, command or pipe",
			exec)
	}

	return err
}

func (shell *Shell) initVar(name *ast.NameNode, value ast.Expr) error {
	obj, err := shell.evalExpr(value)
	if err != nil {
		return err
	}
	return shell.newvar(name, obj)
}

func (shell *Shell) executeVarAssign(v *ast.VarAssignDeclNode) error {
	assign := v.Assign
	if len(assign.Names) != len(assign.Values) {
		return errors.NewEvalError(shell.filename,
			assign, "Invalid multiple assignment. Different amount of variables and values: %s",
			assign,
		)
	}

	for i := 0; i < len(assign.Names); i++ {
		name := assign.Names[i]
		value := assign.Values[i]

		err := shell.initVar(name, value)
		if err != nil {
			return err
		}
	}

	return nil
}

func (shell *Shell) executeVarExecAssign(v *ast.VarExecAssignDeclNode) (err error) {
	assign := v.ExecAssign
	exec := assign.Command()
	switch exec.Type() {
	case ast.NodeFnInv:
		var values []sh.Obj
		values, err = shell.executeExecAssignFn(assign)
		if err != nil {
			return err
		}
		shell.newvars(assign.Names, values)
	case ast.NodeCommand, ast.NodePipe:
		var stdout, stderr, status sh.Obj
		stdout, stderr, status, err = shell.executeExecAssignCmd(assign)
		if err != nil {
			return err
		}

		shell.newcmdvars(assign.Names, stdout, stderr, status)
	default:
		err = errors.NewEvalError(shell.filename,
			exec, "Invalid node type (%v). Expected function call, command or pipe",
			exec)
	}

	return err
}

func (shell *Shell) executeAssignment(v *ast.AssignNode) error {
	if len(v.Names) != len(v.Values) {
		return errors.NewEvalError(shell.filename,
			v, "Invalid multiple assignment. Different amount of variables and values: %s",
			v,
		)
	}

	for i := 0; i < len(v.Names); i++ {
		name := v.Names[i]
		value := v.Values[i]

		obj, err := shell.evalExpr(value)
		if err != nil {
			return err
		}

		err = shell.setvar(name, obj)
		if err != nil {
			return err
		}
	}

	return nil
}

func (shell *Shell) evalIfArgument(arg ast.Node) (sh.Obj, error) {
	var (
		obj sh.Obj
		err error
	)

	obj, err = shell.evalExpr(arg)
	if err != nil {
		return nil, err
	} else if obj == nil {
		return nil, errors.NewEvalError(shell.filename,
			arg, "lvalue doesn't yield value (%s)", arg)
	}

	return obj, nil
}

func (shell *Shell) evalIfArguments(n *ast.IfNode) (string, string, error) {
	var (
		lobj, robj sh.Obj
		err        error
	)

	lobj, err = shell.evalIfArgument(n.Lvalue())

	if err != nil {
		return "", "", err
	}

	robj, err = shell.evalIfArgument(n.Rvalue())

	if err != nil {
		return "", "", err
	}

	if lobj.Type() != sh.StringType {
		return "", "", errors.NewEvalError(shell.filename,
			n, "lvalue is not comparable: (%v) -> %s.", lobj, lobj.Type())
	}

	if robj.Type() != sh.StringType {
		return "", "", errors.NewEvalError(shell.filename,
			n, "rvalue is not comparable: (%v) -> %s.", lobj, lobj.Type())
	}

	lobjstr := lobj.(*sh.StrObj)
	robjstr := robj.(*sh.StrObj)

	return lobjstr.Str(), robjstr.Str(), nil
}

func (shell *Shell) executeIfEqual(n *ast.IfNode) ([]sh.Obj, error) {
	lstr, rstr, err := shell.evalIfArguments(n)

	if err != nil {
		return nil, err
	}

	if lstr == rstr {
		return shell.executeTree(n.IfTree(), false)
	} else if n.ElseTree() != nil {
		return shell.executeTree(n.ElseTree(), false)
	}

	return nil, nil
}

func (shell *Shell) executeIfNotEqual(n *ast.IfNode) ([]sh.Obj, error) {
	lstr, rstr, err := shell.evalIfArguments(n)

	if err != nil {
		return nil, err
	}

	if lstr != rstr {
		return shell.executeTree(n.IfTree(), false)
	} else if n.ElseTree() != nil {
		return shell.executeTree(n.ElseTree(), false)
	}

	return nil, nil
}

func (shell *Shell) executeFnInv(n *ast.FnInvNode) ([]sh.Obj, error) {
	var fnDef sh.FnDef

	fnName := n.Name()
	if len(fnName) > 1 && fnName[0] == '$' {
		argVar := ast.NewVarExpr(token.NewFileInfo(n.Line(), n.Column()), fnName)

		obj, err := shell.evalVariable(argVar)
		if err != nil {
			return nil, err
		}

		if obj.Type() != sh.FnType {
			return nil, errors.NewEvalError(shell.filename,
				n, "Variable '%s' is not a function.", fnName)
		}

		objfn := obj.(*sh.FnObj)
		fnDef = objfn.Fn()
	} else {
		fnObj, err := shell.GetFn(fnName)
		if err != nil {
			return nil, errors.NewEvalError(shell.filename,
				n, err.Error())
		}
		fnDef = fnObj.Fn()
	}

	fn := fnDef.Build()
	args, err := shell.evalArgExprs(n.Args())
	if err != nil {
		return nil, err
	}

	err = fn.SetArgs(args)
	if err != nil {
		return nil, errors.NewEvalError(shell.filename,
			n, err.Error())
	}

	fn.SetStdin(shell.stdin)
	fn.SetStdout(shell.stdout)
	fn.SetStderr(shell.stderr)

	err = fn.Start()
	if err != nil {
		return nil, errors.NewEvalError(shell.filename,
			n, err.Error())
	}

	err = fn.Wait()
	if err != nil {
		return nil, errors.NewEvalError(shell.filename,
			n, err.Error())
	}

	return fn.Results(), nil
}

func (shell *Shell) executeInfLoop(tr *ast.Tree) ([]sh.Obj, error) {
	var (
		err  error
		objs []sh.Obj
	)

	for {
		objs, err = shell.executeTree(tr, false)

		runtime.Gosched()

		type (
			interruptedError interface {
				Interrupted() bool
			}

			stopWalkingError interface {
				StopWalking() bool
			}
		)

		if errInterrupted, ok := err.(interruptedError); ok && errInterrupted.Interrupted() {
			break
		}

		if errStopWalking, ok := err.(stopWalkingError); ok && errStopWalking.StopWalking() {
			return objs, err
		}

		shell.Lock()

		if shell.getIntr() {
			shell.setIntr(false)

			if err != nil {
				err = newErrInterrupted(err.Error())
			} else {
				err = newErrInterrupted("loop interrupted")
			}
		}

		shell.Unlock()

		if err != nil {
			break
		}
	}

	return nil, err
}

func (shell *Shell) executeFor(n *ast.ForNode) ([]sh.Obj, error) {
	shell.Lock()
	shell.looping = true
	shell.Unlock()

	defer func() {
		shell.Lock()
		defer shell.Unlock()

		shell.looping = false
	}()

	if n.InExpr() == nil {
		return shell.executeInfLoop(n.Tree())
	}

	id := n.Identifier()
	inExpr := n.InExpr()

	var (
		obj sh.Obj
		err error
	)

	if inExpr.Type() == ast.NodeVarExpr {
		obj, err = shell.evalVariable(inExpr.(*ast.VarExpr))
	} else if inExpr.Type() == ast.NodeListExpr {
		obj, err = shell.evalList(inExpr.(*ast.ListExpr))
	} else if inExpr.Type() == ast.NodeFnInv {
		var objs []sh.Obj
		objs, err = shell.executeFnInv(inExpr.(*ast.FnInvNode))
		if err != nil {
			return nil, err
		}

		if len(objs) != 1 {
			return nil, errors.NewEvalError(shell.filename,
				inExpr, "Functions with multiple returns do not work as for 'in expression' yet: %v", inExpr)
		}

		obj = objs[0]
	} else {
		return nil, errors.NewEvalError(shell.filename,
			inExpr, "Invalid expression in for loop: %s", inExpr.Type())
	}

	if err != nil {
		return nil, err
	}

	col, err := sh.NewCollection(obj)
	if err != nil {
		return nil, errors.NewEvalError(shell.filename,
			inExpr, "error[%s] trying to iterate", err)
	}

	for i := 0; i < col.Len(); i++ {
		val, err := col.Get(i)
		if err != nil {
			return nil, errors.NewEvalError(shell.filename,
				inExpr, "unexpected error[%s] during iteration", err)
		}
		shell.Newvar(id, val)
		objs, err := shell.executeTree(n.Tree(), false)

		type (
			interruptedError interface {
				Interrupted() bool
			}

			stopWalkingError interface {
				StopWalking() bool
			}
		)

		if errInterrupted, ok := err.(interruptedError); ok && errInterrupted.Interrupted() {
			return nil, err
		}

		if errStopWalking, ok := err.(stopWalkingError); ok && errStopWalking.StopWalking() {
			return objs, err
		}

		shell.Lock()

		if shell.getIntr() {
			shell.setIntr(false)
			shell.Unlock()

			if err != nil {
				return nil, newErrInterrupted(err.Error())
			}

			return nil, newErrInterrupted("loop interrupted")
		}

		shell.Unlock()

		if err != nil {
			return nil, err
		}
	}

	return nil, nil
}

func (shell *Shell) executeFnDecl(n *ast.FnDeclNode) error {
	fnDef, err := newUserFnDef(n.Name(), shell, n.Args(), n.Tree())
	if err != nil {
		return err
	}

	shell.Newvar(n.Name(), sh.NewFnObj(fnDef))
	shell.logf("Function %s declared on '%s'", n.Name(), shell.name)
	return nil
}

func (shell *Shell) executeBindFn(n *ast.BindFnNode) error {
	if !shell.Interactive() {
		return errors.NewEvalError(shell.filename,
			n, "'bindfn' is not allowed in non-interactive mode.")
	}

	fnDef, err := shell.GetFn(n.Name())
	if err != nil {
		return errors.NewEvalError(shell.filename,
			n, err.Error())
	}

	shell.Setbindfn(n.CmdName(), fnDef.Fn())
	return nil
}

func (shell *Shell) executeIf(n *ast.IfNode) ([]sh.Obj, error) {
	op := n.Op()

	if op == "==" {
		return shell.executeIfEqual(n)
	} else if op == "!=" {
		return shell.executeIfNotEqual(n)
	}

	return nil, fmt.Errorf("invalid operation '%s'", op)
}

func validateDirs(nashpath string, nashroot string) error {
	if nashpath == nashroot {
		return fmt.Errorf("invalid nashpath and nashroot, they are both[%s] but they must differ", nashpath)
	}
	err := validateDir(nashpath)
	if err != nil {
		return fmt.Errorf("invalid nashpath, user's config won't be loaded: error: %s", err)
	}
	err = validateDir(nashroot)
	if err != nil {
		return fmt.Errorf("invalid nashroot, stdlib/stdbin won't be available: error: %s", err)
	}
	return nil
}

func validateDir(dir string) error {
	dir, err := filepath.EvalSymlinks(dir)
	if err != nil {
		return err
	}

	info, err := os.Stat(dir)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is a file, expected a dir", dir)
	}
	if !filepath.IsAbs(dir) {
		return fmt.Errorf("%s is a relative path, expected a absolute path", dir)
	}
	return nil
}
