package sysexec

import (
	"fmt"
	"strings"
)

// ErrNeedsElevation is returned when a privileged command cannot run because the
// process lacks the rights and the platform offers no way to ask for them.
type ErrNeedsElevation struct {
	Hint string
}

func (e *ErrNeedsElevation) Error() string { return e.Hint }

// shellQuote renders args as a POSIX shell command line, single-quoting each
// element so that no user-supplied value can break out into the shell.
func shellQuote(args []string) string {
	quoted := make([]string, 0, len(args))
	for _, a := range args {
		quoted = append(quoted, "'"+strings.ReplaceAll(a, "'", `'\''`)+"'")
	}
	return strings.Join(quoted, " ")
}

// appleScriptQuote renders s as an AppleScript string literal body.
func appleScriptQuote(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return s
}

// joinAnd renders several argv slices as one POSIX shell command line chained
// with &&, so a batch stops at the first failing command. Each argument is
// single-quoted, so no user-supplied value can break out into the shell.
func joinAnd(cmds [][]string) string {
	parts := make([]string, 0, len(cmds))
	for _, cmd := range cmds {
		if len(cmd) == 0 {
			continue
		}
		parts = append(parts, shellQuote(cmd))
	}
	return strings.Join(parts, " && ")
}

func elevationHint(action string) error {
	return &ErrNeedsElevation{Hint: fmt.Sprintf(
		"%s için yönetici yetkisi gerekiyor. Uygulamayı yönetici olarak yeniden başlatın.", action)}
}
