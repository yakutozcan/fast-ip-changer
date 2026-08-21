package diagnostics

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// withTestEndpoint points the public-IP lookup at srv for the duration of the
// test. No test may reach the real service.
func withTestEndpoint(t *testing.T, url string) {
	t.Helper()
	previous := publicIPURL
	publicIPURL = url
	t.Cleanup(func() { publicIPURL = previous })
}

func TestLookupPublicIP(t *testing.T) {
	tests := []struct {
		name   string
		status int
		body   string
		want   string
	}{
		{name: "ipv4 body", status: http.StatusOK, body: "203.0.113.7", want: "203.0.113.7"},
		{name: "surrounding whitespace is trimmed", status: http.StatusOK, body: "\n 203.0.113.7 \n", want: "203.0.113.7"},
		{name: "ipv6 body", status: http.StatusOK, body: "2001:db8::1", want: "2001:db8::1"},
		// Anything that is not an address is discarded rather than shown: the
		// value goes straight into the UI, and a captive portal or an error page
		// is a far more likely response than a malicious one.
		{name: "html error page", status: http.StatusOK, body: "<html>nope</html>", want: ""},
		{name: "empty body", status: http.StatusOK, body: "", want: ""},
		{name: "server error", status: http.StatusInternalServerError, body: "203.0.113.7", want: ""},
		{name: "rate limited", status: http.StatusTooManyRequests, body: "203.0.113.7", want: ""},
		// The read is capped at publicIPMaxBytes, so an oversized response is
		// truncated mid-value and then fails the address check.
		{name: "oversized body", status: http.StatusOK, body: strings.Repeat("1", publicIPMaxBytes*2), want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer srv.Close()
			withTestEndpoint(t, srv.URL)

			if got := lookupPublicIP(context.Background()); got != tt.want {
				t.Errorf("lookupPublicIP() = %q, want %q", got, tt.want)
			}
		})
	}
}

// A hanging service must not hold up the quick check, and must not surface as an
// error either: an unknown public IP is simply empty.
func TestLookupPublicIPHonoursTimeout(t *testing.T) {
	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-release
	}))
	defer func() {
		close(release)
		srv.Close()
	}()
	withTestEndpoint(t, srv.URL)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	start := time.Now()
	if got := lookupPublicIP(ctx); got != "" {
		t.Errorf("lookupPublicIP() = %q, want empty", got)
	}
	if elapsed := time.Since(start); elapsed > publicIPTimeout {
		t.Errorf("lookupPublicIP blocked for %v, longer than the %v cap", elapsed, publicIPTimeout)
	}
}

// The privacy claim the README makes: nothing contacts the third-party service
// unless the caller asked for it. This is the test that keeps that true.
func TestQuickCheckDoesNotContactServiceUnlessOptedIn(t *testing.T) {
	var hits atomic.Int64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		_, _ = w.Write([]byte("203.0.113.7"))
	}))
	defer srv.Close()
	withTestEndpoint(t, srv.URL)

	res := QuickCheck(context.Background(), "", false)

	if got := hits.Load(); got != 0 {
		t.Errorf("public-IP endpoint was contacted %d time(s) with checkPublicIP=false", got)
	}
	if res.PublicIP != "" {
		t.Errorf("PublicIP = %q, want empty when the lookup is not opted into", res.PublicIP)
	}
	// An empty gateway means "no gateway to probe", so that half of the result
	// stays zeroed instead of reporting a failed ping.
	if res.GatewayOk || res.GatewayLatency != "" {
		t.Errorf("gateway result = (%v, %q), want zero values for an empty gateway",
			res.GatewayOk, res.GatewayLatency)
	}
}

func TestQuickCheckReportsPublicIPWhenOptedIn(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("203.0.113.7"))
	}))
	defer srv.Close()
	withTestEndpoint(t, srv.URL)

	if got := QuickCheck(context.Background(), "", true).PublicIP; got != "203.0.113.7" {
		t.Errorf("PublicIP = %q, want 203.0.113.7", got)
	}
}
