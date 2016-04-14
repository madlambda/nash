package cnt

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
)

var debugLevel bool

func debug(format string, a ...interface{}) {
	if debugLevel {
		fmt.Printf(format, a...)
	}
}

// ExecuteString executes the commands specified by string content
func ExecuteString(path, content string, debugval bool) error {
	debugLevel = debugval

	parser := NewParser(path, content)

	tr, err := parser.Parse()

	if err != nil {
		return err
	}

	return ExecuteTree(tr, debugval)
}

// Execute the cnt file at given path
func Execute(path string, debugval bool) error {
	content, err := ioutil.ReadFile(path)

	if err != nil {
		return err
	}

	return ExecuteString(path, string(content), debugval)
}

// ExecuteTree evaluates the given tree
func ExecuteTree(tr *Tree, debugval bool) error {
	if tr == nil || tr.Root == nil {
		return errors.New("nothing parsed")
	}

	debugLevel = debugval

	root := tr.Root

	for _, node := range root.Nodes {
		debug("Executing: %v\n", node)

		switch node.Type() {
		case NodeComment:
			continue
		case NodeCommand:
			err := execute(node.(*CommandNode))

			if err != nil {
				return err
			}
		case NodeRfork:
			err := executeRfork(node.(*RforkNode))

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

func execute(c *CommandNode) error {
	var (
		err error
		out bytes.Buffer
	)

	cmdPath := c.name

	if c.name[0] != '/' {
		cmdPath, err = exec.LookPath(c.name)

		if err != nil {
			return err
		}
	}

	debug("Executing: %s\n", cmdPath)

	args := make([]string, len(c.args))

	for i := 0; i < len(c.args); i++ {
		args[i] = c.args[i].val
	}

	cmd := exec.Command(cmdPath, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = &out

	err = cmd.Start()

	if err != nil {
		return err
	}

	err = cmd.Wait()

	if err != nil {
		return err
	}

	fmt.Printf("%s", out.Bytes())

	return nil
}

func executeCd(cd *CdNode) error {
	path, err := cd.Dir()

	if err != nil {
		return err
	}

	return os.Chdir(path)
}
