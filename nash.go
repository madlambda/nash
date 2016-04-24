package nash

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type (
	// Env is the environment map of lists
	Env map[string][]string

	// Shell is the core data structure.
	Shell struct {
		debug     bool
		log       LogFn
		nashdPath string

		stdin  io.Reader
		stdout io.Writer
		stderr io.Writer

		env       Env
		multiline bool
	}
)

const logNS = "nash.Shell"

// NewShell creates a new shell object
func NewShell(debug bool) *Shell {
	return &Shell{
		log:       NewLog(logNS, debug),
		nashdPath: nashdAutoDiscover(),
		stdout:    os.Stdout,
		stderr:    os.Stderr,
		stdin:     os.Stdin,
		env:       NewEnv(),
	}
}

// NewEnv creates a new environment with default values
func NewEnv() Env {
	env := make(Env)
	env["*"] = os.Args
	env["PID"] = append(make([]string, 0, 1), strconv.Itoa(os.Getpid()))
	env["HOME"] = append(make([]string, 0, 1), os.Getenv("HOME"))
	env["PATH"] = append(make([]string, 0, 128), os.Getenv("PATH"))

	if os.Getenv("PROMPT") != "" {
		env["PROMPT"] = append(make([]string, 0, 1), os.Getenv("PROMPT"))
	}

	return env
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
	if tr == nil || tr.Root == nil {
		return errors.New("nothing parsed")
	}

	root := tr.Root

	for _, node := range root.Nodes {
		sh.log("Executing: %v\n", node)

		switch node.Type() {
		case NodeComment:
			continue // ignore comment
		case NodeAssignment:
			err := sh.executeAssignment(node.(*AssignmentNode))

			if err != nil {
				return err
			}
		case NodeCommand:
			err := sh.execute(node.(*CommandNode))

			if err != nil {
				return err
			}
		case NodeRfork:
			err := sh.executeRfork(node.(*RforkNode))

			if err != nil {
				return err
			}
		case NodeCd:
			err := executeCd(node.(*CdNode))

			if err != nil {
				return err
			}
		default:
			fmt.Printf("invalid command")
		}
	}

	return nil
}

func (sh *Shell) executeAssignment(v *AssignmentNode) error {
	sh.env[v.name] = v.list
	return nil
}

func (sh *Shell) execute(c *CommandNode) error {
	var (
		err         error
		ignoreError bool
	)

	cmdPath := c.name

	if len(c.name) > 1 && c.name[0] == '-' {
		ignoreError = true
		c.name = c.name[1:]

		sh.log("Ignoring error\n")
	}

	if c.name[0] != '/' {
		cmdPath, err = exec.LookPath(c.name)

		if err != nil {
			return err
		}
	}

	sh.log("Executing: %s\n", cmdPath)

	args := make([]string, len(c.args))

	for i := 0; i < len(c.args); i++ {
		argval := c.args[i].val

		// variable substitution
		if len(argval) > 0 && argval[0] == '$' && sh.env[argval[1:]] != nil {
			arglist := sh.env[argval[1:]]

			if len(arglist) == 1 {
				args[i] = arglist[0]
			} else if len(arglist) > 1 {
				args[i] = strings.Join(arglist, " ")
			}
		} else {
			args[i] = argval
		}
	}

	cmd := exec.Command(cmdPath, args...)

	stdoutDone := make(chan bool)
	stderrDone := make(chan bool)

	if sh.stdout != os.Stdout {
		stdout, err := cmd.StdoutPipe()

		if err != nil {
			return err
		}

		go func() {
			defer close(stdoutDone)

			io.Copy(sh.stdout, stdout)
		}()
	} else {
		close(stdoutDone)
		cmd.Stdout = sh.stdout
	}

	if sh.stderr != os.Stderr {
		stderr, err := cmd.StderrPipe()

		if err != nil {
			return err
		}

		go func() {
			defer close(stderrDone)
			io.Copy(sh.stderr, stderr)
		}()
	} else {
		close(stderrDone)
		cmd.Stderr = sh.stderr
	}

	cmd.Stdin = sh.stdin

	err = cmd.Start()

	if err != nil {
		return err
	}

	<-stdoutDone
	<-stderrDone

	err = cmd.Wait()

	if err != nil && !ignoreError {
		return err
	}

	return nil
}

func executeCd(cd *CdNode) error {
	path, err := cd.Dir()

	if err != nil {
		return err
	}

	return os.Chdir(path)
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