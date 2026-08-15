package printer

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/kavix/kurl/client"
)

func TestOutputPathSavesRawBodyOnly(t *testing.T) {
	body := `{"b":2,"a":1}`
	req, err := http.NewRequest(http.MethodGet, "https://example.com/data.json", nil)
	if err != nil {
		t.Fatal(err)
	}
	result := &client.Result{
		Request: req,
		Response: &http.Response{
			StatusCode:    http.StatusOK,
			Status:        "200 OK",
			Proto:         "HTTP/1.1",
			Header:        http.Header{"Content-Type": []string{"application/json"}},
			Body:          io.NopCloser(strings.NewReader(body)),
			ContentLength: int64(len(body)),
		},
	}
	outputPath := filepath.Join(t.TempDir(), "nested", "response.json")

	var stdout bytes.Buffer
	err = Render(&stdout, result, Options{OutputPath: outputPath}, time.Millisecond)
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}

	saved, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("expected output file: %v", err)
	}
	if string(saved) != body {
		t.Fatalf("saved body = %q, want raw %q", string(saved), body)
	}
	if strings.Contains(stdout.String(), body) {
		t.Fatalf("stdout should not duplicate the raw body when -o is used, got %q", stdout.String())
	}
}
