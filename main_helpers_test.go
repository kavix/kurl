package main

import (
	"testing"
)

func TestJoinURL(t *testing.T) {
	tests := []struct {
		baseURL  string
		path     string
		expected string
	}{
		{"http://localhost:8080/v1", "/users", "http://localhost:8080/v1/users"},
		{"http://localhost:8080/v1/", "users", "http://localhost:8080/v1/users"},
		{"http://localhost:8080/v1", "", "http://localhost:8080/v1"},
		{"", "users", "users"},
	}

	for _, tt := range tests {
		got := joinURL(tt.baseURL, tt.path)
		if got != tt.expected {
			t.Errorf("joinURL(%q, %q) = %q; want %q", tt.baseURL, tt.path, got, tt.expected)
		}
	}
}

func TestMergeHeaders(t *testing.T) {
	profile := []string{
		"X-Environment: dev",
		"Authorization: Bearer profile-token",
		"Accept: application/json",
	}

	cli := []string{
		"Authorization: Bearer cli-token",
		"Content-Type: application/json",
	}

	merged := mergeHeaders(profile, cli)

	expected := []string{
		"X-Environment: dev",
		"Authorization: Bearer cli-token",
		"Accept: application/json",
		"Content-Type: application/json",
	}

	if len(merged) != len(expected) {
		t.Fatalf("expected length %d, got %d", len(expected), len(merged))
	}

	for i, h := range merged {
		if h != expected[i] {
			t.Errorf("at index %d: expected %q, got %q", i, expected[i], h)
		}
	}
}

func TestIsValidRequestName(t *testing.T) {
	valid := []string{"req1", "my-req", "my_req_123", "GET-users"}
	invalid := []string{"", "req 1", "req/1", "req@name", "req.json"}

	for _, name := range valid {
		if !isValidRequestName(name) {
			t.Errorf("expected %q to be valid", name)
		}
	}

	for _, name := range invalid {
		if isValidRequestName(name) {
			t.Errorf("expected %q to be invalid", name)
		}
	}
}

func TestLooksLikeURLAndMethodToken(t *testing.T) {
	if !looksLikeURL("http://example.com") || !looksLikeURL("https://example.com") {
		t.Errorf("looksLikeURL failed for standard URLs")
	}
	if looksLikeURL("example.com") {
		t.Errorf("looksLikeURL failed for scheme-less domain")
	}

	if !isMethodToken("GET") || !isMethodToken("POST") || !isMethodToken("DELETE") {
		t.Errorf("isMethodToken failed for valid HTTP methods")
	}
	if isMethodToken("get") || isMethodToken("GET/POST") || isMethodToken("api.com") || isMethodToken("") {
		t.Errorf("isMethodToken failed for invalid method tokens")
	}
}
