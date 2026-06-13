package main

import "testing"

func TestParseCLIRedirectDefaults(t *testing.T) {
	opts, err := parseCLI([]string{"https://example.com"})
	if err != nil {
		t.Fatalf("parseCLI returned error: %v", err)
	}
	if !opts.followRedirects {
		t.Error("followRedirects should default to true")
	}
	if opts.maxRedirects != defaultMaxRedirects {
		t.Errorf("maxRedirects default = %d, want %d", opts.maxRedirects, defaultMaxRedirects)
	}
}

func TestParseCLIRedirectFlags(t *testing.T) {
	cases := []struct {
		name       string
		args       []string
		wantFollow bool
		wantMax    int
	}{
		{"no-follow", []string{"--no-follow-redirects", "https://example.com"}, false, defaultMaxRedirects},
		{"explicit-follow", []string{"--follow-redirects", "https://example.com"}, true, defaultMaxRedirects},
		{"max", []string{"--max-redirects", "5", "https://example.com"}, true, 5},
		{"max-eq", []string{"--max-redirects=3", "https://example.com"}, true, 3},
		{"max-zero", []string{"--max-redirects", "0", "https://example.com"}, true, 0},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			opts, err := parseCLI(c.args)
			if err != nil {
				t.Fatalf("parseCLI(%v) error: %v", c.args, err)
			}
			if opts.followRedirects != c.wantFollow {
				t.Errorf("followRedirects = %v, want %v", opts.followRedirects, c.wantFollow)
			}
			if opts.maxRedirects != c.wantMax {
				t.Errorf("maxRedirects = %d, want %d", opts.maxRedirects, c.wantMax)
			}
		})
	}
}

func TestParseCLIMaxRedirectsInvalid(t *testing.T) {
	for _, v := range []string{"abc", "-1"} {
		if _, err := parseCLI([]string{"--max-redirects", v, "https://example.com"}); err == nil {
			t.Errorf("expected error for --max-redirects %q", v)
		}
	}
}
