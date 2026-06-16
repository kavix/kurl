package client

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestPhase(t *testing.T) {
	base := time.Now()
	later := base.Add(20 * time.Millisecond)

	if got := phase(base, later); got != 20*time.Millisecond {
		t.Errorf("phase(base, base+20ms) = %v, want 20ms", got)
	}
	if got := phase(time.Time{}, later); got != 0 {
		t.Errorf("phase with zero start = %v, want 0", got)
	}
	if got := phase(base, time.Time{}); got != 0 {
		t.Errorf("phase with zero end = %v, want 0", got)
	}
	if got := phase(later, base); got != 0 {
		t.Errorf("phase with end before start = %v, want 0", got)
	}
}

func TestFetchPopulatesTiming(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// A small delay so server-processing and TTFB are clearly measurable.
		time.Sleep(10 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, "ok")
	}))
	defer server.Close()

	result, err := Fetch(Options{Method: "GET", URL: server.URL, Timeout: 5 * time.Second, Timing: true})
	if err != nil {
		t.Fatalf("Fetch returned error: %v", err)
	}
	defer result.Response.Body.Close()
	io.Copy(io.Discard, result.Response.Body)

	if result.Timing == nil {
		t.Fatal("expected Timing to be populated when Options.Timing is set")
	}
	tm := result.Timing

	if tm.TCPConnection <= 0 {
		t.Errorf("expected a positive TCP connection time, got %v", tm.TCPConnection)
	}
	if tm.TimeToFirstByte <= 0 {
		t.Errorf("expected a positive time-to-first-byte, got %v", tm.TimeToFirstByte)
	}
	if tm.ServerProcessing < 0 {
		t.Errorf("server processing should never be negative, got %v", tm.ServerProcessing)
	}
	// Plain HTTP: no TLS handshake should be recorded.
	if tm.TLSHandshake != 0 {
		t.Errorf("expected zero TLS handshake for http, got %v", tm.TLSHandshake)
	}
	// TTFB is the outermost phase measured by the client; it must bound the
	// inner phases it contains.
	if tm.ServerProcessing > tm.TimeToFirstByte {
		t.Errorf("server processing %v exceeds time-to-first-byte %v", tm.ServerProcessing, tm.TimeToFirstByte)
	}
}

func TestFetchWithoutTimingLeavesNil(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "ok")
	}))
	defer server.Close()

	result, err := Fetch(Options{Method: "GET", URL: server.URL, Timeout: 5 * time.Second})
	if err != nil {
		t.Fatalf("Fetch returned error: %v", err)
	}
	defer result.Response.Body.Close()

	if result.Timing != nil {
		t.Errorf("expected Timing to stay nil when Options.Timing is unset, got %+v", result.Timing)
	}
}
