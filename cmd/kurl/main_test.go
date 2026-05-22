package main

import (
	"net/http"
	"strings"
	"testing"

	"brew-terminal-curl/internal/response"
)

func TestFormatBodyPrettyPrintsJSON(t *testing.T) {
	body := []byte(`{"name":"kurl","nested":{"ok":true}}`)

	rendered := response.FormatBody(body, "application/json")

	if !strings.Contains(rendered, "\n  \"nested\"") {
		t.Fatalf("expected pretty printed json, got: %s", rendered)
	}
}

func TestFromHTTPResponseKeepsStatus(t *testing.T) {
	headers := http.Header{}
	headers.Set("Content-Type", "application/json")

	parsed := response.FromHTTPResponse("HTTP/2 200", headers, []byte(`{"ok":true}`))

	if parsed.StatusLine != "HTTP/2 200" {
		t.Fatalf("unexpected status line: %q", parsed.StatusLine)
	}
	if !strings.Contains(parsed.Body, "\n  \"ok\"") {
		t.Fatalf("expected pretty json body, got: %s", parsed.Body)
	}
}
