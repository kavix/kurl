package client

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// A server slower than the configured timeout must surface a clear,
// timeout-specific error rather than a raw Go transport error.
func TestFetchTimeoutMessage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	_, err := Fetch(Options{
		Method:  http.MethodGet,
		URL:     srv.URL, // explicit scheme -> single fetch path
		Timeout: 20 * time.Millisecond,
	})
	if err == nil {
		t.Fatal("expected a timeout error, got nil")
	}
	if !strings.Contains(err.Error(), "timed out") {
		t.Fatalf("expected a clear timeout message, got %q", err.Error())
	}
}

// A response that arrives within the timeout must not be misreported as a
// timeout.
func TestFetchWithinTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	result, err := Fetch(Options{
		Method:  http.MethodGet,
		URL:     srv.URL,
		Timeout: 2 * time.Second,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	result.Response.Body.Close()
}
