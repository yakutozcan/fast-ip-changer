// Package diagnostics runs the connectivity probes the UI shows: ping,
// traceroute and the start-up quick check. Every system command goes through
// pkg/sysexec so it inherits a deadline and never flashes a console window on
// Windows.
package diagnostics

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/yakutozcan/fast-ip-changer/pkg/sysexec"
)

const (
	// pingWaitPerPacket is the per-reply wait handed to ping. It is also the
	// budget used to derive the overall deadline of a run.
	pingWaitPerPacket = 2 * time.Second
	// pingOverhead covers name resolution and process start-up on top of the
	// per-packet budget.
	pingOverhead = 5 * time.Second

	// quickPingWait keeps the start-up quick check snappy: a gateway that does
	// not answer within a second is reported as down.
	quickPingWait = 1 * time.Second
	// quickCheckTimeout bounds the whole quick check, which runs its probes
	// concurrently.
	quickCheckTimeout = 5 * time.Second

	// traceRouteMaxHops mirrors the -m/-h argument below.
	traceRouteMaxHops = 15
	// traceRouteTimeout is generous enough for 15 hops of unanswered probes but
	// still finite, so a black-holed route cannot wedge the UI.
	traceRouteTimeout = 45 * time.Second

	publicIPTimeout = 2 * time.Second
	// publicIPMaxBytes caps the response read; an IPv6 literal is 45 bytes.
	publicIPMaxBytes = 64

	internetProbeHost = "1.1.1.1"
)

// QuickCheckResult is the payload the dashboard renders on start-up.
type QuickCheckResult struct {
	GatewayOk       bool   `json:"gatewayOk"`
	GatewayLatency  string `json:"gatewayLatency"`
	InternetOk      bool   `json:"internetOk"`
	InternetLatency string `json:"internetLatency"`
	PublicIP        string `json:"publicIp"`
}

// run tracks the single long-running diagnostic the UI can cancel.
type run struct {
	cancel context.CancelFunc
}

var (
	runMu   sync.Mutex
	current *run
)

// publicIPURL is only contacted when the caller explicitly opts in. It is a var
// rather than a const so the tests can redirect it at a local test server; the
// app never reassigns it.
var publicIPURL = "https://api.ipify.org"

var publicIPClient = &http.Client{Timeout: publicIPTimeout}

// begin registers a cancellable diagnostic as the current one, aborting any
// previous run, and returns the context plus a finish func that must be
// deferred by the caller.
func begin(parent context.Context, timeout time.Duration) (context.Context, func()) {
	if parent == nil {
		parent = context.Background()
	}
	ctx, cancel := context.WithTimeout(parent, timeout)
	r := &run{cancel: cancel}

	runMu.Lock()
	prev := current
	current = r
	runMu.Unlock()

	if prev != nil {
		prev.cancel()
	}

	return ctx, func() {
		runMu.Lock()
		// Only clear the slot if a newer run has not taken it over already.
		if current == r {
			current = nil
		}
		runMu.Unlock()
		cancel()
	}
}

// Cancel aborts the ping or traceroute currently in flight. It is a no-op when
// nothing is running, so the UI can wire it straight to a button.
func Cancel() {
	runMu.Lock()
	r := current
	current = nil
	runMu.Unlock()

	if r != nil {
		r.cancel()
	}
}

// pingArgs builds the ping arguments for goos. The flag spelling and, more
// importantly, the unit of the per-reply timeout differ per platform: Windows
// -w and macOS -W are milliseconds, while Linux -W is whole seconds.
func pingArgs(goos, host string, count int, wait time.Duration) []string {
	if count <= 0 {
		count = 1
	}

	ms := int(wait.Milliseconds())
	if ms < 1 {
		ms = 1
	}

	switch goos {
	case "windows":
		return []string{"-n", strconv.Itoa(count), "-w", strconv.Itoa(ms), host}
	case "darwin":
		return []string{"-c", strconv.Itoa(count), "-W", strconv.Itoa(ms), host}
	default:
		// Round up so a sub-second wait never collapses to "-W 0".
		secs := int((wait + time.Second - 1) / time.Second)
		if secs < 1 {
			secs = 1
		}
		return []string{"-c", strconv.Itoa(count), "-W", strconv.Itoa(secs), host}
	}
}

func traceRouteArgs(goos, host string) (string, []string) {
	hops := strconv.Itoa(traceRouteMaxHops)
	if goos == "windows" {
		return "tracert", []string{"-d", "-h", hops, host}
	}
	return "traceroute", []string{"-m", hops, "-q", "1", "-w", "1", host}
}

// validateHost rejects empty targets and values that would be read as flags by
// ping/traceroute.
func validateHost(host string) (string, error) {
	host = strings.TrimSpace(host)
	if host == "" {
		return "", errors.New("hedef IP veya adres boş olamaz")
	}
	if strings.HasPrefix(host, "-") || strings.ContainsAny(host, " \t\r\n") {
		return "", fmt.Errorf("geçersiz hedef adres: %s", host)
	}
	return host, nil
}

