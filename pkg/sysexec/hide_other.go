//go:build !windows

package sysexec

import "os/exec"

func hideWindow(_ *exec.Cmd) {}
