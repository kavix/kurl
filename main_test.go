package main

import (
	"testing"
	"time"
)

func TestParseTimeout(t *testing.T) {
	cases := []struct {
		in   string
		want time.Duration
	}{
		{"30", 30 * time.Second}, // bare number -> seconds (documented default)
		{"0", 0},                 // zero disables the timeout
		{"0.5", 500 * time.Millisecond},
		{"5s", 5 * time.Second},
		{"1m", time.Minute}, // regression: previously became 1ms
		{"500ms", 500 * time.Millisecond},
		{"1m30s", 90 * time.Second},
		{" 2s ", 2 * time.Second}, // surrounding whitespace tolerated
	}

	for _, c := range cases {
		got, err := parseTimeout(c.in)
		if err != nil {
			t.Errorf("parseTimeout(%q) returned error: %v", c.in, err)
			continue
		}
		if got != c.want {
			t.Errorf("parseTimeout(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestParseTimeoutInvalid(t *testing.T) {
	for _, in := range []string{"", "notaduration", "-5", "-1s", "5x"} {
		if _, err := parseTimeout(in); err == nil {
			t.Errorf("parseTimeout(%q) expected an error, got nil", in)
		}
	}
}

// The flag must actually reach cliOptions.timeout, for both the "--timeout v"
// and "--timeout=v" / "-t=v" spellings.
func TestParseCLITimeoutFlag(t *testing.T) {
	cases := []struct {
		args []string
		want time.Duration
	}{
		{[]string{"--timeout", "1m", "https://example.com"}, time.Minute},
		{[]string{"-t", "750ms", "https://example.com"}, 750 * time.Millisecond},
		{[]string{"--timeout=2s", "https://example.com"}, 2 * time.Second},
		{[]string{"-t=15", "https://example.com"}, 15 * time.Second},
	}

	for _, c := range cases {
		opts, err := parseCLI(c.args)
		if err != nil {
			t.Errorf("parseCLI(%v) returned error: %v", c.args, err)
			continue
		}
		if opts.timeout != c.want {
			t.Errorf("parseCLI(%v) timeout = %v, want %v", c.args, opts.timeout, c.want)
		}
	}
}

func TestParseCLITimeoutInvalid(t *testing.T) {
	if _, err := parseCLI([]string{"--timeout", "notaduration", "https://example.com"}); err == nil {
		t.Fatal("expected parseCLI to reject an invalid --timeout value")
	}
}

// Without an explicit --timeout, the default must remain 30s.
func TestParseCLITimeoutDefault(t *testing.T) {
	opts, err := parseCLI([]string{"https://example.com"})
	if err != nil {
		t.Fatalf("parseCLI returned error: %v", err)
	}
	if opts.timeout != 30*time.Second {
		t.Fatalf("default timeout = %v, want 30s", opts.timeout)
	}
}
