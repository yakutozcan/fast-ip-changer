// Package sysexec wraps os/exec with the behaviour every platform helper in this
// app needs: a timeout on every call, no console window flash on Windows, and
// error values that carry the command's stderr instead of a bare "exit status 1".
package sysexec

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// DefaultTimeout bounds a single system command. Anything slower than this
// (an unreachable traceroute hop, a wedged netsh) is reported as a timeout
// rather than hanging the UI.
const DefaultTimeout = 20 * time.Second

// Command builds an *exec.Cmd bound to ctx, with the platform tweaks applied.
func Command(ctx context.Context, name string, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, name, args...)
	hideWindow(cmd)
	return cmd
}

// Output runs a command and returns its stdout. On failure the error carries
// stderr so callers can surface something more useful than the exit code.
func Output(ctx context.Context, name string, args ...string) (string, error) {
	cmd := Command(ctx, name, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return stdout.String(), wrap(ctx, name, err, stderr.String())
	}
	return stdout.String(), nil
}

// Merged runs a command and returns stdout, falling back to stderr when the
// command wrote its diagnostics there. Used by the ping/traceroute views, which
// want the raw text either way.
func Merged(ctx context.Context, name string, args ...string) (string, error) {
	cmd := Command(ctx, name, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	out := stdout.String()
	if strings.TrimSpace(out) == "" && stderr.Len() > 0 {
		out = stderr.String()
	}
	if err != nil {
		return out, wrap(ctx, name, err, stderr.String())
	}
	return out, nil
}

// Run runs a command and discards its output, keeping only the error.
func Run(ctx context.Context, name string, args ...string) error {
	_, err := Output(ctx, name, args...)
	return err
}

// WithTimeout derives a context capped at DefaultTimeout when the parent has no
// deadline of its own. The returned cancel func must always be called.
func WithTimeout(parent context.Context) (context.Context, context.CancelFunc) {
	if parent == nil {
		parent = context.Background()
	}
	if _, ok := parent.Deadline(); ok {
		return context.WithCancel(parent)
	}
	return context.WithTimeout(parent, DefaultTimeout)
}

func wrap(ctx context.Context, name string, err error, stderr string) error {
	if ctx.Err() == context.DeadlineExceeded {
		return fmt.Errorf("%s zaman aşımına uğradı", name)
	}
	if ctx.Err() == context.Canceled {
		return fmt.Errorf("%s iptal edildi", name)
	}
	if msg := strings.TrimSpace(stderr); msg != "" {
		return fmt.Errorf("%s başarısız: %s", name, msg)
	}
	return fmt.Errorf("%s başarısız: %w", name, err)
}
