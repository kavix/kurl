package printer

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"brew-terminal-curl/client"
)

func TestRenderTiming(t *testing.T) {
	tm := &client.Timing{
		DNSLookup:        45 * time.Millisecond,
		TCPConnection:    30 * time.Millisecond,
		TLSHandshake:     120 * time.Millisecond,
		ServerProcessing: 200 * time.Millisecond,
		TimeToFirstByte:  395 * time.Millisecond,
	}

	var buf bytes.Buffer
	if err := RenderTiming(&buf, tm, 15*time.Millisecond, 410*time.Millisecond, false); err != nil {
		t.Fatalf("RenderTiming returned error: %v", err)
	}
	out := buf.String()

	for _, want := range []string{
		"TIMING",
		"DNS Lookup", "45ms",
		"TCP Connection", "30ms",
		"TLS Handshake", "120ms",
		"Server Processing", "200ms",
		"Content Transfer", "15ms",
		"Total", "410ms",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("RenderTiming output missing %q\n--- output ---\n%s", want, out)
		}
	}
}
