//go:build windows

package sysexec

import (
	"context"

	"golang.org/x/sys/windows"
)

// IsElevated reports whether the process token is elevated. The app manifest
// requests elevation at launch, so this is normally true; it is false when the
// binary is started from a stripped manifest or a non-admin account.
func IsElevated() bool {
	return windows.GetCurrentProcessToken().IsElevated()
}

// RunPrivileged runs a command that needs administrator rights. Windows cannot
// elevate a single child process without a UAC re-launch, so the whole app is
// manifested as requireAdministrator; here we only verify and report.
func RunPrivileged(ctx context.Context, name string, args ...string) error {
	if !IsElevated() {
		return elevationHint(name)
	}
	return Run(ctx, name, args...)
}

// RunPrivilegedBatch runs several privileged commands in order, stopping at the
// first failure. Only macOS needs to collapse them into one prompt; elsewhere
// the process is either already elevated or cannot elevate at all.
func RunPrivilegedBatch(ctx context.Context, cmds [][]string) error {
	for _, cmd := range cmds {
		if len(cmd) == 0 {
			continue
		}
		if err := RunPrivileged(ctx, cmd[0], cmd[1:]...); err != nil {
			return err
		}
	}
	return nil
}