// PingHost pings host count times and returns the raw command output, which the
// UI shows verbatim. A host that does not answer is not an error: ping exits
// non-zero but still explains itself in its output.
func PingHost(ctx context.Context, host string, count int) (string, error) {
	host, err := validateHost(host)
	if err != nil {
		return "", err
	}
	if count <= 0 {
		count = 4
	}

	timeout := time.Duration(count)*pingWaitPerPacket + pingOverhead
	ctx, finish := begin(ctx, timeout)
	defer finish()

	out, err := sysexec.Merged(ctx, "ping", pingArgs(runtime.GOOS, host, count, pingWaitPerPacket)...)
	if cerr := contextError(ctx, "Ping"); cerr != nil {
		return out, cerr
	}
	if err != nil && strings.TrimSpace(out) == "" {
		return "", err
	}
	return out, nil
}

// TraceRouteHost traces the route to host and returns the raw command output.
func TraceRouteHost(ctx context.Context, host string) (string, error) {
	host, err := validateHost(host)
	if err != nil {
		return "", err
	}

	ctx, finish := begin(ctx, traceRouteTimeout)
	defer finish()

	name, args := traceRouteArgs(runtime.GOOS, host)
	out, err := sysexec.Merged(ctx, name, args...)
	if cerr := contextError(ctx, "Traceroute"); cerr != nil {
		return out, cerr
	}
	if err != nil && strings.TrimSpace(out) == "" {
		return "", err
	}
	return out, nil
}

// contextError turns a cancelled or expired run into a message the UI can show.
func contextError(ctx context.Context, label string) error {
	switch {
	case errors.Is(ctx.Err(), context.Canceled):
		return fmt.Errorf("%s iptal edildi", label)
	case errors.Is(ctx.Err(), context.DeadlineExceeded):
		return fmt.Errorf("%s zaman aşımına uğradı", label)
	default:
		return nil
	}
}

// QuickCheck probes the gateway and the internet, and — only when the caller
// opts in — looks up the public IP through a third-party service. The three
// probes run concurrently because this sits on the app's start-up path.
func QuickCheck(ctx context.Context, gateway string, checkPublicIP bool) QuickCheckResult {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, quickCheckTimeout)
	defer cancel()

	var (
		wg         sync.WaitGroup
		gwOk       bool
		gwLatency  string
		netOk      bool
		netLatency string
		publicIP   string
	)

	if gateway = strings.TrimSpace(gateway); gateway != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			gwOk, gwLatency = singlePing(ctx, gateway)
		}()
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		netOk, netLatency = singlePing(ctx, internetProbeHost)
	}()

	if checkPublicIP {
		wg.Add(1)
		go func() {
			defer wg.Done()
			publicIP = lookupPublicIP(ctx)
		}()
	}

	// Each goroutine owns one local; Wait publishes them all to this goroutine.
	wg.Wait()

	return QuickCheckResult{
		GatewayOk:       gwOk,
		GatewayLatency:  gwLatency,
		InternetOk:      netOk,
		InternetLatency: netLatency,
		PublicIP:        publicIP,
	}
}

// singlePing sends one probe to host. The first return says whether the host
// answered — that is, whether ping exited successfully; the second is the
// latency, which stays empty when the reply carried nothing parseable.
func singlePing(ctx context.Context, host string) (bool, string) {
	host, err := validateHost(host)
	if err != nil {
		return false, ""
	}

	out, err := sysexec.Merged(ctx, "ping", pingArgs(runtime.GOOS, host, 1, quickPingWait)...)
	if err != nil {
		return false, ""
	}
	latency, _ := parsePingLatency(out)
	return true, latency
}

var (
	// Matches the per-reply latency in English ("time=12 ms", "time<1ms") and
	// Turkish ("süre=12ms", "süre<1ms") ping output.
	latencyRe = regexp.MustCompile(`(?:time|süre)\s*(<|=)\s*([0-9]+(?:[.,][0-9]+)?)`)
	// Matches the average in a BSD/macOS summary line:
	// "round-trip min/avg/max/stddev = 19.274/19.274/19.274/0.000 ms".
	summaryRe = regexp.MustCompile(`=\s*[0-9.]+/([0-9.]+)/`)
)

// parsePingLatency extracts a display-ready latency from ping output. The
// second return reports whether anything could be parsed at all.
func parsePingLatency(out string) (string, bool) {
	if m := latencyRe.FindStringSubmatch(strings.ToLower(out)); len(m) > 2 {
		value := strings.Replace(m[2], ",", ".", 1)
		if m[1] == "<" {
			return "<" + value + " ms", true
		}
		return value + " ms", true
	}
	if m := summaryRe.FindStringSubmatch(out); len(m) > 1 {
		return m[1] + " ms", true
	}
	return "", false
}

// lookupPublicIP asks an external service for the public IP. Callers must have
// opted in: this is the one probe that discloses the user's address to a third
// party. An empty string means "unknown"; the failure is not worth surfacing.
func lookupPublicIP(ctx context.Context) string {
	ctx, cancel := context.WithTimeout(ctx, publicIPTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, publicIPURL, nil)
	if err != nil {
		return ""
	}

	resp, err := publicIPClient.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ""
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, publicIPMaxBytes))
	if err != nil {
		return ""
	}

	ip := strings.TrimSpace(string(body))
	if net.ParseIP(ip) == nil {
		return ""
	}
	return ip
}
