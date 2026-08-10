package printer

import (
	"bytes"
	"strings"
	"testing"
)

func TestPrettyJSON(t *testing.T) {
	input := `{"name":"kurl","version":1.0,"active":true,"tags":["cli","http"],"author":null}`
	var buf bytes.Buffer
	n, err := PrettyJSON(&buf, strings.NewReader(input), false)
	if err != nil {
		t.Fatalf("unexpected error formatting JSON: %v", err)
	}

	if n <= 0 {
		t.Fatalf("expected written count > 0, got %d", n)
	}

	out := buf.String()
	if !strings.Contains(out, `"name": "kurl"`) {
		t.Errorf("expected pretty printed key-value, got:\n%s", out)
	}
	if !strings.Contains(out, `"tags": [`) {
		t.Errorf("expected formatted array, got:\n%s", out)
	}
	if !strings.Contains(out, `null`) {
		t.Errorf("expected null value formatting, got:\n%s", out)
	}
}

func TestPrettyJSONColored(t *testing.T) {
	input := `{"key": "value", "num": 42}`
	var buf bytes.Buffer
	_, err := PrettyJSON(&buf, strings.NewReader(input), true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "\033[") {
		t.Fatalf("expected color escape codes when color enabled, got:\n%s", out)
	}
}
