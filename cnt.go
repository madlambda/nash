package cnt

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
)

type (
	Shell struct {
		debug    bool
		log      LogFn
		cntdPath string

		stdin  io.Reader
		stdout io.Writer
		stderr io.Writer
	}
)

const logNS = "cnt.Shell"

func NewShell(debug bool) *Shell {
	return &Shell{
		log:      NewLog(logNS, debug),
		cntdPath: cntdAutoDiscover(),
		stdout:   os.Stdout,
		stderr:   os.Stderr,
		stdin:    os.Stdin,
	}
}

func (sh *Shell) SetDebug(debug bool) {
	sh.log = NewLog(logNS, debug)
}

func (sh *Shell) SetCntdPath(path string) {
	sh.cntdPath = path
}

func (sh *Shell) SetStdin(in io.Reader) {
	sh.stdin = in
}

func (sh *Shell) SetStdout(out io.Writer) {
	sh.stdout = out
}

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

// Execute the cnt file at given path
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
			continue
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

func (sh *Shell) execute(c *CommandNode) error {
	var (
		err         error
		ignoreError bool
	)

	cmdPath := c.name

	if c.name[0] == '-' {
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
		args[i] = c.args[i].val
	}

	cmd := exec.Command(cmdPath, args...)
	cmd.Stdin = sh.stdin
	cmd.Stdout = sh.stdout
	cmd.Stderr = sh.stderr

	err = cmd.Start()

	if err != nil {
		return err
	}

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

func cntdAutoDiscover() string {
	path, err := os.Readlink("/proc/self/exe")

	if err != nil {
		path = os.Args[0]

		if _, err := os.Stat(path); err != nil {
			return ""
		}
	}

	return path
}
