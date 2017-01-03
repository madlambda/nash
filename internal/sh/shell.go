package sh

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"os/signal"
	"path"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"github.com/NeowayLabs/nash/ast"
	"github.com/NeowayLabs/nash/errors"
	"github.com/NeowayLabs/nash/parser"
	"github.com/NeowayLabs/nash/sh"
	"github.com/NeowayLabs/nash/token"
)

const (
	logNS     = "nashell.Shell"
	defPrompt = "\033[31mλ>\033[0m "
)

type (
	// Env is the environment map of lists
	Env map[string]sh.Obj
	Var Env
	Fns map[string]sh.Fn

	StatusCode uint8

	// Shell is the core data structure.
	Shell struct {
		name        string
		debug       bool
		interactive bool
		lambdas     uint
		logf        LogFn
		nashdPath   string
		isFn        bool
		filename    string // current file being executed or imported

		interrupted bool
		looping     bool

		stdin  io.Reader
		stdout io.Writer
		stderr io.Writer

		env  Env
		vars Var
		fns  Fns

		builtins Fns
		binds    Fns

		root   *ast.Tree
		parent *Shell

		repr string // string representation

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

// NewShell creates a new shell object
func NewShell() (*Shell, error) {
	shell := &Shell{
		name:        "parent scope",
		interactive: false,
		isFn:        false,
		logf:        NewLog(logNS, false),
		nashdPath:   nashdAutoDiscover(),
		stdout:      os.Stdout,
		stderr:      os.Stderr,
		stdin:       os.Stdin,
		env:         make(Env),
		vars:        make(Var),
		fns:         make(Fns),
		builtins:    make(Fns),
		binds:       make(Fns),
		Mutex:       &sync.Mutex{},
		filename:    "<interactive>",
	}

	err := shell.setup()

	if err != nil {
		return nil, err
	}

	shell.setupSignals()

	return shell, nil
}

// NewSubShell creates a nashell.Shell that inherits the parent shell stdin,
// stdout, stderr and mutex lock.
// Every variable and function lookup is done first in the subshell and then, if
// not found, in the parent shell recursively.
func NewSubShell(name string, parent *Shell) (*Shell, error) {
	if parent == nil {
		return nil, errors.NewError("A sub Shell requires a parent shell")
	}

	sh := &Shell{
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
		fns:       make(Fns),
		binds:     make(Fns),
		builtins:  nil, // subshell does not have builtins
		Mutex:     parent.Mutex,
		filename:  parent.filename,
	}

	return sh, nil
}

// initEnv creates a new environment from old one
func (shell *Shell) initEnv(processEnv []string) error {
	largs := make([]sh.Obj, len(os.Args))

	for i := 0; i < len(os.Args); i++ {
		largs[i] = sh.NewStrObj(os.Args[i])
	}

	argv := sh.NewListObj(largs)

	shell.Setenv("argv", argv)
	shell.Setvar("argv", argv)

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

			shell.Setvar(p[0], value)
			shell.Setenv(p[0], value)
		}
	}

	pidVal := sh.NewStrObj(strconv.Itoa(os.Getpid()))

	shell.Setenv("PID", pidVal)
	shell.Setvar("PID", pidVal)

	if _, ok := shell.Getenv("SHELL"); !ok {
		shellVal := sh.NewStrObj(nashdAutoDiscover())
		shell.Setenv("SHELL", shellVal)
		shell.Setvar("SHELL", shellVal)
	}

	cwd, err := os.Getwd()

	if err != nil {
		return err
	}

	cwdObj := sh.NewStrObj(cwd)
	shell.Setenv("PWD", cwdObj)
	shell.Setvar("PWD", cwdObj)

	return nil
}

// Reset internal state
func (shell *Shell) Reset() {
	shell.fns = make(Fns)
	shell.vars = make(Var)
	shell.env = make(Env)
	shell.binds = make(Fns)
}

// SetDebug enable/disable debug in the shell
func (shell *Shell) SetDebug(d bool) {
	shell.debug = d
	shell.logf = NewLog(logNS, d)
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

	shell.Setvar(name, value)

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

			shell.Setvar(p[0], value)
			shell.Setenv(p[0], value)
		}
	}
}

