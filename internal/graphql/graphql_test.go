package graphql

import (
	"encoding/json"
	"testing"
)

func TestBuildPayloadIntrospect(t *testing.T) {
	payload, err := BuildPayload("", "", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var p Payload
	if err := json.Unmarshal(payload, &p); err != nil {
		t.Fatalf("invalid json payload: %v", err)
	}

	if p.Query == "" {
		t.Errorf("expected non-empty introspection query")
	}
}

func TestBuildPayloadWithVariables(t *testing.T) {
	query := `query { user(id: $id) { name } }`
	vars := `{"id": "123"}`

	payload, err := BuildPayload(query, vars, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var p Payload
	if err := json.Unmarshal(payload, &p); err != nil {
		t.Fatalf("invalid json payload: %v", err)
	}

	if p.Query != query {
		t.Errorf("query mismatch")
	}
	if p.Variables["id"] != "123" {
		t.Errorf("variables mismatch")
	}
}

func TestGenerateQueryForType(t *testing.T) {
	query := GenerateQueryForType("User")
	expected := "query GetUser {\n  User {\n    id\n    name\n  }\n}"
	if query != expected {
		t.Errorf("expected %q, got %q", expected, query)
	}
}
