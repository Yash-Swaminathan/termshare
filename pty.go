package main

import (
	"os"
	"os/exec"

	"github.com/creack/pty"
)

// StartPTY launches a shell attached to a fresh pty and returns the master file plus the cmd.
func StartPTY() (*os.File, *exec.Cmd, error) {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "bash"
	}
	cmd := exec.Command(shell)
	f, err := pty.Start(cmd)
	if err != nil {
		return nil, nil, err
	}
	return f, cmd, nil
}
