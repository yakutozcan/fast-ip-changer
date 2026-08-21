package diagnostics

import (
	"reflect"
	"testing"
	"time"
)

func TestParsePingLatency(t *testing.T) {
	tests := []struct {
		name string
		out  string
		want string
		ok   bool
	}{
		{
			name: "macOS english reply",
			out: `PING 1.1.1.1 (1.1.1.1): 56 data bytes
64 bytes from 1.1.1.1: icmp_seq=0 ttl=57 time=19.274 ms

--- 1.1.1.1 ping statistics ---
1 packets transmitted, 1 packets received, 0.0% packet loss
round-trip min/avg/max/stddev = 19.274/19.274/19.274/0.000 ms`,
			want: "19.274 ms",
			ok:   true,
		},
		{
			name: "linux english reply",
			out: `PING 1.1.1.1 (1.1.1.1) 56(84) bytes of data.
64 bytes from 1.1.1.1: icmp_seq=1 ttl=57 time=18.4 ms`,
			want: "18.4 ms",
			ok:   true,
		},
		{
			name: "windows english reply",
			out: `Pinging 192.168.1.1 with 32 bytes of data:
Reply from 192.168.1.1: bytes=32 time=12ms TTL=64`,
			want: "12 ms",
			ok:   true,
		},
		{
			name: "windows english sub-millisecond reply",
			out:  `Reply from 192.168.1.1: bytes=32 time<1ms TTL=64`,
			want: "<1 ms",
			ok:   true,
		},
		{
			name: "windows turkish reply",
			out: `192.168.1.1 adresine ping komutu gönderiliyor 32 bayt veri:
192.168.1.1 üzerinden yanıt: bayt=32 süre=12ms TTL=64`,
			want: "12 ms",
			ok:   true,
		},
		{
			name: "windows turkish sub-millisecond reply",
			out:  `192.168.1.1 üzerinden yanıt: bayt=32 süre<1ms TTL=64`,
			want: "<1 ms",
			ok:   true,
		},
		{
			name: "windows turkish uppercase reply",
			out:  `192.168.1.1 ÜZERINDEN YANIT: BAYT=32 SÜRE=8MS TTL=64`,
			want: "8 ms",
			ok:   true,
		},
		{
			name: "macOS summary only",
			out: `--- 192.168.1.1 ping statistics ---
1 packets transmitted, 1 packets received, 0.0% packet loss
round-trip min/avg/max/stddev = 1.234/2.345/3.456/0.111 ms`,
			want: "2.345 ms",
			ok:   true,
		},
		{
			name: "comma decimal separator",
			out:  `192.168.1.1 üzerinden yanıt: bayt=32 süre=1,5ms TTL=64`,
			want: "1.5 ms",
			ok:   true,
		},
		{
			name: "destination unreachable",
			out: `Pinging 10.0.0.9 with 32 bytes of data:
Reply from 10.0.0.1: Destination host unreachable.`,
			want: "",
			ok:   false,
		},
		{
			name: "request timed out",
			out:  "İstek zaman aşımına uğradı.",
			want: "",
			ok:   false,
		},
		{
			name: "garbage",
			out:  "?!?! not ping output at all",
			want: "",
			ok:   false,
		},
		{
			name: "empty",
			out:  "",
			want: "",
			ok:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := parsePingLatency(tt.out)
			if ok != tt.ok {
				t.Fatalf("parsePingLatency ok = %v, want %v", ok, tt.ok)
			}
			if got != tt.want {
				t.Errorf("parsePingLatency = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPingArgs(t *testing.T) {
	tests := []struct {
		name  string
		goos  string
		count int
		wait  time.Duration
		want  []string
	}{
		{
			// Windows -w is milliseconds.
			name:  "windows",
			goos:  "windows",
			count: 4,
			wait:  2 * time.Second,
			want:  []string{"-n", "4", "-w", "2000", "1.1.1.1"},
		},
		{
			// macOS -W is milliseconds.
			name:  "darwin",
			goos:  "darwin",
			count: 4,
			wait:  2 * time.Second,
			want:  []string{"-c", "4", "-W", "2000", "1.1.1.1"},
		},
		{
			// Linux -W is SECONDS: passing 2000 here would wait ~33 minutes.
			name:  "linux",
			goos:  "linux",
			count: 4,
			wait:  2 * time.Second,
			want:  []string{"-c", "4", "-W", "2", "1.1.1.1"},
		},
		{
			// Sub-second waits must round up rather than collapse to "-W 0".
			name:  "linux rounds sub-second wait up",
			goos:  "linux",
			count: 1,
			wait:  1500 * time.Millisecond,
			want:  []string{"-c", "1", "-W", "2", "1.1.1.1"},
		},
		{
			name:  "linux keeps at least one second",
			goos:  "linux",
			count: 1,
			wait:  10 * time.Millisecond,
			want:  []string{"-c", "1", "-W", "1", "1.1.1.1"},
		},
		{
			name:  "darwin keeps at least one millisecond",
			goos:  "darwin",
			count: 1,
			wait:  10 * time.Microsecond,
			want:  []string{"-c", "1", "-W", "1", "1.1.1.1"},
		},
		{
			name:  "non-positive count falls back to one",
			goos:  "linux",
			count: 0,
			wait:  time.Second,
			want:  []string{"-c", "1", "-W", "1", "1.1.1.1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pingArgs(tt.goos, "1.1.1.1", tt.count, tt.wait)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("pingArgs(%q) = %v, want %v", tt.goos, got, tt.want)
			}
		})
	}
}

// TestPingArgsLinuxWaitIsSeconds is the regression test for the bug where the
// non-Windows branch passed a millisecond value to Linux's second-based -W,
// turning "wait 1 second" into "wait 1000 seconds".
func TestPingArgsLinuxWaitIsSeconds(t *testing.T) {
	linux := pingArgs("linux", "1.1.1.1", 1, time.Second)
	darwin := pingArgs("darwin", "1.1.1.1", 1, time.Second)

	if linux[3] != "1" {
		t.Errorf("linux -W = %q, want %q (seconds)", linux[3], "1")
	}
	if darwin[3] != "1000" {
		t.Errorf("darwin -W = %q, want %q (milliseconds)", darwin[3], "1000")
	}
}

func TestTraceRouteArgs(t *testing.T) {
	name, args := traceRouteArgs("windows", "1.1.1.1")
	if name != "tracert" {
		t.Errorf("windows command = %q, want %q", name, "tracert")
	}
	if want := []string{"-d", "-h", "15", "1.1.1.1"}; !reflect.DeepEqual(args, want) {
		t.Errorf("windows args = %v, want %v", args, want)
	}

	name, args = traceRouteArgs("darwin", "1.1.1.1")
	if name != "traceroute" {
		t.Errorf("darwin command = %q, want %q", name, "traceroute")
	}
	if want := []string{"-m", "15", "-q", "1", "-w", "1", "1.1.1.1"}; !reflect.DeepEqual(args, want) {
		t.Errorf("darwin args = %v, want %v", args, want)
	}
}

func TestValidateHost(t *testing.T) {
	tests := []struct {
		in      string
		want    string
		wantErr bool
	}{
		{in: "  1.1.1.1 ", want: "1.1.1.1"},
		{in: "example.com", want: "example.com"},
		{in: "", wantErr: true},
		{in: "   ", wantErr: true},
		{in: "-c 100", wantErr: true},
		{in: "1.1.1.1 -f", wantErr: true},
	}

	for _, tt := range tests {
		got, err := validateHost(tt.in)
		if tt.wantErr {
			if err == nil {
				t.Errorf("validateHost(%q) expected an error", tt.in)
			}
			continue
		}
		if err != nil {
			t.Errorf("validateHost(%q) unexpected error: %v", tt.in, err)
			continue
		}
		if got != tt.want {
			t.Errorf("validateHost(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

// TestCancelWithNothingRunning covers the UI wiring its cancel button straight
// to Cancel: pressing it while idle, or twice in a row, must not panic.
func TestCancelWithNothingRunning(t *testing.T) {
	Cancel()
	Cancel()

	runMu.Lock()
	inFlight := current
	runMu.Unlock()

	if inFlight != nil {
		t.Fatalf("current run = %v, want nil after Cancel", inFlight)
	}
}
