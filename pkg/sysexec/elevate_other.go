//go:build !windows && !darwin

package sysexec

import (
	"context"
	"os"
)

// IsElevated reports whether the process already runs as root.
func IsElevated() bool { return os.Geteuid() == 0 }

// RunPrivileged runs a command that needs administrator rights. These platforms
// are unsupported (pkg/network refuses to configure anything here), so there is
// no elevation mechanism to reach for: either the process is already root or the
// caller is told to restart elevated.
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
