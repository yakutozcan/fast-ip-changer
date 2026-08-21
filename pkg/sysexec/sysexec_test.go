package sysexec

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestShellQuote(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"plain", []string{"networksetup", "-setdhcp", "Ethernet"}, `'networksetup' '-setdhcp' 'Ethernet'`},
		{"space in arg", []string{"networksetup", "-setdhcp", "Wi-Fi 2"}, `'networksetup' '-setdhcp' 'Wi-Fi 2'`},
		{"single quote", []string{"echo", "it's"}, `'echo' 'it'\''s'`},
		{"command substitution is inert", []string{"echo", "$(rm -rf /)"}, `'echo' '$(rm -rf /)'`},
		{"semicolon is inert", []string{"echo", "a; rm -rf /"}, `'echo' 'a; rm -rf /'`},
		{"backtick is inert", []string{"echo", "`whoami`"}, "'echo' '`whoami`'"},
		{"empty arg", []string{"echo", ""}, `'echo' ''`},
		{"no args", nil, ``},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shellQuote(tt.args); got != tt.want {
				t.Errorf("shellQuote(%q)\n got: %s\nwant: %s", tt.args, got, tt.want)
			}
		})
	}
}

// A quoted argument must never leave its single-quoted context: every quote in
// the result has to be either a delimiter or an escaped literal.
func TestShellQuoteNeverEscapesQuoting(t *testing.T) {
	hostile := []string{
		`'; rm -rf / #`,
		`'"'"'`,
		`\'`,
		"a'b'c",
	}
	for _, arg := range hostile {
		got := shellQuote([]string{arg})
		if !strings.HasPrefix(got, "'") || !strings.HasSuffix(got, "'") {
			t.Errorf("shellQuote(%q) = %s: not wrapped in single quotes", arg, got)
		}
		// Stripping the escaped-quote sequences must leave no bare quote behind.
		inner := strings.TrimSuffix(strings.TrimPrefix(got, "'"), "'")
		if strings.Contains(strings.ReplaceAll(inner, `'\''`, ""), "'") {
			t.Errorf("shellQuote(%q) = %s: contains an unescaped quote", arg, got)
		}
	}
}

func TestAppleScriptQuote(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"plain", `'networksetup' '-setdhcp'`, `'networksetup' '-setdhcp'`},
		{"double quote escaped", `say "hi"`, `say \"hi\"`},
		{"backslash escaped", `a\b`, `a\\b`},
		{"backslash before quote", `a\"b`, `a\\\"b`},
		{"empty", ``, ``},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := appleScriptQuote(tt.in); got != tt.want {
				t.Errorf("appleScriptQuote(%q)\n got: %s\nwant: %s", tt.in, got, tt.want)
			}
		})
	}
}

// Backslashes must be doubled before quotes are escaped, otherwise the
// backslash pass would mangle the escapes the quote pass just added.
func TestAppleScriptQuoteOrdering(t *testing.T) {
	got := appleScriptQuote(`"`)
	if got != `\"` {
		t.Fatalf("appleScriptQuote(%q) = %q, want %q", `"`, got, `\"`)
	}
}

func TestWithTimeoutAddsDeadline(t *testing.T) {
	ctx, cancel := WithTimeout(context.Background())
	defer cancel()

	deadline, ok := ctx.Deadline()
	if !ok {
		t.Fatal("WithTimeout did not set a deadline on a bare context")
	}
	if remaining := time.Until(deadline); remaining > DefaultTimeout {
		t.Errorf("deadline %v exceeds DefaultTimeout %v", remaining, DefaultTimeout)
	}
}

func TestWithTimeoutPreservesExistingDeadline(t *testing.T) {
	parent, parentCancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer parentCancel()

	ctx, cancel := WithTimeout(parent)
	defer cancel()

	deadline, ok := ctx.Deadline()
	if !ok {
		t.Fatal("derived context lost the parent deadline")
	}
	if remaining := time.Until(deadline); remaining > time.Second {
		t.Errorf("deadline %v was replaced by the default instead of inherited", remaining)
	}
}

