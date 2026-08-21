//go:build darwin

package sysexec

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

// IsElevated reports whether the process already runs as root.
func IsElevated() bool { return os.Geteuid() == 0 }

// RunPrivileged runs a command with administrator rights. When the process is
// not already root it is re-issued through osascript, which shows the standard
// macOS authorisation dialog. Arguments are shell-quoted, so user-entered values
// cannot escape into the shell that AppleScript spawns.
func RunPrivileged(ctx context.Context, name string, args ...string) error {
	if IsElevated() {
		return Run(ctx, name, args...)
	}

	// `do shell script` runs with a minimal PATH; resolve the binary up front.
	resolved, err := exec.LookPath(name)
	if err != nil {
		resolved = name
	}

	shellCmd := shellQuote(append([]string{resolved}, args...))
	script := fmt.Sprintf(`do shell script "%s" with administrator privileges`,
		appleScriptQuote(shellCmd))

	if _, err := Output(ctx, "osascript", "-e", script); err != nil {
		return err
	}
	return nil
}

// RunPrivilegedBatch runs several privileged commands behind a SINGLE
// authorisation prompt. Applying a static IP needs both -setmanual and
// -setdnsservers; issuing them separately would ask the user for their password
// twice. Commands are chained with && so the batch stops at the first failure.
func RunPrivilegedBatch(ctx context.Context, cmds [][]string) error {
	if len(cmds) == 0 {
		return nil
	}

	if IsElevated() {
		for _, cmd := range cmds {
			if len(cmd) == 0 {
				continue
			}
			if err := Run(ctx, cmd[0], cmd[1:]...); err != nil {
				return err
			}
		}
		return nil
	}

	// `do shell script` runs with a minimal PATH; resolve every binary up front.
	resolved := make([][]string, 0, len(cmds))
	for _, cmd := range cmds {
		if len(cmd) == 0 {
			continue
		}
		bin, err := exec.LookPath(cmd[0])
		if err != nil {
			bin = cmd[0]
		}
		resolved = append(resolved, append([]string{bin}, cmd[1:]...))
	}

	script := fmt.Sprintf(`do shell script "%s" with administrator privileges`,
		appleScriptQuote(joinAnd(resolved)))

	_, err := Output(ctx, "osascript", "-e", script)
	return err
}