func (shell *Shell) Getvar(name string) (sh.Obj, bool) {
	if value, ok := shell.vars[name]; ok {
		return value, ok
	}

	if shell.parent != nil {
		return shell.parent.Getvar(name)
	}

	return nil, false
}

func (shell *Shell) GetBuiltin(name string) (sh.Fn, bool) {
	shell.logf("Looking for builtin '%s' on shell '%s'\n", name, shell.name)

	if shell.parent != nil {
		return shell.parent.GetBuiltin(name)
	}

	if fn, ok := shell.builtins[name]; ok {
		return fn, true
	}

	return nil, false
}

func (shell *Shell) GetFn(name string) (sh.Fn, bool) {
	shell.logf("Looking for function '%s' on shell '%s'\n", name, shell.name)

	if fn, ok := shell.fns[name]; ok {
		return fn, ok
	}

	if shell.parent != nil {
		return shell.parent.GetFn(name)
	}

	return nil, false
}

func (shell *Shell) Setbindfn(name string, value sh.Fn) {
	shell.binds[name] = value
}

func (shell *Shell) Getbindfn(cmdName string) (sh.Fn, bool) {
	if fn, ok := shell.binds[cmdName]; ok {
		return fn, true
	}

	if shell.parent != nil {
		return shell.parent.Getbindfn(cmdName)
	}

	return nil, false
}

func (shell *Shell) Setvar(name string, value sh.Obj) {
	shell.vars[name] = value
}

func (shell *Shell) IsFn() bool { return shell.isFn }

func (shell *Shell) SetIsFn(b bool) { shell.isFn = b }

// SetNashdPath sets an alternativa path to nashd
func (shell *Shell) SetNashdPath(path string) {
	shell.nashdPath = path
}

// SetStdin sets the stdin for commands
func (shell *Shell) SetStdin(in io.Reader) { shell.stdin = in }

// SetStdout sets stdout for commands
func (shell *Shell) SetStdout(out io.Writer) { shell.stdout = out }

// SetStderr sets stderr for commands
func (shell *Shell) SetStderr(err io.Writer) { shell.stderr = err }

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

func (shell *Shell) String() string {
	if shell.repr != "" {
		return shell.repr
	}

	var out bytes.Buffer

	shell.dump(&out)

	return string(out.Bytes())
}

func (shell *Shell) setupBuiltin() {
	lenfn := NewLenFn(shell)
	shell.builtins["len"] = lenfn
	shell.Setvar("len", sh.NewFnObj(lenfn))

	appendfn := NewAppendFn(shell)
	shell.builtins["append"] = appendfn
	shell.Setvar("append", sh.NewFnObj(appendfn))

	splitfn := NewSplitFn(shell)
	shell.builtins["split"] = splitfn
	shell.Setvar("split", sh.NewFnObj(splitfn))

	chdir := NewChdir(shell)
	shell.builtins["chdir"] = chdir
	shell.Setvar("chdir", sh.NewFnObj(chdir))
}

func (shell *Shell) setupDefaultBindings() error {
	// only one builtin fn... no need for advanced machinery yet
	err := shell.Exec(shell.name, `fn nash_builtin_cd(path) {
            if $path == "" {
                    path = $HOME
            }

            chdir($path)
        }

        bindfn nash_builtin_cd cd`)

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
		shell.Setvar("PROMPT", pobj)
	}

	shell.setupBuiltin()
	return err
}

