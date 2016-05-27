package nash

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
)

type (
	// Env is the environment map of lists
	Env map[string][]string
	Var Env
	Fns map[string]*Shell
	Bns Fns

	// Shell is the core data structure.
	Shell struct {
		name      string
		debug     bool
		lambdas   uint
		log       LogFn
		nashdPath string
		dotDir    string

		interrupted bool

		stdin  io.Reader
		stdout io.Writer
		stderr io.Writer

		argNames []string
		env      Env
		vars     Var
		fns      Fns
		binds    Fns

		root   *Tree
		parent *Shell

		repr string // string representation

		*sync.Mutex
	}

	errIgnore struct {
		*nashError
	}

	errInterrupted struct {
		*nashError
	}
)

func newErrIgnore(format string, arg ...interface{}) error {
	e := &errIgnore{
		nashError: newError(format, arg...),
	}

	return e
}

func (e *errIgnore) Ignore() bool { return true }

func newErrInterrupted(format string, arg ...interface{}) error {
	return &errInterrupted{
		nashError: newError(format, arg...),
	}
}

func (e *errInterrupted) Interrupted() bool { return true }

const (
	logNS     = "nash.Shell"
	defPrompt = "\033[31mλ>\033[0m "
)

// NewShell creates a new shell object
func NewShell(debug bool) (*Shell, error) {
	env, vars, err := NewEnv()

	if err != nil {
		return nil, err
	}

	if env["PROMPT"] == nil {
		env["PROMPT"] = append(make([]string, 0, 1), defPrompt)
		vars["PROMPT"] = append(make([]string, 0, 1), defPrompt)
	}

	sh := &Shell{
		name:      "parent scope",
		debug:     debug,
		log:       NewLog(logNS, debug),
		nashdPath: nashdAutoDiscover(),
		stdout:    os.Stdout,
		stderr:    os.Stderr,
		stdin:     os.Stdin,
		env:       env,
		vars:      vars,
		fns:       make(Fns),
		binds:     make(Fns),
		argNames:  make([]string, 0, 16),
		Mutex:     &sync.Mutex{},
	}

	sh.setup()

	return sh, nil
}

// NewEnv creates a new environment from old one
func NewEnv() (Env, Var, error) {
	env := make(Env)
	vars := make(Var)
	processEnv := os.Environ()

	env["argv"] = os.Args
	vars["argv"] = os.Args

	for _, penv := range processEnv {
		var value []string
		p := strings.Split(penv, "=")

		if len(p) == 1 {
			value = make([]string, 0, 1)
		} else if len(p) > 1 {
			value = append(make([]string, 0, 256), p[1:]...)
		}

		env[p[0]] = value
		vars[p[0]] = value
	}

	pidVal := append(make([]string, 0, 1), strconv.Itoa(os.Getpid()))

	env["PID"] = pidVal
	vars["PID"] = pidVal

	shellVal := append(make([]string, 0, 1), os.Args[0])
	env["SHELL"] = shellVal
	vars["SHELL"] = shellVal

	cwd, err := os.Getwd()

	if err != nil {
		return nil, nil, err
	}

	env["PWD"] = append(make([]string, 0, 1), cwd)
	vars["PWD"] = append(make([]string, 0, 1), cwd)

	return env, vars, nil
}

func (sh *Shell) SetName(a string) {
	sh.name = a
}

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

func (sh *Shell) GetEnv(name string) ([]string, bool) {
	if sh.parent != nil {
		return sh.parent.GetEnv(name)
	}

	value, ok := sh.env[name]
	return value, ok
}

func (sh *Shell) SetEnv(name string, value []string) {
	if sh.parent != nil {
		sh.parent.SetEnv(name, value)
		return
	}

	sh.env[name] = value
}

func (sh *Shell) SetEnviron(env Env) {
	sh.env = env
}

func (sh *Shell) GetVar(name string) ([]string, bool) {
	if value, ok := sh.vars[name]; ok {
		return value, ok
	}

	if sh.parent != nil {
		return sh.parent.GetVar(name)
	}

	return nil, false
}