func TestWithTimeoutHandlesNilParent(t *testing.T) {
	//nolint:staticcheck // deliberately passing nil to prove it does not panic
	ctx, cancel := WithTimeout(nil)
	defer cancel()

	if _, ok := ctx.Deadline(); !ok {
		t.Error("nil parent should still yield a deadline")
	}
}

// requireShell skips a test that needs a POSIX shell. The round-trip tests below
// are only meaningful against a real shell, and CI runs the suite on Windows too,
// where one is not guaranteed to be on PATH.
func requireShell(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("sh"); err != nil {
		t.Skip("no POSIX shell on PATH")
	}
}

func TestOutputReturnsStderrInError(t *testing.T) {
	requireShell(t)

	ctx, cancel := WithTimeout(context.Background())
	defer cancel()

	_, err := Output(ctx, "sh", "-c", "echo boom >&2; exit 1")
	if err == nil {
		t.Fatal("expected an error from a failing command")
	}
	if !strings.Contains(err.Error(), "boom") {
		t.Errorf("error %q does not carry stderr", err)
	}
}

func TestOutputReportsTimeout(t *testing.T) {
	requireShell(t)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := Output(ctx, "sh", "-c", "sleep 5")
	if err == nil {
		t.Fatal("expected a timeout error")
	}
	if !strings.Contains(err.Error(), "zaman aşımı") {
		t.Errorf("error %q is not reported as a timeout", err)
	}
}

func TestMergedFallsBackToStderr(t *testing.T) {
	requireShell(t)

	ctx, cancel := WithTimeout(context.Background())
	defer cancel()

	out, _ := Merged(ctx, "sh", "-c", "echo only-on-stderr >&2")
	if !strings.Contains(out, "only-on-stderr") {
		t.Errorf("Merged returned %q, want stderr content", out)
	}
}

func TestMergedPrefersStdout(t *testing.T) {
	requireShell(t)

	ctx, cancel := WithTimeout(context.Background())
	defer cancel()

	out, err := Merged(ctx, "sh", "-c", "echo from-stdout; echo from-stderr >&2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "from-stdout") {
		t.Errorf("Merged returned %q, want stdout content", out)
	}
	if strings.Contains(out, "from-stderr") {
		t.Errorf("Merged returned %q, should not append stderr when stdout is non-empty", out)
	}
}

func TestErrNeedsElevationMessage(t *testing.T) {
	var err error = &ErrNeedsElevation{Hint: "yönetici gerekli"}
	if err.Error() != "yönetici gerekli" {
		t.Errorf("Error() = %q", err.Error())
	}
}

// The quoting is only trustworthy if a real shell agrees, so round-trip a set of
// hostile arguments through sh and assert each one arrives verbatim.
func TestShellQuoteRoundTripsThroughRealShell(t *testing.T) {
	requireShell(t)

	args := []string{
		`plain`,
		`with space`,
		`it's`,
		`$(echo pwned)`,
		"`echo pwned`",
		`a; echo pwned`,
		`a && echo pwned`,
		`"double"`,
		`back\slash`,
		`$HOME`,
		`*`,
		`newline
here`,
	}

	for _, arg := range args {
		t.Run(arg, func(t *testing.T) {
			ctx, cancel := WithTimeout(context.Background())
			defer cancel()

			// printf %s emits the argument with no interpretation of its own.
			script := shellQuote([]string{"printf", "%s", arg})
			out, err := Output(ctx, "sh", "-c", script)
			if err != nil {
				t.Fatalf("shell rejected %s: %v", script, err)
			}
			if out != arg {
				t.Errorf("round trip changed the argument\n got: %q\nwant: %q\nvia: %s", out, arg, script)
			}
		})
	}
}

