package nash

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"github.com/NeowayLabs/nash/ast"
	"github.com/NeowayLabs/nash/errors"
	"github.com/NeowayLabs/nash/parser"
)

type (
	// Env is the environment map of lists
	Env map[string]*Obj
	Var Env
	Fns map[string]*Shell
	Bns Fns

	// Shell is the core data structure.
	Shell struct {
		name        string
		debug       bool
		lambdas     uint
		logf        LogFn
		nashdPath   string
		dotDir      string
		isFn        bool
		currentFile string // current file being executed or imported

		interrupted bool
		looping     bool

		stdin  io.Reader
		stdout io.Writer
		stderr io.Writer

		argNames []string
		env      Env
		vars     Var
		fns      Fns
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

const (
	logNS     = "nash.Shell"
	defPrompt = "\033[31mλ>\033[0m "
)

// NewShell creates a new shell object
func NewShell() (*Shell, error) {
	sh := &Shell{
		name:      "parent scope",
		isFn:      false,
		logf:      NewLog(logNS, false),
		nashdPath: nashdAutoDiscover(),
		stdout:    os.Stdout,
		stderr:    os.Stderr,
		stdin:     os.Stdin,
		env:       make(Env),
		vars:      make(Var),
		fns:       make(Fns),
		binds:     make(Fns),
		argNames:  make([]string, 0, 16),
		Mutex:     &sync.Mutex{},
	}

	err := sh.setup()

	if err != nil {
		return nil, err
	}

	sh.setupSignals()

	return sh, nil
}

// NewSubShell creates a nash.Shell that inherits the parent shell stdin,
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
		argNames:  make([]string, 0, 16),
		Mutex:     parent.Mutex,
	}

	return sh, nil
}

// initEnv creates a new environment from old one
func (sh *Shell) initEnv() error {
	processEnv := os.Environ()

	argv := NewListObj(os.Args)

	sh.Setenv("argv", argv)
	sh.Setvar("argv", argv)

	for _, penv := range processEnv {
		var value *Obj
		p := strings.Split(penv, "=")

		if len(p) == 2 {
			value = NewStrObj(p[1])

			sh.Setvar(p[0], value)
			sh.Setenv(p[0], value)
		}
	}

	pidVal := NewStrObj(strconv.Itoa(os.Getpid()))

	sh.Setenv("PID", pidVal)
	sh.Setvar("PID", pidVal)

	if _, ok := sh.Getenv("SHELL"); !ok {
		shellVal := NewStrObj(nashdAutoDiscover())
		sh.Setenv("SHELL", shellVal)
		sh.Setvar("SHELL", shellVal)
	}

	cwd, err := os.Getwd()

	if err != nil {
		return err
	}

	cwdObj := NewStrObj(cwd)
	sh.Setenv("PWD", cwdObj)
	sh.Setvar("PWD", cwdObj)

	return nil
}

// Reset internal state
func (sh *Shell) Reset() {
	sh.fns = make(Fns)
	sh.vars = make(Var)
	sh.env = make(Env)
	sh.binds = make(Fns)
}

// SetDebug enable/disable debug in the shell
func (sh *Shell) SetDebug(d bool) {
	sh.debug = d
	sh.logf = NewLog(logNS, d)
}

func (sh *Shell) SetName(a string) {
	sh.name = a
}

func (sh *Shell) Name() string { return sh.name }

func (sh *Shell) SetParent(a *Shell) {
	sh.parent = a
}

func (sh *Shell) Environ() Env {
	if sh.parent != nil {
		// Note(i4k): stack overflow, refactor this!
		return sh.parent.Environ()
	}

	return sh.env
}

func (sh *Shell) Getenv(name string) (*Obj, bool) {
	if sh.parent != nil {
		return sh.parent.Getenv(name)
	}

	value, ok := sh.env[name]
	return value, ok
}

func (sh *Shell) Setenv(name string, value *Obj) {
	if sh.parent != nil {
		sh.parent.Setenv(name, value)
		return
	}

	sh.env[name] = value

	os.Setenv(name, value.String())
}

func (sh *Shell) SetEnviron(env Env) {
	sh.env = env
}

func (sh *Shell) GetVar(name string) (*Obj, bool) {
	if value, ok := sh.vars[name]; ok {
		return value, ok
	}

	if sh.parent != nil {
		return sh.parent.GetVar(name)
	}

	return nil, false
}

func (sh *Shell) GetFn(name string) (*Shell, bool) {
	sh.logf("Looking for function '%s' on shell '%s'\n", name, sh.name)

	if fn, ok := sh.fns[name]; ok {
		return fn, ok
	}

	if sh.parent != nil {
		return sh.parent.GetFn(name)
	}

	return nil, false
}

func (sh *Shell) Setbindfn(name string, value *Shell) {
	sh.binds[name] = value
}

func (sh *Shell) Getbindfn(cmdName string) (*Shell, bool) {
	if fn, ok := sh.binds[cmdName]; ok {
		return fn, true
	}

	if sh.parent != nil {
		return sh.parent.Getbindfn(cmdName)
	}

	return nil, false
}

func (sh *Shell) Setvar(name string, value *Obj) {
	sh.vars[name] = value
}

