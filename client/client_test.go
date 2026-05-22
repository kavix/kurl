package client

import "testing"

func TestHasExplicitScheme(t *testing.T) {
	if !hasExplicitScheme("https://example.com") {
		t.Fatal("expected https URL to be explicit")
	}
	if hasExplicitScheme("example.com") {
		t.Fatal("expected bare host to be non-explicit")
	}
}