// AppleScript quoting sits on top of shell quoting, so verify the composed form
// survives an actual osascript parse (macOS only; `do shell script` without the
// elevation clause needs no privileges).
func TestAppleScriptQuoteComposesWithShellQuote(t *testing.T) {
	requireShell(t)

	if runtime.GOOS != "darwin" {
		t.Skip("osascript is only available on macOS")
	}

	args := []string{`it's`, `$(echo pwned)`, `"double"`, `back\slash`, `a; echo pwned`}
	for _, arg := range args {
		t.Run(arg, func(t *testing.T) {
			ctx, cancel := WithTimeout(context.Background())
			defer cancel()

			shellCmd := shellQuote([]string{"printf", "%s", arg})
			script := fmt.Sprintf(`do shell script "%s"`, appleScriptQuote(shellCmd))

			out, err := Output(ctx, "osascript", "-e", script)
			if err != nil {
				t.Fatalf("osascript rejected %s: %v", script, err)
			}
			if strings.TrimRight(out, "\n") != arg {
				t.Errorf("round trip changed the argument\n got: %q\nwant: %q", out, arg)
			}
		})
	}
}

func TestJoinAnd(t *testing.T) {
	tests := []struct {
		name string
		cmds [][]string
		want string
	}{
		{"single", [][]string{{"networksetup", "-setdhcp", "Wi-Fi"}}, `'networksetup' '-setdhcp' 'Wi-Fi'`},
		{
			"two chained",
			[][]string{{"networksetup", "-setmanual", "Wi-Fi", "10.0.0.2", "255.255.255.0", "10.0.0.1"}, {"networksetup", "-setdnsservers", "Wi-Fi", "1.1.1.1"}},
			`'networksetup' '-setmanual' 'Wi-Fi' '10.0.0.2' '255.255.255.0' '10.0.0.1' && 'networksetup' '-setdnsservers' 'Wi-Fi' '1.1.1.1'`,
		},
		{"empty slices skipped", [][]string{{"a"}, {}, {"b"}}, `'a' && 'b'`},
		{"none", nil, ``},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := joinAnd(tt.cmds); got != tt.want {
				t.Errorf("joinAnd()\n got: %s\nwant: %s", got, tt.want)
			}
		})
	}
}

// A batch must stop at the first failure rather than running the rest, so a
// rejected -setmanual can never be followed by a DNS change.
func TestJoinAndShortCircuitsInRealShell(t *testing.T) {
	requireShell(t)

	ctx, cancel := WithTimeout(context.Background())
	defer cancel()

	script := joinAnd([][]string{
		{"printf", "first"},
		{"false"},
		{"printf", "third"},
	})

	out, err := Merged(ctx, "sh", "-c", script)
	if err == nil {
		t.Error("expected a non-zero exit from the chained batch")
	}
	if out != "first" {
		t.Errorf("output = %q, want only %q — the batch did not stop at the failure", out, "first")
	}
}

func TestRunPrivilegedBatchNoCommands(t *testing.T) {
	ctx, cancel := WithTimeout(context.Background())
	defer cancel()

	if err := RunPrivilegedBatch(ctx, nil); err != nil {
		t.Errorf("empty batch returned %v, want nil", err)
	}
}

// Run is the discard-output wrapper the network package uses, so it has to carry
// the same stderr-bearing error that Output does.
func TestRunCarriesStderr(t *testing.T) {
	requireShell(t)

	ctx, cancel := WithTimeout(context.Background())
	defer cancel()

	err := Run(ctx, "sh", "-c", "echo run-boom >&2; exit 3")
	if err == nil {
		t.Fatal("expected an error from a failing command")
	}
	if !strings.Contains(err.Error(), "run-boom") {
		t.Errorf("error %q does not carry stderr", err)
	}
}

// A cancelled run has to be distinguishable from a timeout: the UI shows the two
// differently, and only one of them is the user's own doing.
func TestRunReportsCancellation(t *testing.T) {
	requireShell(t)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()
	defer cancel()

	err := Run(ctx, "sh", "-c", "sleep 5")
	if err == nil {
		t.Fatal("expected an error from a cancelled command")
	}
	if !strings.Contains(err.Error(), "iptal edildi") {
		t.Errorf("error %q is not reported as a cancellation", err)
	}
}