func (shell *Shell) setupSignals() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT)

	go func() {
		for {
			sig := <-sigs

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
			partValues, err := shell.evalVariable(part)

			if err != nil {
				return "", err
			}

			if partValues.Type() == sh.ListType {
				return "", errors.NewEvalError(shell.filename,
					part, "Concat of list variables is not allowed: %v = %v",
					part, partValues)
			} else if partValues.Type() != sh.StringType {
				return "", errors.NewEvalError(shell.filename, part,
					"Invalid concat element: %v", partValues)
			}

			strval := partValues.(*sh.StrObj)
			pathStr += strval.Str()
		case ast.NodeStringExpr:
			str, ok := part.(*ast.StringExpr)

			if !ok {
				return "", errors.NewEvalError(shell.filename, part,
					"Failed to eval string: %s", part)
			}

			pathStr += str.Value()
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

func (shell *Shell) executeNode(node ast.Node, builtin bool) (sh.Obj, error) {
	var (
		obj sh.Obj
		err error
	)

	shell.logf("Executing node: %v\n", node)

	switch node.Type() {
	case ast.NodeImport:
		err = shell.executeImport(node.(*ast.ImportNode))
	case ast.NodeComment:
		// ignore
	case ast.NodeSetenv:
		err = shell.executeSetenv(node.(*ast.SetenvNode))
	case ast.NodeAssignment:
		err = shell.executeAssignment(node.(*ast.AssignmentNode))
	case ast.NodeExecAssign:
		err = shell.executeExecAssign(node.(*ast.ExecAssignNode))
	case ast.NodeCommand:
		err = shell.executeCommand(node.(*ast.CommandNode))
	case ast.NodePipe:
		err = shell.executePipe(node.(*ast.PipeNode))
	case ast.NodeRfork:
		err = shell.executeRfork(node.(*ast.RforkNode))
	case ast.NodeIf:
		obj, err = shell.executeIf(node.(*ast.IfNode))
	case ast.NodeFnDecl:
		err = shell.executeFnDecl(node.(*ast.FnDeclNode))
	case ast.NodeFnInv:
		// invocation ignoring output
		_, err = shell.executeFnInv(node.(*ast.FnInvNode))
	case ast.NodeFor:
		obj, err = shell.executeFor(node.(*ast.ForNode))
	case ast.NodeBindFn:
		err = shell.executeBindFn(node.(*ast.BindFnNode))
	case ast.NodeDump:
		err = shell.executeDump(node.(*ast.DumpNode))
	case ast.NodeReturn:
		if shell.IsFn() {
			obj, err = shell.executeReturn(node.(*ast.ReturnNode))
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

	return obj, err
}

func (shell *Shell) ExecuteTree(tr *ast.Tree) (sh.Obj, error) {
	return shell.executeTree(tr, true)
}

// executeTree evaluates the given tree
func (shell *Shell) executeTree(tr *ast.Tree, stopable bool) (sh.Obj, error) {
	if tr == nil || tr.Root == nil {
		return nil, errors.NewError("empty abstract syntax tree to execute")
	}

	root := tr.Root

	for _, node := range root.Nodes {
		obj, err := shell.executeNode(node, false)

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
				return obj, nil
			}

			return obj, err
		}
	}

	return nil, nil
}

func (shell *Shell) executeReturn(n *ast.ReturnNode) (sh.Obj, error) {
	if n.Return() == nil {
		return nil, newErrStopWalking()
	}

	obj, err := shell.evalExpr(n.Return())

	if err != nil {
		return nil, err
	}

	return obj, newErrStopWalking()
}

func (shell *Shell) executeImport(node *ast.ImportNode) error {
	arg := node.Path()

	obj, err := shell.evalExpr(arg)

	if err != nil {
		return err
	}

	if obj.Type() != sh.StringType {
		return errors.NewEvalError(shell.filename,
			arg,
			"Invalid type on import argument: %s", obj.Type())
	}

	objstr := obj.(*sh.StrObj)
	fname := objstr.Str()

	shell.logf("Importing '%s'", fname)

	var tries []string

	var hasExt bool

	if len(fname) > 3 && fname[len(fname)-3:] == ".sh" {
		hasExt = true
	}

	if (len(fname) > 0 && fname[0] == '/') ||
		(len(fname) > 1 && fname[0] == '.' && fname[1] == '/') ||
		(len(fname) > 2 && fname[0] == '.' && fname[1] == '.' &&
			fname[2] == '/') {
		tries = append(tries, fname)

		if !hasExt {
			tries = append(tries, fname+".sh")
		}
	}

	if shell.filename != "" {
		tries = append(tries, path.Dir(shell.filename)+"/"+fname)

		if !hasExt {
			tries = append(tries, path.Dir(shell.filename)+"/"+fname+".sh")
		}
	}

	nashPath, ok := shell.Getenv("NASHPATH")

	if !ok {
		return errors.NewError("NASHPATH environment variable not set on shell %s", shell.name)
	} else if nashPath.Type() != sh.StringType {
		return errors.NewError("NASHPATH must be n string")
	}

	dotDir := nashPath.String()

	tries = append(tries, dotDir+"/lib/"+fname)

	if !hasExt {
		tries = append(tries, dotDir+"/lib/"+fname+".sh")
	}

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

	return errors.NewEvalError(shell.filename, node,
		"Failed to import path '%s'. The locations below have been tried:\n \"%s\"",
		fname,
		strings.Join(tries, `", "`))
}

// executePipe executes a pipe of ast.Command's. Each command can be
// a path command in the operating system or a function bind to a
// command name.
// The error of each command can be suppressed prepending it with '-' (dash).
// The error returned will be a string representing the errors (or none) of
// each command separated by '|'. The $status of pipe execution will be
// the $status of each command separated by '|'.
func (shell *Shell) executePipe(pipe *ast.PipeNode) error {
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
		return errors.NewEvalError(shell.filename,
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
		cmd.SetStderr(shell.stderr)

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

		cmd.SetStderr(shell.stderr)

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

	shell.Setvar("status", sh.NewStrObj("0"))
	return nil

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

	if len(uniqCodes) == 1 {
		// if all status are the same
		shell.Setvar("status", sh.NewStrObj(uniqCode))
	} else {
		shell.Setvar("status", sh.NewStrObj(strings.Join(cods, "|")))
	}

	if igns[errIndex] {
		return nil
	}

	return err
}

func (shell *Shell) openRedirectLocation(location ast.Expr) (io.WriteCloser, error) {
	var (
		protocol string
	)

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
		return os.OpenFile(locationStr, os.O_RDWR|os.O_CREATE, 0644)
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
			file, err := os.OpenFile("/dev/null", os.O_RDWR, 0644)

			if err != nil {
				return closeAfterWait, err
			}

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
			file, err := os.OpenFile("/dev/null", os.O_RDWR, 0644)

			if err != nil {
				return closeAfterWait, err
			}

			cmd.SetStderr(file)
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

	if fn, ok := shell.Getbindfn(cmdName); ok {
		shell.logf("Executing bind %s", cmdName)
		shell.logf("%s bind to %s", cmdName, fn)

		if !shell.Interactive() {
			err = errors.NewEvalError(shell.filename,
				c, "'%s' is a bind to '%s'. "+
					"No binds allowed in non-interactive mode.",
				cmdName,
				fn.Name())
			return nil, ignoreError, err
		}

		if len(c.Args()) > len(fn.ArgNames()) {
			err = errors.NewEvalError(shell.filename,
				c, "Too much arguments for"+
					" function '%s'. It expects %d args, but given %d. Arguments: %q",
				fn.Name(),
				len(fn.ArgNames()),
				len(c.Args()), c.Args())
			return nil, ignoreError, err
		}

		for i := 0 + len(c.Args()); i < len(fn.ArgNames()); i++ {
			// fill missing args with empty string
			// safe?
			c.SetArgs(append(c.Args(), ast.NewStringExpr(token.NewFileInfo(0, 0), "", true)))
		}

		return fn, ignoreError, nil
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

	return cmd, ignoreError, nil
}

func (shell *Shell) executeCommand(c *ast.CommandNode) error {
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

	args, err = cmdArgs(c.Args(), shell)

	if err != nil {
		goto cmdError
	}

	err = cmd.SetArgs(args)

	if err != nil {
		goto cmdError
	}

	cmd.SetStdin(shell.stdin)
	cmd.SetStdout(shell.stdout)
	cmd.SetStderr(shell.stderr)

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

	shell.Setvar("status", sh.NewStrObj("0"))

	return nil

cmdError:
	shell.Setvar("status", sh.NewStrObj(getErrStatus(err, status)))

	if ignoreError {
		return newErrIgnore(err.Error())
	}

	return err
}

func (shell *Shell) evalList(argList *ast.ListExpr) (sh.Obj, error) {
	values := make([]sh.Obj, 0, len(argList.List()))

	for _, arg := range argList.List() {
		obj, err := shell.evalExpr(arg)

		if err != nil {
			return nil, err
		}

		values = append(values, obj)
	}

	return sh.NewListObj(values), nil
}

func (shell *Shell) evalIndexedVar(indexVar *ast.IndexExpr) (sh.Obj, error) {
	var (
		indexNum int
	)

	variable := indexVar.Var()
	index := indexVar.Index()

	v, err := shell.evalVariable(variable)

	if err != nil {
		return nil, err
	}

	if v.Type() != sh.ListType {
		return nil, errors.NewEvalError(shell.filename, variable, "Invalid indexing of non-list variable: %s (%+v)", v.Type(), v)
	}

	if index.Type() == ast.NodeIntExpr {
		idxArg := index.(*ast.IntExpr)
		indexNum = idxArg.Value()
	} else if index.Type() == ast.NodeVarExpr {
		idxObj, err := shell.evalVariable(index)

		if err != nil {
			return nil, err
		}

		if idxObj.Type() != sh.StringType {
			return nil, errors.NewEvalError(shell.filename,
				index, "Invalid object type on index value: %s", idxObj.Type())
		}

		objstr := idxObj.(*sh.StrObj)
		indexNum, err = strconv.Atoi(objstr.Str())

		if err != nil {
			return nil, err
		}
	}

	vlist := v.(*sh.ListObj)
	values := vlist.List()

	if indexNum < 0 || indexNum >= len(values) {
		return nil, errors.NewEvalError(shell.filename,
			variable,
			"Index out of bounds. len(%s) == %d, but given %d", variable.Name(), len(values), indexNum)
	}

	return values[indexNum], nil
}

func (shell *Shell) evalVariable(a ast.Expr) (sh.Obj, error) {
	var (
		v  sh.Obj
		ok bool
	)

	if a.Type() == ast.NodeIndexExpr {
		return shell.evalIndexedVar(a.(*ast.IndexExpr))
	}

	if a.Type() != ast.NodeVarExpr {
		return nil, errors.NewEvalError(shell.filename,
			a, "Invalid eval of non variable argument: %s", a)
	}

	vexpr := a.(*ast.VarExpr)
	varName := vexpr.Name()

	if v, ok = shell.Getvar(varName[1:]); !ok {
		return nil, errors.NewEvalError(shell.filename,
			a, "Variable %s not set on shell %s", varName, shell.name)
	}

	return v, nil
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
			obj, err := shell.executeFnInv(fnInv)

			if err != nil {
				return nil, err
			}

			if obj == nil {
				return nil, errors.NewEvalError(shell.filename,
					expr,
					"Function used in"+
						" expression but do not return any value: %s",
					fnInv)
			}

			return obj, nil
		}
	}

	return nil, errors.NewEvalError(shell.filename,
		expr, "Failed to eval expression: %+v", expr)
}

func (shell *Shell) executeSetenv(v *ast.SetenvNode) error {
	var (
		varValue sh.Obj
		ok       bool
		assign   = v.Assignment()
		err      error
	)

	if assign != nil {
		switch assign.Type() {
		case ast.NodeAssignment:
			err = shell.executeAssignment(assign.(*ast.AssignmentNode))
		case ast.NodeExecAssign:
			err = shell.executeExecAssign(assign.(*ast.ExecAssignNode))
		default:
			err = errors.NewEvalError(shell.filename,
				v, "Failed to eval setenv, invalid assignment type: %+v",
				assign)
		}

		if err != nil {
			return err
		}
	}

	varName := v.Identifier()

	if varValue, ok = shell.Getvar(varName); !ok {
		return fmt.Errorf("Variable '%s' not set on shell %s", varName, shell.name)
	}

	shell.Setenv(varName, varValue)

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

func (shell *Shell) executeExecAssign(v *ast.ExecAssignNode) error {
	var (
		varOut bytes.Buffer
		err    error
	)

	bkStdout := shell.stdout

	shell.SetStdout(&varOut)

	defer shell.SetStdout(bkStdout)

	assign := v.Command()

	switch assign.Type() {
	case ast.NodeCommand:
		err = shell.executeCommand(assign.(*ast.CommandNode))
	case ast.NodePipe:
		err = shell.executePipe(assign.(*ast.PipeNode))
	case ast.NodeFnInv:
		fnValues, err := shell.executeFnInv(assign.(*ast.FnInvNode))

		if err != nil {
			return err
		}

		if fnValues == nil {
			return errors.NewEvalError(shell.filename,
				v, "Invalid assignment from function that does not return values: %s", assign)
		}

		shell.Setvar(v.Identifier(), fnValues)
		return nil
	default:
		err = errors.NewEvalError(shell.filename,
			assign, "Unexpected node in assignment: %s", assign.String())
	}

	output := varOut.Bytes()

	if len(output) > 0 && output[len(output)-1] == '\n' {
		// remove the trailing new line
		// Why? because it's what user wants in 99% of times...

		output = output[0 : len(output)-1]
	}

	shell.Setvar(v.Identifier(), sh.NewStrObj(string(output)))

	return err
}

func (shell *Shell) executeAssignment(v *ast.AssignmentNode) error {
	var err error

	obj, err := shell.evalExpr(v.Value())

	if err != nil {
		return err
	}

	shell.Setvar(v.Identifier(), obj)
	return nil
}

func (shell *Shell) evalIfArgument(arg ast.Node) (sh.Obj, error) {
	var (
		obj sh.Obj
		err error
	)

	if arg.Type() == ast.NodeFnInv {
		obj, err = shell.executeFnInv(arg.(*ast.FnInvNode))
	} else {
		obj, err = shell.evalExpr(arg)
	}

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

func (shell *Shell) executeIfEqual(n *ast.IfNode) (sh.Obj, error) {
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

func (shell *Shell) executeIfNotEqual(n *ast.IfNode) (sh.Obj, error) {
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

func (shell *Shell) executeFn(fn sh.Fn, nodeArgs []ast.Expr) (sh.Obj, error) {
	args, err := shell.evalExprs(nodeArgs)

	if err != nil {
		return nil, err
	}

	err = fn.SetArgs(args)

	if err != nil {
		return nil, err
	}

	err = fn.Start()

	if err != nil {
		return nil, err
	}

	err = fn.Wait()

	if err != nil {
		return nil, err
	}

	return fn.Results(), nil
}

func (shell *Shell) executeFnInv(n *ast.FnInvNode) (sh.Obj, error) {
	var (
		fn sh.Runner
		ok bool
	)

	fnName := n.Name()

	if len(fnName) > 1 && fnName[0] == '$' {
		argVar := ast.NewVarExpr(token.NewFileInfo(n.Line(), n.Column()), fnName)

		obj, err := shell.evalVariable(argVar)

		if err != nil {
			return nil, err
		}

		if obj.Type() != sh.FnType {
			return nil, errors.NewEvalError(shell.filename,
				n, "Variable '%s' isnt a function.", fnName)
		}

		objfn := obj.(*sh.FnObj)
		fn = objfn.Fn()
	} else {
		fn, ok = shell.GetBuiltin(fnName)

		if !ok {
			fn, ok = shell.GetFn(fnName)

			if !ok {
				return nil, errors.NewEvalError(shell.filename,
					n, "no such function '%s'", fnName)
			}
		}
	}

	args, err := shell.evalExprs(n.Args())

	if err != nil {
		return nil, err
	}

	err = fn.SetArgs(args)

	if err != nil {
		return nil, err
	}

	err = fn.Start()

	if err != nil {
		return nil, err
	}

	err = fn.Wait()

	if err != nil {
		return nil, err
	}

	return fn.Results(), nil
}

func (shell *Shell) executeInfLoop(tr *ast.Tree) (sh.Obj, error) {
	var (
		err error
		obj sh.Obj
	)

	for {
		obj, err = shell.executeTree(tr, false)

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
			return obj, err
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

func (shell *Shell) executeFor(n *ast.ForNode) (sh.Obj, error) {
	shell.Lock()
	shell.looping = true
	shell.Unlock()

	defer func() {
		shell.Lock()
		defer shell.Unlock()

		shell.looping = false
	}()

	if n.InVar() == "" {
		return shell.executeInfLoop(n.Tree())
	}

	id := n.Identifier()
	inVar := n.InVar()

	argVar := ast.NewVarExpr(token.NewFileInfo(n.Line(), n.Column()), inVar)

	obj, err := shell.evalVariable(argVar)

	if err != nil {
		return nil, err
	}

	if obj.Type() != sh.ListType {
		return nil, errors.NewEvalError(shell.filename,
			argVar, "Invalid variable type in for range: %s", obj.Type())
	}

	objlist := obj.(*sh.ListObj)

	for _, val := range objlist.List() {
		shell.Setvar(id, val)

		obj, err = shell.executeTree(n.Tree(), false)

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
			return obj, err
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
	fn, err := NewUserFn(n.Name(), shell)

	if err != nil {
		return err
	}

	fn.SetRepr(n.String())

	args := n.Args()

	for i := 0; i < len(args); i++ {
		arg := args[i]

		fn.AddArgName(arg)
	}

	fn.SetTree(n.Tree())

	fnName := n.Name()

	if fnName == "" {
		fnName = fmt.Sprintf("lambda %d", int(shell.lambdas))
		shell.lambdas++
	}

	shell.fns[fnName] = fn

	shell.Setvar(fnName, sh.NewFnObj(fn))
	shell.logf("Function %s declared on '%s'", fnName, shell.name)

	return nil
}

func (shell *Shell) dumpVar(file io.Writer) {
	for n, v := range shell.vars {
		printVar(file, n, v)
	}
}

func (shell *Shell) dumpEnv(file io.Writer) {
	for n := range shell.env {
		printEnv(file, n)
	}
}

func (shell *Shell) dumpFns(file io.Writer) {
	for _, f := range shell.fns {
		fmt.Fprintf(file, "%s\n\n", f.String())
	}
}

func (shell *Shell) dump(out io.Writer) {
	shell.dumpVar(out)
	shell.dumpEnv(out)
	shell.dumpFns(out)
}

func (shell *Shell) executeDump(n *ast.DumpNode) error {
	var (
		err    error
		file   io.Writer
		obj    sh.Obj
		objstr *sh.StrObj
	)

	fnameArg := n.Filename()

	if fnameArg == nil {
		file = shell.stdout
		goto execDump
	}

	obj, err = shell.evalExpr(fnameArg)

	if err != nil {
		return err
	}

	if obj.Type() != sh.StringType {
		return errors.NewEvalError(shell.filename,
			fnameArg,
			"dump does not support argument of type %s", obj.Type())
	}

	objstr = obj.(*sh.StrObj)
	file, err = os.OpenFile(objstr.Str(), os.O_CREATE|os.O_RDWR, 0644)

	if err != nil {
		return err
	}

execDump:
	shell.dump(file)

	return nil
}

func (shell *Shell) executeBindFn(n *ast.BindFnNode) error {
	if !shell.Interactive() {
		return errors.NewEvalError(shell.filename,
			n, "'bindfn' is not allowed in non-interactive mode.")
	}

	if fn, ok := shell.GetFn(n.Name()); ok {
		shell.Setbindfn(n.CmdName(), fn)
	} else {
		return errors.NewEvalError(shell.filename,
			n, "No such function '%s'", n.Name())
	}

	return nil
}

func (shell *Shell) executeIf(n *ast.IfNode) (sh.Obj, error) {
	op := n.Op()

	if op == "==" {
		return shell.executeIfEqual(n)
	} else if op == "!=" {
		return shell.executeIfNotEqual(n)
	}

	return nil, fmt.Errorf("Invalid operation '%s'.", op)
}
