package filter

import (
	"testing"
)

func TestApplyFilterObject(t *testing.T) {
	input := []byte(`{"user":{"name":"Alice","role":"admin"}}`)

	out, err := ApplyFilter(input, ".user.name")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(out) != `"Alice"` {
		t.Errorf("expected %q, got %q", `"Alice"`, string(out))
	}
}

func TestApplyFilterArray(t *testing.T) {
	input := []byte(`{"users":[{"name":"Alice"},{"name":"Bob"}]}`)

	out, err := ApplyFilter(input, ".users[1].name")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(out) != `"Bob"` {
		t.Errorf("expected %q, got %q", `"Bob"`, string(out))
	}
}

func TestFilterKeys(t *testing.T) {
	input := []byte(`{"name":"Alice","role":"admin","secret":"12345"}`)

	out, err := FilterKeys(input, "name, role")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := `{"name":"Alice","role":"admin"}`
	if string(out) != expected {
		t.Errorf("expected %s, got %s", expected, string(out))
	}
}

func TestFlattenArray(t *testing.T) {
	input := []byte(`[[1, 2], [3, [4, 5]]]`)

	out, err := FlattenArray(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := `[1,2,3,4,5]`
	if string(out) != expected {
		t.Errorf("expected %s, got %s", expected, string(out))
	}
}
