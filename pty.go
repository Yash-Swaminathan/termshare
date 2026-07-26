package main

import (
	"os"
	"os/exec"

	"github.com/creack/pty"
)

// StartPTY launches a shell attached to a fresh pty.
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

// SetPTYSize resizes the pty so the shell reflows to the client window.
func SetPTYSize(f *os.File, cols, rows uint16) error {
	return pty.Setsize(f, &pty.Winsize{Cols: cols, Rows: rows})
}
