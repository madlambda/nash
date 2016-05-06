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

	// Shell is the core data structure.
	Shell struct {
		debug     bool
		log       LogFn
		nashdPath string
		dotDir    string

		stdin  io.Reader
		stdout io.Writer
		stderr io.Writer

		env       Env
		vars      Var
		multiline bool
	}
)

const (
	logNS = "nash.Shell"
)

// NewShell creates a new shell object
func NewShell(debug bool) *Shell {
	return &Shell{
		log:       NewLog(logNS, debug),
		nashdPath: nashdAutoDiscover(),
		stdout:    os.Stdout,
		stderr:    os.Stderr,
		stdin:     os.Stdin,
		env:       NewEnv(),
		vars:      make(Var),
	}
}

// NewEnv creates a new environment from old one
func NewEnv() Env {
	env := make(Env)
	processEnv := os.Environ()

	env["*"] = os.Args

	for _, penv := range processEnv {
		p := strings.Split(penv, "=")

		if len(p) == 1 {
			env[p[0]] = make([]string, 0, 1)
		} else if len(p) > 1 {
			env[p[0]] = append(make([]string, 0, 256), p[1:]...)
		}
	}

	env["PID"] = append(make([]string, 0, 1), strconv.Itoa(os.Getpid()))

	return env
}

func (sh *Shell) Env() Env {
	return sh.env
}

func (sh *Shell) SetEnv(env Env) {
	sh.env = env
}

// Prompt returns the environment prompt or the default one
func (sh *Shell) Prompt() string {
	if sh.env["PROMPT"] != nil && len(sh.env["PROMPT"]) > 0 {
		return sh.env["PROMPT"][0]
	}

	return "\033[31mλ>\033[0m "
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
func (sh *Shell) Execute(path string) error {
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
	return sh.Execute(node.path.val)
}

func (sh *Shell) executeShowEnv(node *ShowEnvNode) error {
	envVars := buildenv(sh.env)
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
	if v, ok := sh.vars[a[1:]]; ok {
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

	if varValue, ok = sh.vars[varName]; !ok {
		return fmt.Errorf("Variable '%s' not set", varName)
	}

	sh.env[varName] = varValue

	return nil
}

func (sh *Shell) concatElements(elem ElemNode) (string, error) {
	value := ""

	for j := 0; j < len(elem.concats); j++ {
		ec := elem.concats[j]

		if len(ec) > 0 && ec[0] == '$' {
			elemstr, err := sh.evalVariable(elem.concats[j])
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
			value = value + ec
		}
	}

	return value, nil
}

// Note(i4k): shit code
func (sh *Shell) executeAssignment(v *AssignmentNode) error {
	elems := v.list
	strelems := make([]string, 0, len(elems))

	for i := 0; i < len(elems); i++ {
		elem := elems[i]

		if len(elem.concats) > 0 {
			value, err := sh.concatElements(elem)

			if err != nil {
				return err
			}

			strelems = append(strelems, value)
		} else {
			strelems = append(strelems, elem.elem)
		}
	}

	sh.vars[v.name] = strelems
	return nil
}

func (sh *Shell) executeCd(cd *CdNode) error {
	var (
		ok       bool
		pathlist []string
	)

	path := cd.Dir()

	if path == "" {
		if pathlist, ok = sh.env["HOME"]; !ok {
			return errors.New("Nash don't know where to cd. No variable $HOME or $home set")
		}

		if len(pathlist) > 0 && pathlist[0] != "" {
			path = pathlist[0]
		} else {
			return fmt.Errorf("Invalid $HOME value: %v", pathlist)
		}
	} else if path[0] == '$' {
		elemstr, err := sh.evalVariable(path)

		if err != nil {
			return err
		}

		if len(elemstr) == 0 {
			return errors.New("Variable $path contains an empty list.")
		}

		if len(elemstr) > 1 {
			return fmt.Errorf("Variable $path contains a list: %q", elemstr)
		}

		path = elemstr[0]
	}

	return os.Chdir(path)
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
