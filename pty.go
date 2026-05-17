package main

import (
	"os"
	"os/exec"
)

// pty.go owns the shell process and its pty master file.
//
// Pointers:
//   - import the lib:  "github.com/creack/pty"   (add to the import block
//     above once you use it — Go won't compile an unused import).
//   - cmd := exec.Command("bash")  // or os.Getenv("SHELL"), fallback "sh"
//   - f, err := pty.Start(cmd)
//       f is an *os.File. Read(f) = shell OUTPUT. Write(f) = shell INPUT.
//       pty.Start also calls cmd.Start for you — don't call it yourself.
//   - resize later with pty.Setsize(f, &pty.Winsize{Rows, Cols}).
//   - pty.Start is Unix-only. You're in WSL, so it builds.
//
// Lifecycle note: reading from f returns an error the moment the shell
// process exits. That error is the signal to tear the Session down
// (see session.go readPTY) — it's expected, not a bug.

// StartPTY launches a shell attached to a fresh pty and returns the
// master file plus the cmd (keep cmd so you can cmd.Wait / cmd.Process.Kill
// on teardown).
func StartPTY() (*os.File, *exec.Cmd, error) {
	// TODO:
	//   shell := os.Getenv("SHELL"); if shell == "" { shell = "bash" }
	//   cmd := exec.Command(shell)
	//   f, err := pty.Start(cmd)
	//   return f, cmd, err
	panic("TODO: implement StartPTY")
}
