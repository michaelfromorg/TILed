package til

import (
	"bytes"
	"os/exec"
)

// Command represents a command to be run
type Command struct {
	*exec.Cmd
}

// NewCommand creates a new command
func NewCommand(name string, args ...string) *Command {
	return &Command{exec.Command(name, args...)}
}

// RunStdOut runs the command and returns the stdout
func (c *Command) RunStdOut() (string, error) {
	var stdout bytes.Buffer
	c.Stdout = &stdout
	err := c.Run()
	return stdout.String(), err
}

// RunStdErr runs the command and returns the stderr
func (c *Command) RunStdErr() (string, error) {
	var stderr bytes.Buffer
	c.Stderr = &stderr
	err := c.Run()
	return stderr.String(), err
}

// RunOutput runs the command and returns both stdout and stderr
func (c *Command) RunOutput() (string, string, error) {
	var stdout, stderr bytes.Buffer
	c.Stdout = &stdout
	c.Stderr = &stderr
	err := c.Run()
	return stdout.String(), stderr.String(), err
}
