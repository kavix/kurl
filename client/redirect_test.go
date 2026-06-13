package client

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// redirectServer returns a server whose /a redirects to /b, and /b returns 200.
func redirectServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/a":
			http.Redirect(w, r, "/b", http.StatusFound)
		default:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("final"))
		}
	}))
}

func TestFetchFollowsRedirects(t *testing.T) {
	srv := redirectServer()
	defer srv.Close()

	result, err := Fetch(Options{
		Method:          http.MethodGet,
		URL:             srv.URL + "/a",
		FollowRedirects: true,
		MaxRedirects:    10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer result.Response.Body.Close()

	if result.Response.StatusCode != http.StatusOK {
		t.Errorf("final status = %d, want 200", result.Response.StatusCode)
	}
	if len(result.Redirects) != 1 {
		t.Fatalf("recorded %d redirect hops, want 1", len(result.Redirects))
	}
}

func TestFetchNoFollowReturnsRedirect(t *testing.T) {
	srv := redirectServer()
	defer srv.Close()

	result, err := Fetch(Options{
		Method:          http.MethodGet,
		URL:             srv.URL + "/a",
		FollowRedirects: false,
		MaxRedirects:    10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer result.Response.Body.Close()

	if result.Response.StatusCode != http.StatusFound {
		t.Errorf("status = %d, want 302 (redirect not followed)", result.Response.StatusCode)
	}
	if len(result.Redirects) != 0 {
		t.Errorf("recorded %d hops, want 0 when not following", len(result.Redirects))
	}
}

func TestFetchMaxRedirectsExceeded(t *testing.T) {
	srv := redirectServer() // /a -> /b is one redirect
	defer srv.Close()

	_, err := Fetch(Options{
		Method:          http.MethodGet,
		URL:             srv.URL + "/a",
		FollowRedirects: true,
		MaxRedirects:    0, // disallow following any redirect
	})
	if err == nil {
		t.Fatal("expected an error when max redirects is exceeded")
	}
	if !strings.Contains(err.Error(), "redirects") {
		t.Errorf("error %q should mention redirects", err.Error())
	}
}
