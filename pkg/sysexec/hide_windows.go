//go:build windows

package sysexec

import (
	"os/exec"
	"syscall"
)

// hideWindow stops the child process from flashing a console window, which is
// what every netsh/ping call would otherwise do in a GUI app.
func hideWindow(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
}
