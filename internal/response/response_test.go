package response

import (
	"strings"
	"testing"
)

func TestParseRawResponse(t *testing.T) {
	input := "HTTP/1.1 200 OK\nContent-Type: text/plain\nX-Test: one\n\nhello world"

	parsed := Parse(input)

	if parsed.Mode != "raw" {
		t.Fatalf("expected mode raw, got %q", parsed.Mode)
	}
	if parsed.StatusLine != "HTTP/1.1 200 OK" {
		t.Fatalf("unexpected status line: %q", parsed.StatusLine)
	}
	if len(parsed.Headers) != 2 {
		t.Fatalf("expected 2 headers, got %d", len(parsed.Headers))
	}
	if parsed.Body != "hello world" {
		t.Fatalf("unexpected body: %q", parsed.Body)
	}
}

func TestParseVerboseResponse(t *testing.T) {
	input := strings.Join([]string{
		"* Connected to example.com (93.184.216.34) port 443 (#0)",
		"> GET / HTTP/1.1",
		"> Host: example.com",
		"< HTTP/2 200",
		"< content-type: text/plain",
		"<",
		"ok",
	}, "\n")

	parsed := Parse(input)

	if parsed.Mode != "verbose" {
		t.Fatalf("expected mode verbose, got %q", parsed.Mode)
	}
	if parsed.RequestLine != "GET / HTTP/1.1" {
		t.Fatalf("unexpected request line: %q", parsed.RequestLine)
	}
	if len(parsed.RequestHeaders) != 1 {
		t.Fatalf("expected 1 request header, got %d", len(parsed.RequestHeaders))
	}
	if parsed.StatusLine != "HTTP/2 200" {
		t.Fatalf("unexpected status line: %q", parsed.StatusLine)
	}
	if len(parsed.Headers) != 1 {
		t.Fatalf("expected 1 response header, got %d", len(parsed.Headers))
	}
	if parsed.Body != "ok" {
		t.Fatalf("unexpected body: %q", parsed.Body)
	}
}

func TestRenderJSON(t *testing.T) {
	parsed := Response{Mode: "raw", StatusLine: "HTTP/1.1 200 OK"}

	rendered, err := RenderJSON(parsed)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(rendered, "\"mode\": \"raw\"") {
		t.Fatalf("json output missing mode: %s", rendered)
	}
}