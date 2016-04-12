package cnt

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
)

var debugLevel int

func debug(format string, a ...interface{}) {
	if debugLevel > 0 {
		fmt.Printf(format, a...)
	}
}

// ExecuteString executes the commands specified by string content
func ExecuteString(path, content string, debugVal int) error {
	debugLevel = debugVal

	parser := NewParser(path, content)

	tr, err := parser.Parse()

	if err != nil {
		return err
	}

	if tr.Root == nil {
		return errors.New("nothing parsed")
	}

	root := tr.Root

	for _, node := range root.Nodes {
		debug("Executing: %v\n", node)

		switch node.Type() {
		case NodeComment:
			continue
		case NodeCommand:
			err := execute(node.(*CommandNode))

			if err != nil {
				fmt.Printf("Command failed: %s", err.Error())
				return err
			}
		case NodeRfork:
			err := executeRfork(node.(*RforkNode))

			if err != nil {
				return err
			}
		default:
			fmt.Printf("invalid command")
		}
	}

	return nil
}

// Execute the cnt file at given path
func Execute(path string, debugval int) error {
	content, err := ioutil.ReadFile(path)

	if err != nil {
		return err
	}

	return ExecuteString(path, string(content), debugval)
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