func (sh *Shell) GetFn(name string) (*Shell, bool) {
	if fn, ok := sh.fns[name]; ok {
		return fn, ok
	}

	if sh.parent != nil {
		return sh.parent.GetFn(name)
	}

	return nil, false
}

func (sh *Shell) SetVar(name string, value []string) {
	sh.vars[name] = value
}

// Prompt returns the environment prompt or the default one
func (sh *Shell) Prompt() string {
	value, ok := sh.GetEnv("PROMPT")

	if ok {
		return value[0]
	}

	return "<no prompt> "
}

// SetDebug enable/disable debug in the shell
func (sh *Shell) SetDebug(debug bool) {
	sh.log = NewLog(logNS, debug)
}

// SetNashdPath sets an alternativa path to nashd
func (sh *Shell) SetNashdPath(path string) {
	sh.nashdPath = path
}

func (sh *Shell) SetDotDir(path string) {
	sh.dotDir = path
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

func (sh *Shell) AddArgName(name string) {
	sh.argNames = append(sh.argNames, name)
}

func (sh *Shell) SetTree(t *Tree) {
	sh.root = t
}

func (sh *Shell) Tree() *Tree { return sh.root }

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

func (sh *Shell) setup() {
	sh.env["NASHPATH"] = append(make([]string, 0, 1), sh.dotDir)
	sh.vars["NASHPATH"] = append(make([]string, 0, 1), sh.dotDir)

	sh.setupSignals()
}

func (sh *Shell) setupSignals() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT)

	go func() {
		for {
			<-sigs

			sh.Lock()
			sh.interrupted = !sh.interrupted
			sh.Unlock()

			//			fmt.Println("^C")
		}
	}()
}

func (sh *Shell) executeConcat(path *Arg) (string, error) {
	var pathStr string

	for i := 0; i < len(path.concat); i++ {
		part := path.concat[i]

		if part.IsConcat() {
			return "", errors.New("Nested concat is not allowed")
		}

		if part.IsVariable() {
			partValues, err := sh.evalVariable(part.Value())

			if err != nil {
				return "", err
			}

			if len(partValues) > 1 {
				return "", fmt.Errorf("Concat of list variables is not allowed: %s = %v", part.Value(), partValues)
			} else if len(partValues) == 0 {
				return "", fmt.Errorf("Variable %s not set", part.Value())
			}

			pathStr += partValues[0]
		} else {
			pathStr += part.Value()
		}
	}

	return pathStr, nil
}

func (sh *Shell) Execute() error {
	if sh.root != nil {
		return sh.ExecuteTree(sh.root)
	}

	return nil
}

// ExecuteString executes the commands specified by string content
func (sh *Shell) ExecuteString(path, content string) error {
	parser := NewParser(path, content)

	tr, err := parser.Parse()

	if err != nil {
		return err
	}

	return sh.ExecuteTree(tr)
}

// Execute the nash file at given path
func (sh *Shell) ExecuteFile(path string) error {
	content, err := ioutil.ReadFile(path)

	if err != nil {
		return err
	}

	return sh.ExecuteString(path, string(content))
}

