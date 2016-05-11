package nash

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

type (
	// Env is the environment map of lists
	Env map[string][]string
	Var Env
	Fns map[string]*Shell

	// Shell is the core data structure.
	Shell struct {
		name      string
		debug     bool
		lambdas   uint
		log       LogFn
		nashdPath string
		dotDir    string

		stdin  io.Reader
		stdout io.Writer
		stderr io.Writer

		argNames  []string
		env       Env
		vars      Var
		fns       Fns
		multiline bool

		root   *Tree
		parent *Shell
	}
)

const (
	logNS     = "nash.Shell"
	defPrompt = "\033[31mλ>\033[0m "
)

// NewShell creates a new shell object
func NewShell(debug bool) *Shell {
	env, vars := NewEnv()

	if env["PROMPT"] == nil {
		env["PROMPT"] = append(make([]string, 0, 1), defPrompt)
		vars["PROMPT"] = append(make([]string, 0, 1), defPrompt)
	}

	return &Shell{
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
		argNames:  make([]string, 0, 16),
	}
}

// NewEnv creates a new environment from old one
func NewEnv() (Env, Var) {
	env := make(Env)
	vars := make(Var)
	processEnv := os.Environ()

	env["*"] = os.Args
	vars["*"] = os.Args

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

	return env, vars
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
		case NodeCommand:
			err = sh.executeCommand(node.(*CommandNode))
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
	return sh.ExecuteFile(node.path.val)
}

func (sh *Shell) executeShowEnv(node *ShowEnvNode) error {
	envVars := buildenv(sh.Environ())
	for _, e := range envVars {
		fmt.Fprintf(sh.stdout, "%s\n", e)
	}

	return nil
}

func (sh *Shell) executeCommand(c *CommandNode) error {
	cmd, err := NewCommand(c.name, sh)

	if err != nil {
		type IgnoreError interface {
			IgnoreError() bool
		}

		if errIgnore, ok := err.(IgnoreError); ok && errIgnore.IgnoreError() {
			fmt.Fprintf(sh.stderr, "ERROR: %s\n", err.Error())

			return nil
		}

		return err
	}

	err = cmd.SetArgs(c.args)

	if err != nil {
		return err
	}

	err = cmd.SetRedirects(c.redirs)

	if err != nil {
		return err
	}

	return cmd.Execute()
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
		for i := 0; i < len(path.concat); i++ {
			part := path.concat[i]

			if part.IsConcat() {
				return errors.New("Nested concat is not allowed")
			}

			if part.IsVariable() {
				partValues, err := sh.evalVariable(part.Value())

				if err != nil {
					return err
				}

				if len(partValues) > 1 {
					return fmt.Errorf("Concat of list variables is not allowed: %s = %v", part.Value(), partValues)
				} else if len(partValues) == 0 {
					return fmt.Errorf("Variable %s not set", part.Value())
				}

				pathStr += partValues[0]
			} else {
				pathStr = pathStr + part.Value()
			}
		}
	} else {
		return fmt.Errorf("Exec error: Invalid path: %v", path)
	}

	fmt.Printf("Executing cd into %s\n", pathStr)

	return os.Chdir(pathStr)
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

func (sh *Shell) executeFnInv(n *FnInvNode) error {
	if fn, ok := sh.fns[n.name]; ok {
		if len(fn.argNames) != len(n.args) {
			return newError("Wrong number of arguments for function %s. Expected %d but found %d",
				n.name, len(fn.argNames), len(n.args))
		}

		for i := 0; i < len(n.args); i++ {
			arg := n.args[i]
			argName := fn.argNames[i]

			if len(arg) > 0 && arg[0] == '$' {
				elemstr, err := sh.evalVariable(arg)

				if err != nil {
					return err
				}

				fn.vars[argName] = elemstr
			} else {
				fn.vars[argName] = append(make([]string, 0, 1), arg)
			}
		}

		return fn.Execute()
	}

	return newError("no such function '%s'", n.name)
}

func (sh *Shell) executeFnDecl(n *FnDeclNode) error {
	fn := NewShell(sh.debug)
	fn.SetName(n.Name())
	fn.SetParent(sh)
	fn.SetStdout(sh.stdout)
	fn.SetStderr(sh.stderr)
	fn.SetStdin(sh.stdin)

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