func (sh *Shell) IsFn() bool { return sh.isFn }

func (sh *Shell) SetIsFn(b bool) { sh.isFn = b }

// Prompt returns the environment prompt or the default one
func (sh *Shell) Prompt() string {
	value, ok := sh.Getenv("PROMPT")

	if ok {
		return value.String()
	}

	return "<no prompt> "
}

// SetNashdPath sets an alternativa path to nashd
func (sh *Shell) SetNashdPath(path string) {
	sh.nashdPath = path
}

func (sh *Shell) SetDotDir(path string) {
	sh.dotDir = path

	obj := NewStrObj(sh.dotDir)

	sh.Setenv("NASHPATH", obj)
	sh.Setvar("NASHPATH", obj)
}

func (sh *Shell) DotDir() string {
	return sh.dotDir
}

// SetStdin sets the stdin for commands
func (sh *Shell) SetStdin(in io.Reader) {
	sh.stdin = in
}

// SetStdout sets stdout for commands
func (sh *Shell) SetStdout(out io.Writer) {
	sh.stdout = out
}

// SetStderr sets stderr for commands
func (sh *Shell) SetStderr(err io.Writer) {
	sh.stderr = err
}

func (sh *Shell) Stdout() io.Writer { return sh.stdout }
func (sh *Shell) Stderr() io.Writer { return sh.stderr }
func (sh *Shell) Stdin() io.Reader  { return sh.stdin }

func (sh *Shell) AddArgName(name string) {
	sh.argNames = append(sh.argNames, name)
}

func (sh *Shell) SetTree(t *ast.Tree) {
	sh.root = t
}

func (sh *Shell) Tree() *ast.Tree { return sh.root }

func (sh *Shell) SetRepr(a string) {
	sh.repr = a
}

func (sh *Shell) String() string {
	if sh.repr != "" {
		return sh.repr
	}

	var out bytes.Buffer

	sh.dump(&out)

	return string(out.Bytes())
}

func (sh *Shell) setup() error {
	err := sh.initEnv()

	if err != nil {
		return err
	}

	if sh.env["PROMPT"] == nil {
		pobj := NewStrObj(defPrompt)
		sh.Setenv("PROMPT", pobj)
		sh.Setvar("PROMPT", pobj)
	}

	return nil
}

func (sh *Shell) setupSignals() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGCHLD)

	go func() {
		for {
			sig := <-sigs

			switch sig {
			case syscall.SIGINT:
				sh.Lock()

				// TODO(i4k): Review implementation when interrupted inside
				// function loops
				if sh.looping {
					sh.setIntr(true)
				}

				sh.Unlock()
			case syscall.SIGCHLD:
				// dont need reaping because we dont have job control yet
				// every command is wait'ed.
			default:
				fmt.Printf("%s\n", sig)
			}
		}
	}()
}

func (sh *Shell) TriggerCTRLC() error {
	p, err := os.FindProcess(os.Getpid())

	if err != nil {
		return err
	}

	return p.Signal(syscall.SIGINT)
}

// setIntr *do not lock*. You must do it yourself!
func (sh *Shell) setIntr(b bool) {
	if sh.parent != nil {
		sh.parent.setIntr(b)
		return
	}

	sh.interrupted = b
}

// getIntr returns true if nash was interrupted by CTRL-C
func (sh *Shell) getIntr() bool {
	if sh.parent != nil {
		return sh.parent.getIntr()
	}

	return sh.interrupted
}

func (sh *Shell) evalConcat(path *ast.Arg) (string, error) {
	var pathStr string

	concat := path.Concat()

	for i := 0; i < len(concat); i++ {
		part := concat[i]

		if part.IsConcat() {
			return "", errors.NewError("Nested concat is not allowed")
		}

		if part.IsVariable() {
			partValues, err := sh.evalVariable(part)

			if err != nil {
				return "", err
			}

			if partValues.Type() == ListType {
				return "", fmt.Errorf("Concat of list variables is not allowed: %s = %v", part.Value(), partValues)
			} else if partValues.Type() != StringType {
				return "", fmt.Errorf("Invalid concat element: %v", partValues)
			}

			pathStr += partValues.Str()
		} else if part.IsQuoted() || part.IsUnquoted() {
			pathStr += part.Value()
		} else if part.IsList() {
			return "", errors.NewError("Concat of lists is not allowed: %+v", part.List())
		}
	}

	return pathStr, nil
}

func (sh *Shell) Execute() (*Obj, error) {
	if sh.root != nil {
		return sh.ExecuteTree(sh.root)
	}

	return nil, nil
}

// ExecuteString executes the commands specified by string content
func (sh *Shell) ExecuteString(path, content string) error {
	parser := parser.NewParser(path, content)

	tr, err := parser.Parse()

	if err != nil {
		return err
	}

	_, err = sh.ExecuteTree(tr)
	return err
}

// Execute the nash file at given path
func (sh *Shell) ExecuteFile(path string) error {
	bkCurFile := sh.currentFile

	content, err := ioutil.ReadFile(path)

	if err != nil {
		return err
	}

	sh.currentFile = path

	defer func() {
		sh.currentFile = bkCurFile
	}()

	return sh.ExecuteString(path, string(content))
}