// ExecuteTree evaluates the given tree
func (sh *Shell) ExecuteTree(tr *Tree) error {
	var err error

	if tr == nil || tr.Root == nil {
		return errors.New("nothing parsed")
	}

	root := tr.Root

	for _, node := range root.Nodes {
		sh.log("Executing: %v\n", node)

		switch node.Type() {
		case NodeImport:
			err = sh.executeImport(node.(*ImportNode))
		case NodeShowEnv:
			err = sh.executeShowEnv(node.(*ShowEnvNode))
		case NodeComment:
			continue // ignore comment
		case NodeSetAssignment:
			err := sh.executeSetAssignment(node.(*SetAssignmentNode))

			if err != nil {
				return err
			}
		case NodeAssignment:
			err = sh.executeAssignment(node.(*AssignmentNode))
		case NodeCmdAssignment:
			err = sh.executeCmdAssignment(node.(*CmdAssignmentNode))
		case NodeCommand:
			err = sh.executeCommand(node.(*CommandNode))
		case NodePipe:
			err = sh.executePipe(node.(*PipeNode))
		case NodeRfork:
			err = sh.executeRfork(node.(*RforkNode))
		case NodeCd:
			err = sh.executeCd(node.(*CdNode))
		case NodeIf:
			err = sh.executeIf(node.(*IfNode))
		case NodeFnDecl:
			err = sh.executeFnDecl(node.(*FnDeclNode))
		case NodeFnInv:
			err = sh.executeFnInv(node.(*FnInvNode))
		case NodeBindFn:
			err = sh.executeBindFn(node.(*BindFnNode))
		case NodeDump:
			err = sh.executeDump(node.(*DumpNode))
		default:
			// should never get here
			return fmt.Errorf("invalid node: %v.", node.Type())
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func (sh *Shell) executeImport(node *ImportNode) error {
	fname := node.Filename()

	tries := []string{
		fname,
		sh.dotDir + "/src/" + fname,
	}

	if len(fname) > 2 && fname[0] == '.' && fname[1] == '/' {
		return sh.ExecuteFile(fname)
	}

	for _, path := range tries {
		_, err := os.Stat(path)

		if err != nil {
			continue
		}

		return sh.ExecuteFile(path)
	}

	return newError("Failed to import path '%s'. The locations below have been tried:\n \"%s\"",
		strings.Join(tries, `", "`))
}

func (sh *Shell) executeShowEnv(node *ShowEnvNode) error {
	envVars := buildenv(sh.Environ())
	for _, e := range envVars {
		fmt.Fprintf(sh.stdout, "%s\n", e)
	}

	return nil
}

func (sh *Shell) executePipe(pipe *PipeNode) error {
	var err error
	nodeCommands := pipe.Commands()

	if len(nodeCommands) <= 1 {
		return newError("Pipe requires at least two commands.")
	}

	cmds := make([]*Command, len(nodeCommands))

	// Create all commands
	for i := 0; i < len(nodeCommands); i++ {
		nodeCmd := nodeCommands[i]
		cmd, err := NewCommand(nodeCmd.name, sh)

		if err != nil {
			return err
		}

		err = cmd.SetArgs(nodeCmd.args)

		if err != nil {
			return err
		}

		cmd.SetPassDone(false)

		if i < (len(nodeCommands) - 1) {
			err = cmd.SetRedirects(nodeCmd.redirs)

			if err != nil {
				return err
			}
		}

		cmds[i] = cmd
	}

	last := len(nodeCommands) - 1

	// Setup the commands. Pointing the stdin of next command to stdout of previous.
	// Except the last one
	for i, cmd := range cmds[:last] {
		//cmd.SetFDMap(0, sh.stdin)
		//cmd.SetFDMap(1, sh.stdout)
		//cmd.SetFDMap(2, sh.stderr)

		// connect commands
		if cmds[i+1].Stdin != os.Stdin {
			return newError("Stdin redirected")
		}

		if cmd.Stdout != os.Stdout {
			return newError("Stdout redirected")
		}

		cmds[i+1].Stdin = nil
		cmd.Stdout = nil

		if cmds[i+1].Stdin, err = cmd.StdoutPipe(); err != nil {
			return err
		}

		cmd.stdoutDone <- true
		cmd.stdinDone <- true
		cmd.stderrDone <- true
	}

	cmds[last].stdinDone <- true

	if sh.stdout != os.Stdout {
		cmds[last].Stdout = nil
		stdout, err := cmds[last].StdoutPipe()

		if err != nil {
			return err
		}

		go func() {
			io.Copy(sh.stdout, stdout)
			cmds[last].stdoutDone <- true
		}()
	} else {
		cmds[last].Stdout = sh.stdout
		cmds[last].stdoutDone <- true
	}

	if sh.stderr != os.Stderr {
		cmds[last].Stderr = nil
		stderr, err := cmds[last].StderrPipe()

		if err != nil {
			return err
		}

		go func() {
			io.Copy(sh.stderr, stderr)
			cmds[last].stderrDone <- true
		}()
	} else {
		cmds[last].Stderr = sh.stderr
		cmds[last].stderrDone <- true
	}

	for _, cmd := range cmds {
		err := cmd.Start()

		if err != nil {
			return err
		}
	}

	for _, cmd := range cmds {
		err := cmd.Wait()

		if err != nil {
			return err
		}
	}

	return nil
}

func (sh *Shell) executeCommand(c *CommandNode) error {
	var ignoreError bool

	sh.Lock()
	sh.interrupted = false
	sh.Unlock()

	if len(c.name) > 1 && c.name[0] == '-' {
		ignoreError = true
		c.name = c.name[1:]

		sh.log("Ignoring error\n")
	}

	cmd, err := NewCommand(c.name, sh)

	if err != nil {
		type NotFound interface {
			NotFound() bool
		}

		sh.log("Command fails: %s", err.Error())

		if errNotFound, ok := err.(NotFound); ok && errNotFound.NotFound() {
			if fn, ok := sh.binds[c.Name()]; ok {
				sh.log("Executing bind %s", c.Name())

				return sh.executeFn(fn, c.args)
			}
		}

		goto cmdError
	}

	err = cmd.SetArgs(c.args)

	if err != nil {
		goto cmdError
	}

	cmd.SetFDMap(0, sh.stdin)
	cmd.SetFDMap(1, sh.stdout)
	cmd.SetFDMap(2, sh.stderr)

	err = cmd.SetRedirects(c.redirs)

	if err != nil {
		goto cmdError
	}

	defer cmd.CloseNetDescriptors()

	err = cmd.Start()

	if err != nil {
		goto cmdError
	}

	err = cmd.Wait()

	if err != nil {
		goto cmdError
	}

	return nil

cmdError:
	if ignoreError {
		return newErrIgnore(err.Error())
	}

	sh.Lock()
	defer sh.Unlock()

	if sh.interrupted {
		sh.interrupted = !sh.interrupted
		return newErrInterrupted(err.Error())
	}

	return err
}

func (sh *Shell) evalVariable(a string) ([]string, error) {
	if v, ok := sh.GetVar(a[1:]); ok {
		return v, nil
	}

	return nil, fmt.Errorf("Variable %s not set", a)
}

func (sh *Shell) executeSetAssignment(v *SetAssignmentNode) error {
	var (
		varValue []string
		ok       bool
	)

	varName := v.varName

	if varValue, ok = sh.GetVar(varName); !ok {
		return fmt.Errorf("Variable '%s' not set", varName)
	}

	sh.SetEnv(varName, varValue)

	return nil
}

func (sh *Shell) concatElements(elem *Arg) (string, error) {
	value := ""

	for i := 0; i < len(elem.concat); i++ {
		ec := elem.concat[i]

		if ec.IsVariable() {
			elemstr, err := sh.evalVariable(ec.val)

			if err != nil {
				return "", err
			}

			if len(elemstr) > 1 {
				return "", errors.New("Impossible to concat list variable and string")
			}

			if len(elemstr) == 1 {
				value = value + elemstr[0]
			}
		} else {
			value = value + ec.val
		}
	}

	return value, nil
}

func (sh *Shell) executeCmdAssignment(v *CmdAssignmentNode) error {
	var (
		varOut bytes.Buffer
		err    error
	)

	bkStdout := sh.stdout

	sh.SetStdout(&varOut)

	assign := v.Command()

	switch assign.Type() {
	case NodeCommand:
		err = sh.executeCommand(assign.(*CommandNode))
	case NodePipe:
		err = sh.executePipe(assign.(*PipeNode))
	case NodeFnInv:
		err = sh.executeFnInv(assign.(*FnInvNode))
	default:
		err = newError("Unexpected node in assignment: %s", assign.String())
	}

	sh.SetStdout(bkStdout)

	if err != nil {
		return err
	}

	strelems := append(make([]string, 0, 1), string(varOut.Bytes()))

	sh.SetVar(v.Name(), strelems)
	return nil
}

// Note(i4k): shit code
func (sh *Shell) executeAssignment(v *AssignmentNode) error {
	var err error

	elems := v.list
	strelems := make([]string, 0, len(elems))

	for i := 0; i < len(elems); i++ {
		elem := elems[i]

		if elem.IsConcat() {
			value, err := sh.concatElements(elem)

			if err != nil {
				return err
			}

			strelems = append(strelems, value)
		} else {
			if elem.IsVariable() {
				strelems, err = sh.evalVariable(elem.val)

				if err != nil {
					return err
				}
			} else {
				strelems = append(strelems, elem.val)
			}
		}
	}

	sh.SetVar(v.name, strelems)
	return nil
}

func (sh *Shell) executeCd(cd *CdNode) error {
	var (
		ok       bool
		pathlist []string
		pathStr  string
	)

	path := cd.Dir()

	if fn, ok := sh.binds["cd"]; ok {
		args := make([]*Arg, 0, 1)

		if path != nil {
			args = append(args, path)
		} else {
			// empty arg
			args = append(args, NewArg(0, ArgQuoted))
		}

		return sh.executeFn(fn, args)
	}

	if path == nil {
		if pathlist, ok = sh.GetEnv("HOME"); !ok {
			return errors.New("Nash don't know where to cd. No variable $HOME or $home set")
		}

		if len(pathlist) > 0 && pathlist[0] != "" {
			pathStr = pathlist[0]
		} else {
			return fmt.Errorf("Invalid $HOME value: %v", pathlist)
		}
	} else if path.IsVariable() {
		elemstr, err := sh.evalVariable(path.Value())

		if err != nil {
			return err
		}

		if len(elemstr) == 0 {
			return errors.New("Variable $path contains an empty list.")
		}

		if len(elemstr) > 1 {
			return fmt.Errorf("Variable $path contains a list: %q", elemstr)
		}

		pathStr = elemstr[0]
	} else if path.IsQuoted() || path.IsUnquoted() {
		pathStr = path.Value()
	} else if path.IsConcat() {
		pathConcat, err := sh.executeConcat(path)

		if err != nil {
			return err
		}

		pathStr += pathConcat
	} else {
		return fmt.Errorf("Exec error: Invalid path: %v", path)
	}

	err := os.Chdir(pathStr)

	if err != nil {
		return err
	}

	pwd, ok := sh.GetVar("PWD")

	if !ok {
		return fmt.Errorf("Variable $PWD is not set")
	}

	sh.SetVar("OLDPWD", pwd)
	sh.SetVar("PWD", append(make([]string, 0, 1), pathStr))
	return nil
}

func (sh *Shell) evalIfArguments(n *IfNode) (string, string, error) {
	var (
		lstr, rstr string
	)

	lvalue := n.Lvalue()
	rvalue := n.Rvalue()

	if len(lvalue.val) > 0 && lvalue.val[0] == '$' {
		variableValue, err := sh.evalVariable(lvalue.val)

		if err != nil {
			return "", "", err
		}

		if len(variableValue) > 1 {
			return "", "", fmt.Errorf("List is not comparable")
		} else if len(variableValue) == 0 {
			lstr = ""
		} else {
			lstr = variableValue[0]
		}
	} else {
		lstr = lvalue.val
	}

	if len(rvalue.val) > 0 && rvalue.val[0] == '$' {
		variableValue, err := sh.evalVariable(rvalue.val)

		if err != nil {
			return "", "", err
		}

		if len(variableValue) > 1 {
			return "", "", fmt.Errorf("List is not comparable")
		} else if len(variableValue) == 0 {
			rstr = ""
		} else {
			rstr = variableValue[0]
		}
	} else {
		rstr = rvalue.val
	}

	return lstr, rstr, nil
}

func (sh *Shell) executeIfEqual(n *IfNode) error {
	lstr, rstr, err := sh.evalIfArguments(n)

	if err != nil {
		return err
	}

	if lstr == rstr {
		return sh.ExecuteTree(n.IfTree())
	} else if n.ElseTree() != nil {
		return sh.ExecuteTree(n.ElseTree())
	}

	return nil
}

func (sh *Shell) executeIfNotEqual(n *IfNode) error {
	lstr, rstr, err := sh.evalIfArguments(n)

	if err != nil {
		return err
	}

	if lstr != rstr {
		return sh.ExecuteTree(n.IfTree())
	} else if n.ElseTree() != nil {
		return sh.ExecuteTree(n.ElseTree())
	}

	return nil
}

func (sh *Shell) executeFn(fn *Shell, args []*Arg) error {
	var err error

	if len(fn.argNames) != len(args) {
		return newError("Wrong number of arguments for function %s. Expected %d but found %d",
			fn.name, len(fn.argNames), len(args))
	}

	for i := 0; i < len(args); i++ {
		arg := args[i]
		argStr := ""
		argName := fn.argNames[i]

		if arg.IsConcat() {
			argStr, err = sh.executeConcat(arg)

			if err != nil {
				return err
			}
		} else {
			argStr = arg.Value()
		}

		if len(argStr) > 0 && argStr[0] == '$' {
			elemstr, err := sh.evalVariable(argStr)

			if err != nil {
				return err
			}

			fn.vars[argName] = elemstr
		} else {
			fn.vars[argName] = append(make([]string, 0, 1), argStr)
		}
	}

	return fn.Execute()
}

func (sh *Shell) executeFnInv(n *FnInvNode) error {
	if fn, ok := sh.GetFn(n.name); ok {
		return sh.executeFn(fn, n.args)
	}

	return newError("no such function '%s'", n.name)
}

func (sh *Shell) executeFnDecl(n *FnDeclNode) error {
	fn, err := NewShell(sh.debug)

	if err != nil {
		return err
	}

	fn.SetName(n.Name())
	fn.SetParent(sh)
	fn.SetStdout(sh.stdout)
	fn.SetStderr(sh.stderr)
	fn.SetStdin(sh.stdin)
	fn.SetRepr(n.String())

	args := n.Args()

	for i := 0; i < len(args); i++ {
		arg := args[i]

		fn.AddArgName(arg)
	}

	fn.SetTree(n.Tree())

	fnName := n.Name()

	if fnName == "" {
		fnName = fmt.Sprintf("lambda %d", strconv.Itoa(int(sh.lambdas)))
		sh.lambdas++
	}

	sh.fns[fnName] = fn

	sh.log("Function %s declared", fnName)

	return nil
}

func (sh *Shell) dumpVar(file io.Writer) {
	for n, v := range sh.vars {
		printVar(file, n, v)
	}
}

func (sh *Shell) dumpEnv(file io.Writer) {
	for n, _ := range sh.env {
		printEnv(file, n)
	}
}

func (sh *Shell) dumpFns(file io.Writer) {
	for _, f := range sh.fns {
		fmt.Fprintf(file, "%s\n\n", f.String())
	}
}

func (sh *Shell) dump(out io.Writer) {
	sh.dumpVar(out)
	sh.dumpEnv(out)
	sh.dumpFns(out)
}

func (sh *Shell) executeDump(n *DumpNode) error {
	var (
		err   error
		fname string
		file  *os.File
	)

	fnameArg := n.Filename()

	if fnameArg == nil {
		file = os.Stdout
		goto execDump
	} else if fnameArg.IsVariable() {
		variableList, err := sh.evalVariable(fnameArg.Value())

		if err != nil {
			return err
		}

		if len(variableList) == 0 || len(variableList) > 1 {
			return newError("Invalid variable used in dump")
		}

		fname = variableList[0]
	} else if fnameArg.IsConcat() {
		fname, err = sh.executeConcat(fnameArg)

		if err != nil {
			return err
		}
	} else {
		fname = fnameArg.Value()
	}

	file, err = os.OpenFile(fname, os.O_CREATE|os.O_RDWR, 0644)

	if err != nil {
		return err
	}

execDump:
	sh.dump(file)

	return nil
}

func (sh *Shell) executeBindFn(n *BindFnNode) error {
	if fn, ok := sh.GetFn(n.Name()); ok {
		sh.binds[n.CmdName()] = fn
	} else {
		return newError("No such function '$s'", n.Name())
	}

	return nil
}

func (sh *Shell) executeIf(n *IfNode) error {
	op := n.Op()

	if op == "==" {
		return sh.executeIfEqual(n)
	} else if op == "!=" {
		return sh.executeIfNotEqual(n)
	}

	return fmt.Errorf("Invalid operation '%s'.", op)
}

func nashdAutoDiscover() string {
	path, err := os.Readlink("/proc/self/exe")

	if err != nil {
		path = os.Args[0]

		if _, err := os.Stat(path); err != nil {
			return ""
		}
	}

	return path
}
